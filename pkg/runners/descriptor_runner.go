package runners

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

const (
	dirPerm  os.FileMode = 0o755
	filePerm os.FileMode = 0o600
)

// descriptorRunner implements Runner using a RunnerDescriptor.
type descriptorRunner struct {
	desc        RunnerDescriptor
	meta        ExerciseMeta
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	wrapperFile string // absolute path to written wrapper; empty if no template
	binaryFile  string // absolute path to compiled binary; empty if not compiled

	relExerciseDir string // precomputed {rel_exercise_dir} value; empty until Prepare resolves it

	// waitErr receives the result of the single cmd.Wait() call, owned by the
	// reaper goroutine started in Open. exited is closed once the process has
	// been reaped (ProcessState populated), letting Run detect a crashed
	// subprocess instead of blocking on its stdout forever.
	waitErr chan error
	exited  chan struct{}
}

func (r *descriptorRunner) String() string { return r.desc.Name }

func (r *descriptorRunner) wrapperExt() string {
	return r.desc.Prepare.WrapperExt
}

const relExerciseDirToken = "{rel_exercise_dir}" //nolint:gosec // template token, not a credential

// descriptorReferencesRelDir reports whether any Prepare-time substitutable field
// references {rel_exercise_dir}, so Prepare only walks for go.mod when needed.
func (r *descriptorRunner) descriptorReferencesRelDir() bool {
	for _, v := range r.desc.Prepare.TemplateVars {
		if strings.Contains(v, relExerciseDirToken) {
			return true
		}
	}

	for _, cmd := range r.desc.Prepare.BuildCommands {
		for _, arg := range cmd {
			if strings.Contains(arg, relExerciseDirToken) {
				return true
			}
		}
	}

	if r.desc.Prepare.TemplatePath != "" {
		if b, err := os.ReadFile(r.desc.Prepare.TemplatePath); err == nil &&
			strings.Contains(string(b), relExerciseDirToken) {
			return true
		}
	}

	return false
}

func (r *descriptorRunner) writeWrapper(langDir, ext, subdir string) error {
	if r.desc.Prepare.TemplatePath == "" {
		return nil
	}

	templateBytes, readErr := os.ReadFile(r.desc.Prepare.TemplatePath)
	if readErr != nil {
		return fmt.Errorf("reading template %s: %w", r.desc.Prepare.TemplatePath, readErr)
	}

	vars := make(map[string]string, len(r.desc.Prepare.TemplateVars))
	for k, v := range r.desc.Prepare.TemplateVars {
		vars[k] = substituteTokens(v, r.meta, ext, subdir, r.relExerciseDir)
	}

	substitutedContent := substituteTokens(string(templateBytes), r.meta, ext, subdir, r.relExerciseDir)

	tpl, parseErr := template.New("").Parse(substitutedContent)
	if parseErr != nil {
		return fmt.Errorf("parsing template: %w", parseErr)
	}

	var buf bytes.Buffer
	if execErr := tpl.Execute(&buf, vars); execErr != nil {
		return fmt.Errorf("executing template: %w", execErr)
	}

	wrapperDir := langDir
	if subdir != "" {
		wrapperDir = filepath.Join(langDir, subdir)
		if mkErr := os.MkdirAll(wrapperDir, dirPerm); mkErr != nil {
			return fmt.Errorf("creating wrapper subdir: %w", mkErr)
		}
	}

	r.wrapperFile = filepath.Join(wrapperDir, wrapperBaseName+ext)

	return os.WriteFile(r.wrapperFile, buf.Bytes(), filePerm)
}

// Prepare creates the language directory, writes the wrapper template (if any),
// and runs any build commands.
func (r *descriptorRunner) Prepare(ctx context.Context) error {
	langDir := r.meta.LangDir()

	if mkErr := os.MkdirAll(langDir, dirPerm); mkErr != nil {
		return fmt.Errorf("creating lang dir: %w", mkErr)
	}

	ext := r.wrapperExt()
	subdir := r.desc.Prepare.WrapperSubdir

	if r.descriptorReferencesRelDir() {
		rel, relErr := moduleRelDir(r.meta.Dir)
		if relErr != nil {
			return fmt.Errorf("resolving {rel_exercise_dir}: %w", relErr)
		}

		r.relExerciseDir = rel
	}

	if err := r.writeWrapper(langDir, ext, subdir); err != nil {
		return err
	}

	if r.desc.Open.Binary != "" {
		wrapperDir := langDir
		if subdir != "" {
			wrapperDir = filepath.Join(langDir, subdir)
		}
		r.binaryFile = filepath.Join(wrapperDir, wrapperBaseName)
	}

	for _, cmdArgs := range r.desc.Prepare.BuildCommands {
		if len(cmdArgs) == 0 {
			continue
		}

		substituted := substituteSlice(cmdArgs, r.meta, ext, subdir, r.relExerciseDir)
		//nolint:gosec // build commands come from user config, not untrusted external input
		cmd := exec.CommandContext(ctx, substituted[0], substituted[1:]...)
		cmd.Dir = langDir

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if runErr := cmd.Run(); runErr != nil {
			return fmt.Errorf("build command %q failed: %w: %s", substituted[0], runErr, stderr.String())
		}
	}

	return nil
}

// Open starts the runner subprocess.
// If Binary is set, the compiled binary is executed directly.
// Otherwise, Interpreter + Args are used.
func (r *descriptorRunner) Open(ctx context.Context) error {
	ext := r.wrapperExt()
	subdir := r.desc.Prepare.WrapperSubdir

	var cmd *exec.Cmd

	if r.desc.Open.Binary != "" {
		binaryPath := substituteTokens(r.desc.Open.Binary, r.meta, ext, subdir, "")
		absPath, absErr := filepath.Abs(binaryPath)
		if absErr != nil {
			return fmt.Errorf("resolving binary path: %w", absErr)
		}

		cmd = exec.CommandContext(ctx, absPath)
	} else {
		args := substituteSlice(r.desc.Open.Args, r.meta, ext, subdir, "")
		//nolint:gosec // interpreter and args come from user config
		cmd = exec.CommandContext(ctx, r.desc.Open.Interpreter, args...)
	}

	cmd.Dir = r.meta.LangDir()
	cmd.Env = append(os.Environ(), substituteSlice(r.desc.Open.Env, r.meta, ext, subdir, "")...)

	// Put the subprocess in its own process group so we can signal the entire
	// tree on timeout. Wrapper runners (e.g. bash) fork children — a $(...) or
	// pipeline doing the real work — and killing only the leader orphans those
	// children, which keep the stdout/stderr pipes open and block cmd.Wait().
	// No-op on platforms without process groups (see process_other.go).
	setProcessGroup(cmd)

	var setupErr error
	r.stdin, setupErr = setupBuffers(cmd)
	if setupErr != nil {
		return fmt.Errorf("setting up buffers: %w", setupErr)
	}

	r.cmd = cmd

	if startErr := cmd.Start(); startErr != nil {
		return startErr
	}

	// Single reaper: cmd.Wait() is called exactly once here. Both Run (via
	// checkWait) and Close observe the result through these channels, avoiding
	// the "no child processes" error that double-reaping causes.
	r.waitErr = make(chan error, 1)
	r.exited = make(chan struct{})

	go func() {
		err := cmd.Wait()
		r.waitErr <- err
		close(r.exited) // ProcessState is set by the time Wait returns
	}()

	return nil
}

// Close stops the subprocess and waits for it to exit. It is a teardown
// operation: its postcondition is "the process is no longer running", so a
// process that has already exited (e.g. killed by a timeout context, or crashed
// during Run) is treated as success. Only a genuine failure to kill a still-live
// process is reported as an error.
func (r *descriptorRunner) Close(ctx context.Context) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	// If the process already exited (e.g. a crash detected during Run, or a
	// timeout context that killed it), the reaper has populated waitErr and
	// closed exited; nothing more to stop.
	select {
	case <-r.exited:
		<-r.waitErr // drain so the buffered send is consumed
		return nil
	default:
	}

	// Close stdin so the subprocess receives EOF and can exit cleanly.
	_ = r.stdin.Close()

	// SIGTERM the whole process group so a wrapper's forked children die too. If
	// the group was reaped between the select above and here (a benign race),
	// signalGroup returns os.ErrProcessDone — the process is already stopped, so
	// fall through to observe the reaper's result rather than treating it as a
	// failure.
	if err := signalGroup(r.cmd.Process.Pid, syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("failed to send SIGTERM to %s process: %w", r.desc.Name, err)
	}

	// The reaper goroutine started in Open owns cmd.Wait(); observe its result
	// through waitErr rather than calling Wait() again (which would double-reap).
	// A non-nil waitErr here is expected — we are tearing the process down on
	// purpose, so its exit status (signal: killed, non-zero exit, etc.) is not a
	// failure of Close. Only a kill that fails for a reason other than "already
	// done" is a genuine error.
	select {
	case <-ctx.Done():
		killErr := signalGroup(r.cmd.Process.Pid, syscall.SIGKILL)
		<-r.waitErr // drain reaper regardless of kill outcome
		if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
			return fmt.Errorf("failed to kill %s process: %w", r.desc.Name, killErr)
		}
	case <-r.waitErr:
	}

	return nil
}

// Run sends a task to the subprocess via stdin and reads the JSON result from stdout.
func (r *descriptorRunner) Run(ctx context.Context, task *protocol.Task) (*protocol.Result, error) {
	taskJSON, marshalErr := json.Marshal(task)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshalling task: %w", marshalErr)
	}

	if _, writeErr := r.stdin.Write(append(taskJSON, '\n')); writeErr != nil {
		return nil, fmt.Errorf("writing task to stdin: %w", writeErr)
	}

	result := new(protocol.Result)
	if readErr := readJSONFromCommand(ctx, result, r.cmd, r.exited); readErr != nil {
		return nil, readErr
	}

	return result, nil
}

// Cleanup removes the wrapper file, binary file, and (if empty) the wrapper subdirectory
// created during Prepare. The subdirectory is only removed if it is empty after file removal;
// a non-empty directory is left in place without error.
func (r *descriptorRunner) Cleanup() error {
	var wErr, bErr error

	if r.wrapperFile != "" {
		wErr = os.Remove(r.wrapperFile)
		if errors.Is(wErr, os.ErrNotExist) {
			wErr = nil
		}
	}

	if r.binaryFile != "" {
		bErr = os.Remove(r.binaryFile)
		if errors.Is(bErr, os.ErrNotExist) {
			bErr = nil
		}
	}

	if wErr != nil || bErr != nil {
		if r.desc.Prepare.WrapperSubdir != "" {
			return fmt.Errorf("%w: files not removed, %s must be cleaned up manually",
				errors.Join(wErr, bErr),
				filepath.Join(r.meta.LangDir(), r.desc.Prepare.WrapperSubdir))
		}

		return errors.Join(wErr, bErr)
	}

	if r.desc.Prepare.WrapperSubdir != "" {
		if dErr := removeIfEmpty(filepath.Join(r.meta.LangDir(), r.desc.Prepare.WrapperSubdir)); dErr != nil {
			return dErr
		}
	}

	return r.removeCleanupPaths()
}

// removeIfEmpty removes dir, tolerating that it is absent (already gone) or still
// contains files (leftover user content). [os.Remove] on a non-empty directory returns
// ENOTEMPTY, which is the emptiness test itself — there is no stat pre-check, which
// avoids both a redundant syscall and a TOCTOU race.
func removeIfEmpty(dir string) error {
	err := os.Remove(dir)
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, syscall.ENOTEMPTY) {
		return nil
	}

	return err
}

// removeCleanupPaths removes the descriptor's declared cleanup_paths (e.g. build-output
// trees such as bin/obj or target), resolved against the lang dir. A missing path is not
// an error since [os.RemoveAll] is a no-op in that case.
func (r *descriptorRunner) removeCleanupPaths() error {
	var errs []error

	for _, p := range r.desc.Prepare.CleanupPaths {
		if p == "" {
			continue
		}

		resolved := substituteTokens(p, r.meta, r.wrapperExt(), r.desc.Prepare.WrapperSubdir, "")
		// Relative cleanup paths are resolved against the lang dir.
		if !filepath.IsAbs(resolved) {
			resolved = filepath.Join(r.meta.LangDir(), resolved)
		}

		if rmErr := os.RemoveAll(resolved); rmErr != nil {
			errs = append(errs, fmt.Errorf("removing cleanup path %q: %w", resolved, rmErr))
		}
	}

	return errors.Join(errs...)
}

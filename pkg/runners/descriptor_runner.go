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
}

func (r *descriptorRunner) String() string { return r.desc.Name }

func (r *descriptorRunner) wrapperExt() string {
	return r.desc.Prepare.WrapperExt
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
		vars[k] = substituteTokens(v, r.meta, ext, subdir)
	}

	substitutedContent := substituteTokens(string(templateBytes), r.meta, ext, subdir)

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

		substituted := substituteSlice(cmdArgs, r.meta, ext, subdir)
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
		binaryPath := substituteTokens(r.desc.Open.Binary, r.meta, ext, subdir)
		absPath, absErr := filepath.Abs(binaryPath)
		if absErr != nil {
			return fmt.Errorf("resolving binary path: %w", absErr)
		}

		cmd = exec.CommandContext(ctx, absPath)
	} else {
		args := substituteSlice(r.desc.Open.Args, r.meta, ext, subdir)
		//nolint:gosec // interpreter and args come from user config
		cmd = exec.CommandContext(ctx, r.desc.Open.Interpreter, args...)
	}

	cmd.Dir = r.meta.LangDir()
	cmd.Env = append(os.Environ(), substituteSlice(r.desc.Open.Env, r.meta, ext, subdir)...)

	var setupErr error
	r.stdin, setupErr = setupBuffers(cmd)
	if setupErr != nil {
		return fmt.Errorf("setting up buffers: %w", setupErr)
	}

	r.cmd = cmd

	return cmd.Start()
}

// Close sends SIGTERM to the subprocess and waits for it to exit,
// killing it if the context deadline expires first.
func (r *descriptorRunner) Close(ctx context.Context) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	// Close stdin so the subprocess receives EOF and can exit cleanly.
	_ = r.stdin.Close()

	if err := r.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to %s process: %w", r.desc.Name, err)
	}

	done := make(chan error, 1)
	go func() {
		done <- r.cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		killErr := r.cmd.Process.Kill()
		<-done // drain goroutine regardless of kill outcome
		if killErr != nil {
			return fmt.Errorf("failed to kill %s process: %w", r.desc.Name, killErr)
		}
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to stop %s process: %w", r.desc.Name, err)
		}
	}

	return nil
}

// Run sends a task to the subprocess via stdin and reads the JSON result from stdout.
func (r *descriptorRunner) Run(_ context.Context, task *protocol.Task) (*protocol.Result, error) {
	taskJSON, marshalErr := json.Marshal(task)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshalling task: %w", marshalErr)
	}

	if _, writeErr := r.stdin.Write(append(taskJSON, '\n')); writeErr != nil {
		return nil, fmt.Errorf("writing task to stdin: %w", writeErr)
	}

	result := new(protocol.Result)
	if readErr := readJSONFromCommand(result, r.cmd); readErr != nil {
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

	if r.binaryFile != "" && r.desc.Open.Binary != "" {
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
		dErr := os.Remove(filepath.Join(r.meta.LangDir(), r.desc.Prepare.WrapperSubdir))
		if dErr != nil && !errors.Is(dErr, os.ErrNotExist) && !errors.Is(dErr, syscall.ENOTEMPTY) {
			return dErr
		}
	}

	return nil
}

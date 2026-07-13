package runners

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

func TestDescriptorRunner_Prepare_WritesTemplate(t *testing.T) {
	templateContent := "hello {year} from {title}"
	templateFile := filepath.Join(t.TempDir(), "test.templ")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0o600))

	exerciseDir := t.TempDir()

	desc := RunnerDescriptor{
		Key:  "py",
		Name: "Python",
		Prepare: PrepareSpec{
			TemplatePath: templateFile,
			WrapperExt:   ".py",
		},
		Open: OpenSpec{Interpreter: "python3"},
	}

	meta := ExerciseMeta{
		Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "py",
	}

	runner := &descriptorRunner{desc: desc, meta: meta}
	require.NoError(t, runner.Prepare(context.Background()))

	langDir := filepath.Join(exerciseDir, "py")
	wrapperPath := filepath.Join(langDir, "runtime-wrapper.py")
	content, err := os.ReadFile(wrapperPath)
	require.NoError(t, err)
	assert.Equal(t, "hello 2015 from foo", string(content))
}

func TestDescriptorRunner_Prepare_NoTemplate(t *testing.T) {
	exerciseDir := t.TempDir()

	desc := RunnerDescriptor{
		Key:  "go",
		Name: "Go",
		Prepare: PrepareSpec{
			BuildCommands: [][]string{{"echo", "build"}},
		},
		Open: OpenSpec{Binary: "{binary_file}"},
	}

	meta := ExerciseMeta{
		Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go",
	}

	runner := &descriptorRunner{desc: desc, meta: meta}
	// With no template_path and a build command that succeeds (echo), Prepare should not error
	require.NoError(t, runner.Prepare(context.Background()))
}

func TestDescriptorRunner_Prepare_TemplateVarsSubstituted(t *testing.T) {
	templateContent := "import {{ index . \"import_path\" }}"
	templateFile := filepath.Join(t.TempDir(), "test.go.tmpl")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0o600))

	exerciseDir := t.TempDir()

	desc := RunnerDescriptor{
		Key:  "go",
		Name: "Go",
		Prepare: PrepareSpec{
			TemplatePath: templateFile,
			WrapperExt:   ".go",
			TemplateVars: map[string]string{
				"import_path": "github.com/me/aoc/{year}/{day}-{title}/go",
			},
		},
		Open: OpenSpec{Binary: "{binary_file}"},
	}

	meta := ExerciseMeta{
		Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go",
	}

	runner := &descriptorRunner{desc: desc, meta: meta}
	require.NoError(t, runner.Prepare(context.Background()))

	langDir := filepath.Join(exerciseDir, "go")
	wrapperPath := filepath.Join(langDir, "runtime-wrapper.go")
	content, err := os.ReadFile(wrapperPath)
	require.NoError(t, err)
	assert.Equal(t, "import github.com/me/aoc/2015/01-foo/go", string(content))
}

// TestDescriptorRunner_Prepare_ShippedGoTemplate guards the contract between the
// shipped go.tmpl (the GoTemplate embed) and the import_path template var that
// `elf runners install` documents. The two must agree: go.tmpl must read the var
// via `index . "import_path"`, because Go's text/template cannot access a
// snake_case map key with `.Field` syntax (it renders "<no value>" instead).
// Regression for the import-path mismatch that broke the Go runner after
// `runners install --force` overwrote a customized template.
func TestDescriptorRunner_Prepare_ShippedGoTemplate(t *testing.T) {
	templateFile := filepath.Join(t.TempDir(), "go.tmpl")
	require.NoError(t, os.WriteFile(templateFile, GoTemplate, 0o600))

	exerciseDir := t.TempDir()

	desc := RunnerDescriptor{
		Key:  "go",
		Name: "Go",
		Prepare: PrepareSpec{
			TemplatePath:  templateFile,
			WrapperExt:    ".go",
			WrapperSubdir: "cmd",
			// The exact var key documented by `elf runners install`.
			TemplateVars: map[string]string{
				"import_path": "github.com/me/aoc/{year}/{day}-{title}/go",
			},
		},
		Open: OpenSpec{Binary: "{binary_file}"},
	}

	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go"}

	runner := &descriptorRunner{desc: desc, meta: meta}
	require.NoError(t, runner.Prepare(context.Background()))

	wrapperPath := filepath.Join(exerciseDir, "go", "cmd", "runtime-wrapper.go")
	content, err := os.ReadFile(wrapperPath)
	require.NoError(t, err)

	rendered := string(content)
	assert.NotContains(t, rendered, "<no value>",
		"shipped go.tmpl must resolve the import_path var; got <no value>")
	assert.Contains(t, rendered, `ex "github.com/me/aoc/2015/01-foo/go"`,
		"shipped go.tmpl should render the resolved import path")
}

func TestDescriptorRunner_Cleanup_RemovesEmptySubdir(t *testing.T) {
	templateContent := "hello"
	templateFile := filepath.Join(t.TempDir(), "test.go.tmpl")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0o600))

	exerciseDir := t.TempDir()

	desc := RunnerDescriptor{
		Key:  "go",
		Name: "Go",
		Prepare: PrepareSpec{
			TemplatePath:  templateFile,
			WrapperExt:    ".go",
			WrapperSubdir: "cmd",
		},
		Open: OpenSpec{Binary: "{binary_file}"},
	}

	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go"}
	runner := &descriptorRunner{desc: desc, meta: meta}
	require.NoError(t, runner.Prepare(context.Background()))

	subdir := filepath.Join(exerciseDir, "go", "cmd")
	require.DirExists(t, subdir)

	require.NoError(t, runner.Cleanup())
	assert.NoDirExists(t, subdir)
}

func TestDescriptorRunner_Cleanup_LeavesNonEmptySubdir(t *testing.T) {
	templateContent := "hello"
	templateFile := filepath.Join(t.TempDir(), "test.go.tmpl")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0o600))

	exerciseDir := t.TempDir()

	desc := RunnerDescriptor{
		Key:  "go",
		Name: "Go",
		Prepare: PrepareSpec{
			TemplatePath:  templateFile,
			WrapperExt:    ".go",
			WrapperSubdir: "cmd",
		},
		Open: OpenSpec{Binary: "{binary_file}"},
	}

	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go"}
	runner := &descriptorRunner{desc: desc, meta: meta}
	require.NoError(t, runner.Prepare(context.Background()))

	subdir := filepath.Join(exerciseDir, "go", "cmd")
	userFile := filepath.Join(subdir, "user.go")
	require.NoError(t, os.WriteFile(userFile, []byte("package main"), 0o600))

	require.NoError(t, runner.Cleanup())
	assert.DirExists(t, subdir)
	assert.FileExists(t, userFile)
}

func TestDescriptorRunner_Cleanup_FileErrorIncludesSubdirPath(t *testing.T) {
	exerciseDir := t.TempDir()
	langDir := filepath.Join(exerciseDir, "go")
	subdir := filepath.Join(langDir, "cmd")
	require.NoError(t, os.MkdirAll(subdir, 0o755))

	wrapperFile := filepath.Join(subdir, "runtime-wrapper.go")
	require.NoError(t, os.WriteFile(wrapperFile, []byte("hello"), 0o600))

	// Make the subdir unwritable so os.Remove on the wrapper file fails with EPERM.
	require.NoError(t, os.Chmod(subdir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(subdir, 0o755) })

	desc := RunnerDescriptor{
		Key:  "go",
		Name: "Go",
		Prepare: PrepareSpec{
			WrapperSubdir: "cmd",
		},
		Open: OpenSpec{Binary: "{binary_file}"},
	}

	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exerciseDir, Key: "go"}
	runner := &descriptorRunner{desc: desc, meta: meta, wrapperFile: wrapperFile}

	err := runner.Cleanup()
	require.Error(t, err)
	assert.Contains(t, err.Error(), subdir)
	assert.Contains(t, err.Error(), "manually")
}

func TestDescriptorRunner_Cleanup_RemovesBuildDirs(t *testing.T) {
	root := t.TempDir()
	meta := ExerciseMeta{Dir: root, Key: "cs"}
	langDir := meta.LangDir()
	for _, d := range []string{"bin", "obj"} {
		require.NoError(t, os.MkdirAll(filepath.Join(langDir, d), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(langDir, d, "f"), []byte("x"), 0o600))
	}
	r := &descriptorRunner{
		desc: RunnerDescriptor{Prepare: PrepareSpec{CleanupPaths: []string{"bin", "obj"}}},
		meta: meta,
	}
	require.NoError(t, r.Cleanup())
	assert.NoDirExists(t, filepath.Join(langDir, "bin"))
	assert.NoDirExists(t, filepath.Join(langDir, "obj"))
}

func TestDescriptorRunner_Open_Interpreter(t *testing.T) {
	// Use a real interpreter that just echoes JSON back — "cat" reads stdin and writes stdout
	// We simulate the runner protocol with a shell script
	scriptContent := `#!/bin/sh
while IFS= read -r line; do
  echo "$line"
done
`
	scriptFile := filepath.Join(t.TempDir(), "echo.sh")
	require.NoError(t, os.WriteFile(scriptFile, []byte(scriptContent), 0o700))

	exerciseDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(exerciseDir, "sh"), 0o755))

	desc := RunnerDescriptor{
		Key:  "sh",
		Name: "Shell",
		Open: OpenSpec{
			Interpreter: "sh",
			Args:        []string{scriptFile},
		},
	}

	meta := ExerciseMeta{Dir: exerciseDir, Key: "sh"}
	runner := &descriptorRunner{desc: desc, meta: meta}

	ctx := context.Background()
	require.NoError(t, runner.Open(ctx))
	require.NotNil(t, runner.cmd)
	require.NotNil(t, runner.stdin)

	_ = runner.Close(ctx)
}

// startReapedRunner builds a descriptorRunner around a process that has already
// exited, wiring a reaper goroutine exactly like Open does but leaving the
// `exited` channel OPEN. This forces Close past its early fast-path so it
// exercises the SIGTERM-on-dead-process race directly.
func startReapedRunner(t *testing.T) *descriptorRunner {
	t.Helper()

	cmd := exec.Command("true") // exits immediately
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())

	waitErr := make(chan error, 1)
	exited := make(chan struct{})

	go func() {
		waitErr <- cmd.Wait()
		// Intentionally do NOT close(exited): we want Close to skip its
		// fast-path and reach the SIGTERM line while the process is dead.
	}()

	// Wait for the reaper to observe exit, then put the result back so Close can
	// drain it. This guarantees SIGTERM lands on an already-finished process
	// (the benign os.ErrProcessDone race).
	select {
	case waited := <-waitErr:
		waitErr <- waited
	case <-time.After(2 * time.Second):
		require.Fail(t, "reaper did not finish in time")
	}

	return &descriptorRunner{
		desc:    RunnerDescriptor{Key: "sh", Name: "Shell"},
		cmd:     cmd,
		stdin:   stdin,
		waitErr: waitErr,
		exited:  exited,
	}
}

func TestDescriptorRunner_Close_SigtermOnDeadProcess(t *testing.T) {
	runner := startReapedRunner(t)

	// The process is already dead; SIGTERM returns os.ErrProcessDone. Close must
	// treat that as a successful teardown, not a fatal error.
	require.NoError(t, runner.Close(context.Background()))
}

func TestDescriptorRunner_Close_TimeoutKilledProcess(t *testing.T) {
	scriptFile := filepath.Join(t.TempDir(), "sleep.sh")
	require.NoError(t, os.WriteFile(scriptFile, []byte("#!/bin/sh\nsleep 60\n"), 0o700))

	exerciseDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(exerciseDir, "sh"), 0o755))

	desc := RunnerDescriptor{
		Key:  "sh",
		Name: "Shell",
		Open: OpenSpec{Interpreter: "sh", Args: []string{scriptFile}},
	}
	runner := &descriptorRunner{desc: desc, meta: ExerciseMeta{Dir: exerciseDir, Key: "sh"}}

	require.NoError(t, runner.Open(context.Background()))

	// Mimic a fired per-task deadline: the context is already cancelled when
	// Close runs, driving the ctx.Done() branch that Kills the process. A
	// timeout-killed process yields a non-nil waitErr (signal: killed), which
	// Close must treat as a successful teardown.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, runner.Close(ctx))
}

// TestDescriptorRunner_Run_PanicInExerciseDoesNotHang is the regression guard for
// issue #48 ("panics in exercise code hang the entire application"). A panic in a
// runner's exercise code terminates the subprocess: it writes a stack trace to
// stderr, writes NO result to stdout, and exits non-zero (verified: a Go panic
// exits with status 2). Historically the host spun forever waiting for a result
// from a process that had already died; the single-reaper + `exited` channel now
// makes Run observe the exit and return an error instead of hanging.
//
// The `sh` wrapper below reproduces exactly that I/O signature — read the task,
// emit a crash message on stderr, exit non-zero, never touch stdout — without
// depending on a compiled Go toolchain. The select-with-timeout fails loudly if
// Run ever hangs again.
func TestDescriptorRunner_Run_PanicInExerciseDoesNotHang(t *testing.T) {
	// Mimics a panicking exercise: consume the task line, crash to stderr, exit 2,
	// write nothing to stdout.
	scriptFile := filepath.Join(t.TempDir(), "panic.sh")
	require.NoError(t, os.WriteFile(scriptFile,
		[]byte("#!/bin/sh\nread task\necho 'panic: boom in exercise code' >&2\nexit 2\n"), 0o700))

	exerciseDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(exerciseDir, "sh"), 0o755))

	desc := RunnerDescriptor{
		Key:  "sh",
		Name: "Shell",
		Open: OpenSpec{Interpreter: "sh", Args: []string{scriptFile}},
	}
	runner := &descriptorRunner{desc: desc, meta: ExerciseMeta{Dir: exerciseDir, Key: "sh"}}

	require.NoError(t, runner.Open(context.Background()))
	t.Cleanup(func() { _ = runner.Close(context.Background()) })

	type runResult struct {
		res *protocol.Result
		err error
	}
	done := make(chan runResult, 1)

	go func() {
		// A raw context (no deadline) proves Run returns on subprocess exit alone,
		// not merely because a timeout fired.
		res, err := runner.Run(context.Background(),
			&protocol.Task{TaskID: "1", Part: protocol.PartOne, Input: "irrelevant"})
		done <- runResult{res: res, err: err}
	}()

	select {
	case got := <-done:
		require.Error(t, got.err, "Run must surface the crash as an error, not a result")
		assert.Nil(t, got.res)
		assert.Contains(t, got.err.Error(), "exit code 2")
		assert.Contains(t, got.err.Error(), "boom in exercise code")
	case <-time.After(5 * time.Second):
		require.Fail(t, "Run hung after the exercise subprocess panicked and exited (issue #48 regression)")
	}
}

func TestPrepare_ResolvesRelExerciseDirToken(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root)
	exDir := filepath.Join(root, "exercises", "2015", "01-foo")
	require.NoError(t, os.MkdirAll(exDir, 0o755))

	tmplPath := filepath.Join(root, "go.tmpl")
	require.NoError(t, os.WriteFile(tmplPath, []byte(`import {{ index . "import_path" }}`), 0o600))

	desc := RunnerDescriptor{
		Key: "go", Name: "Go",
		Prepare: PrepareSpec{
			TemplatePath:  tmplPath,
			WrapperExt:    ".go",
			WrapperSubdir: "cmd",
			TemplateVars:  map[string]string{"import_path": "mod/{rel_exercise_dir}/go"},
		},
	}
	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exDir, Key: "go"}
	r := desc.ToCreator()(meta).(*descriptorRunner)

	require.NoError(t, r.Prepare(context.Background()))

	got, err := os.ReadFile(filepath.Join(exDir, "go", "cmd", "runtime-wrapper.go"))
	require.NoError(t, err)
	assert.Contains(t, string(got), "import mod/exercises/2015/01-foo/go")
}

func TestPrepare_RelExerciseDirNoGoMod(t *testing.T) {
	root := t.TempDir() // no go.mod
	exDir := filepath.Join(root, "exercises", "2015", "01-foo")
	require.NoError(t, os.MkdirAll(exDir, 0o755))

	tmplPath := filepath.Join(root, "go.tmpl")
	require.NoError(t, os.WriteFile(tmplPath, []byte(`import {{ index . "import_path" }}`), 0o600))

	desc := RunnerDescriptor{
		Key: "go", Name: "Go",
		Prepare: PrepareSpec{
			TemplatePath: tmplPath, WrapperExt: ".go", WrapperSubdir: "cmd",
			TemplateVars: map[string]string{"import_path": "mod/{rel_exercise_dir}/go"},
		},
	}
	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exDir, Key: "go"}
	r := desc.ToCreator()(meta).(*descriptorRunner)

	err := r.Prepare(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no go.mod found")
}

func TestPrepare_NoRelDirReference_SkipsResolution(t *testing.T) {
	// No go.mod anywhere, but the descriptor never references the token, so
	// Prepare must NOT walk for go.mod and must NOT fail (non-Go runner case).
	root := t.TempDir()
	exDir := filepath.Join(root, "exercises", "2015", "01-foo")
	require.NoError(t, os.MkdirAll(exDir, 0o755))

	tmplPath := filepath.Join(root, "py.tmpl")
	require.NoError(t, os.WriteFile(tmplPath, []byte(`print("hi {{ index . "x" }}")`), 0o600))

	desc := RunnerDescriptor{
		Key: "py", Name: "Python",
		Prepare: PrepareSpec{
			TemplatePath: tmplPath, WrapperExt: ".py",
			TemplateVars: map[string]string{"x": "{title}"},
		},
	}
	meta := ExerciseMeta{Year: 2015, Day: 1, Title: "foo", Dir: exDir, Key: "py"}
	r := desc.ToCreator()(meta).(*descriptorRunner)

	assert.NoError(t, r.Prepare(context.Background()))
}

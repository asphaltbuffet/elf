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

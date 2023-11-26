package runners

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"
	"time"
)

var project string

const (
	goRunnerName                    string = "Go"
	golangInstallation              string = "go"
	golangWrapperFilename           string = "runtime-wrapper.go"
	golangWrapperExecutableFilename string = "runtime-wrapper"
	golangBuildpathBase             string = "github.com/asphaltbuffet/elf/exercises/%s/%s"
)

type golangRunner struct {
	dir                string
	cmd                *exec.Cmd
	wrapperFilepath    string
	executableFilepath string
	stdin              io.WriteCloser
}

func newGolangRunner(dir string) Runner {
	return &golangRunner{
		dir:                dir,
		wrapperFilepath:    filepath.Join(dir, golangWrapperFilename),
		executableFilepath: filepath.Join(dir, golangWrapperExecutableFilename),
	}
}

//go:embed interface/go.tmpl
var golangInterfaceFile []byte

// Start compiles the exercise code and starts the executable.
func (g *golangRunner) Start() error {
	slog.LogAttrs(context.TODO(), slog.LevelDebug, "setting up runner",
		slog.String("dir", g.dir),
	)

	// windows requires .exe extension
	if runtime.GOOS == "windows" {
		g.executableFilepath += ".exe"
	}

	project = getModuleName()

	slog.LogAttrs(context.TODO(), slog.LevelDebug, "paths created",
		slog.String("dir", g.dir),
		slog.String("project", "project"),
	)

	tokens := strings.Split(filepath.ToSlash(g.dir), "/")
	buildPath := filepath.Join(tokens[len(tokens)-3:]...)

	// determine package import path
	// should be like: "github.com/asphaltbuffet/advent-of-code/exercises/2015/01-notQuiteLisp/go"
	importPath := filepath.Join(project, buildPath, "go")

	// generate wrapper code from template
	var wrapperContent []byte
	{
		tpl := template.Must(template.New("").Parse(string(golangInterfaceFile)))
		b := new(bytes.Buffer)

		err := tpl.Execute(b, struct{ ImportPath string }{importPath})
		if err != nil {
			return err
		}

		wrapperContent = b.Bytes()
	}

	// write wrapped code
	if err := os.WriteFile(g.wrapperFilepath, wrapperContent, 0o600); err != nil {
		return err
	}

	slog.LogAttrs(context.Background(), slog.LevelDebug, "building runner",
		slog.String("wrapper", g.wrapperFilepath),
		slog.String("executable", g.executableFilepath),
		slog.String("buildPath", buildPath),
		slog.String("project", project),
		slog.String("importPath", importPath),
	)

	stderrBuffer := new(bytes.Buffer)

	tidycmd := exec.Command(golangInstallation, "mod", "tidy")

	tidycmd.Stderr = stderrBuffer
	if err := tidycmd.Run(); err != nil {
		return fmt.Errorf("tidy failed: %w: %s", err, stderrBuffer.String())
	}

	//nolint:gosec // no user input
	cmd := exec.Command(golangInstallation, "build",
		"-tags", "runtime",
		"-o", g.executableFilepath,
		g.wrapperFilepath)

	cmd.Stderr = stderrBuffer
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w: %s", err, stderrBuffer.String())
	}

	if !cmd.ProcessState.Success() {
		return errors.New("compilation failed")
	}

	absExecPath, err := filepath.Abs(g.executableFilepath)
	if err != nil {
		return err
	}

	// run executable for exercise (wrapped)

	g.cmd = exec.Command(absExecPath)
	cmd.Dir = g.dir

	stdin, err := setupBuffers(g.cmd)
	if err != nil {
		return err
	}

	g.stdin = stdin

	return g.cmd.Start()
}

func (g *golangRunner) Stop() error {
	const processExitTimeout time.Duration = 5 * time.Second

	if g.cmd == nil || g.cmd.Process == nil {
		return nil
	}

	// First try to send a SIGTERM.
	if err := g.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to go process: %w", err)
	}

	// Wait for the process to exit, but not forever.
	done := make(chan error, 1)
	go func() {
		_, err := g.cmd.Process.Wait()
		done <- err
	}()

	// wait up to 5 seconds for the process to exit.
	select {
	case <-time.After(processExitTimeout):
		if err := g.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill go process: %w", err)
		}
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to stop go process: %w", err)
		}
	}

	return nil
}

func (g *golangRunner) Cleanup() error {
	var wrapperErr, execErr error

	if g.wrapperFilepath != "" {
		wrapperErr = os.Remove(g.wrapperFilepath)
	}

	if g.executableFilepath != "" {
		execErr = os.Remove(g.executableFilepath)
	}

	return errors.Join(wrapperErr, execErr)
}

func (g *golangRunner) Run(task *Task) (*Result, error) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("marshalling task to json: %w", err)
	}

	_, err = g.stdin.Write(append(taskJSON, '\n'))
	if err != nil {
		return nil, fmt.Errorf("writing task to stdin: %w", err)
	}

	r := new(Result)

	if jsonErr := readJSONFromCommand(r, g.cmd); jsonErr != nil {
		return nil, jsonErr
	}

	return r, nil
}

// String returns a string representation of the runner type.
func (g *golangRunner) String() string {
	return goRunnerName
}

func getModuleName() string {
	errBuf := new(bytes.Buffer)
	outBuf := new(bytes.Buffer)

	cmd := exec.Command(golangInstallation, "list", "-m")
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		panic("failed to get module name: " + errBuf.String())
	}

	return strings.Trim(outBuf.String(), "\n")
}

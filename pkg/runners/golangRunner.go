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

	"github.com/asphaltbuffet/elf/pkg/protocol"
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
func (g *golangRunner) Prepare(ctx context.Context) error {
	//nolint:sloglint // runner has no logger, uses global for debug
	slog.LogAttrs(ctx, slog.LevelDebug, "setting up runner", slog.String("dir", g.dir))

	if runtime.GOOS == "windows" {
		g.executableFilepath += ".exe"
	}

	project = getModuleName()

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

	if err := os.WriteFile(g.wrapperFilepath, wrapperContent, 0o600); err != nil {
		return err
	}

	//nolint:sloglint // runner has no logger, uses global for debug
	slog.LogAttrs(ctx, slog.LevelDebug, "building runner",
		slog.String("wrapper", g.wrapperFilepath),
		slog.String("executable", g.executableFilepath),
		slog.String("importPath", importPath),
	)

	stderrBuffer := new(bytes.Buffer)

	tidycmd := exec.CommandContext(ctx, golangInstallation, "mod", "tidy")
	tidycmd.Stderr = stderrBuffer

	if err := tidycmd.Run(); err != nil {
		return fmt.Errorf("tidy failed: %w: %s", err, stderrBuffer.String())
	}

	//nolint:gosec // no user input
	cmd := exec.CommandContext(ctx, golangInstallation, "build",
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

	return nil
}

func (g *golangRunner) Open(ctx context.Context) error {
	absExecPath, err := filepath.Abs(g.executableFilepath)
	if err != nil {
		return err
	}

	g.cmd = exec.CommandContext(ctx, absExecPath)
	g.cmd.Dir = g.dir

	stdin, err := setupBuffers(g.cmd)
	if err != nil {
		return err
	}

	g.stdin = stdin

	return g.cmd.Start()
}

func (g *golangRunner) Close(ctx context.Context) error {
	if g.cmd == nil || g.cmd.Process == nil {
		return nil
	}

	if err := g.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to go process: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := g.cmd.Process.Wait()
		done <- err
	}()

	select {
	case <-ctx.Done():
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

func (g *golangRunner) Run(_ context.Context, task *protocol.Task) (*protocol.Result, error) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("marshalling task to json: %w", err)
	}

	_, err = g.stdin.Write(append(taskJSON, '\n'))
	if err != nil {
		return nil, fmt.Errorf("writing task to stdin: %w", err)
	}

	r := new(protocol.Result)

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

	cmd := exec.CommandContext(context.Background(), golangInstallation, "list", "-m")
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		panic("failed to get module name: " + errBuf.String())
	}

	return strings.Trim(outBuf.String(), "\n")
}

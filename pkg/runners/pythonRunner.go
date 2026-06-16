package runners

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

const (
	pythonRunnerName      string = "Python"
	python3Installation   string = "python3"
	pythonWrapperFilename string = "runtime-wrapper.py"
)

type pythonRunner struct {
	cmd             *exec.Cmd
	dir             string
	stdin           io.WriteCloser
	wrapperFilepath string
}

func newPythonRunner(dir string) Runner {
	return &pythonRunner{
		dir:             dir,
		wrapperFilepath: filepath.Join(dir, pythonWrapperFilename),
	}
}

//go:embed interface/python.templ
var pythonInterface []byte

func (p *pythonRunner) Prepare(_ context.Context) error {
	return os.WriteFile(p.wrapperFilepath, pythonInterface, 0o600)
}

func (p *pythonRunner) Open(ctx context.Context) error {
	absDir, err := filepath.Abs(p.dir)
	if err != nil {
		return err
	}

	pythonPathVar := strings.Join([]string{
		filepath.Join(absDir, "../../..", "lib"),
		filepath.Join(absDir, "py"),
	}, ":")

	p.cmd = exec.CommandContext(
		ctx,
		python3Installation,
		"-B",
		pythonWrapperFilename,
	)
	p.cmd.Env = append(p.cmd.Env, "PYTHONPATH="+pythonPathVar)
	p.cmd.Dir = p.dir

	stdin, err := setupBuffers(p.cmd)
	if err != nil {
		return err
	}

	p.stdin = stdin

	return p.cmd.Start()
}

func (p *pythonRunner) Close(ctx context.Context) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to python process: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := p.cmd.Process.Wait()
		done <- err
	}()

	select {
	case <-ctx.Done():
		if err := p.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill python process: %w", err)
		}
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to stop python process: %w", err)
		}
	}

	return nil
}

func (p *pythonRunner) Cleanup() error {
	err := os.Remove(p.wrapperFilepath)

	if errors.Is(err, os.ErrNotExist) {
		// already gone, maybe log this?
		return nil
	}

	return err
}

func (p *pythonRunner) Run(_ context.Context, task *protocol.Task) (*protocol.Result, error) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("marshalling task to json: %w", err)
	}

	_, err = p.stdin.Write(append(taskJSON, '\n'))
	if err != nil {
		return nil, fmt.Errorf("writing task to stdin: %w", err)
	}

	r := new(protocol.Result)
	if jsonErr := readJSONFromCommand(r, p.cmd); jsonErr != nil {
		return nil, jsonErr
	}

	return r, nil
}

func (p *pythonRunner) String() string {
	return pythonRunnerName
}

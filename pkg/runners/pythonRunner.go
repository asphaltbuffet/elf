package runners

import (
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
	"time"
)

const (
	pythonRunnerName      string = "Python"
	python3Installation   string = "python3"
	pythonWrapperFilename string = "runtime-wrapper.py"
)

type pythonRunner struct {
	dir             string
	cmd             *exec.Cmd
	wrapperFilepath string
	stdin           io.WriteCloser
}

func newPythonRunner(dir string) Runner {
	return &pythonRunner{
		dir:             dir,
		wrapperFilepath: filepath.Join(dir, pythonWrapperFilename),
	}
}

//go:embed interface/python.templ
var pythonInterface []byte

func (p *pythonRunner) Start() error {
	// Save interaction code
	if err := os.WriteFile(p.wrapperFilepath, pythonInterface, 0o600); err != nil {
		return err
	}

	// Sort out PYTHONPATH
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	absDir, err := filepath.Abs(p.dir)
	if err != nil {
		return err
	}

	pythonPathVar := strings.Join([]string{
		filepath.Join(cwd, "lib"),   // so we can use aocpy
		filepath.Join(absDir, "py"), // so we can import stuff in the exercises directory
	}, ":")

	p.cmd = exec.Command(python3Installation, "-B", pythonWrapperFilename) // -B prevents .pyc files from being written
	p.cmd.Env = append(p.cmd.Env, "PYTHONPATH="+pythonPathVar)
	p.cmd.Dir = p.dir

	stdin, err := setupBuffers(p.cmd)
	if err != nil {
		return err
	}

	p.stdin = stdin

	return p.cmd.Start()
}

func (p *pythonRunner) Stop() error {
	const processExitTimeout time.Duration = 5 * time.Second

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	// First try to send a SIGTERM.
	if err := p.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to python process: %w", err)
	}

	// Wait for the process to exit, but not forever.
	done := make(chan error, 1)
	go func() {
		_, err := p.cmd.Process.Wait()
		done <- err
	}()

	// wait up to 5 seconds for the process to exit.
	select {
	case <-time.After(processExitTimeout):
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

	switch {
	case errors.Is(err, os.ErrNotExist):
		// already gone, maybe log this?
		fallthrough

	case err == nil:
		return nil

	default:
		return err
	}
}

func (p *pythonRunner) Run(task *Task) (*Result, error) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("marshalling task to json: %w", err)
	}

	_, err = p.stdin.Write(append(taskJSON, '\n'))
	if err != nil {
		return nil, fmt.Errorf("writing task to stdin: %w", err)
	}

	res := new(Result)
	if jsonErr := readJSONFromCommand(res, p.cmd); jsonErr != nil {
		return nil, jsonErr
	}

	return res, nil
}

func (p *pythonRunner) String() string {
	return pythonRunnerName
}

package runners

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

var project string

const (
	golangInstallation              = "go"
	golangWrapperFilename           = "runtime-wrapper.go"
	golangWrapperExecutableFilename = "runtime-wrapper"
	golangBuildpathBase             = "github.com/asphaltbuffet/elf/exercises/%s/%s"
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
		dir: dir,
	}
}

//go:embed interface/go.tmpl
var golangInterfaceFile []byte

// Start compiles the exercise code and starts the executable.
func (g *golangRunner) Start() error {
	// exercises/<year>/<day>-<title>/go
	g.wrapperFilepath = filepath.Join(g.dir, golangWrapperFilename)
	fmt.Println("[DEBUG] wrapper file path: ", g.wrapperFilepath)
	g.executableFilepath = filepath.Join(g.dir, golangWrapperExecutableFilename)

	// windows requires .exe extension
	if runtime.GOOS == "windows" {
		g.executableFilepath += ".exe"
	}

	fmt.Println("[DEBUG] exec file path: ", g.executableFilepath)

	// determine package import path
	buildPath := filepath.Join(".", g.dir)
	fmt.Println("[DEBUG] build path: ", buildPath)

	project = getModuleName()

	// should be like: "github.com/asphaltbuffet/advent-of-code/exercises/2015/01-notQuiteLisp/go"
	importPath := filepath.Join(project, buildPath, "go")
	fmt.Println("[DEBUG] import path: ", importPath)

	// generate wrapper code from template
	var wrapperContent []byte
	{
		tpl := template.Must(template.New("").Parse(string(golangInterfaceFile)))
		b := new(bytes.Buffer)
		err := tpl.Execute(b, struct {
			ImportPath string
		}{importPath})
		if err != nil {
			return err
		}
		wrapperContent = b.Bytes()
	}

	// write wrapped code
	if err := os.WriteFile(g.wrapperFilepath, wrapperContent, 0o600); err != nil {
		return err
	}

	stderrBuffer := new(bytes.Buffer)

	tidycmd := exec.Command(golangInstallation, "mod", "tidy")

	tidycmd.Stderr = stderrBuffer
	if err := tidycmd.Run(); err != nil {
		return fmt.Errorf("tidy failed: %w: %s", err, stderrBuffer.String())
	}

	//nolint:gosec // no user input
	cmd := exec.Command(
		golangInstallation,
		"build",
		"-tags", "runtime",
		"-o", g.executableFilepath,
		g.wrapperFilepath)

	cmd.Stderr = stderrBuffer
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %s: %s", err, stderrBuffer.String())
	}

	if !cmd.ProcessState.Success() {
		return errors.New("compilation failed")
	}

	absExecPath, err := filepath.Abs(g.executableFilepath)
	if err != nil {
		return err
	}

	// run executable for exercise (wrapped)
	//nolint:gosec // no user input
	g.cmd = exec.Command(absExecPath)
	cmd.Dir = g.dir

	if stdin, err := setupBuffers(g.cmd); err != nil {
		return err
	} else {
		g.stdin = stdin
	}

	return g.cmd.Start()
}

func (g *golangRunner) Stop() error {
	if g.cmd == nil || g.cmd.Process == nil {
		return nil
	}

	return g.cmd.Process.Kill()
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
		return nil, err
	}

	_, err = g.stdin.Write(append(taskJSON, '\n'))
	if err != nil {
		return nil, err
	}

	res := new(Result)

	if err := readJSONFromCommand(res, g.cmd); err != nil {
		return nil, err
	}

	return res, nil
}

func getModuleName() string {
	const golangInstallation string = "go"

	stderrBuffer := new(bytes.Buffer)
	stdoutBuffer := new(bytes.Buffer)

	cmd := exec.Command(golangInstallation, "list", "-m")
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer

	if err := cmd.Run(); err != nil {
		panic("failed to get module name: " + stderrBuffer.String())
	}

	return stdoutBuffer.String()
}

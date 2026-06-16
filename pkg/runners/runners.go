package runners

import "github.com/asphaltbuffet/elf/pkg/protocol"

// Runner is an interface defining methods for starting, stopping,
// cleaning up, and running tasks.
type Runner interface {
	// Start initializes the runner.
	Start() error
	// Stop terminates the runner.
	Stop() error
	// Cleanup handles any cleanup operations required after running a task.
	Cleanup() error
	// Run executes a given task and returns the result or an error.
	Run(task *protocol.Task) (*protocol.Result, error)

	String() string
}

// ResultOrError holds either the result of a task or an error.
type ResultOrError struct {
	Result *protocol.Result
	Error  error
}

// RunnerCreator is a function type that takes a directory string
// as input and returns a Runner.
type RunnerCreator func(dir string) Runner

// Available maps runner type strings (like "go" or "py") to their respective
// RunnerCreator functions.
var Available = map[string]RunnerCreator{
	"go": newGolangRunner,
	"py": newPythonRunner,
}

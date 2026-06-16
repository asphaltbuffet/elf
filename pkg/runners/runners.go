package runners

import (
	"context"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// Runner is an interface defining methods for starting, stopping,
// cleaning up, and running tasks.
type Runner interface {
	// Prepare compiles or otherwise readies the runner before launch.
	// May be a no-op for pre-built or discovered runners.
	Prepare(ctx context.Context) error

	// Open starts the runner subprocess.
	Open(ctx context.Context) error

	// Run executes a single task and returns the result.
	Run(ctx context.Context, task *protocol.Task) (*protocol.Result, error)

	// Close stops the runner subprocess gracefully.
	Close(ctx context.Context) error

	// Cleanup removes any artifacts created by Prepare. May be a no-op.
	Cleanup() error

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

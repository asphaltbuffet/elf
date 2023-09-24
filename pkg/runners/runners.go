package runners

// Part represents a section or segment of a task or process.
type Part uint8

const (
	PartOne   Part = iota + 1 // PartOne is the first part of the exercise.
	PartTwo                   // PartTwo is the second part of the exercise.
	Visualize                 // Visualize is the optional visualization of exercise processing.
)

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
	Run(task *Task) (*Result, error)
}

// ResultOrError holds either the result of a task or an error.
// It is useful for communicating results and errors from asynchronous operations.
type ResultOrError struct {
	Result *Result
	Error  error
}

// RunnerCreator is a function type that takes a directory string
// as input and returns a Runner. This allows for dynamic creation
// of different types of runners based on the provided directory.
type RunnerCreator func(dir string) Runner

// Available maps runner type strings (like "go" or "py") to their respective
// RunnerCreator functions. This allows for the dynamic creation of runners
// based on the runner type.
var Available = map[string]RunnerCreator{
	"go": newGolangRunner,
	"py": newPythonRunner,
}

// RunnerNames maps runner type strings (like "go" or "py") to more
// human-friendly names (like "Golang" or "Python").
var RunnerNames = map[string]string{
	"go": "Golang",
	"py": "Python",
}

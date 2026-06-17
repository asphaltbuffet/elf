package runners

import (
	"context"
	"path/filepath"

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

// ExerciseMeta carries the identity fields of an Exercise to a RunnerCreator.
type ExerciseMeta struct {
	Year  int
	Day   int
	Title string
	Dir   string // exercise root path (e.g. "exercises/2015/01-foo")
	Key   string // language key / subdirectory name (e.g. "py")
}

// LangDir returns the language subdirectory path: Dir/Key.
func (m ExerciseMeta) LangDir() string {
	return filepath.Join(m.Dir, m.Key)
}

// RunnerCreator constructs a Runner for a given exercise.
type RunnerCreator func(meta ExerciseMeta) Runner

// Available is the runtime runner registry, populated at startup from config.
var Available = map[string]RunnerCreator{}

// RegisterFromDescriptors populates Available from a slice of RunnerDescriptors.
// Existing entries are overwritten.
func RegisterFromDescriptors(descs []RunnerDescriptor) {
	for _, d := range descs {
		desc := d // capture loop variable
		Available[desc.Key] = desc.ToCreator()
	}
}

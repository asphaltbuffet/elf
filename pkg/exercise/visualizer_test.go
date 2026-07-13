package exercise

import (
	"context"
	"log/slog"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// TestExercise_Visualize_HappyPath is a characterization test: it captures the
// OBSERVABLE contract of Visualize (returned results + emitted events) before
// the inlined tail is replaced with a runTaskList call, so the refactor can be
// verified against unchanged behavior rather than assumed safe.
func TestExercise_Visualize_HappyPath(t *testing.T) {
	fs := afero.NewMemMapFs()

	e := &Exercise{
		Title:    "Test Exercise",
		Language: "go",
		Year:     2015,
		Day:      1,
		Path:     "exercises/2015/01-test-exercise",
		Data: &Data{
			InputFileName: "input.txt",
		},
	}

	require.NoError(t, fs.MkdirAll(e.Path, 0o755))
	require.NoError(t, afero.WriteFile(fs, "exercises/2015/01-test-exercise/input.txt", []byte("puzzle input"), 0o644))

	runner := &fakeRunner{
		name: "Go",
		runFn: func(_ context.Context, task *protocol.Task) (*protocol.Result, error) {
			return &protocol.Result{TaskID: task.TaskID, Ok: true, Output: "drew.svg", Duration: 0.1}, nil
		},
	}

	var events []tasks.Event
	cb := func(ev tasks.Event) { events = append(events, ev) }

	results, err := e.Visualize(context.Background(), fs, slog.Default(), runner, "/out", cb)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, tasks.StatusUnverified, results[0].Status) // Visualize has no expected answer
	assert.Equal(t, "drew.svg", results[0].Output)

	// metaEvent + Planned + Started + Finished, in that order.
	require.Len(t, events, 4)
	assert.Equal(t, tasks.EventMeta, events[0].Kind)
	assert.Equal(t, tasks.EventPlanned, events[1].Kind)
	assert.Equal(t, tasks.EventStarted, events[2].Kind)
	assert.Equal(t, tasks.EventFinished, events[3].Kind)
}

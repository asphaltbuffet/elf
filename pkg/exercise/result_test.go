package exercise

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func Test_buildResult(t *testing.T) {
	type args struct {
		r *protocol.Result
	}

	tests := []struct {
		name string
		args args
		want tasks.Result
	}{
		{
			name: "sucessful run",
			args: args{
				r: &protocol.Result{
					TaskID:   "solve.1",
					Ok:       true,
					Output:   "good output",
					Duration: 0.042,
				},
			},
			want: tasks.Result{
				ID:       "solve.1",
				Type:     tasks.Solve,
				Part:     1,
				SubPart:  0,
				Status:   tasks.StatusPassed,
				Output:   "good output",
				Expected: "good output",
				Duration: 0.042,
			},
		},
		{
			name: "not ok",
			args: args{
				r: &protocol.Result{
					TaskID:   "solve.2",
					Ok:       false,
					Output:   "error text",
					Duration: 0.042,
				},
			},
			want: tasks.Result{
				ID:       "solve.2",
				Type:     tasks.Solve,
				Part:     2,
				SubPart:  0,
				Status:   tasks.StatusError,
				Output:   "⤷ saying:error text",
				Expected: "",
				Duration: 0.042,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildResult(tt.args.r, "good output")

			assert.Equal(t, tt.want, got)
		})
	}
}

// fakeRunner is a minimal runners.Runner for driving runTaskList directly.
// runFn lets each test decide how Run behaves per call. restartRunner calls
// Close then Open (result.go:111-121); we count each Close as one restart
// because runTaskList's only Close call is inside restartRunner (the driver
// does not manage the prologue's defer-Close — these direct tests never invoke
// the prologue). This makes restarts an unambiguous count of timeout-restart
// cycles, independent of Open-ordering subtleties.
type fakeRunner struct {
	name     string
	runFn    func(ctx context.Context, task *protocol.Task) (*protocol.Result, error)
	restarts int
}

func (f *fakeRunner) String() string                  { return f.name }
func (f *fakeRunner) Prepare(_ context.Context) error { return nil }
func (f *fakeRunner) Open(_ context.Context) error    { return nil }
func (f *fakeRunner) Close(_ context.Context) error   { f.restarts++; return nil }
func (f *fakeRunner) Cleanup() error                  { return nil }
func (f *fakeRunner) Run(ctx context.Context, task *protocol.Task) (*protocol.Result, error) {
	return f.runFn(ctx, task)
}

func TestExercise_runTaskList_HappyPath(t *testing.T) {
	e := &Exercise{} // taskTimeout zero → runWithTimeout skips the deadline

	runner := &fakeRunner{
		name: "Go",
		runFn: func(_ context.Context, task *protocol.Task) (*protocol.Result, error) {
			// protocol.Result.Ok must be true or buildResult marks the task StatusError.
			return &protocol.Result{TaskID: task.TaskID, Ok: true, Output: "42", Duration: 0.1}, nil
		},
	}

	plan := []plannedTask{
		{
			task:     &protocol.Task{TaskID: tasks.MakeTaskID(tasks.Solve, protocol.PartOne), Part: protocol.PartOne},
			expected: "42",
		},
	}

	var events []tasks.Event
	cb := func(ev tasks.Event) { events = append(events, ev) }

	results, err := e.runTaskList(context.Background(), runner, plan, cb)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "42", results[0].Output)
	assert.Equal(t, tasks.StatusPassed, results[0].Status) // output "42" == expected "42"

	// One Planned, one Started, one Finished — in that order.
	// NOTE: the discriminator field on tasks.Event is Kind (not Type — Type holds the TaskType).
	require.Len(t, events, 3)
	assert.Equal(t, tasks.EventPlanned, events[0].Kind)
	assert.Equal(t, tasks.EventStarted, events[1].Kind)
	assert.Equal(t, tasks.EventFinished, events[2].Kind)
}

func TestExercise_runTaskList_TimeoutRestartsAndContinues(t *testing.T) {
	e := &Exercise{taskTimeout: 10 * time.Millisecond}

	var calls int
	runner := &fakeRunner{
		name: "Go",
		runFn: func(ctx context.Context, task *protocol.Task) (*protocol.Result, error) {
			calls++
			if calls == 1 {
				// First task: block past the deadline so runWithTimeout returns errTaskTimeout.
				<-ctx.Done()
				return nil, ctx.Err()
			}
			// Second task: succeed immediately.
			return &protocol.Result{TaskID: task.TaskID, Ok: true, Output: "ok", Duration: 0.01}, nil
		},
	}

	plan := []plannedTask{
		{
			task:     &protocol.Task{TaskID: tasks.MakeTaskID(tasks.Solve, protocol.PartOne), Part: protocol.PartOne},
			expected: "",
		},
		{
			task:     &protocol.Task{TaskID: tasks.MakeTaskID(tasks.Solve, protocol.PartTwo), Part: protocol.PartTwo},
			expected: "",
		},
	}

	results, err := e.runTaskList(context.Background(), runner, plan, nil)

	require.NoError(t, err)
	require.Len(t, results, 2)
	// First result is the timeout; second is the success. Loop continued past the timeout.
	assert.Equal(t, tasks.StatusTimeout, results[0].Status, "first task timed out")
	assert.Equal(t, "ok", results[1].Output)
	assert.Equal(t, 1, runner.restarts, "runner restarted exactly once after the timeout")
}

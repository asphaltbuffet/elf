package exercise

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

type plannedTask struct {
	task     *protocol.Task
	expected string
}

// metaEvent builds the run-level metadata event a renderer uses for chrome:
// the exercise identity for the header and the runner's human-readable name
// (e.g. "Rust") rather than the lookup key (e.g. "rs").
func (e *Exercise) metaEvent(runner runners.Runner) tasks.Event {
	return tasks.MetaEvent(tasks.Meta{
		Year:     e.Year,
		Day:      e.Day,
		Number:   e.Number,
		Title:    e.Title,
		Language: runner.String(),
	})
}

// buildResult constructs a tasks.Result from a runner result and expected value.
// This is pure data logic with no rendering side effects.
func buildResult(r *protocol.Result, expected string) tasks.Result {
	taskType, part, subpart := tasks.ParseTaskID(r.TaskID)

	result := tasks.Result{
		ID:       r.TaskID,
		Type:     taskType,
		Part:     part,
		SubPart:  subpart,
		Duration: r.Duration,
	}

	switch {
	case taskType == tasks.Benchmark:
		result.Status = tasks.StatusPassed
		result.Output = r.Output
		result.Expected = ""
		result.Duration = r.Duration

	case !r.Ok:
		result.Status = tasks.StatusError
		result.Output = fmt.Sprint("⤷ saying:", r.Output)

	case expected == "":
		result.Status = tasks.StatusUnverified
		result.Output = r.Output

	case r.Output == expected:
		result.Status = tasks.StatusPassed
		result.Output = r.Output
		result.Expected = expected

	case r.Output != expected:
		result.Status = tasks.StatusFailed
		result.Output = fmt.Sprintf("⤷ got %q, but expected %q", r.Output, expected)

	default:
		result.Status = tasks.StatusInvalid
		result.Output = r.Output
		result.Expected = expected
	}

	return result
}

// RenderReport writes a scaffold Report to w: one indented row per file, showing the relative path
// followed by a colorized outcome (green added, yellow replaced, faint skipped). The exercise
// directory itself is not printed here — the caller prints it as a header.
func RenderReport(w io.Writer, r Report) {
	pathStyle := lipgloss.NewStyle().PaddingLeft(ReportIndent).Width(ReportPathWidth)

	for _, e := range r {
		fmt.Fprintln(w, pathStyle.Render(e.Path), scaffoldOutcomeStyle(e.Outcome))
	}
}

// timeoutResult synthesizes a TIMEOUT result for a task that exceeded its deadline.
func timeoutResult(taskID string) tasks.Result {
	taskType, part, subpart := tasks.ParseTaskID(taskID)

	return tasks.Result{
		ID:      taskID,
		Type:    taskType,
		Part:    part,
		SubPart: subpart,
		Status:  tasks.StatusTimeout,
	}
}

// errTaskTimeout is returned by runWithTimeout when the per-task deadline fires.
var errTaskTimeout = errors.New("task timed out")

// restartRunner closes the killed runner and reopens it so subsequent tasks
// can run. This is necessary because Kill() terminates the subprocess; the
// runner's exited channel is closed and its cmd is dead, so the next Run call
// would fail immediately with exit code -1.
func restartRunner(ctx context.Context, r runners.Runner) error {
	if err := r.Close(ctx); err != nil {
		return fmt.Errorf("closing runner after timeout: %w", err)
	}

	if err := r.Open(ctx); err != nil {
		return fmt.Errorf("reopening runner after timeout: %w", err)
	}

	return nil
}

// runWithTimeout executes a single task, wrapping it in a deadline context when timeout > 0.
// If the task exceeds the deadline, errTaskTimeout is returned so callers can render a
// TIMEOUT result instead of propagating a fatal error.
func runWithTimeout(
	ctx context.Context,
	r runners.Runner,
	task *protocol.Task,
	timeout time.Duration,
) (*protocol.Result, error) {
	if timeout <= 0 {
		return r.Run(ctx, task)
	}

	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := r.Run(taskCtx, task)
	if err != nil && errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
		return nil, errTaskTimeout
	}

	return result, err
}

// runTaskList runs an ordered list of tasks against a prepared, open runner,
// emitting the Planned/Started/Finished event stream a renderer consumes. It
// owns the timeout-and-restart policy: a task that exceeds e.taskTimeout yields
// a timeoutResult, the runner is restarted, and the loop continues with the
// next task. Any other run error aborts the list. This is the single loop
// shared by Solve (main tasks), Test, and Visualize.
func (e *Exercise) runTaskList(
	ctx context.Context,
	runner runners.Runner,
	plan []plannedTask,
	cb func(tasks.Event),
) ([]tasks.Result, error) {
	// Announce the full task list up front so a renderer can show pending rows.
	if cb != nil {
		for _, t := range plan {
			tt, part, sub := tasks.ParseTaskID(t.task.TaskID)
			cb(tasks.PlannedEvent(t.task.TaskID, tt, part, sub, ""))
		}
	}

	results := make([]tasks.Result, 0, len(plan))

	for _, t := range plan {
		if cb != nil {
			tt, part, sub := tasks.ParseTaskID(t.task.TaskID)
			cb(tasks.StartedEvent(t.task.TaskID, tt, part, sub, ""))
		}

		result, err := runWithTimeout(ctx, runner, t.task, e.taskTimeout)
		if errors.Is(err, errTaskTimeout) {
			r := timeoutResult(t.task.TaskID)
			if cb != nil {
				cb(tasks.FinishedEvent(r, ""))
			}

			results = append(results, r)

			if restartErr := restartRunner(ctx, runner); restartErr != nil {
				return nil, restartErr
			}

			continue
		} else if err != nil {
			return nil, err
		}

		r := buildResult(result, t.expected)
		if cb != nil {
			cb(tasks.FinishedEvent(r, ""))
		}

		results = append(results, r)
	}

	return results, nil
}

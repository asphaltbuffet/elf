package exercise

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

type testTask struct {
	task     *protocol.Task
	expected string
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

// renderResult writes a styled representation of a task result to w.
// It reconstructs the CLI visual output from the result data alone.
func renderResult(w io.Writer, result tasks.Result) {
	if result.Type == tasks.Benchmark {
		return
	}

	dur, err := time.ParseDuration(fmt.Sprintf("%fs", result.Duration))
	if err != nil {
		panic(err)
	}

	name := taskStyle(int(result.Part), result.SubPart)

	var output, extra, followUpText lipgloss.Style
	var printExtra bool

	switch result.Status {
	case tasks.StatusInvalid:
		// nothing to render for invalid results

	case tasks.StatusError:
		output = lipgloss.NewStyle().
			Bold(true).Align(lipgloss.Center).
			Foreground(lipgloss.Color("9")).
			SetString("ERROR")

		extra = extraStyle.Foreground(bad).SetString(result.Output)
		printExtra = true

	case tasks.StatusUnverified:
		output = statusStyle.Foreground(newAns).Background(lipgloss.Color("0")).SetString("NEW")
		followUpText = timeStyle.SetString(dur.String())

		extra = extraStyle.SetString("⤷ " + result.Output)
		printExtra = true

	case tasks.StatusPassed:
		output = lipgloss.NewStyle().Bold(true).Align(lipgloss.Right).Foreground(lipgloss.Color("46")).SetString("PASS")
		followUpText = timeStyle.SetString(dur.String())

		if result.Type == tasks.Solve {
			extra = extraStyle.Foreground(lipgloss.Color("7")).SetString("⤷ " + result.Output)
			printExtra = true
		}

	case tasks.StatusFailed:
		output = statusStyle.Foreground(bad).SetString("FAIL")
		extra = extraStyle.Foreground(bad).SetString(result.Output)
		printExtra = true

	case tasks.StatusTimeout:
		output = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).SetString("TIMEOUT")
	}

	fmt.Fprintln(w, name, output, followUpText)

	if printExtra {
		fmt.Fprintln(w, extra)
	}
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

func handleTaskResult(w io.Writer, r *protocol.Result, expected string) tasks.Result {
	result := buildResult(r, expected)
	renderResult(w, result)

	return result
}

// timeoutResult synthesizes a TIMEOUT result for a task that exceeded its deadline,
// renders it inline, and returns it so callers can continue to the next task.
func timeoutResult(w io.Writer, taskID string) tasks.Result {
	taskType, part, subpart := tasks.ParseTaskID(taskID)

	r := tasks.Result{
		ID:      taskID,
		Type:    taskType,
		Part:    part,
		SubPart: subpart,
		Status:  tasks.StatusTimeout,
	}

	renderResult(w, r)

	return r
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

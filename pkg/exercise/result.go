package exercise

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/pkg/protocol"
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

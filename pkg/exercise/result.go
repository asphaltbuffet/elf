package exercise

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

type testTask struct {
	task     *runners.Task
	expected string
}

//nolint:funlen // this function is long, but it's mostly formatting
func handleTaskResult(w io.Writer, r *runners.Result, expected string) tasks.Result {
	taskType, part, subpart := tasks.ParseTaskID(r.TaskID)

	result := tasks.Result{
		ID:       r.TaskID,
		Type:     taskType,
		Part:     part,
		SubPart:  subpart,
		Duration: r.Duration,
	}

	dur, err := time.ParseDuration(fmt.Sprintf("%fs", r.Duration)) // TODO: store duration as time.Duration
	if err != nil {
		panic(err)
	}

	name := taskStyle(int(part), subpart)

	var output, extra, followUpText lipgloss.Style
	var printExtra bool

	switch {
	case taskType == tasks.Benchmark:
		// for now, we assume benchmarks are always successful
		result.Status = tasks.StatusPassed
		result.Output = r.Output
		result.Expected = "" // no expected output for benchmarks
		result.Duration = r.Duration
	case !r.Ok:
		result.Status = tasks.StatusError
		result.Output = fmt.Sprint("⤷ saying:", r.Output)

		output = lipgloss.NewStyle().
			Bold(true).Align(lipgloss.Center).
			Foreground(lipgloss.Color("9")).
			SetString("ERROR")

		extra = extraStyle.Foreground(bad).SetString("⤷ saying: " + r.Output)
		printExtra = true

	case expected == "":
		result.Status = tasks.StatusUnverified
		result.Output = r.Output

		output = statusStyle.Foreground(newAns).Background(lipgloss.Color("0")).SetString("NEW")
		followUpText = timeStyle.SetString(dur.String())

		extra = extraStyle.SetString("⤷ " + r.Output)
		printExtra = true

	case r.Output == expected:
		result.Status = tasks.StatusPassed
		result.Output = r.Output
		result.Expected = expected

		output = lipgloss.NewStyle().Bold(true).Align(lipgloss.Right).Foreground(lipgloss.Color("46")).SetString("PASS")
		followUpText = timeStyle.SetString(dur.String())

		if taskType == tasks.Solve {
			extra = extraStyle.Foreground(lipgloss.Color("7")).SetString("⤷ " + r.Output)
			printExtra = true
		}

	case r.Output != expected:
		result.Status = tasks.StatusFailed
		result.Output = fmt.Sprintf("⤷ got %q, but expected %q", r.Output, expected)

		output = statusStyle.Foreground(bad).SetString("FAIL")
		extra = extraStyle.Foreground(bad).SetString(result.Output)
		printExtra = true

	default:
		result.Status = tasks.StatusInvalid
		result.Output = r.Output
		result.Expected = expected
	}

	if taskType != tasks.Benchmark {
		fmt.Fprintln(w, name, output, followUpText)

		// show extra info
		if printExtra {
			fmt.Fprintln(w, extra)
		}
	}

	return result
}

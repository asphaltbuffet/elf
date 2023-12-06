package advent

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func (e *Exercise) Test() error {
	if e == nil || *e == (Exercise{}) {
		slog.Error("exercise is nil")
		return fmt.Errorf("exercise is nil")
	}

	testerLog := slog.With(slog.String("fn", "Test"), slog.String("exercise", e.Title))
	testerLog.Debug("solving", slog.String("language", e.Language))

	if err := e.runner.Start(); err != nil {
		testerLog.Error("starting runner", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(os.Stdout, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	if err := runTests(e.runner, e.Data); err != nil {
		testerLog.Error("running tests", tint.Err(err))
		return err
	}

	return nil
}

func makeTestID(part runners.Part, n int) string {
	return fmt.Sprintf("test.%d.%d", part, n)
}

func parseTestID(id string) (runners.Part, int) {
	var a runners.Part
	var b int

	_, err := fmt.Sscanf(id, "test.%d.%d", &a, &b)
	if err != nil {
		panic(err)
	}

	return a, b
}

type testTask struct {
	task     *runners.Task
	expected string
}

func runTests(runner runners.Runner, data *Data) error {
	var tasks []testTask

	tasks = append(tasks, makeTestTasks(runners.PartOne, data.TestCases.One)...)
	tasks = append(tasks, makeTestTasks(runners.PartTwo, data.TestCases.Two)...)

	for _, t := range tasks {
		result, err := runner.Run(t.task)
		if err != nil {
			return err
		}

		handleTestResult(result, t.expected)
	}

	return nil
}

func makeTestTasks(p runners.Part, tests []*Test) []testTask {
	var tasks []testTask

	for i, t := range tests {
		tasks = append(tasks, testTask{
			task: &runners.Task{
				TaskID:    makeTestID(p, i),
				Part:      p,
				Input:     t.Input,
				OutputDir: "",
			},
			expected: t.Expected,
		})
	}

	return tasks
}

// func handleTestResult(r *runners.Result, expected string) {
// 	part, n := parseTestID(r.TaskID)

// 	name := taskStyle(int(part), n)

// 	passed := r.Output == expected

	var status, followUpText lipgloss.Style

	switch {
	case !r.Ok:
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			SetString("DID NOT COMPLETE")

// 		followUpText = lipgloss.NewStyle().
// 			Faint(true).
// 			Italic(true).
// 			Foreground(lipgloss.Color("7")).
// 			SetString(fmt.Sprintf("saying %q", r.Output))

// 	case passed:
// 		status = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).SetString("PASS")

// 	default:
// 		status = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).SetString("FAIL")
// 	}

// 	if followUpText.String() == "" {
// 		followUpText = lipgloss.NewStyle().
// 			Faint(true).
// 			Italic(true).
// 			Foreground(lipgloss.Color("7")).
// 			SetString(fmt.Sprintf("in %s", humanize.SIWithDigits(r.Duration, 1, "s")))
// 	}

// 	fmt.Println(name, status, followUpText)

// 	if !passed && r.Ok {
// 		extra := lipgloss.NewStyle().
// 			Bold(true).
// 			Foreground(lipgloss.Color("1")).
// 			PaddingLeft(4). //nolint:gomnd // hard-coded padding for now
// 			SetString(fmt.Sprintf("⤷ expected %q, got %q", expected, r.Output))

		fmt.Println(extra)
	}
}

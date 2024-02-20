package advent

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func (e *Exercise) Solve(skipTests bool) ([]tasks.Result, error) {
	logger := e.logger.With(slog.String("exercise", e.Title))
	logger.Debug("solving", slog.String("language", e.Language))

	results := []tasks.Result{}

	inputFile := filepath.Join(e.Path, e.Data.InputFileName)
	input, err := afero.ReadFile(e.appFs, inputFile)
	if err != nil {
		logger.Error("reading input file", slog.String("path", inputFile), tint.Err(err))
		return nil, err
	}

	e.Data.InputData = string(input)

	if err = e.runner.Start(); err != nil {
		logger.Error("starting runner", tint.Err(err))
		return nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(os.Stdout, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	if !skipTests {
		fmt.Printf("Testing (%s)...\n", e.runner)

		var tr []tasks.Result

		tr, err = runTests(e.runner, e.Data)
		if err != nil {
			logger.Error("running tests", tint.Err(err))
			return nil, err
		}

		results = append(results, tr...)
	}

	fmt.Printf("Solving (%s)...\n", e.runner)

	mainResults, err := runMainTasks(e.runner, e.Data)
	if err != nil {
		logger.Error("running main tasks", tint.Err(err))
		return nil, err
	}

	results = append(results, mainResults...)

	return results, nil
}

func runMainTasks(runner runners.Runner, data *Data) ([]tasks.Result, error) {
	var solveTasks []testTask

	solveTasks = append(solveTasks, makeMainTasks(runners.PartOne, data)...)
	solveTasks = append(solveTasks, makeMainTasks(runners.PartTwo, data)...)

	results := make([]tasks.Result, 0, len(solveTasks))

	for _, t := range solveTasks {
		result, err := runner.Run(t.task)
		if err != nil {
			slog.Error("running task", slog.String("id", t.task.TaskID), tint.Err(err))
			return nil, err
		}

		r := handleTaskResult(os.Stdout, result, t.expected)
		results = append(results, r)
	}

	return results, nil
}

func makeMainTasks(part runners.Part, data *Data) []testTask {
	var solveTasks []testTask
	var expected string

	if part == runners.PartOne {
		expected = data.Answers.One
	} else {
		expected = data.Answers.Two
	}

	solveTasks = append(solveTasks, testTask{
		task: &runners.Task{
			TaskID:    tasks.MakeTaskID(tasks.Solve, part),
			Part:      part,
			Input:     data.InputData,
			OutputDir: "",
		},
		expected: expected,
	})

	return solveTasks
}

func handleTaskResult(w io.Writer, r *runners.Result, expected string) tasks.Result {
	taskType, part, subpart := tasks.ParseTaskID(r.TaskID)

	result := tasks.Result{
		ID:       r.TaskID,
		Type:     taskType,
		Part:     part,
		SubPart:  subpart,
		Duration: r.Duration,
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
		// followUpText = timeStyle.SetString(humanize.SIWithDigits(r.Duration, 1, "s"))
		followUpText = timeStyle.SetString(fmt.Sprintf("%.2f ms", r.Duration*1000))

		extra = extraStyle.SetString("⤷ " + r.Output)
		printExtra = true

	case r.Output == expected:
		result.Status = tasks.StatusPassed
		result.Output = r.Output
		result.Expected = expected

		output = lipgloss.NewStyle().Bold(true).Align(lipgloss.Right).Foreground(lipgloss.Color("46")).SetString("PASS")
		// followUpText = timeStyle.SetString(humanize.SIWithDigits(r.Duration, 1, "s"))
		followUpText = timeStyle.SetString(fmt.Sprintf("%.2f ms", r.Duration*1000))

		if taskType == tasks.Solve {
			extra = extraStyle.Foreground(lipgloss.Color("7")).SetString("⤷ " + r.Output)
			printExtra = true
		}

	case r.Output != expected:
		result.Status = tasks.StatusFailed
		result.Output = fmt.Sprintf("⤷ got %q, but expected %q", r.Output, expected)

		output = statusStyle.Foreground(bad).SetString("FAIL")
		// followUpText = mainNoteStyle(humanize.SIWithDigits(r.Duration, 1, "s"), r.Ok)

		extra = extraStyle.Foreground(bad).SetString()
		printExtra = true

	default:
		result.Status = tasks.StatusInvalid
		result.Output = r.Output
		result.Expected = expected
	}

	slog.Debug("handling result", slog.Group("result", "id", r.TaskID, "ok", r.Ok, "output", r.Output))

	if taskType != tasks.Benchmark {
		fmt.Fprintln(w, name, output, followUpText)

		// show extra info
		if printExtra {
			fmt.Fprintln(w, extra)
		}
	}

	return result
}

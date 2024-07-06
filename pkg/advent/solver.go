package advent

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

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

	fmt.Fprintln(e.writer, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	if !skipTests {
		fmt.Fprintf(e.writer, "Testing (%s)...\n", e.runner)

		var tr []tasks.Result

		tr, err = e.runTests()
		if err != nil {
			logger.Error("running tests", tint.Err(err))
			return nil, err
		}

		results = append(results, tr...)
	}

	fmt.Fprintf(e.writer, "Solving (%s)...\n", e.runner)

	mainResults, err := e.runMainTasks()
	if err != nil {
		logger.Error("running main tasks", tint.Err(err))
		return nil, err
	}

	results = append(results, mainResults...)

	return results, nil
}

func (e *Exercise) runMainTasks() ([]tasks.Result, error) {
	var solveTasks []testTask

	solveTasks = append(solveTasks, makeMainTasks(runners.PartOne, e.Data)...)
	solveTasks = append(solveTasks, makeMainTasks(runners.PartTwo, e.Data)...)

	results := make([]tasks.Result, 0, len(solveTasks))

	for _, t := range solveTasks {
		result, err := e.runner.Run(t.task)
		if err != nil {
			return nil, err
		}

		r := handleTaskResult(e.writer, result, t.expected)
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
		extra = extraStyle.Foreground(bad).SetString()
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

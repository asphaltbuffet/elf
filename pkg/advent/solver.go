package advent

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func (e *Exercise) Solve(skipTests bool) error {
	solverLog := slog.With(slog.String("fn", "Solve"), slog.String("exercise", e.Title))
	solverLog.Debug("solving", slog.String("language", e.Language))

	input, err := os.ReadFile(e.Data.InputFile)
	if err != nil {
		solverLog.Error("reading input file", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	e.Data.Input = string(input)

	if err = e.runner.Start(); err != nil {
		solverLog.Error("starting runner", slog.String("path", e.Data.InputFile), tint.Err(err))
		return err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(os.Stdout, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	if !skipTests {
		fmt.Fprintln(os.Stdout, taskHeaderStyle("Testing..."))

		if err = runTests(e.runner, e.Data); err != nil {
			solverLog.Error("running tests", tint.Err(err))
			return err
		}
	}

	fmt.Fprintln(os.Stdout, taskHeaderStyle("Solving..."))

	if err = runMainTasks(e.runner, e.Data); err != nil {
		solverLog.Error("running main tasks", tint.Err(err))
		return err
	}

	return nil
}

func makeMainID(part runners.Part) string {
	return fmt.Sprintf("main.%d", part)
}

func parseMainID(id string) runners.Part {
	var p runners.Part

	_, err := fmt.Sscanf(id, "main.%d", &p)
	if err != nil {
		slog.Error("parsing main id", slog.Group("task", "id", id), tint.Err(err))
		panic(err)
	}

	return p
}

func runMainTasks(runner runners.Runner, data *Data) error {
	var tasks []testTask

	tasks = append(tasks, makeMainTasks(runners.PartOne, data)...)
	tasks = append(tasks, makeMainTasks(runners.PartTwo, data)...)

	for _, t := range tasks {
		result, err := runner.Run(t.task)
		if err != nil {
			slog.Error("running task",
				slog.Group("result", "id", result.TaskID, "ok", result.Ok, "output", result.Output),
				tint.Err(err))
			return err
		}

		handleMainResult(os.Stdout, result, t.expected)
	}

	return nil
}

func makeMainTasks(p runners.Part, data *Data) []testTask {
	var tasks []testTask
	var exp string

	if p == runners.PartOne {
		exp = data.Answers.One
	} else {
		exp = data.Answers.Two
	}

	tasks = append(tasks, testTask{
		task: &runners.Task{
			TaskID:    makeMainID(p),
			Part:      p,
			Input:     data.Input,
			OutputDir: "",
		},
		expected: exp,
	})

	return tasks
}

type TaskType string

const (
	TaskTest      TaskType = "test"
	TaskMain      TaskType = "main"
	TaskBenchmark TaskType = "benchmark"
	TaskVisual    TaskType = "vis"
	TaskInvalid   TaskType = "invalid"
)

type part runners.Part

func parseTaskID(id string) (TaskType, part, int) {
	tokens := strings.Split(id, ".")

	switch t := TaskType(tokens[0]); t {
	case TaskTest, TaskVisual: // test/visual
		p, err := strconv.Atoi(tokens[1])
		if err != nil {
			slog.Error("invalid part type", slog.String("id", id))
			panic("invalid part type")
		}

		n, err := strconv.Atoi(tokens[2])
		if err != nil {
			slog.Error("invalid sub-test number", slog.String("id", id))
			panic("invalid sub-test number")
		}

		return t, part(p), n

	case TaskMain, TaskBenchmark: // main/benchmark
		p, err := strconv.Atoi(tokens[1])
		if err != nil {
			slog.Error("invalid part type", slog.String("id", id))
			panic("invalid part type")
		}

		return t, part(p), -1

	default:
		slog.Error("invalid task type", slog.String("id", id))
		return TaskInvalid, 0, 0

	}
}

type resultType int

const (
	resultTypeUnknown resultType = iota
	resultTypePassed
	resultTypeNew
	resultTypeFailed
	resultTypeError
)

func handleMainResult(w io.Writer, r *runners.Result, expected string) {
	var status resultType

	_, part, _ := parseTaskID(r.TaskID)

	name := taskStyle(int(part), -1)

	if r.Ok && r.Output == expected {
		status = resultTypePassed
	} else if r.Ok && expected == "" {
		status = resultTypeNew
	} else if r.Ok && r.Output != expected {
		status = resultTypeFailed
	} else if !r.Ok {
		status = resultTypeError
	} else {
		status = resultTypeUnknown // shouldn't be able to get here
	}

	var output, followUpText lipgloss.Style

	switch status {
	case resultTypeError:
		output = mainResultStyle("did not complete", r.Ok)
		followUpText = mainNoteStyle(r.Output, r.Ok)

	case resultTypeNew: // TODO: differentiate between new and passed
		output = lipgloss.NewStyle().Bold(true).SetString(r.Output)
		followUpText = mainNoteStyle(humanize.SIWithDigits(r.Duration, 1, "s"), r.Ok)

	case resultTypePassed:
		output = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).SetString(r.Output)
		followUpText = mainNoteStyle(humanize.SIWithDigits(r.Duration, 1, "s"), r.Ok)

	case resultTypeFailed:
		output = mainResultStyle(r.Output, r.Ok)
		followUpText = mainNoteStyle(humanize.SIWithDigits(r.Duration, 1, "s"), r.Ok)
	}

	slog.Debug("handling main result", slog.Group("result", "id", r.TaskID, "ok", r.Ok, "output", r.Output))

	fmt.Fprintln(w, name, output, followUpText)

	// show extra info for failures
	if status == resultTypeFailed {
		extra := lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("1")).
			Foreground(lipgloss.Color("0")).
			PaddingLeft(4). //nolint:gomnd // hard-coded padding for now
			SetString(fmt.Sprintf("⤷ accepted answer is %q", expected))

		fmt.Println(extra)
	}
}

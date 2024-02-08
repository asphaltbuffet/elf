package advent

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func (e *Exercise) Solve(skipTests bool) ([]TaskResult, error) {
	solverLog := slog.With(slog.String("exercise", e.Title))
	solverLog.Debug("solving", slog.String("language", e.Language))

	results := []TaskResult{}

	inputFile := filepath.Join(e.Path, e.Data.InputFileName)
	input, err := os.ReadFile(inputFile)
	if err != nil {
		solverLog.Error("reading input file", slog.String("path", inputFile), tint.Err(err))
		return nil, err
	}

	e.Data.Input = string(input)

	if err = e.runner.Start(); err != nil {
		solverLog.Error("starting runner", tint.Err(err))
		return nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(os.Stdout, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	if !skipTests {
		fmt.Printf("Testing (%s)...\n", e.runner)

		var tr []TaskResult

		tr, err = runTests(e.runner, e.Data)
		if err != nil {
			solverLog.Error("running tests", tint.Err(err))
			return nil, err
		}

		results = append(results, tr...)
	}

	fmt.Printf("Solving (%s)...\n", e.runner)

	mainResults, err := runMainTasks(e.runner, e.Data)
	if err != nil {
		solverLog.Error("running main tasks", tint.Err(err))
		return nil, err
	}

	results = append(results, mainResults...)

	return results, nil
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

func runMainTasks(runner runners.Runner, data *Data) ([]TaskResult, error) {
	var tasks []testTask

	tasks = append(tasks, makeMainTasks(runners.PartOne, data)...)
	tasks = append(tasks, makeMainTasks(runners.PartTwo, data)...)

	results := make([]TaskResult, 0, len(tasks))

	for _, t := range tasks {
		result, err := runner.Run(t.task)
		if err != nil {
			slog.Error("running task",
				slog.Group("result", "id", result.TaskID, "ok", result.Ok, "output", result.Output),
				tint.Err(err))
			return nil, err
		}

		r := handleTaskResult(os.Stdout, result, t.expected)
		results = append(results, r)
	}

	return results, nil
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

type (
	TaskPart    runners.Part
	TaskSubPart int
)

func parseTaskID(id string) (TaskType, TaskPart, TaskSubPart) {
	tokens := strings.Split(id, ".")

	switch t := TaskType(tokens[0]); t {
	case TaskBenchmark, TaskTest, TaskVisual: // test/visual
		p, err := strconv.ParseUint(tokens[1], 10, 8)
		// p, err := strconv.Atoi(tokens[1])
		if err != nil {
			slog.Error("invalid part type", slog.String("id", id))
			panic("invalid part type")
		}

		n, err := strconv.Atoi(tokens[2])
		if err != nil {
			slog.Error("invalid sub-test number", slog.String("id", id))
			panic("invalid sub-test number")
		}

		return t, TaskPart(p), TaskSubPart(n)

	case TaskMain: // main
		p, err := strconv.ParseUint(tokens[1], 10, 8)
		// p, err := strconv.Atoi(tokens[1])
		if err != nil {
			slog.Error("invalid part type", slog.String("id", id))
			panic("invalid part type")
		}

		return t, TaskPart(p), TaskSubPart(-1)

	case TaskInvalid:
		fallthrough

	default:
		slog.Error("invalid task type", slog.String("id", id))
		return TaskInvalid, TaskPart(0), TaskSubPart(0)
	}
}

//go:generate stringer -type=TaskStatus
type TaskStatus int

const (
	Invalid TaskStatus = iota
	Passed
	Unverified
	Failed
	Error
)

type TaskResult struct {
	ID       string
	Type     TaskType
	Part     TaskPart
	SubPart  TaskSubPart
	Status   TaskStatus
	Output   string
	Expected string
	Duration float64
}

func handleTaskResult(w io.Writer, r *runners.Result, expected string) TaskResult {
	taskType, part, subpart := parseTaskID(r.TaskID)

	result := TaskResult{
		ID:       r.TaskID,
		Type:     taskType,
		Part:     part,
		SubPart:  subpart,
		Duration: r.Duration,
	}

	name := taskStyle(int(part), int(subpart))

	var output, extra, followUpText lipgloss.Style
	var printExtra bool

	switch {
	case taskType == TaskBenchmark:
		// for now, we assume benchmarks are always successful
		result.Status = Passed
		result.Output = r.Output
		result.Expected = "" // no expected output for benchmarks
		result.Duration = r.Duration
	case !r.Ok:
		result.Status = Error
		result.Output = fmt.Sprint("⤷ saying:", r.Output)

		output = lipgloss.NewStyle().
			Bold(true).Align(lipgloss.Center).
			Foreground(lipgloss.Color("9")).
			SetString("ERROR")

		extra = extraStyle.Foreground(bad).SetString("⤷ saying: " + r.Output)
		printExtra = true

	case expected == "":
		result.Status = Unverified
		result.Output = r.Output

		output = statusStyle.Foreground(newAns).Background(lipgloss.Color("0")).SetString("NEW")
		// followUpText = timeStyle.SetString(humanize.SIWithDigits(r.Duration, 1, "s"))
		followUpText = timeStyle.SetString(fmt.Sprintf("%.2f ms", r.Duration*1000))

		extra = extraStyle.SetString("⤷ " + r.Output)
		printExtra = true

	case r.Output == expected:
		result.Status = Passed
		result.Output = r.Output
		result.Expected = expected

		output = lipgloss.NewStyle().Bold(true).Align(lipgloss.Right).Foreground(lipgloss.Color("46")).SetString("PASS")
		// followUpText = timeStyle.SetString(humanize.SIWithDigits(r.Duration, 1, "s"))
		followUpText = timeStyle.SetString(fmt.Sprintf("%.2f ms", r.Duration*1000))

		if taskType == TaskMain {
			extra = extraStyle.Foreground(lipgloss.Color("7")).SetString("⤷ " + r.Output)
			printExtra = true
		}

	case r.Output != expected:
		result.Status = Failed
		result.Output = fmt.Sprintf("⤷ got %q, but expected %q", r.Output, expected)

		output = statusStyle.Foreground(bad).SetString("FAIL")
		// followUpText = mainNoteStyle(humanize.SIWithDigits(r.Duration, 1, "s"), r.Ok)

		extra = extraStyle.Foreground(bad).SetString()
		printExtra = true

	default:
		result.Status = Invalid
		result.Output = r.Output
		result.Expected = expected
	}

	slog.Debug("handling result", slog.Group("result", "id", r.TaskID, "ok", r.Ok, "output", r.Output))

	if taskType != TaskBenchmark {
		fmt.Fprintln(w, name, output, followUpText)

		// show extra info
		if printExtra {
			fmt.Fprintln(w, extra)
		}
	}

	return result
}

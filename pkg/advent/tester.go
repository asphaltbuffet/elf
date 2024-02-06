package advent

import (
	"fmt"
	"log/slog"
	"os"

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

	_, err := runTests(e.runner, e.Data)
	if err != nil {
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

func runTests(runner runners.Runner, data *Data) ([]TaskResult, error) {
	var tasks []testTask

	tasks = append(tasks, makeTestTasks(runners.PartOne, data.TestCases.One)...)
	tasks = append(tasks, makeTestTasks(runners.PartTwo, data.TestCases.Two)...)

	results := make([]TaskResult, 0, len(tasks))

	for _, t := range tasks {
		result, err := runner.Run(t.task)
		if err != nil {
			slog.Error("running test task",
				slog.Group("result", "id", result.TaskID, "ok", result.Ok, "output", result.Output),
				tint.Err(err))
			return nil, err
		}

		r := handleTaskResult(os.Stdout, result, t.expected)
		results = append(results, r)
	}

	return results, nil
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

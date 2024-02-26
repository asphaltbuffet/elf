package advent

import (
	"fmt"
	"log/slog"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func (e *Exercise) Test() ([]tasks.Result, error) {
	if *e == (Exercise{}) {
		return nil, fmt.Errorf("exercise is empty")
	}

	logger := e.logger.With(slog.String("fn", "Test"), slog.String("exercise", e.Title))
	logger.Debug("testing", slog.String("language", e.Language))

	if err := e.runner.Start(); err != nil {
		logger.Error("starting runner",
			slog.String("path", e.Path),
			slog.String("implementation", e.runner.String()),
			tint.Err(err))

		return nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(e.writer, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	results, err := e.runTests()
	if err != nil {
		logger.Error("running tests", tint.Err(err))

		return nil, err
	}

	return results, nil
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

func (e *Exercise) runTests() ([]tasks.Result, error) {
	var testTasks []testTask

	testTasks = append(testTasks, makeTestTasks(runners.PartOne, e.Data.TestCases.One)...)
	testTasks = append(testTasks, makeTestTasks(runners.PartTwo, e.Data.TestCases.Two)...)

	results := make([]tasks.Result, 0, len(testTasks))

	for _, t := range testTasks {
		result, err := e.runner.Run(t.task)
		if err != nil {
			e.logger.Error("running test task", tint.Err(err))
			return nil, err
		}

		r := handleTaskResult(e.writer, result, t.expected)
		results = append(results, r)
	}

	return results, nil
}

func makeTestTasks(p runners.Part, tests []*Test) []testTask {
	var testTasks []testTask

	for i, t := range tests {
		testTasks = append(testTasks, testTask{
			task: &runners.Task{
				TaskID:    tasks.MakeTaskID(tasks.Test, p, i),
				Part:      p,
				Input:     t.Input,
				OutputDir: "",
			},
			expected: t.Expected,
		})
	}

	return testTasks
}

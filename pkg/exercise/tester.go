package exercise

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Test runs the exercise test cases and returns pass/fail results for each.
func (e *Exercise) Test() ([]tasks.Result, error) {
	ctx := context.Background()

	if e.Year == 0 && e.Day == 0 && e.Title == "" {
		return nil, errors.New("exercise is empty")
	}

	logger := e.logger.With(slog.String("fn", "Test"), slog.String("exercise", e.Title))
	logger.Debug("testing", slog.String("language", e.Language))

	if err := e.runner.Prepare(ctx); err != nil {
		logger.Error("preparing runner",
			slog.String("path", e.Path),
			slog.String("implementation", e.runner.String()),
			tint.Err(err))

		return nil, err
	}

	if err := e.runner.Open(ctx); err != nil {
		logger.Error("opening runner",
			slog.String("path", e.Path),
			slog.String("implementation", e.runner.String()),
			tint.Err(err))

		return nil, err
	}

	defer func() {
		_ = e.runner.Close(ctx)
		_ = e.runner.Cleanup()
	}()

	fmt.Fprintln(e.writer, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	results, err := e.runTests(ctx)
	if err != nil {
		logger.Error("running tests", tint.Err(err))

		return nil, err
	}

	return results, nil
}

func (e *Exercise) runTests(ctx context.Context) ([]tasks.Result, error) {
	var testTasks []testTask

	testTasks = append(testTasks, makeTestTasks(protocol.PartOne, e.Data.TestCases.One)...)
	testTasks = append(testTasks, makeTestTasks(protocol.PartTwo, e.Data.TestCases.Two)...)

	results := make([]tasks.Result, 0, len(testTasks))

	for _, t := range testTasks {
		result, err := e.runner.Run(ctx, t.task)
		if err != nil {
			e.logger.ErrorContext(ctx, "running test task", tint.Err(err))
			return nil, err
		}

		r := handleTaskResult(e.writer, result, t.expected)
		if e.onResult != nil {
			e.onResult(r)
		}
		results = append(results, r)
	}

	return results, nil
}

func makeTestTasks(p protocol.Part, tests []*Test) []testTask {
	var testTasks []testTask

	for i, t := range tests {
		testTasks = append(testTasks, testTask{
			task: &protocol.Task{
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

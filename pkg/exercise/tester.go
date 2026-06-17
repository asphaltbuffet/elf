package exercise

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Test runs the exercise test cases and returns pass/fail results for each.
func (e *Exercise) Test(
	ctx context.Context,
	logger *slog.Logger,
	runner runners.Runner,
	w io.Writer,
	cb func(tasks.Result),
) ([]tasks.Result, error) {
	if e.Year == 0 && e.Day == 0 && e.Title == "" {
		return nil, errors.New("exercise is empty")
	}

	logger = logger.With(slog.String("fn", "Test"), slog.String("exercise", e.Title))
	logger.DebugContext(ctx, "testing", slog.String("language", e.Language))

	defer func() {
		_ = runner.Close(ctx)
		_ = runner.Cleanup()
	}()

	if err := runner.Prepare(ctx); err != nil {
		logger.ErrorContext(ctx, "preparing runner",
			slog.String("path", e.Path),
			slog.String("implementation", runner.String()),
			tint.Err(err))

		return nil, err
	}

	if err := runner.Open(ctx); err != nil {
		logger.ErrorContext(ctx, "opening runner",
			slog.String("path", e.Path),
			slog.String("implementation", runner.String()),
			tint.Err(err))

		return nil, err
	}

	fmt.Fprintln(w, headerStyle(fmt.Sprintf("ADVENT OF CODE %d\nDay %d: %s", e.Year, e.Day, e.Title)))

	results, err := e.runTests(ctx, runner, w, cb)
	if err != nil {
		logger.ErrorContext(ctx, "running tests", tint.Err(err))

		return nil, err
	}

	return results, nil
}

func (e *Exercise) runTests(
	ctx context.Context,
	runner runners.Runner,
	w io.Writer,
	cb func(tasks.Result),
) ([]tasks.Result, error) {
	var testTasks []testTask

	testTasks = append(testTasks, makeTestTasks(protocol.PartOne, e.Data.TestCases.One)...)
	testTasks = append(testTasks, makeTestTasks(protocol.PartTwo, e.Data.TestCases.Two)...)

	results := make([]tasks.Result, 0, len(testTasks))

	for _, t := range testTasks {
		result, err := runner.Run(ctx, t.task)
		if err != nil {
			return nil, err
		}

		r := handleTaskResult(w, result, t.expected)
		if cb != nil {
			cb(r)
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

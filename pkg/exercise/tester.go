package exercise

import (
	"context"
	"errors"
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
	cb func(tasks.Event),
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

	if cb != nil {
		cb(e.metaEvent(runner))
	}

	results, err := e.runTests(ctx, runner, cb)
	if err != nil {
		logger.ErrorContext(ctx, "running tests", tint.Err(err))

		return nil, err
	}

	return results, nil
}

func (e *Exercise) runTests(
	ctx context.Context,
	runner runners.Runner,
	cb func(tasks.Event),
) ([]tasks.Result, error) {
	var testTasks []plannedTask

	for _, part := range e.declaredParts() {
		testTasks = append(testTasks, makeTestTasks(part, e.testCasesFor(part))...)
	}

	// Announce the full task list up front so a renderer can show pending rows.
	if cb != nil {
		for _, t := range testTasks {
			tt, part, sub := tasks.ParseTaskID(t.task.TaskID)
			cb(tasks.PlannedEvent(t.task.TaskID, tt, part, sub, ""))
		}
	}

	results := make([]tasks.Result, 0, len(testTasks))

	for _, t := range testTasks {
		if cb != nil {
			tt, part, sub := tasks.ParseTaskID(t.task.TaskID)
			cb(tasks.StartedEvent(t.task.TaskID, tt, part, sub, ""))
		}

		result, err := runWithTimeout(ctx, runner, t.task, e.taskTimeout)
		if errors.Is(err, errTaskTimeout) {
			r := timeoutResult(t.task.TaskID)
			if cb != nil {
				cb(tasks.FinishedEvent(r, ""))
			}

			results = append(results, r)

			if restartErr := restartRunner(ctx, runner); restartErr != nil {
				return nil, restartErr
			}

			continue
		} else if err != nil {
			return nil, err
		}

		r := buildResult(result, t.expected)
		if cb != nil {
			cb(tasks.FinishedEvent(r, ""))
		}

		results = append(results, r)
	}

	return results, nil
}

// testCasesFor returns the test cases for the given part. declaredParts only
// ever yields PartOne or PartTwo, so PartTwo is the sole special case.
func (e *Exercise) testCasesFor(part protocol.Part) []*Test {
	if part == protocol.PartTwo {
		return e.Data.TestCases.Two
	}

	return e.Data.TestCases.One
}

func makeTestTasks(p protocol.Part, tests []*Test) []plannedTask {
	var testTasks []plannedTask

	for i, t := range tests {
		testTasks = append(testTasks, plannedTask{
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

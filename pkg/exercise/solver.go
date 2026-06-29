package exercise

import (
	"context"
	"errors"
	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Solve runs the exercise solution and optionally skips the pre-solve test run.
func (e *Exercise) Solve(
	ctx context.Context,
	fs afero.Fs,
	logger *slog.Logger,
	runner runners.Runner,
	cb func(tasks.Event),
	skipTests bool,
) ([]tasks.Result, error) {
	logger = logger.With(slog.String("exercise", e.Title))
	logger.DebugContext(ctx, "solving", slog.String("language", e.Language))

	input, err := e.readInput(fs)
	if err != nil {
		logger.ErrorContext(ctx, "reading input file", slog.String("path", e.Data.InputFileName), tint.Err(err))
		return nil, err
	}

	e.Data.InputData = input

	defer func() {
		_ = runner.Close(ctx)
		_ = runner.Cleanup()
	}()

	if err = runner.Prepare(ctx); err != nil {
		logger.ErrorContext(ctx, "preparing runner", tint.Err(err))
		return nil, err
	}

	if err = runner.Open(ctx); err != nil {
		logger.ErrorContext(ctx, "opening runner", tint.Err(err))
		return nil, err
	}

	if cb != nil {
		cb(e.metaEvent(runner))
	}

	results := []tasks.Result{}

	if !skipTests {
		var tr []tasks.Result

		tr, err = e.runTests(ctx, runner, cb)
		if err != nil {
			logger.ErrorContext(ctx, "running tests", tint.Err(err))
			return nil, err
		}

		results = append(results, tr...)
	}

	mainResults, err := e.runMainTasks(ctx, runner, cb)
	if err != nil {
		logger.ErrorContext(ctx, "running main tasks", tint.Err(err))
		return nil, err
	}

	results = append(results, mainResults...)

	return results, nil
}

func (e *Exercise) runMainTasks(
	ctx context.Context,
	runner runners.Runner,
	cb func(tasks.Event),
) ([]tasks.Result, error) {
	var solveTasks []testTask

	solveTasks = append(solveTasks, makeMainTasks(protocol.PartOne, e.Data)...)
	solveTasks = append(solveTasks, makeMainTasks(protocol.PartTwo, e.Data)...)

	// Announce the full task list up front so a renderer can show pending rows.
	if cb != nil {
		for _, t := range solveTasks {
			tt, part, sub := tasks.ParseTaskID(t.task.TaskID)
			cb(tasks.PlannedEvent(t.task.TaskID, tt, part, sub, ""))
		}
	}

	results := make([]tasks.Result, 0, len(solveTasks))

	for _, t := range solveTasks {
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

func makeMainTasks(part protocol.Part, data *Data) []testTask {
	var solveTasks []testTask

	var expected string

	if part == protocol.PartOne {
		expected = data.Answers.One
	} else {
		expected = data.Answers.Two
	}

	solveTasks = append(solveTasks, testTask{
		task: &protocol.Task{
			TaskID:    tasks.MakeTaskID(tasks.Solve, part),
			Part:      part,
			Input:     data.InputData,
			OutputDir: "",
		},
		expected: expected,
	})

	return solveTasks
}

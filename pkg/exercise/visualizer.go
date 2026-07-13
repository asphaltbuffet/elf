package exercise

import (
	"context"
	"log/slog"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Visualize runs the visualization task for the exercise and returns the result.
func (e *Exercise) Visualize(
	ctx context.Context,
	fs afero.Fs,
	logger *slog.Logger,
	runner runners.Runner,
	outdir string,
	cb func(tasks.Event),
) ([]tasks.Result, error) {
	logger = logger.With(slog.String("exercise", e.Title))
	logger.DebugContext(ctx, "visualizing", slog.String("language", e.Language))

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

	plan := []plannedTask{
		{
			task: &protocol.Task{
				TaskID:    tasks.MakeTaskID(tasks.Visualize, protocol.Visualize, 0),
				Part:      protocol.Visualize,
				Input:     e.Data.InputData,
				OutputDir: outdir,
			},
			expected: "",
		},
	}

	return e.runTaskList(ctx, runner, plan, cb)
}

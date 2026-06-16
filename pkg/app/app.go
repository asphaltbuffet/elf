// Package app wires together exercise loading, runner construction, and task
// execution. Callers provide path/language/IO; App provides infrastructure.
package app

import (
	"context"
	"io"
	"log/slog"

	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// App holds shared infrastructure used across CLI commands and TUI screens.
type App struct {
	FS     afero.Fs
	Logger *slog.Logger
}

// New constructs an App from a config.Config, extracting FS and Logger.
func New(cfg config.Config) *App {
	return &App{
		FS:     cfg.GetFs(),
		Logger: cfg.GetLogger(),
	}
}

// Solve loads the exercise at path, constructs a runner for language, and
// executes a solve run. customInput overrides input.txt when non-empty.
func (a *App) Solve(
	ctx context.Context,
	path, language, customInput string,
	w io.Writer,
	cb func(tasks.Result),
	skipTests bool,
) ([]tasks.Result, error) {
	rc, ok := runners.Available[language]
	if !ok {
		return nil, exercise.ErrNoRunner
	}

	ex, err := exercise.Load(path, language, customInput, a.FS, a.Logger)
	if err != nil {
		return nil, err
	}

	return ex.Solve(ctx, a.FS, a.Logger, rc(path), w, cb, skipTests)
}

// Test loads the exercise at path and runs its test suite.
func (a *App) Test(
	ctx context.Context,
	path, language, customInput string,
	w io.Writer,
	cb func(tasks.Result),
) ([]tasks.Result, error) {
	rc, ok := runners.Available[language]
	if !ok {
		return nil, exercise.ErrNoRunner
	}

	ex, err := exercise.Load(path, language, customInput, a.FS, a.Logger)
	if err != nil {
		return nil, err
	}

	return ex.Test(ctx, a.Logger, rc(path), w, cb)
}

// Benchmark loads the exercise at path and benchmarks all available
// implementations for the given number of iterations.
func (a *App) Benchmark(
	ctx context.Context,
	path, language string,
	w io.Writer,
	cb func(tasks.Result),
	iterations int,
) ([]tasks.Result, error) {
	ex, err := exercise.Load(path, language, "", a.FS, a.Logger)
	if err != nil {
		return nil, err
	}

	bmk := exercise.NewBenchmarker(ex)

	return bmk.Benchmark(ctx, a.FS, a.Logger, w, cb, iterations)
}

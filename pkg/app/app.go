// Package app wires together exercise loading, runner construction, and task
// execution. Callers provide path/language/IO; App provides infrastructure.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/analyze"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// App holds shared infrastructure used across CLI commands and TUI screens.
type App struct {
	FS          afero.Fs
	Logger      *slog.Logger
	cfg         config.Config
	language    string
	baseDir     string
	taskTimeout time.Duration
}

// RegisterRunners populates runners.Available from the given config.
// Must be called before any exercise operation.
func RegisterRunners(cfg config.Config) {
	runners.RegisterFromDescriptors(cfg.GetRunners())
}

// New constructs an App from a config.Config, extracting FS, Logger, and display values.
func New(cfg config.Config) *App {
	return &App{
		FS:          cfg.GetFs(),
		Logger:      cfg.GetLogger(),
		cfg:         cfg,
		language:    cfg.GetLanguage(),
		baseDir:     cfg.GetBaseDir(),
		taskTimeout: cfg.GetTaskTimeout(),
	}
}

// SetTaskTimeout overrides the per-task timeout after construction.
// A value <=0 disables the timeout.
func (a *App) SetTaskTimeout(d time.Duration) {
	a.taskTimeout = d
}

// Language returns the configured default language.
func (a *App) Language() string { return a.language }

// BaseDir returns the configured exercise base directory.
func (a *App) BaseDir() string { return a.baseDir }

// GetFs returns the filesystem used by the App.
func (a *App) GetFs() afero.Fs { return a.FS }

// GetLogger returns the logger used by the App.
func (a *App) GetLogger() *slog.Logger { return a.Logger }

// Solve loads the exercise at path, constructs a runner for language, and
// executes a solve run. customInput overrides input.txt when non-empty.
func (a *App) Solve(
	ctx context.Context,
	path, language, customInput string,
	cb func(tasks.Event),
	skipTests bool,
) ([]tasks.Result, error) {
	ex, err := exercise.Load(path, language, customInput, a.FS, a.Logger, exercise.WithTaskTimeout(a.taskTimeout))
	if err != nil {
		return nil, err
	}

	// Load guarantees language is registered; the lookup cannot fail here.
	rc := runners.Available[language]

	return ex.Solve(
		ctx,
		a.FS,
		a.Logger,
		rc(runners.ExerciseMeta{Year: ex.Year, Day: ex.Day, Title: ex.Title, Dir: path, Key: language}),
		cb,
		skipTests,
	)
}

// Test loads the exercise at path and runs its test suite.
func (a *App) Test(
	ctx context.Context,
	path, language, customInput string,
	cb func(tasks.Event),
) ([]tasks.Result, error) {
	ex, err := exercise.Load(path, language, customInput, a.FS, a.Logger, exercise.WithTaskTimeout(a.taskTimeout))
	if err != nil {
		return nil, err
	}

	// Load guarantees language is registered; the lookup cannot fail here.
	rc := runners.Available[language]

	return ex.Test(
		ctx,
		a.Logger,
		rc(runners.ExerciseMeta{Year: ex.Year, Day: ex.Day, Title: ex.Title, Dir: path, Key: language}),
		cb,
	)
}

// Benchmark loads the exercise at path and benchmarks all available
// implementations for the given number of iterations.
func (a *App) Benchmark(
	ctx context.Context,
	path, language string,
	cb func(tasks.Event),
	iterations int,
) ([]tasks.Result, error) {
	ex, err := exercise.Load(path, language, "", a.FS, a.Logger, exercise.WithTaskTimeout(a.taskTimeout))
	if err != nil {
		return nil, err
	}

	bmk := exercise.NewBenchmarker(ex)

	return bmk.Benchmark(ctx, a.FS, a.Logger, cb, iterations)
}

// Visualize runs the visualization task for the exercise at path.
func (a *App) Visualize(
	ctx context.Context,
	path, language, outdir string,
	cb func(tasks.Event),
) ([]tasks.Result, error) {
	ex, err := exercise.Load(path, language, "", a.FS, a.Logger, exercise.WithTaskTimeout(a.taskTimeout))
	if err != nil {
		return nil, err
	}

	// Load guarantees language is registered; the lookup cannot fail here.
	rc := runners.Available[language]

	return ex.Visualize(
		ctx,
		a.FS,
		a.Logger,
		rc(runners.ExerciseMeta{Year: ex.Year, Day: ex.Day, Title: ex.Title, Dir: path, Key: language}),
		outdir,
		cb,
	)
}

// Add makes the exercise at url exist in the workspace, writing files in lang. It is the App-level
// entry point for the `download` command: it builds an exercise.Adder from the App's config and the
// per-call arguments, runs it, and returns the scaffold report and the resolved exercise path.
func (a *App) Add(url, lang string, forced *exercise.Overwrites) (exercise.Report, string, error) {
	adder, err := exercise.NewAdder(a.cfg,
		exercise.WithURL(url),
		exercise.WithLanguage(lang),
		exercise.WithOverwrites(forced),
	)
	if err != nil {
		return nil, "", fmt.Errorf("creating adder: %w", err)
	}

	if err = adder.Add(); err != nil {
		return nil, "", fmt.Errorf("adding challenge: %w", err)
	}

	return adder.Report(), adder.FilePath(), nil
}

// AddProblem scaffolds a Project Euler Problem in the workspace: it builds a
// ProblemAdder and runs it. The App-level entry point for `elf add euler`.
func (a *App) AddProblem(number int, lang, title string) (exercise.Report, string, error) {
	adder, err := exercise.NewProblemAdder(a.cfg,
		exercise.WithProblemNumber(number),
		exercise.WithProblemLanguage(lang),
		exercise.WithProblemTitle(title),
	)
	if err != nil {
		return nil, "", fmt.Errorf("creating problem adder: %w", err)
	}

	if err = adder.Add(); err != nil {
		return nil, "", fmt.Errorf("adding problem: %w", err)
	}

	return adder.Report(), adder.FilePath(), nil
}

// Analyze renders run-time graphs from persisted benchmark data under dir, writing the graph to
// out (or a default location when out is empty). It is the App-level entry point for the `analyze`
// command: it builds an analyze.Analyzer and runs it.
func (a *App) Analyze(dir, out string) (string, error) {
	az, err := analyze.NewAnalyzer(a.Logger, analyze.WithDirectory(dir), analyze.WithOutput(out))
	if err != nil {
		return "", fmt.Errorf("creating analyzer: %w", err)
	}

	if err = az.Graph(); err != nil {
		return "", fmt.Errorf("generating graph: %w", err)
	}

	return az.Output, nil
}

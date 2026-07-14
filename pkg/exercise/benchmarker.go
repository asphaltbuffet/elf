package exercise

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/protocol"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

// Benchmarker runs an exercise solution repeatedly to collect timing statistics.
type Benchmarker struct {
	*Exercise

	exerciseBaseDir string
}

// NewBenchmarker creates a Benchmarker from an already-loaded Exercise.
func NewBenchmarker(ex *Exercise) *Benchmarker {
	return &Benchmarker{Exercise: ex}
}

// Benchmark runs each language implementation for the given number of iterations and returns timing results.
func (b *Benchmarker) Benchmark(
	ctx context.Context,
	fs afero.Fs,
	logger *slog.Logger,
	cb func(tasks.Event),
	iterations int,
) ([]tasks.Result, error) {
	normFactor := NormalizationFactor()

	// TODO: add way to specify which implementations to run (e.g. --impls go,py or --impls all)
	impls, err := b.GetImplementations(fs)
	if err != nil {
		return nil, fmt.Errorf("get impls: %w", err)
	}

	input, err := b.readInput(fs)
	if err != nil {
		logger.ErrorContext(ctx, "reading input file", slog.String("path", b.Data.InputFileName), tint.Err(err))
		return nil, err
	}

	b.Data.InputData = input

	benchmarks := make([]*ImplementationData, 0, len(impls))

	results := []tasks.Result{}

	if cb != nil {
		// Benchmark spans multiple implementations, so there is no single runner
		// name; the header carries the exercise identity only.
		cb(tasks.MetaEvent(tasks.Meta{Year: b.Year, Day: b.Day, Number: b.Number, Title: b.Title}))
	}

	for _, impl := range impls {
		logger.DebugContext(ctx, "running benchmark", slog.String("impl", impl))

		implRunner, ok := runners.Available[impl]
		if !ok {
			return nil, errNoRunner(impl)
		}

		b.Language = impl
		runner := implRunner(runners.ExerciseMeta{
			Year:  b.Year,
			Day:   b.Day,
			Title: b.Title,
			Dir:   b.Path,
			Key:   impl,
		})

		// Plan this impl's tasks now that its display name is known, so each
		// progress bar is labelled with the human-readable runner name.
		emitPlannedForImpl(cb, b.declaredParts(), runner.String(), iterations)

		var implData *ImplementationData

		var implResults []tasks.Result
		implResults, implData, err = b.runBenchmark(ctx, logger, runner, cb, iterations)
		if err != nil {
			// A runner that fails to start is skipped, not fatal: the other
			// implementations are still worth benchmarking.
			if errors.Is(err, ErrRunnerStart) {
				logger.WarnContext(ctx, "skipping runner that failed to start",
					slog.String("impl", impl), tint.Err(err))

				continue
			}

			return nil, err
		}

		results = append(results, implResults...)
		benchmarks = append(benchmarks, implData)

		logger.DebugContext(ctx, "benchmarking complete", "lang", impl, "iterations", iterations)
	}

	var benchmarkData []BenchmarkData
	benchmarkData = append(benchmarkData, BenchmarkData{
		Date:            time.Now().UTC(),
		Day:             b.Day,
		Number:          b.Number,
		Title:           b.Title,
		Year:            b.Year,
		Runs:            iterations,
		Implementations: benchmarks,
		Normalization:   normFactor,
	})

	outfile := filepath.Join(b.Path, "benchmark.json")

	// TODO: add flag to append/overwrite/fail?

	jsonData, err := json.MarshalIndent(benchmarkData, "", "  ")
	if err != nil {
		logger.ErrorContext(ctx, "marshalling benchmark data", tint.Err(err))
		return nil, err
	}

	return results, afero.WriteFile(fs, outfile, jsonData, 0o600)
}

// NormalizationFactor returns a CPU-speed calibration factor for comparable cross-machine benchmarks.
func NormalizationFactor() float64 {
	start := time.Now()
	m := map[int]string{}

	for i := 1; i < math.MaxInt16; i++ {
		m[i] = fmt.Sprintf("%2.3f", 1/float64(i))

		if _, ok := m[i/3]; ok {
			delete(m, i/2) //nolint:mnd // hard-coding for now
		}
	}

	elapsed := time.Since(start)

	return elapsed.Seconds()
}

func (b *Benchmarker) runBenchmark(
	ctx context.Context,
	logger *slog.Logger,
	runner runners.Runner,
	cb func(tasks.Event),
	iterations int,
) ([]tasks.Result, *ImplementationData, error) {
	parts := b.declaredParts()

	var (
		benchmarkTasks []*protocol.Task
		metricsResults = make(map[protocol.Part][]float64, len(parts)*iterations)
		results        = make([]tasks.Result, 0, len(parts)*iterations)
	)

	for i := range iterations {
		for _, part := range parts {
			benchmarkTasks = append(benchmarkTasks, &protocol.Task{
				TaskID: tasks.MakeTaskID(tasks.Benchmark, part, i),
				Part:   part,
				Input:  b.Data.InputData,
			})
		}
	}

	defer func() {
		_ = runner.Close(ctx)
		_ = runner.Cleanup()
	}()

	// Prepare/Open failures are runner-setup failures (e.g. a missing
	// interpreter), wrapped in ErrRunnerStart so the caller can skip this
	// runner and benchmark the rest rather than aborting the whole run.
	if err := runner.Prepare(ctx); err != nil {
		logger.ErrorContext(ctx, "prepare runner", tint.Err(err))
		return nil, nil, fmt.Errorf("%w: %w", ErrRunnerStart, err)
	}

	if err := runner.Open(ctx); err != nil {
		logger.ErrorContext(ctx, "open runner", tint.Err(err))
		return nil, nil, fmt.Errorf("%w: %w", ErrRunnerStart, err)
	}

	for _, t := range benchmarkTasks {
		if err := b.runBenchmarkTask(ctx, logger, runner, cb, t, metricsResults, &results); err != nil {
			return nil, nil, err
		}
	}

	stats, err := calculateMetrics(metricsResults)
	if err != nil {
		logger.ErrorContext(ctx, "getting stats from results", tint.Err(err))
		return results, nil, err
	}

	return results,
		&ImplementationData{
			Name:    runner.String(),
			PartOne: stats[protocol.PartOne],
			PartTwo: stats[protocol.PartTwo],
		}, nil
}

// emitPlannedForImpl announces every iteration's tasks for a single
// implementation, tagging each event with the runner's display name so a
// renderer can group them into one progress bar per (runner, Part). It iterates
// the exercise's declared parts, so a single-part Problem never announces a
// phantom Part Two. Emitted per-impl (not as an up-front batch) because the
// display name only exists once the runner is constructed; ADR-0010 events are
// incremental, so a bar appears when its first Planned arrives.
func emitPlannedForImpl(cb func(tasks.Event), parts []protocol.Part, lang string, iterations int) {
	if cb == nil {
		return
	}

	for i := range iterations {
		for _, part := range parts {
			cb(tasks.PlannedEvent(
				tasks.MakeTaskID(tasks.Benchmark, part, i),
				tasks.Benchmark, part, i, lang,
			))
		}
	}
}

// runBenchmarkTask executes one benchmark task and records its result.
// Returns a non-nil error only on fatal runner failure; timeout is handled inline.
func (b *Benchmarker) runBenchmarkTask(
	ctx context.Context,
	logger *slog.Logger,
	runner runners.Runner,
	cb func(tasks.Event),
	t *protocol.Task,
	metricsResults map[protocol.Part][]float64,
	results *[]tasks.Result,
) error {
	taskType, taskPart, taskSubPart := tasks.ParseTaskID(t.TaskID)

	lang := runner.String()

	if cb != nil {
		cb(tasks.StartedEvent(t.TaskID, taskType, taskPart, taskSubPart, lang))
	}

	benchResult, runErr := runWithTimeout(ctx, runner, t, b.taskTimeout)
	if errors.Is(runErr, errTaskTimeout) {
		r := timeoutResult(t.TaskID)
		if cb != nil {
			cb(tasks.FinishedEvent(r, lang))
		}

		*results = append(*results, r)

		if restartErr := restartRunner(ctx, runner); restartErr != nil {
			return restartErr
		}

		return nil
	} else if runErr != nil {
		logger.ErrorContext(ctx, "running benchmark", tint.Err(runErr))
		return runErr
	}

	// A benchmark iteration's measurement is its duration, not its output
	// string (ADR-0011): any non-timeout result is a valid sample and must emit
	// exactly one Finished so the progress bar can reach 100%.
	r := buildResult(benchResult, "")
	if cb != nil {
		cb(tasks.FinishedEvent(r, lang))
	}

	*results = append(*results, r)

	metricsResults[r.Part] = append(metricsResults[r.Part], benchResult.Duration)

	return nil
}

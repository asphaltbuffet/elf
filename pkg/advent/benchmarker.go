package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/montanaflynn/stats"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"

	"github.com/asphaltbuffet/elf/pkg/krampus"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

type Benchmarker struct {
	exerciseBaseDir string
	*Exercise
}

type BenchmarkData struct {
	Date time.Time `json:"run-date,omitempty"`
	// Dir             string                `json:"dir"`
	Title           string                `json:"title"`
	Year            int                   `json:"year,omitempty"`
	Day             int                   `json:"day"`
	Runs            int                   `json:"numRuns"`
	Normalization   float64               `json:"normalization,omitempty"`
	Implementations []*ImplementationData `json:"implementations"`
}

type ImplementationData struct {
	Name    string    `json:"name"`
	PartOne *PartData `json:"part-one"`
	PartTwo *PartData `json:"part-two,omitempty"`
}

type PartData struct {
	Mean float64   `json:"mean"`
	Min  float64   `json:"min"`
	Max  float64   `json:"max"`
	Data []float64 `json:"data,omitempty"`
}

func NewBenchmarker(config krampus.ExerciseConfiguration, options ...func(*Benchmarker)) (*Benchmarker, error) {
	b := &Benchmarker{
		Exercise: &Exercise{
			appFs:    config.GetFs(),
			Language: "go",
			logger:   config.GetLogger().With(slog.String("fn", "benchmark")),
		},
	}

	for _, option := range options {
		option(b)
	}

	switch {
	case b.Path != "":
		if err := b.Exercise.loadInfo(config.GetFs()); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("instantiate exercise: %w", ErrNotFound)
	}

	return b, nil
}

func WithExerciseDir(dir string) func(*Benchmarker) {
	return func(b *Benchmarker) {
		b.Path = dir
	}
}

func (b *Benchmarker) Benchmark(afs afero.Fs, iterations int) ([]tasks.Result, error) {
	logger := b.logger
	e := b.Exercise
	normFactor := calcNormalizationFactor()

	impls, err := e.GetImplementations(afs)
	if err != nil {
		return nil, fmt.Errorf("get impls: %w", err)
	}

	inputFile := filepath.Join(e.Path, e.Data.InputFileName)
	input, err := afero.ReadFile(afs, inputFile)
	if err != nil {
		logger.Error("reading input file", slog.String("path", inputFile), tint.Err(err))
		return nil, err
	}

	e.Data.InputData = string(input)

	benchmarks := make([]*ImplementationData, 0, len(impls))

	results := []tasks.Result{}

	for _, impl := range impls {
		logger.Debug("running benchmark", slog.String("impl", impl))
		implRunner, ok := runners.Available[impl]
		if !ok {
			return nil, fmt.Errorf("no runner available for implementation %s", impl)
		}

		e.Language = impl
		e.runner = implRunner(e.Path)

		var implData *ImplementationData

		var implResults []tasks.Result
		implResults, implData, err = b.runBenchmark(iterations)
		if err != nil {
			return nil, err
		}

		results = append(results, implResults...)
		benchmarks = append(benchmarks, implData)

		logger.Debug("benchmarking complete", "lang", impl, "iterations", iterations)
		fmt.Println()
	}

	var benchmarkData []BenchmarkData
	benchmarkData = append(benchmarkData, BenchmarkData{
		Date:            time.Now().UTC(),
		Day:             e.Day,
		Title:           e.Title,
		Year:            e.Year,
		Runs:            iterations,
		Implementations: benchmarks,
		Normalization:   normFactor,
	})

	outfile := filepath.Join(e.Path, "benchmark.json")

	// TODO: add flag to append/overwrite/fail?

	jsonData, err := json.MarshalIndent(benchmarkData, "", "  ")
	if err != nil {
		logger.Error("marshalling benchmark data", tint.Err(err))
		return nil, err
	}

	return results, afero.WriteFile(afs, outfile, jsonData, 0o600)
}

func calcNormalizationFactor() float64 {
	start := time.Now()
	m := map[int]string{}

	for i := 1; i < math.MaxInt16; i++ {
		m[i] = fmt.Sprintf("%2.3f", 1/float64(i))

		if _, ok := m[i/3]; ok {
			delete(m, i/2)
		}
	}

	elapsed := time.Since(start)

	return elapsed.Seconds()
}

func (b *Benchmarker) runBenchmark(iterations int) ([]tasks.Result, *ImplementationData, error) {
	e := b.Exercise
	logger := b.logger

	var (
		benchmarkTasks []*runners.Task
		metricsResults = make(map[runners.Part][]float64, 2*iterations)
		results        = make([]tasks.Result, 0, 2*iterations)
	)

	// generate all the tasks needed for this benchmark run
	for i := 0; i < iterations; i++ {
		benchmarkTasks = append(
			benchmarkTasks,
			&runners.Task{
				TaskID: tasks.MakeTaskID(tasks.Benchmark, runners.PartOne, i),
				Part:   runners.PartOne,
				Input:  e.Data.InputData,
			},
			&runners.Task{
				TaskID: tasks.MakeTaskID(tasks.Benchmark, runners.PartTwo, i),
				Part:   runners.PartTwo,
				Input:  e.Data.InputData,
			})
	}

	progBar := progressbar.NewOptions(
		len(benchmarkTasks),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription(
			fmt.Sprintf("Benchmarking %q (%s)", e.Title, e.runner),
		),
	)

	if err := e.runner.Start(); err != nil {
		logger.Error("starting runner", tint.Err(err))
		return nil, nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	for _, t := range benchmarkTasks {
		benchResult, err := e.runner.Run(t)
		if err != nil {
			logger.Error("running benchmark", tint.Err(err))
			return nil, nil, err
		}

		if benchResult.Ok && benchResult.Output != "" {
			r := handleTaskResult(os.Stdout, benchResult, "")
			results = append(results, r)

			metricsResults[r.Part] = append(metricsResults[r.Part], benchResult.Duration)
		}

		if err = progBar.Add(1); err != nil {
			logger.Error("updating progress bar", tint.Err(err))
			return nil, nil, err
		}
	}

	stats, err := b.calculateMetrics(metricsResults)
	if err != nil {
		logger.Error("getting stats from results", tint.Err(err))
		return results, nil, err
	}

	return results,
		&ImplementationData{
			Name:    e.runner.String(),
			PartOne: stats[runners.PartOne],
			PartTwo: stats[runners.PartTwo],
		}, nil
}

func (b *Benchmarker) calculateMetrics(results map[runners.Part][]float64) (map[runners.Part]*PartData, error) {
	logger := b.logger
	metrics := make(map[runners.Part]*PartData)

	logger.Debug("calculating stats", "results", results)

	for part, durations := range results {
		data := stats.LoadRawData(durations)

		mean, err := data.Mean()
		if err != nil {
			return nil, err
		}

		max, err := data.Max()
		if err != nil {
			return nil, err
		}

		min, err := data.Min()
		if err != nil {
			return nil, err
		}

		metrics[part] = &PartData{
			Mean: mean,
			Min:  min,
			Max:  max,
			Data: durations,
		}
	}

	return metrics, nil
}

func (b *BenchmarkData) String() string {
	return fmt.Sprintf("BenchmarkData{Date: %s, AOC %d/%02d, Runs: %3d, Normalization: %.6f, Implementations: %s}",
		b.Date.Local().Format(time.DateOnly), b.Year, b.Day, b.Runs, b.Normalization, b.Implementations)
}

func (i *ImplementationData) String() string {
	return fmt.Sprintf("%s{%d PartOne, %d PartTwo}",
		i.Name, len(i.PartOne.Data), len(i.PartTwo.Data))
}

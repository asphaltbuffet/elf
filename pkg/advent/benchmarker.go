package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/montanaflynn/stats"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

type BenchmarkData struct {
	Date time.Time `json:"run-date,omitempty"`
	// Dir             string                `json:"dir"`
	Title           string                `json:"title"`
	Year            int                   `json:"year,omitempty"`
	Day             int                   `json:"day"`
	Runs            int                   `json:"numRuns"`
	Implementations []*ImplementationData `json:"implementations"`
}

type ImplementationData struct {
	Name    string    `json:"name"`
	PartOne *PartData `json:"part-one"`
	PartTwo *PartData `json:"part-two,omitempty"`
}

type PartData struct {
	Mean float64 `json:"mean"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

var benchmarkLog = slog.With(slog.String("fn", "Benchmark"))

func (e *Exercise) Benchmark(iterations int) ([]TaskResult, error) {
	impls, err := e.GetImplementations()
	if err != nil {
		return nil, err
	}

	input, err := os.ReadFile(e.Data.InputFile)
	if err != nil {
		benchmarkLog.Error("reading input file", slog.String("path", e.Data.InputFile), tint.Err(err))
		return nil, err
	}

	e.Data.Input = string(input)

	benchmarks := make([]*ImplementationData, 0, len(impls))

	results := []TaskResult{}

	for _, impl := range impls {
		benchmarkLog.Debug("running benchmark", slog.String("impl", impl))
		implRunner, ok := runners.Available[impl]
		if !ok {
			return nil, fmt.Errorf("no runner available for implementation %s", impl)
		}

		e.Language = impl
		e.runner = implRunner(e.Path)

		var implData *ImplementationData

		var implResults []TaskResult
		implResults, implData, err = e.runBenchmark(iterations)
		if err != nil {
			return nil, err
		}

		results = append(results, implResults...)
		benchmarks = append(benchmarks, implData)

		benchmarkLog.Debug("benchmarking complete", "lang", impl, "iterations", iterations)
	}

	var benchmarkData []BenchmarkData
	benchmarkData = append(benchmarkData, BenchmarkData{
		Date:            time.Now().UTC(),
		Day:             e.Day,
		Title:           e.Title,
		Year:            e.Year,
		Runs:            iterations,
		Implementations: benchmarks,
	})

	outfile := filepath.Join(e.Path, "benchmark.json")

	// TODO: add flag to append/overwrite/fail?

	jsonData, err := json.MarshalIndent(benchmarkData, "", "  ")
	if err != nil {
		benchmarkLog.Error("marshalling benchmark data", tint.Err(err))
		return nil, err
	}

	return results, os.WriteFile(outfile, jsonData, 0o600)
}

func makeBenchmarkID(part runners.Part, subPart int) string {
	return fmt.Sprintf("benchmark.%d.%d", part, subPart)
}

func (e *Exercise) runBenchmark(iterations int) ([]TaskResult, *ImplementationData, error) {
	var (
		tasks          []*runners.Task
		metricsResults = make(map[runners.Part][]float64, 2*iterations)
		results        = make([]TaskResult, 0, 2*iterations)
	)

	// generate all the tasks needed for this benchmark run
	for i := 0; i < iterations; i++ {
		tasks = append(
			tasks,
			&runners.Task{
				TaskID: makeBenchmarkID(runners.PartOne, i),
				Part:   runners.PartOne,
				Input:  e.Data.Input,
			},
			&runners.Task{
				TaskID: makeBenchmarkID(runners.PartTwo, i),
				Part:   runners.PartTwo,
				Input:  e.Data.Input,
			})
	}

	fmt.Printf("Benchmarking (%s)\n", e.runner)
	// progBar := progressbar.NewOptions(
	// 	len(tasks),
	// 	progressbar.OptionSetDescription(
	// 		fmt.Sprintf("Running %s benchmarks", e.runner.String()),
	// 	),
	// )

	if err := e.runner.Start(); err != nil {
		benchmarkLog.Error("starting runner", tint.Err(err))
		return nil, nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	for _, t := range tasks {
		benchResult, err := e.runner.Run(t)
		if err != nil {
			benchmarkLog.Error("running benchmark", tint.Err(err))
			return nil, nil, err
		}

		if !benchResult.Ok || benchResult.Output == "" {
			continue
		}

		r := handleTaskResult(os.Stdout, benchResult, "")
		results = append(results, r)

		metricsResults[runners.Part(r.Part)] = append(metricsResults[runners.Part(r.Part)], benchResult.Duration)

		// progBar.Add(1)
	}

	stats, err := resultsToStats(metricsResults)
	if err != nil {
		benchmarkLog.Error("getting stats from results", tint.Err(err))
		return results, nil, err
	}

	return results,
		&ImplementationData{
			Name:    e.runner.String(),
			PartOne: stats[runners.PartOne],
			PartTwo: stats[runners.PartTwo],
		}, nil
}

func resultsToStats(results map[runners.Part][]float64) (map[runners.Part]*PartData, error) {
	metrics := make(map[runners.Part]*PartData)

	benchmarkLog.Debug("calculating stats", "results", results)

	for part, durations := range results {
		data := stats.LoadRawData(durations)

		mean, _ := data.Mean()
		rmean, _ := stats.Round(mean, 9)
		max, _ := data.Max()
		rmax, _ := stats.Round(max, 9)
		min, _ := data.Min()
		rmin, _ := stats.Round(min, 9)

		metrics[part] = &PartData{
			Mean: rmean,
			Min:  rmin,
			Max:  rmax,
		}
	}

	return metrics, nil
}

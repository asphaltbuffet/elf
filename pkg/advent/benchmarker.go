package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/montanaflynn/stats"
	"github.com/schollz/progressbar/v3"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

type BenchmarkData struct {
	Date            time.Time             `json:"date,omitempty"`
	Dir             string                `json:"dir"`
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
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

var benchmarkLog = slog.With(slog.String("fn", "Benchmark"))

func (e *Exercise) Benchmark(iterations int) error {
	impls, err := e.GetImplementations()
	benchmarks := make([]*ImplementationData, 0, len(impls))

	for _, impl := range impls {
		benchmarkLog.Debug("running benchmark", slog.String("impl", impl))
		r, ok := runners.Available[impl]
		if !ok {
			return fmt.Errorf("no runner available for implementation %s", impl)
		}

		e.Language = impl
		e.runner = r(e.path)

		var d *ImplementationData
		d, err = e.runBenchmark(iterations)
		if err != nil {
			return err
		}

		benchmarks = append(benchmarks, d)
		benchmarkLog.Debug("benchmarking complete", "lang", impl, "iterations", iterations)
	}

	var benchmarkData []BenchmarkData
	benchmarkData = append(benchmarkData, BenchmarkData{
		Date:            time.Now().UTC(),
		Day:             e.Day,
		Dir:             e.Dir(),
		Year:            e.Year,
		Runs:            iterations,
		Implementations: benchmarks,
	})

	outfile := filepath.Join(e.path, "benchmark.json")

	// TODO: add flag to append/overwrite/fail?

	jsonData, err := json.MarshalIndent(benchmarkData, "", "  ")
	if err != nil {
		benchmarkLog.Error("marshalling benchmark data", tint.Err(err))
		return err
	}

	return os.WriteFile(outfile, jsonData, 0o600)
}

func makeBenchmarkID(part runners.Part, subPart int) string {
	if subPart == -1 {
		return fmt.Sprintf("benchmark.part.%d", part)
	}

	return fmt.Sprintf("benchmark.part.%d.%d", part, subPart)
}

func (e *Exercise) runBenchmark(iterations int) (*ImplementationData, error) {
	var (
		tasks   []*runners.Task
		results = make(map[runners.Part][]float64, 2*iterations)
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
				TaskID: makeBenchmarkID(runners.PartOne, i),
				Part:   runners.PartTwo,
				Input:  e.Data.Input,
			})
	}

	progBar := progressbar.NewOptions(
		len(tasks),
		progressbar.OptionSetDescription(
			fmt.Sprintf("Running %s benchmarks", e.runner.String()),
		),
	)

	if err := e.runner.Start(); err != nil {
		benchmarkLog.Error("starting runner", tint.Err(err))
		return nil, err
	}

	defer func() {
		_ = e.runner.Stop()
		_ = e.runner.Cleanup()
	}()

	for _, task := range tasks {
		r, err := e.runner.Run(task)
		if err != nil {
			benchmarkLog.Error("running benchmark", tint.Err(err))
			return nil, err
		}

		p := idToPart(r.TaskID)
		results[p] = append(results[p], r.Duration)

		progBar.Add(1)
	}

	stats, err := resultsToStats(results)
	if err != nil {
		benchmarkLog.Error("getting stats from results", tint.Err(err))
		return nil, err
	}

	return &ImplementationData{
		Name:    e.runner.String(),
		PartOne: stats[runners.PartOne],
		PartTwo: stats[runners.PartTwo],
	}, nil
}

func idToPart(id string) runners.Part {
	parts := strings.Split(id, ".")
	if len(parts) == 3 {
		return runners.PartOne
	}

	return runners.PartTwo
}

func resultsToStats(results map[runners.Part][]float64) (map[runners.Part]*PartData, error) {
	metrics := make(map[runners.Part]*PartData)

	if len(results[runners.PartOne]) == 0 {
		return nil, fmt.Errorf("no results for part one")
	}

	if len(results[runners.PartTwo]) == 0 {
		return nil, fmt.Errorf("no results for part two")
	}

	for part, durations := range results {
		data := stats.LoadRawData(durations)

		mean, _ := data.Mean()
		max, _ := data.Max()
		min, _ := data.Min()

		metrics[part] = &PartData{
			Mean: mean,
			Min:  min,
			Max:  max,
		}
	}

	return metrics, nil
}

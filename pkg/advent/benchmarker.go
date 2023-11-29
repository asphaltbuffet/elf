package advent

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"

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

	benchmarkData := &BenchmarkData{
		Date:            time.Now().UTC(),
		Day:             e.Day,
		Dir:             e.Dir(),
		Year:            e.Year,
		Runs:            iterations,
		Implementations: benchmarks,
	}

	outfile := filepath.Join(e.path, "benchmark.json")

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
	panic("not implemented")
}

package exercise

import (
	"errors"
	"fmt"
	"time"

	"github.com/montanaflynn/stats"

	"github.com/asphaltbuffet/elf/pkg/protocol"
)

// BenchmarkData holds the aggregated timing results for all implementations of a single exercise.
type BenchmarkData struct {
	Date time.Time `json:"run-date"`
	// Dir             string                `json:"dir"`
	Title           string                `json:"title"`
	Year            int                   `json:"year,omitempty"`
	Day             int                   `json:"day"`
	Runs            int                   `json:"numRuns"`
	Normalization   float64               `json:"normalization,omitempty"`
	Implementations []*ImplementationData `json:"implementations"`
}

// ImplementationData holds per-language timing data for both parts of an exercise.
type ImplementationData struct {
	Name    string    `json:"name"`
	PartOne *PartData `json:"part-one"`
	PartTwo *PartData `json:"part-two,omitempty"`
}

// PartData holds statistical timing metrics (mean, min, max, raw samples) for one exercise part.
type PartData struct {
	Mean float64   `json:"mean"`
	Min  float64   `json:"min"`
	Max  float64   `json:"max"`
	Data []float64 `json:"data,omitempty"`
}

// ErrRunnerStart is returned when a language runner fails to initialize.
var ErrRunnerStart = errors.New("runner start error")

func (b *BenchmarkData) String() string {
	return fmt.Sprintf("BenchmarkData{Date: %s, AOC %d/%02d, Runs: %3d, Normalization: %.6f, Implementations: %s}",
		b.Date.Local().Format(time.DateOnly), b.Year, b.Day, b.Runs, b.Normalization, b.Implementations)
}

func (i *ImplementationData) String() string {
	return fmt.Sprintf("%s{%d PartOne, %d PartTwo}",
		i.Name, len(i.PartOne.Data), len(i.PartTwo.Data))
}

func calculateMetrics(results map[protocol.Part][]float64) (map[protocol.Part]*PartData, error) {
	metrics := make(map[protocol.Part]*PartData)

	for part, durations := range results {
		data := stats.LoadRawData(durations)

		mean, err := data.Mean()
		if err != nil {
			return nil, err
		}

		maxVal, err := data.Max()
		if err != nil {
			return nil, err
		}

		minVal, err := data.Min()
		if err != nil {
			return nil, err
		}

		metrics[part] = &PartData{
			Mean: mean,
			Min:  minVal,
			Max:  maxVal,
			Data: durations,
		}
	}

	return metrics, nil
}

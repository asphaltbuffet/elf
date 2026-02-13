package analyze

import (
	"io"
	"log/slog"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

// testLogger returns a logger that discards output.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// makeBenchmarkData builds valid BenchmarkData for the given year and days.
// Each day gets both "Golang" and "Python" implementations with both parts populated.
func makeBenchmarkData(year int, days ...int) []*advent.BenchmarkData {
	out := make([]*advent.BenchmarkData, 0, len(days))

	for _, d := range days {
		out = append(out, &advent.BenchmarkData{
			Year: year,
			Day:  d,
			Implementations: []*advent.ImplementationData{
				{
					Name: "Golang",
					PartOne: &advent.PartData{
						Mean: 0.001 * float64(d),
						Min:  0.0005 * float64(d),
						Max:  0.002 * float64(d),
						Data: []float64{0.001, 0.0012, 0.0008},
					},
					PartTwo: &advent.PartData{
						Mean: 0.002 * float64(d),
						Min:  0.001 * float64(d),
						Max:  0.004 * float64(d),
						Data: []float64{0.002, 0.0025, 0.0015},
					},
				},
				{
					Name: "Python",
					PartOne: &advent.PartData{
						Mean: 0.05 * float64(d),
						Min:  0.04 * float64(d),
						Max:  0.07 * float64(d),
						Data: []float64{0.05, 0.055, 0.045},
					},
					PartTwo: &advent.PartData{
						Mean: 0.1 * float64(d),
						Min:  0.08 * float64(d),
						Max:  0.15 * float64(d),
						Data: []float64{0.1, 0.12, 0.08},
					},
				},
			},
		})
	}

	return out
}

// makeBenchmarkDataNilPartTwo is like makeBenchmarkData but sets PartTwo = nil
// on the first implementation of each day (for branch coverage).
func makeBenchmarkDataNilPartTwo(year int, days ...int) []*advent.BenchmarkData {
	data := makeBenchmarkData(year, days...)
	for _, bd := range data {
		bd.Implementations[0].PartTwo = nil
	}

	return data
}

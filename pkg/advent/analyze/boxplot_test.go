package analyze

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

func Test_benchmarkToPlotterValues(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := benchmarkToPlotterValues(nil)
		assert.Empty(t, got)
	})

	t.Run("empty input", func(t *testing.T) {
		got := benchmarkToPlotterValues([]*advent.BenchmarkData{})
		assert.Empty(t, got)
	})

	t.Run("single day", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1)
		got := benchmarkToPlotterValues(data)

		require.Contains(t, got, "Golang")
		require.Contains(t, got["Golang"], 1)
		require.Contains(t, got["Golang"][1], 0) // part one
		require.Contains(t, got["Golang"][1], 1) // part two
		assert.Len(t, got["Golang"][1][0], 3, "part one should have 3 data points")
		assert.Len(t, got["Golang"][1][1], 3, "part two should have 3 data points")
	})

	t.Run("multiple days", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1, 5)
		got := benchmarkToPlotterValues(data)

		require.Contains(t, got["Golang"], 1)
		require.Contains(t, got["Golang"], 5)
	})

	t.Run("nil PartTwo", func(t *testing.T) {
		data := makeBenchmarkDataNilPartTwo(2024, 1)
		got := benchmarkToPlotterValues(data)

		// Golang has nil PartTwo → part two values should be empty
		assert.Empty(t, got["Golang"][1][1])
		// Python still has data in part two
		assert.Len(t, got["Python"][1][1], 3)
	})
}

func Test_addDayPartsToPlot(t *testing.T) {
	t.Run("empty dayMap", func(t *testing.T) {
		p := plot.New()
		err := addDayPartsToPlot(p, map[int]map[int]plotter.Values{})
		require.NoError(t, err)
	})

	t.Run("single day both parts", func(t *testing.T) {
		p := plot.New()
		dayMap := map[int]map[int]plotter.Values{
			1: {
				0: {0.001, 0.002, 0.003},
				1: {0.01, 0.02, 0.03},
			},
		}

		err := addDayPartsToPlot(p, dayMap)
		require.NoError(t, err)
	})

	t.Run("multiple days", func(t *testing.T) {
		p := plot.New()
		dayMap := map[int]map[int]plotter.Values{
			1: {
				0: {0.001, 0.002},
				1: {0.01, 0.02},
			},
			2: {
				0: {0.003, 0.004},
				1: {0.03, 0.04},
			},
		}

		err := addDayPartsToPlot(p, dayMap)
		require.NoError(t, err)
	})
}

func Test_makePlotForEachImplementation(t *testing.T) {
	data := makeBenchmarkData(2024, 1)
	pValues := benchmarkToPlotterValues(data)

	plots, err := makePlotForEachImplementation(2024, pValues)
	require.NoError(t, err)

	assert.Contains(t, plots, "golang-benchmarks.png")
	assert.Contains(t, plots, "python-benchmarks.png")
	assert.Contains(t, plots["golang-benchmarks.png"].Title.Text, "Golang")
	assert.Contains(t, plots["python-benchmarks.png"].Title.Text, "Python")
}

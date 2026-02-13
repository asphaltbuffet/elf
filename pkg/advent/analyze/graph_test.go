package analyze

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

func Test_benchmarkToPlotterXYs(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := benchmarkToPlotterXYs(nil)
		assert.Empty(t, got)
	})

	t.Run("empty input", func(t *testing.T) {
		got := benchmarkToPlotterXYs([]*advent.BenchmarkData{})
		assert.Empty(t, got)
	})

	t.Run("single day with two implementations", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1)
		got := benchmarkToPlotterXYs(data)

		require.Contains(t, got, "Golang")
		require.Contains(t, got, "Python")
		assert.Len(t, got["Golang"], 2, "should have 2 parts")
		assert.Len(t, got["Golang"][0], 1, "part one should have 1 XY point")
		assert.Len(t, got["Golang"][1], 1, "part two should have 1 XY point")
		assert.InDelta(t, 1.0, got["Golang"][0][0].X, 0.001)
		assert.InDelta(t, 0.001, got["Golang"][0][0].Y, 0.001)
	})

	t.Run("multiple days", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1, 2, 3)
		got := benchmarkToPlotterXYs(data)

		assert.Len(t, got["Golang"][0], 3)
		assert.Len(t, got["Python"][0], 3)
	})

	t.Run("nil PartTwo", func(t *testing.T) {
		data := makeBenchmarkDataNilPartTwo(2024, 1)
		got := benchmarkToPlotterXYs(data)

		// Golang has nil PartTwo → part two slice should be empty
		assert.Empty(t, got["Golang"][1])
		// Python still has PartTwo
		assert.Len(t, got["Python"][1], 1)
	})
}

func Test_NewBenchmarkPlots(t *testing.T) {
	plots, err := NewBenchmarkPlots(2024)
	require.NoError(t, err)

	require.Len(t, plots, 1, "should have 1 row")
	require.Len(t, plots[0], 2, "should have 2 columns")

	assert.Contains(t, plots[0][0].Title.Text, "2024")
	assert.Contains(t, plots[0][0].Title.Text, "Part One")
	assert.Contains(t, plots[0][1].Title.Text, "Part Two")

	// Verify Y scale is LogScale
	assert.IsType(t, plot.LogScale{}, plots[0][0].Y.Scale)
	assert.IsType(t, plot.LogScale{}, plots[0][1].Y.Scale)
}

func Test_generateLineGraph(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		err := generateLineGraph([]*advent.BenchmarkData{}, "out.png")
		require.Error(t, err)
		assert.ErrorContains(t, err, "no benchmark data")
	})

	t.Run("valid data writes file", func(t *testing.T) {
		dir := t.TempDir()
		outfile := filepath.Join(dir, "test-graph.png")
		data := makeBenchmarkData(2024, 1, 2, 3)

		err := generateLineGraph(data, outfile)
		require.NoError(t, err)

		info, err := os.Stat(outfile)
		require.NoError(t, err)
		assert.Positive(t, info.Size())
	})
}

func Test_Graph(t *testing.T) {
	tests := []struct {
		name      string
		data      []*advent.BenchmarkData
		assertion require.ErrorAssertionFunc
		errMsg    string
	}{
		{
			name:      "line graph",
			data:      makeBenchmarkData(2024, 1),
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)

			a := &Analyzer{
				Data:   tt.data,
				Output: filepath.Join(dir, "test-output.png"),
				logger: testLogger(),
			}

			err := a.Graph()
			tt.assertion(t, err)

			if tt.errMsg != "" && err != nil {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}

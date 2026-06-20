package analyze

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func Test_styleBox(t *testing.T) {
	bp, err := plotter.NewBoxPlot(vg.Points(20), 0, plotter.Values{0.001, 0.002, 0.003})
	require.NoError(t, err)

	styleBox(bp, "Golang")

	lc := colorForLang("Golang")
	assert.Equal(t, lc, bp.BoxStyle.Color, "box outline = language color")
	assert.Equal(t, lc, bp.MedianStyle.Color, "median = language color")
	assert.Equal(t, lc, bp.WhiskerStyle.Color, "whisker = language color")
	assert.Equal(t, lc, bp.GlyphStyle.Color, "outlier glyph = language color")
	assert.Equal(t, lighten(lc), bp.FillColor, "fill = lightened language color")

	fill, ok := bp.FillColor.(color.NRGBA)
	require.True(t, ok, "fill should be NRGBA")
	assert.Less(t, fill.A, uint8(255), "fill is translucent")
}

func Test_buildConsistencyFacets(t *testing.T) {
	data := makeBenchmarkData(2015, 1) // Golang + Python, both parts, 3 samples each

	grid, err := buildConsistencyFacets(data)
	require.NoError(t, err)

	require.Len(t, grid, 2, "two rows = two parts")
	require.Len(t, grid[0], 2, "two columns = two languages (Golang, Python)")

	// Part One, first language cell exists and is titled with the language name.
	require.NotNil(t, grid[0][0])
	assert.Contains(t, grid[0][0].Title.Text, "median", "cell subtitle names the absolute median")

	// Left column carries the Y label; it names percent-of-median.
	assert.Contains(t, grid[0][0].Y.Label.Text, "%", "left column labels the percent axis")
}

func Test_buildConsistencyFacets_noData(t *testing.T) {
	_, err := buildConsistencyFacets(nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "no benchmark data")
}

func Test_buildConsistencyFacets_missingPartTwo(t *testing.T) {
	data := makeBenchmarkDataNilPartTwo(2015, 1) // first impl has nil PartTwo

	grid, err := buildConsistencyFacets(data)
	require.NoError(t, err)

	// languages are sorted; "Golang" < "Python". makeBenchmarkDataNilPartTwo nils
	// PartTwo on the first implementation of each day (Golang). So Part Two (row 1),
	// Golang column (0) must be a blank (nil) cell, grid still aligned.
	require.Len(t, grid, 2)
	require.Len(t, grid[1], 2)
	assert.Nil(t, grid[1][0], "Golang Part Two is missing → blank cell")
	assert.NotNil(t, grid[1][1], "Python Part Two present")
}

func Test_generateBoxPlot(t *testing.T) {
	t.Run("writes a png for a single exercise", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "run-times.png")
		data := makeBenchmarkData(2015, 1) // one day, Golang + Python, both parts

		err := generateBoxPlot(data, out)
		require.NoError(t, err)

		info, statErr := os.Stat(out)
		require.NoError(t, statErr)
		assert.Positive(t, info.Size())
	})

	t.Run("handles a missing part two", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "run-times.png")
		data := makeBenchmarkDataNilPartTwo(2015, 1)

		err := generateBoxPlot(data, out)
		require.NoError(t, err)
	})

	t.Run("errors with no data", func(t *testing.T) {
		err := generateBoxPlot(nil, filepath.Join(t.TempDir(), "x.png"))
		require.Error(t, err)
		assert.ErrorContains(t, err, "no benchmark data")
	})
}

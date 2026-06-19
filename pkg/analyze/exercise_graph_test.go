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

func Test_buildBoxPlot_isRelative(t *testing.T) {
	data := makeBenchmarkData(2015, 1) // Golang + Python, both parts; Golang faster

	p, err := buildBoxPlot(data)
	require.NoError(t, err)

	// Relative axis: marker is RelativeLogTicks and label names "Relative".
	assert.IsType(t, RelativeLogTicks{}, p.Y.Tick.Marker, "Y axis must use relative ticks")
	assert.Contains(t, p.Y.Label.Text, "Relative", "Y label names the relative view")
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

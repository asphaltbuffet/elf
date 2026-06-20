package analyze

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/vg"
)

func Test_redlineColor(t *testing.T) {
	// Reference line uses Okabe-Ito vermillion to stay within the CVD-safe
	// palette rather than pure red.
	got := redlineColor()
	r, g, b, a := got.RGBA()
	expected := color.RGBA{R: 213, G: 94, B: 0, A: 255}
	actual := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
	assert.Equal(t, expected, actual)
}

func Test_saveGridPNG(t *testing.T) {
	p := plot.New()
	p.Title.Text = "cell"
	// 1x2 grid with one nil (blank) cell.
	grid := [][]*plot.Plot{{p, nil}}

	out := filepath.Join(t.TempDir(), "grid.png")
	err := saveGridPNG(grid, out, font.Length(6*vg.Inch), font.Length(3*vg.Inch), 100)
	require.NoError(t, err)

	info, statErr := os.Stat(out)
	require.NoError(t, statErr)
	assert.Positive(t, info.Size())
}

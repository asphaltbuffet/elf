package components

import (
	"image/color"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_channelRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		colors    []rgbColor
		wantRange int
		wantCh    int
	}{
		{name: "single", colors: []rgbColor{{r: 100, g: 100, b: 100}}, wantRange: 0, wantCh: 0},
		{name: "red", colors: []rgbColor{{r: 0, g: 50, b: 50}, {r: 200, g: 50, b: 50}}, wantRange: 200, wantCh: 0},
		{name: "green", colors: []rgbColor{{r: 50, g: 0, b: 50}, {r: 50, g: 255, b: 50}}, wantRange: 255, wantCh: 1},
		{name: "blue", colors: []rgbColor{{r: 50, g: 50, b: 0}, {r: 50, g: 50, b: 200}}, wantRange: 200, wantCh: 2},
		{name: "rg eq", colors: []rgbColor{{r: 0, g: 0, b: 0}, {r: 100, g: 100, b: 50}}, wantRange: 100, wantCh: 0},
		{name: "gb eq", colors: []rgbColor{{r: 0, g: 0, b: 0}, {r: 50, g: 100, b: 100}}, wantRange: 100, wantCh: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotRange, gotCh := channelRange(tt.colors)
			assert.Equal(t, tt.wantRange, gotRange)
			assert.Equal(t, tt.wantCh, gotCh)
		})
	}
}

func Test_sortByChannel(t *testing.T) {
	t.Parallel()
	base := []rgbColor{{r: 30, g: 10, b: 50}, {r: 10, g: 30, b: 20}, {r: 20, g: 20, b: 10}}
	t.Run("red", func(t *testing.T) {
		t.Parallel()
		c := make([]rgbColor, len(base))
		copy(c, base)
		sortByChannel(c, 0)
		assert.True(t, sort.SliceIsSorted(c, func(i, j int) bool { return c[i].r < c[j].r }))
	})
	t.Run("green", func(t *testing.T) {
		t.Parallel()
		c := make([]rgbColor, len(base))
		copy(c, base)
		sortByChannel(c, 1)
		assert.True(t, sort.SliceIsSorted(c, func(i, j int) bool { return c[i].g < c[j].g }))
	})
	t.Run("blue", func(t *testing.T) {
		t.Parallel()
		c := make([]rgbColor, len(base))
		copy(c, base)
		sortByChannel(c, 2)
		assert.True(t, sort.SliceIsSorted(c, func(i, j int) bool { return c[i].b < c[j].b }))
	})
}

func Test_averageColor(t *testing.T) {
	t.Parallel()
	assert.Equal(t, color.RGBA{A: 255}, averageColor(nil))
	assert.Equal(t, color.RGBA{R: 100, G: 150, B: 200, A: 255}, averageColor([]rgbColor{{r: 100, g: 150, b: 200}}))
	assert.Equal(t, color.RGBA{R: 100, G: 50, B: 25, A: 255}, averageColor([]rgbColor{{}, {r: 200, g: 100, b: 50}}))
}

func Test_medianCut(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		p := medianCut(nil, 4)
		require.Len(t, p, 1)
		assert.Equal(t, color.Black, p[0])
	})
	t.Run("single", func(t *testing.T) {
		t.Parallel()
		require.Len(t, medianCut([]rgbColor{{r: 128, g: 64, b: 32}}, 4), 1)
	})
	t.Run("two", func(t *testing.T) {
		t.Parallel()
		require.Len(t, medianCut([]rgbColor{{}, {r: 255, g: 255, b: 255}}, 2), 2)
	})
	t.Run("many", func(t *testing.T) {
		t.Parallel()
		colors := make([]rgbColor, 100)
		for i := range colors {
			colors[i] = rgbColor{r: uint8(i * 2), g: uint8(i), b: uint8(255 - i)}
		}
		p := medianCut(colors, 4)
		assert.LessOrEqual(t, len(p), 4)
	})
	t.Run("identical", func(t *testing.T) {
		t.Parallel()
		p := medianCut([]rgbColor{{r: 50, g: 50, b: 50}, {r: 50, g: 50, b: 50}}, 4)
		require.Len(t, p, 1)
		assert.Equal(t, color.RGBA{R: 50, G: 50, B: 50, A: 255}, p[0])
	})
	t.Run("one bucket", func(t *testing.T) {
		t.Parallel()
		p := medianCut([]rgbColor{{}, {r: 100, g: 100, b: 100}}, 1)
		require.Len(t, p, 1)
		assert.Equal(t, color.RGBA{R: 50, G: 50, B: 50, A: 255}, p[0])
	})
}

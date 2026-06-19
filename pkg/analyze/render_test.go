package analyze

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
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

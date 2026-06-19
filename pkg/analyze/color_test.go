package analyze

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_colorForLang(t *testing.T) {
	t.Run("is deterministic for the same key", func(t *testing.T) {
		first := colorForLang("Golang")
		second := colorForLang("Golang")
		assert.Equal(t, first, second)
	})

	t.Run("differs for different keys in the palette range", func(t *testing.T) {
		// Not a guarantee for all pairs (collisions are allowed), but these
		// two well-known languages must not collide.
		assert.NotEqual(t, colorForLang("Golang"), colorForLang("Python"))
	})

	t.Run("never returns the zero/transparent color", func(t *testing.T) {
		for _, name := range []string{"Golang", "Python", "Rust", "Ruby", "Zig", ""} {
			c := colorForLang(name)
			r, g, b, a := c.RGBA()
			assert.NotEqual(t, uint32(0), a, "alpha must be opaque for %q", name)
			assert.NotEqual(t, color.RGBA{}, c, "must not be zero color for %q", name)
			_, _, _ = r, g, b
		}
	})
}

func Test_lighten(t *testing.T) {
	t.Run("preserves RGB and reduces alpha", func(t *testing.T) {
		// vermillion from the palette, opaque
		in := color.RGBA{R: 213, G: 94, B: 0, A: 255}
		got := lighten(in)

		assert.Equal(t, uint8(213), got.R, "R preserved")
		assert.Equal(t, uint8(94), got.G, "G preserved")
		assert.Equal(t, uint8(0), got.B, "B preserved")
		assert.Less(t, got.A, uint8(255), "alpha reduced for translucency")
		assert.Positive(t, got.A, "alpha not fully transparent")
	})

	t.Run("works on a colorForLang result", func(t *testing.T) {
		c := colorForLang("Golang")
		r, g, b, _ := c.RGBA()
		got := lighten(c)
		assert.Equal(t, uint8(r>>8), got.R)
		assert.Equal(t, uint8(g>>8), got.G)
		assert.Equal(t, uint8(b>>8), got.B)
		assert.Less(t, got.A, uint8(255))
	})
}

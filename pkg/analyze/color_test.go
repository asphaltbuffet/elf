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

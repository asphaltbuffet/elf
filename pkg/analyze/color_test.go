package analyze

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_colorForLang(t *testing.T) {
	t.Run("is deterministic for the same key", func(t *testing.T) {
		first := colorForLang("Go")
		second := colorForLang("Go")
		assert.Equal(t, first, second)
	})

	t.Run("known languages never collide", func(t *testing.T) {
		names := make([]string, 0, len(knownLangIndex))
		for name := range knownLangIndex {
			names = append(names, name)
		}
		for i := range names {
			for j := i + 1; j < len(names); j++ {
				assert.NotEqual(t, colorForLang(names[i]), colorForLang(names[j]),
					"collision: %q and %q", names[i], names[j])
			}
		}
	})

	t.Run("known languages return a pinned palette entry", func(t *testing.T) {
		assert.Equal(t, langPalette[knownLangIndex["Bash"]], colorForLang("Bash"))
		assert.Equal(t, langPalette[knownLangIndex["Rust"]], colorForLang("Rust"))
	})

	t.Run("unknown language falls back to hash and is deterministic", func(t *testing.T) {
		first := colorForLang("UnknownLang")
		assert.NotEqual(t, color.RGBA{}, first)
		assert.Equal(t, first, colorForLang("UnknownLang"))
	})

	t.Run("never returns the zero/transparent color", func(t *testing.T) {
		for _, name := range []string{"Bash", "Go", "Python", "Rust", "Ruby", "Zig", ""} {
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
		in := langPalette[0] // red, opaque
		got := lighten(in)

		assert.Equal(t, in.R, got.R, "R preserved")
		assert.Equal(t, in.G, got.G, "G preserved")
		assert.Equal(t, in.B, got.B, "B preserved")
		assert.Less(t, got.A, uint8(255), "alpha reduced for translucency")
		assert.Positive(t, got.A, "alpha not fully transparent")
	})

	t.Run("works on a colorForLang result", func(t *testing.T) {
		c := colorForLang("Go")
		r, g, b, _ := c.RGBA()
		got := lighten(c)
		assert.Equal(t, uint8(r>>8), got.R)
		assert.Equal(t, uint8(g>>8), got.G)
		assert.Equal(t, uint8(b>>8), got.B)
		assert.Less(t, got.A, uint8(255))
	})
}

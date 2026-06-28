package analyze

import (
	"hash/fnv"
	"image/color"
)

// langPalette is a fixed, opaque 12-color qualitative palette. Known languages
// are pinned to specific indices via knownLangIndex; unknown languages fall back
// to hashing into this slice (collisions accepted for unrecognised names).
//
// Colors are drawn from ColorBrewer's 12-color qualitative Set, chosen for
// perceptual distinctness across a wide gamut. Adjacent entries are visually
// separable; the set is broadly accessible but not strictly CVD-safe at all
// palette positions.
// Reference: https://colorbrewer2.org/#type=qualitative
//
//nolint:mnd // palette color definitions
var langPalette = []color.RGBA{
	{R: 228, G: 26, B: 28, A: 255},   // 0  red
	{R: 55, G: 126, B: 184, A: 255},  // 1  blue
	{R: 77, G: 175, B: 74, A: 255},   // 2  green
	{R: 152, G: 78, B: 163, A: 255},  // 3  purple
	{R: 255, G: 127, B: 0, A: 255},   // 4  orange
	{R: 255, G: 255, B: 51, A: 255},  // 5  yellow
	{R: 166, G: 86, B: 40, A: 255},   // 6  brown
	{R: 247, G: 129, B: 191, A: 255}, // 7  pink
	{R: 153, G: 153, B: 153, A: 255}, // 8  grey
	{R: 0, G: 190, B: 190, A: 255},   // 9  cyan
	{R: 190, G: 0, B: 190, A: 255},   // 10 magenta
	{R: 0, G: 128, B: 0, A: 255},     // 11 dark green
}

// knownLangIndex pins display names (RunnerDescriptor.Name, written into
// benchmark.json as impl.Name) to specific palette indices, guaranteeing no
// two known languages share a color. Add a new entry here whenever a new
// Runner Descriptor is shipped; choose an index not already in use.
//
// Languages listed but not yet active are commented out — uncomment when the
// corresponding Runner Descriptor is added.
var knownLangIndex = map[string]int{
	"Bash":    0,
	"Go":      1,
	"Python":  2, //nolint:mnd // palette index
	"Rust":    3, //nolint:mnd // palette index
	"Fortran": 4, //nolint:mnd // palette index
	// "Perl":   5,
	// "Lua":    6,
	// "Kotlin": 7,
}

// colorForLang returns a deterministic, opaque color for a language display
// name. Known languages (those in knownLangIndex) always map to a unique
// palette entry. Unknown names fall back to FNV-32a hash mod palette length —
// same color across all graphs for that name, but uniqueness not guaranteed.
func colorForLang(name string) color.Color {
	if idx, ok := knownLangIndex[name]; ok {
		return langPalette[idx]
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(name))

	// len(langPalette) is a small fixed constant, so the conversion cannot overflow.
	return langPalette[h.Sum32()%uint32(len(langPalette))] //nolint:gosec // bounded by fixed palette length
}

// boxFillAlpha is the alpha applied to box-plot fills so the language color
// reads as a light tint over the white background while the box outline,
// median, and whiskers stay crisp. Non-premultiplied ([color.NRGBA]).
const boxFillAlpha uint8 = 90 // visual alpha constant

// lighten returns c as a translucent [color.NRGBA], preserving its R/G/B so the
// hue (and thus the language it encodes) is unchanged; only alpha is reduced.
// Using NRGBA (non-premultiplied) keeps the visible R/G/B equal to the palette
// color when composited over white.
func lighten(c color.Color) color.NRGBA {
	r, g, b, _ := c.RGBA() // 16-bit, premultiplied; palette colors are opaque
	return color.NRGBA{
		R: uint8(r >> 8), //nolint:mnd,gosec // 16-bit channel shifted to 8-bit
		G: uint8(g >> 8), //nolint:mnd,gosec // 16-bit channel shifted to 8-bit
		B: uint8(b >> 8), //nolint:mnd,gosec // 16-bit channel shifted to 8-bit
		A: boxFillAlpha,
	}
}

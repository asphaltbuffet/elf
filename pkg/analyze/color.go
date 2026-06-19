package analyze

import (
	"hash/fnv"
	"image/color"
)

// langPalette is a fixed, opaque palette. Colors are assigned to languages by
// hashing the language key into this slice, so a given language is the same
// color in every graph. Collisions past len(langPalette) are accepted.
//
// These are the chromatic colors of the Okabe-Ito palette, chosen because they
// remain distinguishable under all common forms of color-vision deficiency
// (deuteranopia, protanopia, tritanopia). Okabe-Ito's eighth color (black) is
// omitted: it would be indistinguishable from the graph axes and reference line.
// Reference: https://jfly.uni-koeln.de/color/
//
//nolint:mnd // palette color definitions
var langPalette = []color.RGBA{
	{R: 230, G: 159, B: 0, A: 255},   // orange
	{R: 86, G: 180, B: 233, A: 255},  // sky blue
	{R: 0, G: 158, B: 115, A: 255},   // bluish green
	{R: 240, G: 228, B: 66, A: 255},  // yellow
	{R: 0, G: 114, B: 178, A: 255},   // blue
	{R: 213, G: 94, B: 0, A: 255},    // vermillion
	{R: 204, G: 121, B: 167, A: 255}, // reddish purple
}

// colorForLang returns a deterministic, opaque color for a language key.
// The same key always maps to the same color, across calls and processes.
func colorForLang(name string) color.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))

	// len(langPalette) is a small fixed constant, so the conversion cannot overflow.
	return langPalette[h.Sum32()%uint32(len(langPalette))] //nolint:gosec // bounded by fixed palette length
}

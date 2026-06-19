package analyze

import (
	"hash/fnv"
	"image/color"
)

// langPalette is a fixed, opaque palette. Colors are assigned to languages by
// hashing the language key into this slice, so a given language is the same
// color in every graph. Collisions past len(langPalette) are accepted.
//
//nolint:mnd // palette color definitions
var langPalette = []color.RGBA{
	{R: 0, G: 173, B: 216, A: 255},
	{R: 55, G: 118, B: 171, A: 255},
	{R: 214, G: 39, B: 40, A: 255},
	{R: 44, G: 160, B: 44, A: 255},
	{R: 148, G: 103, B: 189, A: 255},
	{R: 255, G: 127, B: 14, A: 255},
	{R: 227, G: 119, B: 194, A: 255},
	{R: 188, G: 189, B: 34, A: 255},
}

// colorForLang returns a deterministic, opaque color for a language key.
// The same key always maps to the same color, across calls and processes.
func colorForLang(name string) color.Color {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))

	// len(langPalette) is a small fixed constant, so the conversion cannot overflow.
	return langPalette[h.Sum32()%uint32(len(langPalette))] //nolint:gosec // bounded by fixed palette length
}

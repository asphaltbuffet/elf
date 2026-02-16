package components

import (
	"image/color"
	"sort"
)

const minSplitSize = 2 // minimum bucket size for median-cut split

type rgbColor struct {
	r, g, b uint8
}

// medianCut reduces a set of colors to numColors using the median-cut algorithm.
func medianCut(colors []rgbColor, numColors int) color.Palette {
	if len(colors) == 0 {
		return color.Palette{color.Black}
	}

	buckets := [][]rgbColor{colors}

	for len(buckets) < numColors {
		// Find the bucket with the widest channel range.
		widest := 0
		widestRange := 0
		widestCh := 0 // 0=R, 1=G, 2=B

		for i, bucket := range buckets {
			if len(bucket) < minSplitSize {
				continue
			}

			rng, ch := channelRange(bucket)
			if rng > widestRange {
				widest = i
				widestRange = rng
				widestCh = ch
			}
		}

		if widestRange == 0 {
			break
		}

		// Sort by the widest channel and split at median.
		bucket := buckets[widest]
		sortByChannel(bucket, widestCh)
		mid := len(bucket) / minSplitSize

		buckets[widest] = bucket[:mid]
		buckets = append(buckets, bucket[mid:])
	}

	// Average each bucket to produce the palette.
	palette := make(color.Palette, len(buckets))
	for i, bucket := range buckets {
		palette[i] = averageColor(bucket)
	}

	return palette
}

// channelRange returns the largest channel range and which channel (0=R,1=G,2=B).
func channelRange(colors []rgbColor) (int, int) {
	var rMin, gMin, bMin uint8 = 255, 255, 255
	var rMax, gMax, bMax uint8

	for _, c := range colors {
		if c.r < rMin {
			rMin = c.r
		}
		if c.r > rMax {
			rMax = c.r
		}
		if c.g < gMin {
			gMin = c.g
		}
		if c.g > gMax {
			gMax = c.g
		}
		if c.b < bMin {
			bMin = c.b
		}
		if c.b > bMax {
			bMax = c.b
		}
	}

	rRange := int(rMax) - int(rMin)
	gRange := int(gMax) - int(gMin)
	bRange := int(bMax) - int(bMin)

	if rRange >= gRange && rRange >= bRange {
		return rRange, 0
	}

	if gRange >= bRange {
		return gRange, 1
	}

	return bRange, 2 //nolint:mnd // channel index for blue
}

func sortByChannel(colors []rgbColor, ch int) {
	sort.Slice(colors, func(i, j int) bool {
		switch ch {
		case 0:
			return colors[i].r < colors[j].r
		case 1:
			return colors[i].g < colors[j].g
		default:
			return colors[i].b < colors[j].b
		}
	})
}

func averageColor(colors []rgbColor) color.RGBA {
	if len(colors) == 0 {
		return color.RGBA{A: 255} //nolint:mnd // fully opaque
	}

	var rSum, gSum, bSum uint64

	for _, c := range colors {
		rSum += uint64(c.r)
		gSum += uint64(c.g)
		bSum += uint64(c.b)
	}

	n := uint64(len(colors))

	return color.RGBA{
		R: uint8(rSum / n), //nolint:gosec // average of uint8 values; result ≤ 255
		G: uint8(gSum / n), //nolint:gosec // average of uint8 values; result ≤ 255
		B: uint8(bSum / n), //nolint:gosec // average of uint8 values; result ≤ 255
		A: 255,             //nolint:mnd   // fully opaque
	}
}

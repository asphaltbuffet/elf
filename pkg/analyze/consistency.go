package analyze

import (
	"sort"

	"gonum.org/v1/plot/plotter"
)

// median returns the median of vals without mutating the input. An empty slice
// returns 0.
func median(vals plotter.Values) float64 {
	if len(vals) == 0 {
		return 0
	}

	s := make(plotter.Values, len(vals))
	copy(s, vals)
	sort.Float64s(s)

	const divisor = 2
	mid := len(s) / divisor
	if len(s)%divisor == 1 {
		return s[mid]
	}

	return (s[mid-1] + s[mid]) / divisor
}

// medianPercents expresses each sample as a percentage of the slice's own
// median (v / median * 100), so the result is dimensionless and centred at 100.
// Returns an empty slice when vals is empty or the median is non-positive (which
// would make the percentage undefined). Does not mutate input.
func medianPercents(vals plotter.Values) plotter.Values {
	const pct = 100.0

	m := median(vals)
	if m <= 0 {
		return plotter.Values{}
	}

	out := make(plotter.Values, len(vals))
	for i, v := range vals {
		out[i] = v / m * pct
	}

	return out
}

package analyze

import (
	"gonum.org/v1/plot/plotter"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// fastestMeans returns the smallest per-language mean for each part index
// (0 = Part One, 1 = Part Two) across the given benchmark data. A part with no
// positive mean is absent from the result, signalling "no baseline available".
// Non-positive means are ignored so a bad sample cannot produce a zero divisor.
func fastestMeans(benchData []*exercise.BenchmarkData) map[int]float64 {
	out := map[int]float64{}

	consider := func(part int, pd *exercise.PartData) {
		if pd == nil || pd.Mean <= 0 {
			return
		}
		if cur, ok := out[part]; !ok || pd.Mean < cur {
			out[part] = pd.Mean
		}
	}

	for _, bd := range benchData {
		for _, impl := range bd.Implementations {
			consider(0, impl.PartOne)
			consider(1, impl.PartTwo)
		}
	}

	return out
}

// relativeSamples divides every sample by its part's baseline mean, producing
// dimensionless "relative runtime" values (the reference language sits at 1×).
// A part with no baseline yields an empty slice. The input map is not mutated.
func relativeSamples(
	samples map[string]map[int]plotter.Values,
	baselines map[int]float64,
) map[string]map[int]plotter.Values {
	out := make(map[string]map[int]plotter.Values, len(samples))

	for lang, parts := range samples {
		out[lang] = map[int]plotter.Values{}

		for part, vals := range parts {
			baseline, ok := baselines[part]
			if !ok {
				out[lang][part] = plotter.Values{}
				continue
			}

			rel := make(plotter.Values, len(vals))
			for i, v := range vals {
				rel[i] = v / baseline
			}
			out[lang][part] = rel
		}
	}

	return out
}

package analyze

import "github.com/asphaltbuffet/elf/pkg/exercise"

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

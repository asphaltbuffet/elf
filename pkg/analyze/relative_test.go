package analyze

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/plotter"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func Test_fastestMeans(t *testing.T) {
	t.Run("picks smallest mean per part", func(t *testing.T) {
		data := []*exercise.BenchmarkData{
			{
				Day: 1,
				Implementations: []*exercise.ImplementationData{
					{Name: "Go", PartOne: &exercise.PartData{Mean: 0.001}, PartTwo: &exercise.PartData{Mean: 0.002}},
					{Name: "Python", PartOne: &exercise.PartData{Mean: 0.05}, PartTwo: &exercise.PartData{Mean: 0.001}},
				},
			},
		}

		got := fastestMeans(data)

		require.Contains(t, got, 0)
		require.Contains(t, got, 1)
		assert.InEpsilon(t, 0.001, got[0], 1e-9, "part one fastest = Go")
		assert.InEpsilon(t, 0.001, got[1], 1e-9, "part two fastest = Python")
	})

	t.Run("absent part is not in the map", func(t *testing.T) {
		data := []*exercise.BenchmarkData{
			{
				Day: 1,
				Implementations: []*exercise.ImplementationData{
					{Name: "Go", PartOne: &exercise.PartData{Mean: 0.001}, PartTwo: nil},
				},
			},
		}

		got := fastestMeans(data)

		assert.Contains(t, got, 0)
		assert.NotContains(t, got, 1, "no part two data → no part two baseline")
	})

	t.Run("ignores non-positive means", func(t *testing.T) {
		data := []*exercise.BenchmarkData{
			{
				Day: 1,
				Implementations: []*exercise.ImplementationData{
					{Name: "Bad", PartOne: &exercise.PartData{Mean: 0}},
					{Name: "Go", PartOne: &exercise.PartData{Mean: 0.001}},
				},
			},
		}

		got := fastestMeans(data)

		assert.InEpsilon(t, 0.001, got[0], 1e-9, "zero mean ignored, Go wins")
	})
}

func Test_relativeSamples(t *testing.T) {
	samples := map[string]map[int]plotter.Values{
		"Go":     {0: {0.001, 0.002}, 1: {0.002}},
		"Python": {0: {0.05}, 1: {0.001}},
	}
	baselines := map[int]float64{0: 0.001, 1: 0.001}

	got := relativeSamples(samples, baselines)

	// Go part one: 0.001/0.001=1, 0.002/0.001=2
	assert.InEpsilon(t, 1.0, got["Go"][0][0], 1e-9)
	assert.InEpsilon(t, 2.0, got["Go"][0][1], 1e-9)
	// Python part one: 0.05/0.001 = 50
	assert.InEpsilon(t, 50.0, got["Python"][0][0], 1e-9)
	// Python part two is the fastest → 1.0
	assert.InEpsilon(t, 1.0, got["Python"][1][0], 1e-9)

	// original is not mutated
	assert.InEpsilon(t, 0.05, samples["Python"][0][0], 1e-9, "input untouched")
}

func Test_relativeSamples_missingBaseline(t *testing.T) {
	samples := map[string]map[int]plotter.Values{
		"Go": {0: {0.001}, 1: {0.002}},
	}
	baselines := map[int]float64{0: 0.001} // no part-two baseline

	got := relativeSamples(samples, baselines)

	assert.InEpsilon(t, 1.0, got["Go"][0][0], 1e-9)
	assert.Empty(t, got["Go"][1], "no baseline → empty relative samples for that part")
}

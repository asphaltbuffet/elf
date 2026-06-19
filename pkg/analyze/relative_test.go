package analyze

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

package analyze

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/plot/plotter"
)

func Test_median(t *testing.T) {
	t.Run("odd count", func(t *testing.T) {
		assert.InEpsilon(t, 2.0, median(plotter.Values{3, 1, 2}), 1e-9)
	})
	t.Run("even count averages middle two", func(t *testing.T) {
		assert.InEpsilon(t, 2.5, median(plotter.Values{1, 2, 3, 4}), 1e-9)
	})
	t.Run("does not mutate input", func(t *testing.T) {
		in := plotter.Values{3, 1, 2}
		_ = median(in)
		assert.Equal(t, plotter.Values{3, 1, 2}, in)
	})
	t.Run("empty is zero", func(t *testing.T) {
		assert.InDelta(t, 0.0, median(plotter.Values{}), 1e-9)
	})
}

func Test_medianPercents(t *testing.T) {
	t.Run("expresses samples as percent of median", func(t *testing.T) {
		// median of {1,2,3} = 2; percents = 50, 100, 150
		got := medianPercents(plotter.Values{1, 2, 3})
		assert.InEpsilon(t, 50.0, got[0], 1e-9)
		assert.InEpsilon(t, 100.0, got[1], 1e-9)
		assert.InEpsilon(t, 150.0, got[2], 1e-9)
	})
	t.Run("empty input yields empty", func(t *testing.T) {
		assert.Empty(t, medianPercents(plotter.Values{}))
	})
	t.Run("non-positive median yields empty", func(t *testing.T) {
		// median of {0,0,0} = 0 → guarded
		assert.Empty(t, medianPercents(plotter.Values{0, 0, 0}))
	})
}

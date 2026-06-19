package analyze

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_dayTicker(t *testing.T) {
	tests := []struct {
		name     string
		min, max float64
		wantLen  int
		wantVals []float64
	}{
		{"single day", 1, 1, 1, []float64{1}},
		{"range 1-5", 1, 5, 5, []float64{1, 2, 3, 4, 5}},
		{"full range 1-25", 1, 25, 25, nil},
		{"empty range (min > max)", 5, 1, 0, nil},
		{"fractional bounds", 0.5, 3.5, 4, []float64{0.5, 1.5, 2.5, 3.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dayTicker(tt.min, tt.max)
			assert.Len(t, got, tt.wantLen)

			if tt.wantVals != nil {
				for i, tick := range got {
					assert.InDelta(t, tt.wantVals[i], tick.Value, 0.001)
					assert.NotEmpty(t, tick.Label)
				}
			}
		})
	}
}

func Test_HumanizedLogTicks_Ticks(t *testing.T) {
	t.Run("micro range", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(1e-6, 1e-3)
		assert.NotEmpty(t, ticks)

		// Verify at least some ticks have labels
		labeled := 0
		for _, tick := range ticks {
			if tick.Label != "" {
				labeled++
			}
		}

		assert.Positive(t, labeled)
	})

	t.Run("seconds range", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(0.1, 100)
		assert.NotEmpty(t, ticks)
	})

	t.Run("same order of magnitude", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(1, 5)
		assert.NotEmpty(t, ticks)
	})

	t.Run("wide range", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(1e-9, 1e3)
		assert.NotEmpty(t, ticks)
	})

	t.Run("panics on min <= 0", func(t *testing.T) {
		h := HumanizedLogTicks{}
		assert.Panics(t, func() { h.Ticks(0, 10) })
	})

	t.Run("panics on max <= 0", func(t *testing.T) {
		h := HumanizedLogTicks{}
		assert.Panics(t, func() { h.Ticks(1, -1) })
	})
}

func Test_RelativeLogTicks(t *testing.T) {
	ticks := RelativeLogTicks{}.Ticks(1, 1000)

	// Collect the labelled (major) ticks.
	labels := map[float64]string{}
	for _, tk := range ticks {
		if tk.Label != "" {
			labels[tk.Value] = tk.Label
		}
	}

	assert.Equal(t, "1×", labels[1])
	assert.Equal(t, "10×", labels[10])
	assert.Equal(t, "100×", labels[100])
	assert.Equal(t, "1000×", labels[1000])
}

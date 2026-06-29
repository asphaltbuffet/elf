package render

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPickDurationUnit(t *testing.T) {
	assert.Equal(t, unitMicro, pickDurationUnit(500*time.Nanosecond), "sub-µs max → µs")
	assert.Equal(t, unitMicro, pickDurationUnit(800*time.Microsecond), "sub-ms max → µs")
	assert.Equal(t, unitMilli, pickDurationUnit(15*time.Millisecond), "ms-range max → ms")
	assert.Equal(t, unitSec, pickDurationUnit(2*time.Second), "s-range max → s")
}

func TestFormatDuration_FixedDecimalsAndUnit(t *testing.T) {
	// The screenshot's run: max is ~15ms, so the unit is ms for ALL rows,
	// and every value shows exactly 3 decimals (decimal points align).
	unit := pickDurationUnit(15325656 * time.Nanosecond)
	assert.Equal(t, unitMilli, unit)

	cases := []struct {
		d    time.Duration
		want string
	}{
		{190984 * time.Nanosecond, "0.191ms"},
		{11634847 * time.Nanosecond, "11.635ms"},
		{185025 * time.Nanosecond, "0.185ms"},
		{15325656 * time.Nanosecond, "15.326ms"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, formatDuration(c.d, unit))
	}
}

func TestFormatDuration_DecimalPointsAlignWhenRightPadded(t *testing.T) {
	// All formatted values, once right-aligned to a common width, must align on
	// the '.' because they share a unit and decimal count.
	unit := unitMilli
	vals := []time.Duration{
		190984 * time.Nanosecond,
		11634847 * time.Nanosecond,
	}

	var rendered []string
	width := 0
	for _, d := range vals {
		s := formatDuration(d, unit)
		rendered = append(rendered, s)
		if len(s) > width {
			width = len(s)
		}
	}

	dotCols := map[int]bool{}
	for _, s := range rendered {
		padded := strings.Repeat(" ", width-len(s)) + s
		dotCols[strings.IndexByte(padded, '.')] = true
	}
	assert.Len(t, dotCols, 1, "decimal points should land in the same column: %v", rendered)
}

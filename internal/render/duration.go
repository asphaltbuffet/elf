package render

import (
	"fmt"
	"time"
)

// durationDecimals is the fixed number of fractional digits shown for every
// duration, so decimal points align vertically within a run.
const durationDecimals = 3

// durationUnit is a display unit for task durations. The whole run uses a single
// unit (chosen from the largest duration) so magnitudes are comparable and the
// decimal points line up.
type durationUnit struct {
	suffix string
	per    time.Duration // one unit expressed as a time.Duration
}

var (
	unitMicro = durationUnit{suffix: "µs", per: time.Microsecond}
	unitMilli = durationUnit{suffix: "ms", per: time.Millisecond}
	unitSec   = durationUnit{suffix: "s", per: time.Second}
)

// pickDurationUnit returns the unit that best fits the largest duration in a run:
// the coarsest unit for which the largest value is still >= 1 of that unit. This
// keeps the biggest row readable (e.g. "15.326ms" rather than "15326.000µs")
// while every smaller row uses the same unit (e.g. "0.191ms").
func pickDurationUnit(maxDur time.Duration) durationUnit {
	switch {
	case maxDur >= time.Second:
		return unitSec
	case maxDur >= time.Millisecond:
		return unitMilli
	default:
		return unitMicro
	}
}

// formatDuration renders d in the given unit with a fixed number of decimals,
// so all rows in a run share one magnitude and align on the decimal point.
func formatDuration(d time.Duration, u durationUnit) string {
	value := float64(d) / float64(u.per)

	return fmt.Sprintf("%.*f%s", durationDecimals, value, u.suffix)
}

// secondsToDuration converts a fractional-seconds Duration field (as carried on
// tasks.Result) into a [time.Duration].
func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

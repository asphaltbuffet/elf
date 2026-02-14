package analyze

import (
	"fmt"
	"math"

	"github.com/dustin/go-humanize"
	"gonum.org/v1/plot"
)

// HumanizedLogTicks is suitable for the Tick.Marker field of an Axis,
// it returns tick marks suitable for a log-scale axis which have been
// humanized.
type HumanizedLogTicks struct {
	// Prec specifies the precision of tick rendering
	// according to the documentation for strconv.FormatFloat.
	Prec int
}

var _ plot.Ticker = HumanizedLogTicks{}

// Ticks returns Ticks in a specified range.
func (t HumanizedLogTicks) Ticks(minVal, maxVal float64) []plot.Tick {
	if minVal <= 0 || maxVal <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}

	val := math.Pow10(int(math.Log10(minVal)))
	maxLimit := math.Pow10(int(math.Ceil(math.Log10(maxVal))) + 1) // add buffer to max so we get label

	var ticks []plot.Tick

	for val < maxLimit {
		for i := 1; i < 10; i++ {
			if i == 1 {
				ticks = append(
					ticks,
					plot.Tick{
						Value: val,
						Label: humanize.SIWithDigits(val, 0, "s"),
					})
			}

			ticks = append(ticks, plot.Tick{Value: val * float64(i)})
		}

		val *= 10
	}

	ticks = append(ticks,
		plot.Tick{
			Value: val,
			Label: humanize.SIWithDigits(val, 0, "s"),
		})

	return ticks
}

func dayTicker(minDay, maxDay float64) []plot.Tick {
	var ticks []plot.Tick

	for i := minDay; i <= maxDay; i++ {
		ticks = append(
			ticks,
			plot.Tick{
				Value: i,
				Label: fmt.Sprintf("%.0f", i),
			},
		)
	}

	return ticks
}

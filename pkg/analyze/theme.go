package analyze

import (
	"image/color"

	xfont "golang.org/x/image/font"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/vg"
)

// Visual style constants for analyze graphs. Centralized so the year line graph
// and the exercise box plot share one "minimal with personality" identity.
const (
	titleFontPt  = 16
	axisFontPt   = 12
	tickFontPt   = 11
	legendFontPt = 11

	axisLineWidthPt = 0.75
	gridGrayLevel   = 0xDD

	seriesLineWidthPt = 2.0 //nolint:unused // visual style constant; consumed by later tasks
	redlineWidthPt    = 2.5 //nolint:unused // visual style constant; consumed by later tasks

	boxOutlineWidthPt = 1.25
	legendPaddingPt   = 4
)

// applyTheme applies the shared, color-independent visual style to a plot:
// white background, bold title, consistent type scale, thin soft axis lines,
// and a padded legend. Color (which encodes language), scales, tickers, grid,
// and series styling stay with the individual graph builders.
func applyTheme(p *plot.Plot) {
	p.BackgroundColor = color.White

	titleFont := p.Title.TextStyle.Font
	titleFont.Weight = xfont.WeightBold
	p.Title.TextStyle.Font = font.From(titleFont, vg.Points(titleFontPt))

	for _, ax := range []*plot.Axis{&p.X, &p.Y} {
		ax.Label.TextStyle.Font = font.From(ax.Label.TextStyle.Font, vg.Points(axisFontPt))
		ax.Tick.Label.Font = font.From(ax.Tick.Label.Font, vg.Points(tickFontPt))
		ax.LineStyle.Width = vg.Points(axisLineWidthPt)
		ax.LineStyle.Color = color.RGBA{R: 80, G: 80, B: 80, A: 255} //nolint:mnd // soft axis gray
	}

	p.Legend.TextStyle.Font = font.From(p.Legend.TextStyle.Font, vg.Points(legendFontPt))
	p.Legend.Padding = vg.Points(legendPaddingPt)
}

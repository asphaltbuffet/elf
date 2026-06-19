package analyze

import (
	"errors"
	"fmt"
	"image/color"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// colorThumb is a Thumbnailer that draws a filled rectangle in a single color,
// used to add BoxPlot entries to a legend (BoxPlot does not implement Thumbnailer).
type colorThumb struct{ c color.Color }

func (t colorThumb) Thumbnail(c *draw.Canvas) {
	pts := []vg.Point{
		{X: c.Min.X, Y: c.Min.Y},
		{X: c.Max.X, Y: c.Min.Y},
		{X: c.Max.X, Y: c.Max.Y},
		{X: c.Min.X, Y: c.Max.Y},
	}
	poly := c.ClipPolygonY(pts)
	c.FillPolygon(t.c, poly)
}

// referenceLineSeconds is AoC's soft running-time target, drawn as a dashed
// reference line on every analyze graph.
const referenceLineSeconds = 15

// referenceMultiple is the relative-runtime baseline: the fastest language for
// each part sits at 1×. It is both the reference line value and the Y-axis
// floor for the relative box plot.
const referenceMultiple = 1.0

// generateBoxPlot renders one exercise's per-language relative-runtime
// distributions, grouped by Part (Part One, then Part Two), one box per language.
func generateBoxPlot(benchData []*exercise.BenchmarkData, outfile string) error {
	const (
		plotWidthInches  = 9 * vg.Inch
		plotHeightInches = 5 * vg.Inch
		plotDPI          = 300
	)

	p, err := buildBoxPlot(benchData)
	if err != nil {
		return err
	}

	return savePlotPNG(p, outfile, plotWidthInches, plotHeightInches, plotDPI)
}

// buildBoxPlot constructs the relative-runtime box plot for one exercise: each
// language's samples are divided by that part's fastest mean, so the fastest
// language sits at 1× and slower languages appear as multiples above it.
func buildBoxPlot(benchData []*exercise.BenchmarkData) (*plot.Plot, error) {
	const redlineDashes = 2

	if len(benchData) == 0 {
		return nil, errors.New("no benchmark data to graph")
	}

	samples, langs := collectBoxSamples(benchData)
	samples = relativeSamples(samples, fastestMeans(benchData))

	p := plot.New()
	p.Title.Text = fmt.Sprintf("Advent of Code %d/%02d: %s",
		benchData[0].Year, benchData[0].Day, benchData[0].Title)
	p.Y.Label.Text = "Relative running time"
	p.Y.Scale = plot.LogScale{}
	p.Y.Tick.Marker = RelativeLogTicks{}
	p.Y.Min = referenceMultiple
	applyTheme(p)

	nominal := []string{"Part One", "Part Two"}

	if err := addBoxGroups(p, samples, langs, nominal); err != nil {
		return nil, err
	}

	// X tick at the centre of each part group.
	center := float64(len(langs)-1) / 2 //nolint:mnd // group centre offset
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{
		{Value: center, Label: nominal[0]},
		{Value: float64(len(langs)+1) + center, Label: nominal[1]},
	})

	// Reference line at 1× — the fastest language for each part.
	refline := plotter.NewFunction(func(_ float64) float64 { return referenceMultiple })
	refline.Color = redlineColor()
	refline.Width = vg.Points(redlineWidthPt)
	refline.Dashes = plotutil.Dashes(redlineDashes)
	p.Add(refline)

	return p, nil
}

// collectBoxSamples aggregates timing data from benchData into per-language,
// per-part sample slices, returning the sample map and a sorted language list.
func collectBoxSamples(benchData []*exercise.BenchmarkData) (map[string]map[int]plotter.Values, []string) {
	samples := map[string]map[int]plotter.Values{}
	langs := []string{}

	for _, bd := range benchData {
		for _, impl := range bd.Implementations {
			if _, ok := samples[impl.Name]; !ok {
				samples[impl.Name] = map[int]plotter.Values{0: {}, 1: {}}
				langs = append(langs, impl.Name)
			}

			if impl.PartOne != nil {
				samples[impl.Name][0] = append(samples[impl.Name][0], impl.PartOne.Data...)
			}

			if impl.PartTwo != nil {
				samples[impl.Name][1] = append(samples[impl.Name][1], impl.PartTwo.Data...)
			}
		}
	}

	sort.Strings(langs)

	return samples, langs
}

// styleBox applies the minimal box-plot styling: a language-colored outline
// (box, median, whiskers, outlier glyphs) over a translucent fill of the same
// color, so the distribution marks stay legible while color still encodes
// language.
func styleBox(bp *plotter.BoxPlot, lang string) {
	lc := colorForLang(lang)
	bp.FillColor = lighten(lc)
	bp.BoxStyle.Color = lc
	bp.BoxStyle.Width = vg.Points(boxOutlineWidthPt)
	bp.MedianStyle.Color = lc
	bp.MedianStyle.Width = vg.Points(boxOutlineWidthPt)
	bp.WhiskerStyle.Color = lc
	bp.WhiskerStyle.Width = vg.Points(boxOutlineWidthPt)
	bp.GlyphStyle.Color = lc
}

// addBoxGroups adds one BoxPlot per language per part-group to p, coloring each
// box by language. Legend entries are added for Part One only (one entry per lang).
func addBoxGroups(p *plot.Plot, samples map[string]map[int]plotter.Values, langs, nominal []string) error {
	const boxWidth = 20

	w := vg.Points(boxWidth)

	for partIdx := range nominal {
		for langIdx, lang := range langs {
			vals := samples[lang][partIdx]
			if len(vals) == 0 {
				continue
			}

			// Group base = partIdx * (len(langs)+1); spread languages by langIdx.
			x := float64(partIdx*(len(langs)+1) + langIdx)

			bp, err := plotter.NewBoxPlot(w, x, vals)
			if err != nil {
				return fmt.Errorf("box plot for %s %s: %w", lang, nominal[partIdx], err)
			}

			styleBox(bp, lang)
			p.Add(bp)

			if partIdx == 0 {
				p.Legend.Add(lang, colorThumb{c: colorForLang(lang)})
			}
		}
	}

	return nil
}

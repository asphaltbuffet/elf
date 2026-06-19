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

// generateBoxPlot renders one exercise's per-language timing distributions,
// grouped by Part (Part One, then Part Two), one box per language per group.
func generateBoxPlot(benchData []*exercise.BenchmarkData, outfile string) error {
	const (
		plotWidthInches  = 9 * vg.Inch
		plotHeightInches = 5 * vg.Inch
		plotDPI          = 300
		redlineDashes    = 2
	)

	if len(benchData) == 0 {
		return errors.New("no benchmark data to graph")
	}

	samples, langs := collectBoxSamples(benchData)

	p := plot.New()
	p.Title.Text = fmt.Sprintf("Advent of Code %d/%02d: %s",
		benchData[0].Year, benchData[0].Day, benchData[0].Title)
	p.Y.Label.Text = "Running time"
	p.Y.Scale = plot.LogScale{}
	p.Y.Tick.Marker = HumanizedLogTicks{}
	p.Y.Min = 0.000001

	nominal := []string{"Part One", "Part Two"}

	if err := addBoxGroups(p, samples, langs, nominal); err != nil {
		return err
	}

	// X tick at the centre of each part group.
	center := float64(len(langs)-1) / 2 //nolint:mnd // group centre offset
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{
		{Value: center, Label: nominal[0]},
		{Value: float64(len(langs)+1) + center, Label: nominal[1]},
	})

	redline := plotter.NewFunction(func(_ float64) float64 { return referenceLineSeconds })
	redline.Color = redlineColor()
	redline.Dashes = plotutil.Dashes(redlineDashes)
	p.Add(redline)

	return savePlotPNG(p, outfile, plotWidthInches, plotHeightInches, plotDPI)
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

			bp.FillColor = colorForLang(lang)
			p.Add(bp)

			if partIdx == 0 {
				p.Legend.Add(lang, colorThumb{c: colorForLang(lang)})
			}
		}
	}

	return nil
}

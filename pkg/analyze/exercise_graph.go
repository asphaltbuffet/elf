package analyze

import (
	"errors"
	"fmt"
	"sort"

	"github.com/dustin/go-humanize"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// referenceLineSeconds is AoC's soft running-time target, drawn as a dashed
// reference line on every analyze graph.
const referenceLineSeconds = 15

// generateBoxPlot renders one exercise's per-language consistency facets: a 2×N
// grid (rows = Parts, columns = languages) where each cell shows that language's
// run samples as a percentage of its own median.
func generateBoxPlot(benchData []*exercise.BenchmarkData, outfile string) error {
	const (
		facetWidthInches  = 4 * vg.Inch
		facetHeightInches = 3 * vg.Inch
		plotDPI           = 150
	)

	grid, err := buildConsistencyFacets(benchData)
	if err != nil {
		return err
	}

	cols := 0
	if len(grid) > 0 {
		cols = len(grid[0])
	}

	w := facetWidthInches * vg.Length(cols)
	h := facetHeightInches * vg.Length(len(grid))

	return saveGridPNG(grid, outfile, font.Length(w), font.Length(h), plotDPI)
}

// buildConsistencyFacets builds the R×N facet grid: rows are the Parts present
// in the data (Part One always; Part Two only when some implementation has Part
// Two data), columns are languages (sorted). Each cell is an independently
// auto-scaled box plot of medianPercents(samples) for that (language, part); a
// missing (language, part) within an existing row is a nil cell so columns stay
// aligned. A single-part Problem renders a 1×N grid — no empty Part Two row.
func buildConsistencyFacets(benchData []*exercise.BenchmarkData) ([][]*plot.Plot, error) {
	if len(benchData) == 0 {
		return nil, errors.New("no benchmark data to graph")
	}

	samples, langs := collectBoxSamples(benchData)

	// Determine which parts are present. Part index 0 (Part One) is always a
	// row; part index 1 (Part Two) only if some language has samples for it.
	partNames := []string{"Part One"}
	hasPartTwo := false
	for _, lang := range langs {
		if len(samples[lang][1]) > 0 {
			hasPartTwo = true
			break
		}
	}
	if hasPartTwo {
		partNames = append(partNames, "Part Two")
	}

	grid := make([][]*plot.Plot, len(partNames))
	for part := range partNames {
		grid[part] = make([]*plot.Plot, len(langs))

		for col, lang := range langs {
			cell, err := buildFacetCell(lang, partNames[part], samples[lang][part], col == 0)
			if err != nil {
				return nil, err
			}

			grid[part][col] = cell // nil when no samples
		}
	}

	return grid, nil
}

// buildFacetCell builds one facet: a box plot of the samples as a percentage of
// their own median, titled with the language and its absolute median. Returns a
// nil plot (blank cell) when there are no samples. leftCol controls whether the
// Y axis label is shown (only the leftmost column carries it).
func buildFacetCell(lang, partName string, samples plotter.Values, leftCol bool) (*plot.Plot, error) {
	if len(samples) == 0 {
		return nil, nil //nolint:nilnil // a nil plot is the intended "blank cell" sentinel
	}

	pct := medianPercents(samples)
	if len(pct) == 0 {
		return nil, nil //nolint:nilnil // non-positive median → blank cell
	}

	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s — %s\n(median %s)", lang, partName,
		humanize.SIWithDigits(median(samples), 1, "s"))
	if leftCol {
		p.Y.Label.Text = "% of own median"
	}
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{}) // suppress X ticks
	applyTheme(p)

	const boxWidth = 40

	bp, err := plotter.NewBoxPlot(vg.Points(boxWidth), 0, pct)
	if err != nil {
		return nil, fmt.Errorf("box plot for %s %s: %w", lang, partName, err)
	}

	styleBox(bp, lang)
	p.Add(bp)

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

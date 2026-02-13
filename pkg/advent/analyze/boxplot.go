package analyze

import (
	"fmt"
	"image/color"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

//nolint:mnd // color definition
var partColor = []color.Color{
	color.RGBA{R: 0, G: 173, B: 216, A: 255},
	color.RGBA{R: 55, G: 118, B: 171, A: 255},
}

// ImplDataMap maps language -> day -> part -> values.
type ImplDataMap map[string]map[int]map[int]plotter.Values

func benchmarkToPlotterValues(benchmarks []*advent.BenchmarkData) ImplDataMap {
	// dataMap is a map of language -> day -> part -> values
	dataMap := make(ImplDataMap)

	for _, b := range benchmarks {
		for _, language := range b.Implementations {
			impl := language

			if _, ok := dataMap[impl.Name]; !ok {
				dataMap[impl.Name] = make(map[int]map[int]plotter.Values)
			}

			if _, ok := dataMap[impl.Name][b.Day]; !ok {
				dataMap[impl.Name][b.Day] = make(map[int]plotter.Values)

				dataMap[impl.Name][b.Day][0] = plotter.Values{}
				dataMap[impl.Name][b.Day][1] = plotter.Values{}
			}

			dataMap[impl.Name][b.Day][0] = append(dataMap[impl.Name][b.Day][0], impl.PartOne.Data...)

			if impl.PartTwo == nil {
				continue
			}

			dataMap[impl.Name][b.Day][1] = append(dataMap[impl.Name][b.Day][1], impl.PartTwo.Data...)
		}
	}

	return dataMap
}

func makePlotForEachImplementation(year int, implData ImplDataMap) (map[string]*plot.Plot, error) {
	// plots maps a filename to a plot
	plots := make(map[string]*plot.Plot)

	for impl, d := range implData {
		p := plot.New()

		p.Title.Text = fmt.Sprintf("Advent of Code %d (%s)", year, impl)

		p.X.Label.Text = "Day"
		p.Y.Label.Text = "Running time"

		p.Y.Scale = plot.LogScale{}
		p.Y.Tick.Marker = HumanizedLogTicks{}

		p.X.Tick.Marker = plot.TickerFunc(dayTicker)

		if err := addDayPartsToPlot(p, d); err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%s-benchmarks.png", strings.ToLower(impl))
		plots[filename] = p
	}

	return plots, nil
}

func addDayPartsToPlot(p *plot.Plot, dayMap map[int]map[int]plotter.Values) error {
	const fontWidth = 10
	const numParts = 2

	for idx := range numParts {
		w := vg.Points(fontWidth)

		// dayMap is a map of day -> part -> values
		for day, partData := range dayMap {
			if _, ok := partData[idx]; !ok {
				continue
			}

			// offset part2 so it doesn't overlap part1
			bp, err := plotter.NewBoxPlot(w, float64(day), partData[idx])
			if err != nil {
				return fmt.Errorf("creating box plot: %w", err)
			}

			bp.FillColor = partColor[idx]
			if idx != 0 {
				bp.Offset = w
			}

			p.Add(bp)
		}
	}

	return nil
}

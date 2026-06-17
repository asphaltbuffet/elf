package analyze

import (
	"errors"
	"fmt"
	"image/color"
	"log/slog"
	"os"
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

//nolint:mnd // color definition
var langColor = map[string]color.Color{
	"Golang": color.RGBA{R: 0, G: 173, B: 216, A: 255},
	"Python": color.RGBA{R: 55, G: 118, B: 171, A: 255},
}

// Graph generates a line graph of benchmark run times and writes it to the configured output file.
func (a *Analyzer) Graph() error {
	return generateLineGraph(a.Data, a.Output)
}

func benchmarkToPlotterXYs(benchmarks []*exercise.BenchmarkData) map[string][]plotter.XYs {
	dataMap := make(map[string][]plotter.XYs)

	for _, bd := range benchmarks {
		for _, impl := range bd.Implementations {
			day := float64(bd.Day)

			if _, ok := dataMap[impl.Name]; !ok {
				dataMap[impl.Name] = make([]plotter.XYs, 2) //nolint:mnd // 2 parts per day
			}

			dataMap[impl.Name][0] = append(dataMap[impl.Name][0], plotter.XY{
				X: day,
				Y: impl.PartOne.Mean,
			})

			if impl.PartTwo == nil {
				continue
			}

			dataMap[impl.Name][1] = append(dataMap[impl.Name][1],
				plotter.XY{
					X: day,
					Y: impl.PartTwo.Mean,
				})
		}
	}

	return dataMap
}

func generateLineGraph(benchData []*exercise.BenchmarkData, outfile string) error {
	const plotWidthInches font.Length = 12.5 * vg.Inch
	const plotHeightInches font.Length = 5 * vg.Inch
	const plotDPI int = 300
	const softYMax float64 = 60

	if len(benchData) == 0 {
		return errors.New("no benchmark data to graph")
	}

	plots, err := NewBenchmarkPlots(benchData[0].Year)
	if err != nil {
		return fmt.Errorf("creating plots: %w", err)
	}

	dataMap := benchmarkToPlotterXYs(benchData)

	for lang, parts := range dataMap {
		for part, xys := range parts {
			var (
				ln *plotter.Line
				pt *plotter.Scatter
			)

			ln, pt, err = plotter.NewLinePoints(xys)
			if err != nil {
				return fmt.Errorf("filling %s part %d plot: %w", lang, part, err)
			}

			ln.Color = langColor[lang]
			pt.Shape = draw.CircleGlyph{}
			pt.Color = langColor[lang]

			plots[0][part].Add(ln, pt)
			plots[0][part].Legend.Add(lang, ln, pt)
		}
	}

	// make sure both plots have the same Y axis for alignment
	yMax := max(plots[0][0].Y.Max, plots[0][1].Y.Max, softYMax)
	plots[0][0].Y.Max = yMax
	plots[0][1].Y.Max = yMax

	yMin := min(plots[0][0].Y.Min, plots[0][1].Y.Min)
	plots[0][0].Y.Min = yMin
	plots[0][1].Y.Min = yMin

	img := vgimg.NewWith(vgimg.UseWH(plotWidthInches, plotHeightInches), vgimg.UseDPI(plotDPI))
	dc := draw.New(img)

	const rows, cols = 1, 2

	t := draw.Tiles{
		Rows:      rows,
		Cols:      cols,
		PadX:      vg.Points(20),
		PadRight:  vg.Points(10),
		PadLeft:   vg.Points(10),
		PadBottom: vg.Points(10),
		PadTop:    vg.Points(10),
	}

	canvases := plot.Align(plots, t, dc)

	for r := range rows {
		for c := range cols {
			if plots[r][c] != nil {
				plots[r][c].Draw(canvases[r][c])
			}
		}
	}

	path, err := filepath.Abs(outfile)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	slog.Info("writing graph", slog.String("path", path)) //nolint:sloglint // standalone function, no logger context

	w, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("creating image file: %w", err)
	}
	defer w.Close()

	png := vgimg.PngCanvas{Canvas: img}

	if _, err = png.WriteTo(w); err != nil {
		return fmt.Errorf("writing image file: %w", err)
	}

	return nil
}

// NewBenchmarkPlots creates a grid of plots for each exercise day in the given year.
func NewBenchmarkPlots(year int) ([][]*plot.Plot, error) {
	const rows, cols = 1, 2
	const yPosRedline = 15
	const redlineDashPattern = 2

	plots := make([][]*plot.Plot, rows)

	for r := range rows {
		plots[r] = make([]*plot.Plot, cols)

		for c := range cols {
			p := plot.New()

			p.X.Label.Text = "Day"

			p.Y.Tick.Marker = HumanizedLogTicks{}
			p.X.Tick.Marker = plot.TickerFunc(dayTicker)
			p.Y.Scale = plot.LogScale{}
			p.Y.Min = 0.000001

			plots[r][c] = p
		}
	}

	part1Plot := plots[0][0]
	part2Plot := plots[0][1]

	part1Plot.Title.Text = fmt.Sprintf(
		"Average Exercise Running Time\nAdvent of Code %d: Part One",
		year)
	part2Plot.Title.Text = fmt.Sprintf(
		"Average Exercise Running Time\nAdvent of Code %d: Part Two",
		year)

	g := plotter.NewGrid()
	g.Vertical.Color = color.Transparent
	part1Plot.Add(g)
	part2Plot.Add(g)

	redline := plotter.NewFunction(func(_ float64) float64 { return yPosRedline })
	redline.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} //nolint:mnd // color definition
	redline.Dashes = plotutil.Dashes(redlineDashPattern)
	part1Plot.Add(redline)
	part2Plot.Add(redline)

	return plots, nil
}

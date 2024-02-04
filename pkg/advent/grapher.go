package advent

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type Plotter struct {
	dir  string
	data []*BenchmarkData
}

var langColor = map[string]color.Color{
	"Golang": color.RGBA{R: 0, G: 173, B: 216, A: 255},
	"Python": color.RGBA{R: 55, G: 118, B: 171, A: 255},
}

func NewGraph(path string) (*Plotter, error) {
	files, err := getBenchmarkFiles(path)
	if err != nil {
		return nil, fmt.Errorf("getting benchmark files: %w", err)
	}

	// load benchmark data from files
	slog.Debug("found benchmark files", "count", len(files))
	benchData := make([]*BenchmarkData, 0, len(files))

	for _, bf := range files {
		var data []*BenchmarkData

		data, err = readBenchmarkFile(bf)
		if err != nil {
			return nil, fmt.Errorf("reading benchmark file: %w", err)
		}

		benchData = append(benchData, data...)
	}

	return &Plotter{
		dir:  path,
		data: benchData,
	}, nil
}

func (p *Plotter) Graph(outfile string) error {
	err := generateLineGraph(p.data, outfile)
	if err != nil {
		return fmt.Errorf("generating graph: %w", err)
	}

	fmt.Printf("wrote %d graph to %s\n", p.data[0].Year, outfile)

	err = generateBoxPlots(p.data, "boxplot.png")
	if err != nil {
		return fmt.Errorf("generating box plot: %w", err)
	}

	return nil
}

func getBenchmarkFiles(dir string) ([]string, error) {
	benchFiles := []string{}

	// get all benchmark.json files recursively
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // expected behavior when walking directories
		}

		if filepath.Base(path) == "benchmark.json" {
			benchFiles = append(benchFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return benchFiles, nil
}

func readBenchmarkFile(path string) ([]*BenchmarkData, error) {
	var bd []*BenchmarkData

	f, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	err = json.Unmarshal(f, &bd)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling json: %w", err)
	}

	return bd, nil
}

func benchmarkToPlotterXYs(benchmarks []*BenchmarkData) map[string][]plotter.XYs {
	dataMap := make(map[string][]plotter.XYs)

	for _, bd := range benchmarks {
		for _, impl := range bd.Implementations {
			impl := impl
			day := float64(bd.Day)

			if _, ok := dataMap[impl.Name]; !ok {
				dataMap[impl.Name] = make([]plotter.XYs, 2)
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
					X: float64(bd.Day),
					Y: impl.PartTwo.Mean,
				})
		}
	}

	return dataMap
}

func generateLineGraph(benchData []*BenchmarkData, outfile string) error {
	if len(benchData) == 0 {
		return fmt.Errorf("no benchmark data to graph")
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
	max := max(plots[0][0].Y.Max, plots[0][1].Y.Max, 60)
	plots[0][0].Y.Max = max
	plots[0][1].Y.Max = max

	min := min(plots[0][0].Y.Min, plots[0][1].Y.Min)
	plots[0][0].Y.Min = min
	plots[0][1].Y.Min = min

	img := vgimg.NewWith(vgimg.UseWH(12.5*vg.Inch, 5*vg.Inch), vgimg.UseDPI(300))
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

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if plots[r][c] != nil {
				plots[r][c].Draw(canvases[r][c])
			}
		}
	}

	path, _ := filepath.Abs(outfile)
	fmt.Printf("writing graph to %s\n", path)

	w, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("creating image file: %w", err)
	}

	png := vgimg.PngCanvas{Canvas: img}

	if _, err = png.WriteTo(w); err != nil {
		return fmt.Errorf("writing image file: %w", err)
	}

	return nil
}

func NewBenchmarkPlots(year int) ([][]*plot.Plot, error) {
	const rows, cols = 1, 2
	plots := make([][]*plot.Plot, rows)

	for j := 0; j < rows; j++ {
		plots[j] = make([]*plot.Plot, cols)

		for i := 0; i < cols; i++ {
			p := plot.New()

			p.X.Label.Text = "Day"

			// p.Y.Label.Text = "Running time (seconds)"
			p.Y.Tick.Marker = HumanizedLogTicks{}
			p.X.Tick.Marker = plot.TickerFunc(func(min, max float64) []plot.Tick {
				ticks := []plot.Tick{}

				for i := min; i <= max; i++ {
					ticks = append(
						ticks,
						plot.Tick{
							Value: i,
							Label: fmt.Sprintf("%.0f", i),
						},
					)
				}

				return ticks
			})
			p.Y.Scale = plot.LogScale{}
			p.Y.Min = 0.000001
			// part1Plot.Y.Max = +10
			// part1Plot.X.Label.Position = draw.PosRight
			// part1Plot.Y.Label.Position = draw.PosTop

			plots[j][i] = p
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

	redline := plotter.NewFunction(func(x float64) float64 { return 15 })
	redline.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	redline.Dashes = plotutil.Dashes(2)
	part1Plot.Add(redline)
	part2Plot.Add(redline)

	return plots, nil
}

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
func (t HumanizedLogTicks) Ticks(min, max float64) []plot.Tick {
	if min <= 0 || max <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}

	val := math.Pow10(int(math.Log10(min)))
	max = math.Pow10(int(math.Ceil(math.Log10(max))) + 1) // add buffer to max so we get label

	var ticks []plot.Tick

	for val < max {
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

type ImplDataMap map[string]map[int]map[int]plotter.Values

func benchmarkToPlotterValues(benchmarks []*BenchmarkData) map[string]map[int]map[int]plotter.Values {
	// dataMap is a map of language -> day -> part -> values
	dataMap := make(map[string]map[int]map[int]plotter.Values)

	for _, b := range benchmarks {
		b := b

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

func generateBoxPlots(benchData []*BenchmarkData, _ string) error {
	if len(benchData) == 0 {
		return fmt.Errorf("no benchmark data to graph")
	}

	// pValues is a map of language -> day -> part -> values
	pValues := benchmarkToPlotterValues(benchData)

	plots, err := makePlotForEachImplementation(benchData[0].Year, pValues)
	if err != nil {
		return fmt.Errorf("creating plots: %w", err)
	}

	for out, p := range plots {
		if err = p.Save(4*vg.Inch, 8*vg.Inch, out); err != nil {
			return fmt.Errorf("saving plot: %w", err)
		}
	}

	return nil
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
		// p.Y.Min = 0.000001

		p.X.Tick.Marker = plot.TickerFunc(dayTicker)

		if err := addDayPartsToPlot(p, d); err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%s-benchmarks.png", strings.ToLower(impl))
		plots[filename] = p
	}

	return plots, nil
}

func dayTicker(min, max float64) []plot.Tick {
	ticks := []plot.Tick{}

	for i := min; i <= max; i++ {
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

func addDayPartsToPlot(p *plot.Plot, dayMap map[int]map[int]plotter.Values) error {
	for idx := 0; idx < 2; idx++ {
		w := vg.Points(10)

		colors := []color.Color{
			color.RGBA{R: 0, G: 173, B: 216, A: 255},
			color.RGBA{R: 55, G: 118, B: 171, A: 255},
		}

		// d is a map of day -> part -> values
		for day, partData := range dayMap {
			if len(partData[idx]) < 2 {
				continue
			}

			// offset part2 so it doesn't overlap part1
			bp, err := plotter.NewBoxPlot(w, float64(day), partData[idx])
			if err != nil {
				return fmt.Errorf("creating box plot: %w", err)
			}

			bp.FillColor = colors[idx]
			if idx != 0 {
				bp.Offset = w
			}

			p.Add(bp)
		}
	}

	return nil
}

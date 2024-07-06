package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/afero"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/analysis"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

type Analyzer struct {
	Data      []*advent.BenchmarkData
	Dir       string
	Output    string
	GraphType analysis.GraphType

	yearly  bool
	daily   bool
	compare bool

	appFs  afero.Fs
	writer io.Writer
	logger *slog.Logger
}

//nolint:mnd // color definition
var langColor = map[string]color.Color{
	"Golang": color.RGBA{R: 0, G: 173, B: 216, A: 255},
	"Python": color.RGBA{R: 55, G: 118, B: 171, A: 255},
}

func NewAnalyzer(config krampus.ExerciseConfiguration, opts ...func(*Analyzer)) (*Analyzer, error) {
	analyzer := &Analyzer{
		appFs:  afero.NewOsFs(),
		writer: os.Stdout,
		logger: config.GetLogger(),
	}

	for _, opt := range opts {
		opt(analyzer)
	}

	if analyzer.Dir == "" {
		return nil, errors.New("no directory specified")
	}

	if analyzer.GraphType == analysis.Invalid {
		analyzer.GraphType = analysis.Line
	}

	err := analyzer.Load()
	if err != nil {
		return nil, fmt.Errorf("loading benchmark data: %w", err)
	}

	return analyzer, nil
}

func WithDirectory(dir string) func(*Analyzer) {
	return func(a *Analyzer) {
		a.Dir = dir
	}
}

func WithYearly(yearly bool) func(*Analyzer) {
	return func(a *Analyzer) {
		a.yearly = yearly
	}
}

func WithOutput(name string) func(*Analyzer) {
	return func(a *Analyzer) {
		a.Output = name
	}
}

func WithDaily(daily bool) func(*Analyzer) {
	return func(a *Analyzer) {
		a.daily = daily
	}
}

func WithCompare(compare bool) func(*Analyzer) {
	return func(a *Analyzer) {
		a.compare = compare
	}
}

func (a *Analyzer) Load() error {
	files, err := getBenchmarkFiles(a.Dir)
	if err != nil {
		return fmt.Errorf("getting benchmark files: %w", err)
	}

	// load benchmark data from files
	a.logger.Debug("found benchmark files", "count", len(files))
	benchData := make([]*advent.BenchmarkData, 0, len(files))

	for _, bf := range files {
		var data []*advent.BenchmarkData

		data, err = readBenchmarkFile(bf)
		if err != nil {
			return fmt.Errorf("reading %s: %w", bf, err)
		}

		benchData = append(benchData, data...)
	}

	a.Data = benchData

	return nil
}

func (a *Analyzer) Graph(gt analysis.GraphType) error {
	switch gt {
	case analysis.Line:
		return generateLineGraph(a.Data, a.Output)

	case analysis.Box:
		return generateBoxPlots(a.Data, a.Output)

	case analysis.Invalid:
		fallthrough

	default:
		return fmt.Errorf("invalid graph type: %s", gt)
	}
}

func (a *Analyzer) Stats() error {
	return advent.ErrNotImplemented
}

func getBenchmarkFiles(dir string) ([]string, error) { //nolint:unparam // expected behavior when walking directories
	benchFiles := []string{}

	// get all benchmark.json files recursively
	_ = filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // expected behavior when walking directories
		}

		if filepath.Base(path) == "benchmark.json" {
			benchFiles = append(benchFiles, path)
		}

		return nil
	})
	// if err != nil {
	// 	return nil, err
	// }

	return benchFiles, nil
}

func readBenchmarkFile(path string) ([]*advent.BenchmarkData, error) {
	var bd []*advent.BenchmarkData

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

func benchmarkToPlotterXYs(benchmarks []*advent.BenchmarkData) map[string][]plotter.XYs {
	dataMap := make(map[string][]plotter.XYs)

	for _, bd := range benchmarks {
		for _, impl := range bd.Implementations {
			impl := impl
			day := float64(bd.Day)

			if _, ok := dataMap[impl.Name]; !ok {
				dataMap[impl.Name] = make([]plotter.XYs, 2) //nolint:mnd
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

func generateLineGraph(benchData []*advent.BenchmarkData, outfile string) error {
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
	max := max(plots[0][0].Y.Max, plots[0][1].Y.Max, softYMax)
	plots[0][0].Y.Max = max
	plots[0][1].Y.Max = max

	min := min(plots[0][0].Y.Min, plots[0][1].Y.Min)
	plots[0][0].Y.Min = min
	plots[0][1].Y.Min = min

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
	const yPosRedline = 15
	const redlineDashPattern = 2

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

	redline := plotter.NewFunction(func(_ float64) float64 { return yPosRedline })
	redline.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255} //nolint:mnd // color definition
	redline.Dashes = plotutil.Dashes(redlineDashPattern)
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

func benchmarkToPlotterValues(benchmarks []*advent.BenchmarkData) map[string]map[int]map[int]plotter.Values {
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

func generateBoxPlots(benchData []*advent.BenchmarkData, _ string) error {
	const plotWidthInches font.Length = 4 * vg.Inch
	const plotHeightInches font.Length = 8 * vg.Inch

	if len(benchData) == 0 {
		return errors.New("no benchmark data to graph")
	}

	// pValues is a map of language -> day -> part -> values
	pValues := benchmarkToPlotterValues(benchData)

	plots, err := makePlotForEachImplementation(benchData[0].Year, pValues)
	if err != nil {
		return fmt.Errorf("creating plots: %w", err)
	}

	for out, p := range plots {
		if err = p.Save(plotWidthInches, plotHeightInches, out); err != nil {
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
	const fontWidth = 10
	const numParts = 2

	for idx := 0; idx < numParts; idx++ {
		w := vg.Points(fontWidth)

		//nolint:mnd // color definition
		colors := []color.Color{
			color.RGBA{R: 0, G: 173, B: 216, A: 255},
			color.RGBA{R: 55, G: 118, B: 171, A: 255},
		}

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

			bp.FillColor = colors[idx]
			if idx != 0 {
				bp.Offset = w
			}

			p.Add(bp)
		}
	}

	return nil
}

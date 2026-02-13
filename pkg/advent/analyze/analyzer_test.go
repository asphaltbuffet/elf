package analyze

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"

	mocks "github.com/asphaltbuffet/elf/mocks/krampus"
	"github.com/asphaltbuffet/elf/pkg/advent"
)

// testLogger returns a logger that discards output.
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// makeBenchmarkData builds valid BenchmarkData for the given year and days.
// Each day gets both "Golang" and "Python" implementations with both parts populated.
func makeBenchmarkData(year int, days ...int) []*advent.BenchmarkData {
	out := make([]*advent.BenchmarkData, 0, len(days))

	for _, d := range days {
		out = append(out, &advent.BenchmarkData{
			Year: year,
			Day:  d,
			Implementations: []*advent.ImplementationData{
				{
					Name: "Golang",
					PartOne: &advent.PartData{
						Mean: 0.001 * float64(d),
						Min:  0.0005 * float64(d),
						Max:  0.002 * float64(d),
						Data: []float64{0.001, 0.0012, 0.0008},
					},
					PartTwo: &advent.PartData{
						Mean: 0.002 * float64(d),
						Min:  0.001 * float64(d),
						Max:  0.004 * float64(d),
						Data: []float64{0.002, 0.0025, 0.0015},
					},
				},
				{
					Name: "Python",
					PartOne: &advent.PartData{
						Mean: 0.05 * float64(d),
						Min:  0.04 * float64(d),
						Max:  0.07 * float64(d),
						Data: []float64{0.05, 0.055, 0.045},
					},
					PartTwo: &advent.PartData{
						Mean: 0.1 * float64(d),
						Min:  0.08 * float64(d),
						Max:  0.15 * float64(d),
						Data: []float64{0.1, 0.12, 0.08},
					},
				},
			},
		})
	}

	return out
}

// makeBenchmarkDataNilPartTwo is like makeBenchmarkData but sets PartTwo = nil
// on the first implementation of each day (for branch coverage).
func makeBenchmarkDataNilPartTwo(year int, days ...int) []*advent.BenchmarkData {
	data := makeBenchmarkData(year, days...)
	for _, bd := range data {
		bd.Implementations[0].PartTwo = nil
	}

	return data
}

func Test_NewAnalyzer(t *testing.T) {
	type args struct {
		opts []func(*Analyzer)
	}

	tests := []struct {
		name      string
		setup     func(*mocks.MockExerciseConfiguration)
		args      args
		want      *Analyzer
		assertion require.ErrorAssertionFunc
	}{
		{
			name:  "no dir",
			setup: func(_ *mocks.MockExerciseConfiguration) {},
			args: args{
				opts: []func(*Analyzer){},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name:  "with directory",
			setup: func(_ *mocks.MockExerciseConfiguration) {},
			args: args{
				opts: []func(*Analyzer){
					WithDirectory("foo/bar"),
				},
			},
			want: &Analyzer{
				Data:   []*advent.BenchmarkData{},
				Dir:    "foo/bar",
				Output: "",
			},
			assertion: require.NoError,
		},
		{
			name:  "with output",
			setup: func(_ *mocks.MockExerciseConfiguration) {},
			args: args{
				opts: []func(*Analyzer){
					WithDirectory("foo/bar"),
					WithOutput("fakeOutput.png"),
				},
			},
			want: &Analyzer{
				Data:   []*advent.BenchmarkData{},
				Dir:    "foo/bar",
				Output: "fakeOutput.png",
			},
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := mocks.NewMockExerciseConfiguration(t)
			mockConfig.EXPECT().GetLogger().Return(testLogger())
			tt.setup(mockConfig)

			got, err := NewAnalyzer(mockConfig, tt.args.opts...)

			tt.assertion(t, err)
			if err == nil {
				assert.EqualExportedValues(t, *tt.want, *got)
			}
		})
	}
}

func Test_dayTicker(t *testing.T) {
	tests := []struct {
		name     string
		min, max float64
		wantLen  int
		wantVals []float64
	}{
		{"single day", 1, 1, 1, []float64{1}},
		{"range 1-5", 1, 5, 5, []float64{1, 2, 3, 4, 5}},
		{"full range 1-25", 1, 25, 25, nil},
		{"empty range (min > max)", 5, 1, 0, nil},
		{"fractional bounds", 0.5, 3.5, 4, []float64{0.5, 1.5, 2.5, 3.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dayTicker(tt.min, tt.max)
			assert.Len(t, got, tt.wantLen)

			if tt.wantVals != nil {
				for i, tick := range got {
					assert.InDelta(t, tt.wantVals[i], tick.Value, 0.001)
					assert.NotEmpty(t, tick.Label)
				}
			}
		})
	}
}

func Test_HumanizedLogTicks_Ticks(t *testing.T) {
	t.Run("micro range", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(1e-6, 1e-3)
		assert.NotEmpty(t, ticks)

		// Verify at least some ticks have labels
		labeled := 0
		for _, tick := range ticks {
			if tick.Label != "" {
				labeled++
			}
		}

		assert.Positive(t, labeled)
	})

	t.Run("seconds range", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(0.1, 100)
		assert.NotEmpty(t, ticks)
	})

	t.Run("same order of magnitude", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(1, 5)
		assert.NotEmpty(t, ticks)
	})

	t.Run("wide range", func(t *testing.T) {
		h := HumanizedLogTicks{}
		ticks := h.Ticks(1e-9, 1e3)
		assert.NotEmpty(t, ticks)
	})

	t.Run("panics on min <= 0", func(t *testing.T) {
		h := HumanizedLogTicks{}
		assert.Panics(t, func() { h.Ticks(0, 10) })
	})

	t.Run("panics on max <= 0", func(t *testing.T) {
		h := HumanizedLogTicks{}
		assert.Panics(t, func() { h.Ticks(1, -1) })
	})
}

func Test_benchmarkToPlotterXYs(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := benchmarkToPlotterXYs(nil)
		assert.Empty(t, got)
	})

	t.Run("empty input", func(t *testing.T) {
		got := benchmarkToPlotterXYs([]*advent.BenchmarkData{})
		assert.Empty(t, got)
	})

	t.Run("single day with two implementations", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1)
		got := benchmarkToPlotterXYs(data)

		require.Contains(t, got, "Golang")
		require.Contains(t, got, "Python")
		assert.Len(t, got["Golang"], 2, "should have 2 parts")
		assert.Len(t, got["Golang"][0], 1, "part one should have 1 XY point")
		assert.Len(t, got["Golang"][1], 1, "part two should have 1 XY point")
		assert.InDelta(t, 1.0, got["Golang"][0][0].X, 0.001)
		assert.InDelta(t, 0.001, got["Golang"][0][0].Y, 0.001)
	})

	t.Run("multiple days", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1, 2, 3)
		got := benchmarkToPlotterXYs(data)

		assert.Len(t, got["Golang"][0], 3)
		assert.Len(t, got["Python"][0], 3)
	})

	t.Run("nil PartTwo", func(t *testing.T) {
		data := makeBenchmarkDataNilPartTwo(2024, 1)
		got := benchmarkToPlotterXYs(data)

		// Golang has nil PartTwo → part two slice should be empty
		assert.Empty(t, got["Golang"][1])
		// Python still has PartTwo
		assert.Len(t, got["Python"][1], 1)
	})
}

func Test_benchmarkToPlotterValues(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := benchmarkToPlotterValues(nil)
		assert.Empty(t, got)
	})

	t.Run("empty input", func(t *testing.T) {
		got := benchmarkToPlotterValues([]*advent.BenchmarkData{})
		assert.Empty(t, got)
	})

	t.Run("single day", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1)
		got := benchmarkToPlotterValues(data)

		require.Contains(t, got, "Golang")
		require.Contains(t, got["Golang"], 1)
		require.Contains(t, got["Golang"][1], 0) // part one
		require.Contains(t, got["Golang"][1], 1) // part two
		assert.Len(t, got["Golang"][1][0], 3, "part one should have 3 data points")
		assert.Len(t, got["Golang"][1][1], 3, "part two should have 3 data points")
	})

	t.Run("multiple days", func(t *testing.T) {
		data := makeBenchmarkData(2024, 1, 5)
		got := benchmarkToPlotterValues(data)

		require.Contains(t, got["Golang"], 1)
		require.Contains(t, got["Golang"], 5)
	})

	t.Run("nil PartTwo", func(t *testing.T) {
		data := makeBenchmarkDataNilPartTwo(2024, 1)
		got := benchmarkToPlotterValues(data)

		// Golang has nil PartTwo → part two values should be empty
		assert.Empty(t, got["Golang"][1][1])
		// Python still has data in part two
		assert.Len(t, got["Python"][1][1], 3)
	})
}

func Test_readBenchmarkFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		got, err := readBenchmarkFile("testdata/valid_benchmark.json")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, 2024, got[0].Year)
		assert.Equal(t, 1, got[0].Day)
		assert.Len(t, got[0].Implementations, 2)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := readBenchmarkFile("testdata/does_not_exist.json")
		require.Error(t, err)
		assert.ErrorContains(t, err, "reading file")
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := readBenchmarkFile("testdata/invalid_benchmark.json")
		require.Error(t, err)
		assert.ErrorContains(t, err, "unmarshalling json")
	})
}

func Test_getBenchmarkFiles(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		dir := t.TempDir()
		got, err := getBenchmarkFiles(dir)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("nested benchmark files", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "2024", "01-test")
		require.NoError(t, os.MkdirAll(sub, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(sub, "benchmark.json"), []byte("[]"), 0o644))

		got, err := getBenchmarkFiles(dir)
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Contains(t, got[0], "benchmark.json")
	})

	t.Run("non-benchmark files ignored", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "other.json"), []byte("{}"), 0o644))

		got, err := getBenchmarkFiles(dir)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("nonexistent dir", func(t *testing.T) {
		got, err := getBenchmarkFiles("/nonexistent/path/that/does/not/exist")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func Test_Load(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		a := &Analyzer{
			Dir:    t.TempDir(),
			logger: testLogger(),
		}

		err := a.Load()
		require.NoError(t, err)
		assert.Empty(t, a.Data)
	})

	t.Run("dir with valid benchmark", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "day01")
		require.NoError(t, os.MkdirAll(sub, 0o755))

		src, err := os.ReadFile("testdata/valid_benchmark.json")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(sub, "benchmark.json"), src, 0o644))

		a := &Analyzer{
			Dir:    dir,
			logger: testLogger(),
		}

		err = a.Load()
		require.NoError(t, err)
		assert.NotEmpty(t, a.Data)
	})

	t.Run("dir with invalid benchmark", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "benchmark.json"), []byte("{bad"), 0o644))

		a := &Analyzer{
			Dir:    dir,
			logger: testLogger(),
		}

		err := a.Load()
		require.Error(t, err)
	})
}

func Test_NewBenchmarkPlots(t *testing.T) {
	plots, err := NewBenchmarkPlots(2024)
	require.NoError(t, err)

	require.Len(t, plots, 1, "should have 1 row")
	require.Len(t, plots[0], 2, "should have 2 columns")

	assert.Contains(t, plots[0][0].Title.Text, "2024")
	assert.Contains(t, plots[0][0].Title.Text, "Part One")
	assert.Contains(t, plots[0][1].Title.Text, "Part Two")

	// Verify Y scale is LogScale
	assert.IsType(t, plot.LogScale{}, plots[0][0].Y.Scale)
	assert.IsType(t, plot.LogScale{}, plots[0][1].Y.Scale)
}

func Test_addDayPartsToPlot(t *testing.T) {
	t.Run("empty dayMap", func(t *testing.T) {
		p := plot.New()
		err := addDayPartsToPlot(p, map[int]map[int]plotter.Values{})
		require.NoError(t, err)
	})

	t.Run("single day both parts", func(t *testing.T) {
		p := plot.New()
		dayMap := map[int]map[int]plotter.Values{
			1: {
				0: {0.001, 0.002, 0.003},
				1: {0.01, 0.02, 0.03},
			},
		}

		err := addDayPartsToPlot(p, dayMap)
		require.NoError(t, err)
	})

	t.Run("multiple days", func(t *testing.T) {
		p := plot.New()
		dayMap := map[int]map[int]plotter.Values{
			1: {
				0: {0.001, 0.002},
				1: {0.01, 0.02},
			},
			2: {
				0: {0.003, 0.004},
				1: {0.03, 0.04},
			},
		}

		err := addDayPartsToPlot(p, dayMap)
		require.NoError(t, err)
	})
}

func Test_makePlotForEachImplementation(t *testing.T) {
	data := makeBenchmarkData(2024, 1)
	pValues := benchmarkToPlotterValues(data)

	plots, err := makePlotForEachImplementation(2024, pValues)
	require.NoError(t, err)

	assert.Contains(t, plots, "golang-benchmarks.png")
	assert.Contains(t, plots, "python-benchmarks.png")
	assert.Contains(t, plots["golang-benchmarks.png"].Title.Text, "Golang")
	assert.Contains(t, plots["python-benchmarks.png"].Title.Text, "Python")
}

func Test_generateLineGraph(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		err := generateLineGraph([]*advent.BenchmarkData{}, "out.png")
		require.Error(t, err)
		assert.ErrorContains(t, err, "no benchmark data")
	})

	t.Run("valid data writes file", func(t *testing.T) {
		dir := t.TempDir()
		outfile := filepath.Join(dir, "test-graph.png")
		data := makeBenchmarkData(2024, 1, 2, 3)

		err := generateLineGraph(data, outfile)
		require.NoError(t, err)

		info, err := os.Stat(outfile)
		require.NoError(t, err)
		assert.Positive(t, info.Size())
	})
}

func Test_Graph(t *testing.T) {
	tests := []struct {
		name      string
		data      []*advent.BenchmarkData
		assertion require.ErrorAssertionFunc
		errMsg    string
	}{
		{
			name:      "line graph",
			data:      makeBenchmarkData(2024, 1),
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)

			a := &Analyzer{
				Data:   tt.data,
				Output: filepath.Join(dir, "test-output.png"),
				logger: testLogger(),
			}

			err := a.Graph()
			tt.assertion(t, err)

			if tt.errMsg != "" && err != nil {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}

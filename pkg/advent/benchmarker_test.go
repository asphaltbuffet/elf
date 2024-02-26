package advent

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	krampusMocks "github.com/asphaltbuffet/elf/mocks/krampus"
	mocks "github.com/asphaltbuffet/elf/mocks/runners"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

func TestBenchmark(t *testing.T) {
	type args struct {
		iterations int
		id         string
		lang       string
		year       int
		day        int
		data       *Data
		path       string
	}

	tests := []struct {
		name    string
		setup   func(_m *mocks.MockRunner)
		args    args
		want    int
		wantErr error
	}{
		{
			name:  "no implementations",
			setup: func(_m *mocks.MockRunner) {},
			args: args{
				iterations: 1,
				id:         "2017-02",
				lang:       "go",
				year:       2017,
				day:        2,
				data:       &Data{},
				path:       "exercises/2017/02-fakeEmptyDay",
			},
			want:    0,
			wantErr: ErrNoImplementations,
		},
		{
			name:  "no input",
			setup: func(_m *mocks.MockRunner) {},
			args: args{
				iterations: 1,
				id:         "2017-03",
				lang:       "go",
				year:       2017,
				day:        3,
				data: &Data{
					InputData:     "",
					InputFileName: "fakeInput.txt",
					TestCases:     TestCase{},
					Answers:       Answer{},
				},
				path: "exercises/2017/03-fakeGoDay",
			},
			want:    0,
			wantErr: ErrNotFound,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up testing
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			// set up mocks
			mockRunner := mocks.NewMockRunner(t)
			tt.setup(mockRunner)

			mockBenchmarker := &Benchmarker{
				Exercise: &Exercise{
					ID:       tt.args.id,
					Title:    "Fake Title",
					Language: tt.args.lang,
					Year:     tt.args.year,
					Day:      tt.args.day,
					URL:      "www.fake.com",
					Data:     tt.args.data,
					Path:     tt.args.path,
					runner:   nil,
					appFs:    testFs,
					logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
				},
				exerciseBaseDir: "",
			}

			// execute the function under test
			got, err := mockBenchmarker.Benchmark(mockBenchmarker.appFs, tt.args.iterations)

			// validate the results
			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				assert.Len(t, got, tt.want)
			}
		})
	}
}

func TestBenchmarkWithMissingInput(t *testing.T) {
	base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	roBase = afero.NewReadOnlyFs(base)

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())

	e := &Exercise{
		ID:       "1111-22",
		Title:    "Fake Title",
		Language: "fakeLang",
		Year:     1111,
		Day:      22,
		Data: &Data{
			InputData:     "",
			InputFileName: "missingInput.txt",
			TestCases:     TestCase{},
			Answers:       Answer{},
		},
		Path:   "",
		runner: nil,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		appFs:  testFs,
	}

	b := &Benchmarker{Exercise: e, exerciseBaseDir: ""}

	// execute the function under test
	_, err := b.Benchmark(testFs, 1)

	require.Error(t, err)
}

func TestNormalizationFactor(t *testing.T) {
	const maxNormTime float64 = 0.5

	t1 := NormalizationFactor()

	// not a great test
	assert.NotZero(t, t1)

	// make sure that it doesn't run too long
	assert.LessOrEqualf(t, t1, maxNormTime, "normalization test takes > %.3fs", maxNormTime)
}

func TestNewBenchmarker(t *testing.T) {
	type args struct {
		options []func(*Benchmarker)
	}

	type wants struct {
		path       string
		exerciseID string
	}

	tests := []struct {
		name      string
		args      args
		wants     wants
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "no options",
			args: args{
				options: []func(*Benchmarker){},
			},
			wants:     wants{},
			assertion: require.Error,
		},
		{
			name: "with invalid path",
			args: args{
				options: []func(*Benchmarker){WithExerciseDir("fake")},
			},
			wants:     wants{path: "fake", exerciseID: ""},
			assertion: require.Error,
		},
		{
			name: "empty path",
			args: args{
				options: []func(*Benchmarker){WithExerciseDir("")},
			},
			wants:     wants{},
			assertion: require.Error,
		},
		{
			name: "valid path",
			args: args{
				options: []func(*Benchmarker){WithExerciseDir("exercises/2017/01-fakeFullDay")},
			},
			wants:     wants{path: "exercises/2017/01-fakeFullDay", exerciseID: "2017-01"},
			assertion: require.NoError,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			mockConfig := krampusMocks.NewMockExerciseConfiguration(t)

			mockConfig.EXPECT().GetFs().Return(testFs)
			mockConfig.EXPECT().GetLogger().Return(slog.New(slog.NewTextHandler(io.Discard, nil)))

			got, err := NewBenchmarker(mockConfig, tt.args.options...)

			tt.assertion(t, err)
			if err == nil {
				require.NotNil(t, got)

				assert.Equal(t, tt.wants.path, got.Path)
				assert.Equal(t, tt.wants.exerciseID, got.ID)
			}
		})
	}
}

func TestBenchmarkerString(t *testing.T) {
	type args struct {
		exercise *Exercise
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "empty exercise",
			args: args{exercise: &Exercise{}},
		},
		{
			name: "valid exercise",
			args: args{exercise: &Exercise{
				ID:    "2015-02",
				Year:  2015,
				Day:   2,
				Title: "Fake Title",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Benchmarker{Exercise: tt.args.exercise}

			assert.Equal(t, tt.args.exercise.String(), b.String())
		})
	}
}

func TestBenchmarkDataString(t *testing.T) {
	type fields struct {
		Date            time.Time
		Title           string
		Year            int
		Day             int
		Runs            int
		Normalization   float64
		Implementations []*ImplementationData
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "no impl data",
			fields: fields{
				Date:            time.Date(2021, time.December, 25, 12, 34, 56, 0, time.UTC),
				Title:           "fake title",
				Year:            2015,
				Day:             2,
				Runs:            69,
				Normalization:   0.42,
				Implementations: []*ImplementationData{},
			},
			want: "BenchmarkData{Date: 2021-12-25, AOC 2015/02, Runs:  69, Normalization: 0.420000, Implementations: []}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BenchmarkData{
				Date:            tt.fields.Date,
				Title:           tt.fields.Title,
				Year:            tt.fields.Year,
				Day:             tt.fields.Day,
				Runs:            tt.fields.Runs,
				Normalization:   tt.fields.Normalization,
				Implementations: tt.fields.Implementations,
			}
			assert.Equal(t, tt.want, b.String())
		})
	}
}

func TestRunBenchmark(t *testing.T) {
	type fields struct {
		exerciseBaseDir string
	}

	type args struct {
		iterations int
	}

	tests := []struct {
		name        string
		setup       func(_m *mocks.MockRunner)
		fields      fields
		args        args
		wantResults []tasks.Result
		wantData    *ImplementationData
		assertion   assert.ErrorAssertionFunc
	}{
		{
			name: "runner start error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Start().Return(fmt.Errorf("fake start error"))
			},
			fields:      fields{exerciseBaseDir: ""},
			args:        args{iterations: 10},
			wantResults: nil,
			wantData:    nil,
			assertion:   assert.Error,
		},
		{
			name: "runner run error",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Start().Return(nil)
				_m.EXPECT().Run(mock.Anything).Return(nil, fmt.Errorf("fake run error"))
			},
			fields:      fields{exerciseBaseDir: ""},
			args:        args{iterations: 10},
			wantResults: nil,
			wantData:    nil,
			assertion:   assert.Error,
		},
		{
			name: "all tasks fail",
			setup: func(_m *mocks.MockRunner) {
				_m.EXPECT().Start().Return(nil)
				_m.EXPECT().Run(mock.Anything).Return(&runners.Result{
					TaskID:   "benchmark.1.1",
					Ok:       false,
					Output:   "fake output",
					Duration: 0.666,
				}, nil)
			},
			fields:      fields{exerciseBaseDir: ""},
			args:        args{iterations: 1},
			wantResults: []tasks.Result{},
			wantData: &ImplementationData{
				Name:    "MOCK",
				PartOne: nil,
				PartTwo: nil,
			},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := mocks.NewMockRunner(t)
			mockRunner.EXPECT().String().Return("MOCK")
			mockRunner.EXPECT().Stop().Return(nil).Maybe()
			mockRunner.EXPECT().Cleanup().Return(nil).Maybe()

			tt.setup(mockRunner)

			b := &Benchmarker{
				Exercise: &Exercise{
					ID:       "2015-01",
					Title:    "Fake Day Title",
					Language: "go",
					Year:     2015,
					Day:      1,
					URL:      "https://fake.com",
					Data:     &Data{},
					Path:     "fake/test/path",
					runner:   mockRunner,
					appFs:    testFs,
					logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
					writer:   io.Discard,
				},
				exerciseBaseDir: tt.fields.exerciseBaseDir,
			}

			got, got1, err := b.runBenchmark(tt.args.iterations)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.wantResults, got)
				assert.Equal(t, tt.wantData, got1)
			}
		})
	}
}

func Test_calculateMetrics(t *testing.T) {
	type args struct {
		results map[runners.Part][]float64
	}

	tests := []struct {
		name      string
		args      args
		want      map[runners.Part]*PartData
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "empty results",
			args: args{
				results: map[runners.Part][]float64{},
			},
			want:      map[runners.Part]*PartData{},
			assertion: require.NoError,
		},
		{
			name: "one result",
			args: args{
				results: map[runners.Part][]float64{
					runners.PartOne: {1.0},
					runners.PartTwo: {2.0},
				},
			},
			want: map[runners.Part]*PartData{
				runners.PartOne: {Mean: 1.0, Min: 1.0, Max: 1.0, Data: []float64{1.0}},
				runners.PartTwo: {Mean: 2.0, Min: 2.0, Max: 2.0, Data: []float64{2.0}},
			},
			assertion: require.NoError,
		},
		{
			name: "multiple results",
			args: args{
				results: map[runners.Part][]float64{
					runners.PartOne: {1.0, 2.0, 3.0},
					runners.PartTwo: {2.0, 3.0, 4.0},
				},
			},
			want: map[runners.Part]*PartData{
				runners.PartOne: {Mean: 2.0, Min: 1.0, Max: 3.0, Data: []float64{1.0, 2.0, 3.0}},
				runners.PartTwo: {Mean: 3.0, Min: 2.0, Max: 4.0, Data: []float64{2.0, 3.0, 4.0}},
			},
			assertion: require.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateMetrics(tt.args.results)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

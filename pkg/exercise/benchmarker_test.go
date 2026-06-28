package exercise

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/runners"
	"github.com/asphaltbuffet/elf/pkg/protocol"
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
			setup: func(_ *mocks.MockRunner) {},
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
			setup: func(_ *mocks.MockRunner) {},
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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			mockRunner := mocks.NewMockRunner(t)
			tt.setup(mockRunner)

			b := &Benchmarker{
				Exercise: &Exercise{
					ID:       tt.args.id,
					Title:    "Fake Title",
					Language: tt.args.lang,
					Year:     tt.args.year,
					Day:      tt.args.day,
					URL:      "www.fake.com",
					Data:     tt.args.data,
					Path:     tt.args.path,
				},
				exerciseBaseDir: "",
			}

			got, err := b.Benchmark(t.Context(), testFs, logger, nil, tt.args.iterations)

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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

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
		Path: "",
	}

	b := &Benchmarker{Exercise: e, exerciseBaseDir: ""}

	_, err := b.Benchmark(t.Context(), testFs, logger, nil, 1)

	require.Error(t, err)
}

func TestNewBenchmarker(t *testing.T) {
	type wants struct {
		path       string
		exerciseID string
	}

	tests := []struct {
		name      string
		path      string
		lang      string
		wants     wants
		assertion require.ErrorAssertionFunc
	}{
		{
			name:      "invalid path",
			path:      "fake",
			lang:      "go",
			wants:     wants{},
			assertion: require.Error,
		},
		{
			name:      "empty path",
			path:      "",
			lang:      "go",
			wants:     wants{},
			assertion: require.Error,
		},
		{
			name:      "valid path",
			path:      "exercises/2017/01-fakeFullDay",
			lang:      "py",
			wants:     wants{path: "exercises/2017/01-fakeFullDay", exerciseID: "2017-01"},
			assertion: require.NoError,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			ex, err := Load(tt.path, tt.lang, "", testFs, logger)

			tt.assertion(t, err)
			if err == nil {
				got := NewBenchmarker(ex)
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

func TestBenchmarker_EmitsBenchmarkEvents(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Use a fixture that has a Go implementation so at least one benchmark task runs.
	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().String().Return("MOCK").Maybe()
	mockRunner.EXPECT().Prepare(mock.Anything).Return(nil)
	mockRunner.EXPECT().Open(mock.Anything).Return(nil)
	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "benchmark.1.0",
		Ok:       true,
		Output:   "42",
		Duration: 0.001,
	}, nil).Times(2) // part1 + part2 for 1 iteration
	mockRunner.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	mockRunner.EXPECT().Cleanup().Return(nil).Maybe()

	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{
		"go": func(_ runners.ExerciseMeta) runners.Runner { return mockRunner },
	})
	t.Cleanup(restore)

	b := &Benchmarker{
		Exercise: &Exercise{
			ID:       "2017-01",
			Title:    "Fake Title",
			Language: "go",
			Year:     2017,
			Day:      1,
			URL:      "www.fake.com",
			Data:     &Data{InputFileName: "input.txt"},
			Path:     "exercises/2017/01-fakeFullDay",
		},
		exerciseBaseDir: "",
	}

	var finished int

	cb := func(ev tasks.Event) {
		if ev.Kind == tasks.EventFinished && ev.Type == tasks.Benchmark {
			finished++
		}
	}

	_, err := b.Benchmark(context.Background(), testFs, logger, cb, 1)
	require.NoError(t, err)
	assert.Positive(t, finished, "expected at least one Benchmark Finished event to be emitted")
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
				_m.EXPECT().Prepare(mock.Anything).Return(errors.New("fake start error"))
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
				_m.EXPECT().Prepare(mock.Anything).Return(nil)
				_m.EXPECT().Open(mock.Anything).Return(nil)
				_m.EXPECT().Run(mock.Anything, mock.Anything).Return(nil, errors.New("fake run error"))
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
				_m.EXPECT().Prepare(mock.Anything).Return(nil)
				_m.EXPECT().Open(mock.Anything).Return(nil)
				_m.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
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

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := mocks.NewMockRunner(t)
			mockRunner.EXPECT().String().Return("MOCK")
			mockRunner.EXPECT().Close(mock.Anything).Return(nil).Maybe()
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
				},
				exerciseBaseDir: tt.fields.exerciseBaseDir,
			}

			got, got1, err := b.runBenchmark(t.Context(), logger, mockRunner, nil, tt.args.iterations)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.wantResults, got)
				assert.Equal(t, tt.wantData, got1)
			}
		})
	}
}

package exercise

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

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

// A benchmark iteration's measurement is its duration, not its output string
// (ADR-0011). An Ok-but-empty-output run must still emit a Finished event and
// contribute a sample so the progress bar can reach 100%.
func TestBenchmarker_EmptyOutputStillEmitsFinishedAndSample(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().String().Return("Go").Maybe()
	mockRunner.EXPECT().Prepare(mock.Anything).Return(nil)
	mockRunner.EXPECT().Open(mock.Anything).Return(nil)
	// Ok, but empty output and a real duration — the old guard dropped these.
	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "benchmark.1.0",
		Ok:       true,
		Output:   "",
		Duration: 0.002,
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

	var gotLang string

	cb := func(ev tasks.Event) {
		if ev.Kind == tasks.EventFinished && ev.Type == tasks.Benchmark {
			finished++

			gotLang = ev.Language
		}
	}

	results, err := b.Benchmark(context.Background(), testFs, logger, cb, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, finished, "both parts must emit Finished despite empty output")
	assert.Len(t, results, 2, "both empty-output iterations must be recorded as samples")
	assert.Equal(t, "Go", gotLang, "benchmark Finished events must carry the runner name")
}

// A single runner that fails to start (Prepare/Open) must be skipped, not abort
// the whole benchmark: the other runners' results still come back.
func TestBenchmarker_SkipsRunnerThatFailsToOpen(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// "go" runs fine; "py" fails to open (e.g. missing interpreter).
	goRunner := mocks.NewMockRunner(t)
	goRunner.EXPECT().String().Return("Go").Maybe()
	goRunner.EXPECT().Prepare(mock.Anything).Return(nil)
	goRunner.EXPECT().Open(mock.Anything).Return(nil)
	goRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID: "benchmark.1.0", Ok: true, Output: "42", Duration: 0.003,
	}, nil).Times(2)
	goRunner.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	goRunner.EXPECT().Cleanup().Return(nil).Maybe()

	pyRunner := mocks.NewMockRunner(t)
	pyRunner.EXPECT().String().Return("Python").Maybe()
	pyRunner.EXPECT().Prepare(mock.Anything).Return(nil).Maybe()
	pyRunner.EXPECT().Open(mock.Anything).Return(errors.New("python3 not found"))
	pyRunner.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	pyRunner.EXPECT().Cleanup().Return(nil).Maybe()

	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{
		"go": func(_ runners.ExerciseMeta) runners.Runner { return goRunner },
		"py": func(_ runners.ExerciseMeta) runners.Runner { return pyRunner },
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

	results, err := b.Benchmark(context.Background(), testFs, logger, nil, 1)
	require.NoError(t, err, "one runner failing to open must not abort the benchmark")
	assert.Len(t, results, 2, "the working runner's two parts must still be benchmarked")
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
			// A benchmark's measurement is its duration, not its Ok flag or
			// output (ADR-0011): a non-timeout Ok:false result is still a
			// recorded sample. Both tasks here decode to PartOne (the mock
			// returns a fixed TaskID), so PartOne carries both samples.
			name: "non-ok results are still measured samples",
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
			fields: fields{exerciseBaseDir: ""},
			args:   args{iterations: 1},
			wantResults: []tasks.Result{
				{
					ID:       "benchmark.1.1",
					Type:     tasks.Benchmark,
					Part:     protocol.PartOne,
					SubPart:  1,
					Status:   tasks.StatusPassed,
					Output:   "fake output",
					Duration: 0.666,
				},
				{
					ID:       "benchmark.1.1",
					Type:     tasks.Benchmark,
					Part:     protocol.PartOne,
					SubPart:  1,
					Status:   tasks.StatusPassed,
					Output:   "fake output",
					Duration: 0.666,
				},
			},
			wantData: &ImplementationData{
				Name: "MOCK",
				PartOne: &PartData{
					Mean: 0.666,
					Min:  0.666,
					Max:  0.666,
					Data: []float64{0.666, 0.666},
				},
				PartTwo: nil,
			},
			assertion: assert.NoError,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := mocks.NewMockRunner(t)
			// String() is only reached once a task runs (for the event language)
			// and on the success return; error paths short-circuit before then.
			mockRunner.EXPECT().String().Return("MOCK").Maybe()
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

func TestBenchmarkData_String_Euler(t *testing.T) {
	bd := BenchmarkData{
		Date:          time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
		Year:          2015,
		Day:           1,
		Number:        42,
		Title:         "Answer",
		Runs:          10,
		Normalization: 1.0,
	}

	got := bd.String()

	assert.Contains(t, got, "Euler #42")
	assert.NotContains(t, got, "AOC")
}

func TestBenchmarkData_String_AoC(t *testing.T) {
	bd := BenchmarkData{
		Date:  time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC),
		Year:  2015,
		Day:   1,
		Title: "Not Quite Lisp",
		Runs:  10,
	}

	got := bd.String()

	assert.Contains(t, got, "AOC 2015/01")
	assert.NotContains(t, got, "Euler")
}

// A Project Euler Problem declares only Part One (declaredParts). The
// benchmarker must not plan or run a phantom Part Two for it.
func TestBenchmark_ProblemHasNoPartTwo(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockRunner := mocks.NewMockRunner(t)
	mockRunner.EXPECT().String().Return("MOCK").Maybe()
	mockRunner.EXPECT().Prepare(mock.Anything).Return(nil)
	mockRunner.EXPECT().Open(mock.Anything).Return(nil)
	mockRunner.EXPECT().Run(mock.Anything, mock.Anything).Return(&protocol.Result{
		TaskID:   "benchmark.1.0",
		Ok:       true,
		Output:   "42",
		Duration: 0.001,
	}, nil).Times(1) // only Part One for 1 iteration
	mockRunner.EXPECT().Close(mock.Anything).Return(nil).Maybe()
	mockRunner.EXPECT().Cleanup().Return(nil).Maybe()

	restore := runners.ResetRegistry(map[string]runners.RunnerCreator{
		"go": func(_ runners.ExerciseMeta) runners.Runner { return mockRunner },
	})
	t.Cleanup(restore)

	b := &Benchmarker{
		Exercise: &Exercise{
			ID:       "euler-1",
			Kind:     KindProblem,
			Number:   1,
			Title:    "Fake Problem",
			Language: "go",
			Data:     &Data{},
			// Reuses the single-implementation (go-only) fixture; Kind:
			// KindProblem is what makes declaredParts() return only Part One.
			Path: "exercises/2017/03-fakeGoDay",
		},
		exerciseBaseDir: "",
	}

	var planned []tasks.Event

	cb := func(e tasks.Event) {
		if e.Kind == tasks.EventPlanned {
			planned = append(planned, e)
		}
	}

	results, err := b.Benchmark(context.Background(), testFs, logger, cb, 1)
	require.NoError(t, err)
	require.NotEmpty(t, planned, "expected at least one Planned event")

	for _, e := range planned {
		assert.Equal(t, protocol.PartOne, e.Part, "problem must only plan Part One")
	}

	require.NotEmpty(t, results)

	for _, r := range results {
		assert.Equal(t, protocol.PartOne, r.Part, "problem must only produce Part One results")
	}
}

package advent

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	krampusMocks "github.com/asphaltbuffet/elf/mocks/krampus"
)

func TestExercise_Benchmark(t *testing.T) {
	type args struct {
		iterations int
		exercise   *Exercise
	}

	tests := []struct {
		name      string
		args      args
		want      int
		wantErr   error
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "no implementations",
			args: args{
				iterations: 1,
				exercise: &Exercise{
					ID:       "2017-02",
					Title:    "",
					Language: "go",
					Year:     2017,
					Day:      2,
					URL:      "",
					Data:     &Data{},
					Path:     "exercises/2017/02-fakeEmptyDay",
					runner:   nil,
					appFs:    nil,
					logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
				},
			},
			want:      0,
			wantErr:   ErrNoImplementations,
			assertion: require.Error,
		},
		{
			name: "no input",
			args: args{
				iterations: 1,
				exercise: &Exercise{
					ID:       "2017-03",
					Title:    "",
					Language: "go",
					Year:     2017,
					Day:      3,
					URL:      "",
					Data: &Data{
						InputData:     "",
						InputFileName: "fakeInput.txt",
						TestCases:     TestCase{},
						Answers:       Answer{},
					},
					Path:   "exercises/2017/03-fakeGoDay",
					runner: nil,
					appFs:  nil,
					logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				},
			},
			want:      0,
			wantErr:   ErrNotFound,
			assertion: require.Error,
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
			mockBenchmarker := &Benchmarker{
				appFs:           testFs,
				exerciseBaseDir: "",
				exercise:        tt.args.exercise,
				lang:            "",
				logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			got, err := mockBenchmarker.Benchmark(mockBenchmarker.appFs, tt.args.iterations)

			t.Log(err)

			tt.assertion(t, err)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Len(t, got, tt.want)
		})
	}
}

func Test_calcNormalizationFactor(t *testing.T) {
	const maxNormTime float64 = 0.5

	t1 := calcNormalizationFactor()

	// not a great test
	assert.NotZero(t, t1)

	// make sure that it doesn't run too long
	assert.LessOrEqualf(t, t1, maxNormTime, "normalization test takes > %.3fs", maxNormTime)
}

func TestNewBenchmarker(t *testing.T) {
	type args struct {
		options []func(*Benchmarker)
		callFs  int
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
			name: "with invalid path",
			args: args{
				options: []func(*Benchmarker){WithExerciseDir("fake")},
				callFs:  1,
			},
			wants:     wants{path: "fake", exerciseID: ""},
			assertion: require.Error,
		},
		{
			name: "valid path",
			args: args{
				options: []func(*Benchmarker){WithExerciseDir("exercises/2017/01-fakeFullDay")},
				callFs:  1,
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

				assert.Equal(t, tt.wants.path, got.exercise.Path)
				assert.Equal(t, tt.wants.exerciseID, got.exercise.ID)
			}
		})
	}
}

func TestBenchmarker_String(t *testing.T) {
	type args struct {
		exercise *Exercise
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty exercise",
			args: args{exercise: &Exercise{}},
			want: "Advent of Code: INVALID EXERCISE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Benchmarker{
				appFs:           nil, // testFs,
				exerciseBaseDir: "",
				exercise:        tt.args.exercise,
				lang:            "",
				logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			assert.Equal(t, tt.want, b.String())
		})
	}
}

func TestBenchmarkData_String(t *testing.T) {
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

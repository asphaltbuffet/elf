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
		id         string
		lang       string
		year       int
		day        int
		data       *Data
		path       string
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
				id:         "2017-02",
				lang:       "go",
				year:       2017,
				day:        2,
				data:       &Data{},
				path:       "exercises/2017/02-fakeEmptyDay",
			},
			want:      0,
			wantErr:   ErrNoImplementations,
			assertion: require.Error,
		},
		{
			name: "no input",
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

				assert.Equal(t, tt.wants.path, got.Path)
				assert.Equal(t, tt.wants.exerciseID, got.ID)
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

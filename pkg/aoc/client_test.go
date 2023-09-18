package aoc

import (
	_ "embed"
	"path/filepath"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

//go:embed testdata/valid_info
var infoJSON []byte

func TestAOCClient_New(t *testing.T) {
	// set up fresh fs for each test
	var err error

	rClient = resty.New().SetBaseURL("https://test.fake")

	baseExercisesDir = "test_exercises"
	appFs, err = makeTestFs()
	require.NoError(t, err)

	got, err := NewAOCClient()

	assert.NoError(t, err)
	assert.NotNil(t, got)

	// we should have 2 runners (Go and Python)
	assert.Equal(t, 2, len(got.Runners))

	// assertions based on structure in test filesystem (see makeTestFs())
	assert.Equal(t, "test_exercises", got.ExercisesDir)
	assert.Equal(t, []int{2015, 2016, 2019}, got.Years)
	assert.Equal(t, map[int]([]int){2015: []int{1, 2}, 2016: []int{1}, 2019: []int{10}}, got.Days)
}

func newTestClient(t *testing.T) *AOCClient {
	t.Helper()

	tc, err := NewAOCClient()
	require.NoError(t, err)

	cfgDir = "test_config"
	baseExercisesDir = "test_exercises"
	appFs, err = makeTestFs()
	require.NoError(t, err)

	rClient = resty.New().SetBaseURL("https://test.fake")

	return tc
}

//go:embed testdata/2015-1PuzzleData.golden
var PuzzleData201501 []byte

func makeTestFs() (afero.Fs, error) {
	appFs = afero.NewMemMapFs()

	// set up input files
	if err := appFs.MkdirAll(filepath.Join("test_config", "inputs"), 0o755); err != nil {
		return nil, err
	}

	if err := afero.WriteFile(
		appFs,
		filepath.Join("test_config", "inputs", "2015-1.txt"),
		[]byte("test input data\ntest input data"),
		0o600,
	); err != nil {
		return nil, err
	}

	// set up cached puzzle files
	if err := appFs.MkdirAll(filepath.Join("test_config", "puzzle_pages"), 0o755); err != nil {
		return nil, err
	}

	if err := afero.WriteFile(appFs, filepath.Join("test_config", "puzzle_pages", "2015-1.txt"), PuzzleData201501, 0o600); err != nil {
		return nil, err
	}

	if err := afero.WriteFile(appFs, filepath.Join("test_config", "puzzle_pages", "2019-10.txt"), PuzzleData201501, 0o600); err != nil {
		return nil, err
	}

	// create app config
	if err := afero.WriteFile(appFs, "elf", []byte("token: 'abcd1234efgh5678'"), 0o644); err != nil {
		return nil, err
	}

	// create exercise dirs
	dirs := []string{
		// these are intentionally out of order to test sorting
		filepath.Join("test_exercises", "2015", "01-testDayOne", "go"),
		filepath.Join("test_exercises", "2019", "10-testDayTen", "py"),
		filepath.Join("test_exercises", "2015", "02-testDayTwo", "py"),
		filepath.Join("test_exercises", "2015", "02-testDayTwo", "go"),
		filepath.Join("test_exercises", "2015", "01-testDayOne", "py"),
		filepath.Join("test_exercises", "2016", "01-testDayOne", "go"),
	}

	for _, d := range dirs {
		if err := appFs.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}

	// create expected files in exercise structure
	err := afero.WriteFile(
		appFs,
		filepath.Join("test_exercises", "2015", "01-testDayOne", "info.json"),
		infoJSON,
		0o644)
	if err != nil {
		return nil, err
	}

	return appFs, nil
}

func TestAOCClient_GetExercise(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		want      *exercise.Exercise
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "exercise doesn't exist",
			args:      args{year: 2020, day: 1},
			want:      nil,
			assertion: assert.Error,
			errText:   "no such exercise",
		},
		{
			name: "exercise exists",
			args: args{year: 2015, day: 1},
			want: &exercise.Exercise{
				Year:  2015,
				Day:   1,
				Title: "Test Day One",
				Dir:   "01-testDayOne",
				Path:  filepath.Join("test_exercises", "2015", "01-testDayOne"),
			},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got, err := ac.GetExercise(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func TestAOCClient_GetExerciseInfo(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		want      *exercise.Info
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "year doesn't exist",
			args:      args{year: 2020, day: 1},
			want:      nil,
			assertion: assert.Error,
			errText:   "no such info",
		},
		{
			name:      "day doesn't exist",
			args:      args{year: 2015, day: 25},
			want:      nil,
			assertion: assert.Error,
			errText:   "no such info",
		},
		{
			name:      "info file doesn't exist",
			args:      args{year: 2015, day: 2},
			want:      nil,
			assertion: assert.Error,
			errText:   "no such info",
		},
		{
			name:      "info file is valid",
			args:      args{year: 2015, day: 1},
			want:      &exercise.Info{InputFile: "test_input.txt"},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got, err := ac.GetExerciseInfo(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want.InputFile, got.InputFile)
			} else {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func TestAOCClient_YearDirs(t *testing.T) {
	ac := newTestClient(t)

	got, err := ac.YearDirs()

	assert.NoError(t, err)
	assert.Equal(t, []string{
		filepath.Join("test_exercises", "2015"),
		filepath.Join("test_exercises", "2016"),
		filepath.Join("test_exercises", "2019"),
	}, got)
}

func TestAOCClient_DayDirs(t *testing.T) {
	type args struct {
		year int
	}

	tests := []struct {
		name      string
		args      args
		want      []string
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "year doesn't exist",
			args:      args{year: 2020},
			want:      nil,
			assertion: assert.Error,
			errText:   "year not found:",
		},
		{
			name:      "year with days",
			args:      args{year: 2015},
			want:      []string{filepath.Join("test_exercises", "2015", "01-testDayOne"), filepath.Join("test_exercises", "2015", "02-testDayTwo")},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got, err := ac.DayDirs(tt.args.year)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func TestAOCClient_ImplementationDirs(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		want      map[string]string
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "year not found",
			args:      args{year: 2020, day: 1},
			want:      nil,
			assertion: assert.Error,
			errText:   "year not found: 2020",
		},
		{
			name:      "day not found",
			args:      args{year: 2015, day: 25},
			want:      nil,
			assertion: assert.Error,
			errText:   "day not found: 25",
		},
		{
			name: "two implementations",
			args: args{year: 2015, day: 1},
			want: map[string]string{
				"go": filepath.Join("test_exercises", "2015", "01-testDayOne", "go"),
				"py": filepath.Join("test_exercises", "2015", "01-testDayOne", "py"),
			},
			assertion: assert.NoError,
		},
		{
			name:      "one implementation",
			args:      args{year: 2016, day: 1},
			want:      map[string]string{"go": filepath.Join("test_exercises", "2016", "01-testDayOne", "go")},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got, err := ac.ImplementationDirs(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func TestAOCClient_MissingDays(t *testing.T) {
	// 2015-1/2, 2016-1, 2019-10
	allDays := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}

	tests := []struct {
		name      string
		want      map[int]([]int)
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "basic test structure",
			want: map[int][]int{
				2015: {3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
				2016: {2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
				2017: allDays,
				2018: allDays,
				2019: {1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
				2020: allDays,
				2021: allDays,
				2022: allDays,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got := ac.MissingDays()

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAOCClient_MissingImplementations(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"none missing", args{year: 2015, day: 1}, []string{}},
		{"one missing", args{year: 2016, day: 1}, []string{"py"}},
		{"exercise doesn't exist", args{year: 2015, day: 10}, []string{"py", "go"}},
		{"year doesn't exist", args{year: 2017, day: 1}, []string{"py", "go"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got := ac.MissingImplementations(tt.args.year, tt.args.day)

			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestAOCClient_GetExerciseInput(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := newTestClient(t)

			got, err := ac.GetExerciseInput(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err == nil {
				assert.Equal(t, tt.want, got)
			} else {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

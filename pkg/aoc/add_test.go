package aoc

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func TestAOCClient_AddExercise(t *testing.T) {
	type args struct {
		year     int
		day      int
		language string
	}

	tests := []struct {
		name      string
		args      args
		want      *exercise.Exercise
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "already exists, return error",
			args: args{2015, 1, "go"},
			want: &exercise.Exercise{
				Year:  2015,
				Day:   1,
				Title: "Test Day One",
				Dir:   "01-testDayOne",
				Path:  filepath.Join("test_exercises", "2015", "01-testDayOne"),
			},
			assertion: assert.Error,
			errText:   "exercise already exists",
		},
		// {
		// 	name: "add go implementation",
		// 	args: args{2019, 10, "go"},
		// 	want: &exercise.Exercise{
		// 		Year:  2019,
		// 		Day:   10,
		// 		Title: "Test Day Ten",
		// 		Dir:   "10-testDayTen",
		// 		Path:  filepath.Join("test_exercises", "2019", "10-testDayTen"),
		// 	},
		// 	assertion: assert.NoError,
		// },
		// {
		// 	name: "missing py implementation",
		// 	args: args{2016, 1, "py"},
		// 	want: &exercise.Exercise{
		// 		Day:  1,
		// 		Name: "Test Day One",
		// 		Dir:  filepath.Join("test_exercises", "2016", "01-testDayOne"),
		// 	},
		// 	assertion: assert.Error,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// recreate for each test to keep testing fs clean
			ac := newTestClient(t)

			got, err := ac.AddExercise(tt.args.year, tt.args.day, tt.args.language)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

//go:embed testdata/2015-1_resp_body
var respBody2015d1 string

func Test_downloadPuzzlePage(t *testing.T) {
	_ = newTestClient(t)
	require.Equal(t, rClient.BaseURL, "https://test.fake")
	httpmock.ActivateNonDefault(rClient.GetClient())

	defer httpmock.DeactivateAndReset()

	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		responder httpmock.Responder
		args      args
		golden    string
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "good request for 2015-1",
			responder: httpmock.NewStringResponder(200, respBody2015d1),
			args: args{
				year: 2015,
				day:  1,
			},
			golden:    "2015-1PuzzleData.golden",
			assertion: assert.NoError,
		},
		{
			name:      "404 response",
			responder: httpmock.NewStringResponder(404, "404 Not Found"),
			args: args{
				year: 2015,
				day:  1,
			},
			golden:    "",
			assertion: assert.Error,
			errText:   "getting puzzle page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()
			httpmock.RegisterResponder("GET", `=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`, tt.responder)

			got, err := downloadPuzzlePage(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				want := goldenValue(t, tt.golden)
				assert.Equal(t, want, got)
			}
		})
	}
}

func goldenValue(t *testing.T, goldenFile string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", goldenFile)) //nolint:gosec // this is test code
	require.NoError(t, err)

	return string(content)
}

func Test_getCachedPuzzlePage(t *testing.T) {
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
		// TODO: add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = newTestClient(t)
			cfgDir = ".config/elf"

			got, err := getCachedPuzzlePage(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_addDay(t *testing.T) {
	_ = newTestClient(t)
	require.Equal(t, rClient.BaseURL, "https://test.fake")
	httpmock.ActivateNonDefault(rClient.GetClient())

	defer httpmock.DeactivateAndReset()

	cfgDir = ".config/elf"

	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		responder httpmock.Responder
		want      *exercise.Exercise
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "create year, day, and exercise files",
			args:      args{year: 2017, day: 1},
			responder: httpmock.NewStringResponder(200, respBody2015d1),
			want: &exercise.Exercise{
				Year:  2017,
				Day:   1,
				Title: "Not Quite Lisp",
				Dir:   "01-notQuiteLisp",
				Path:  filepath.Join("test_exercises", "2017", "01-notQuiteLisp"),
			},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()
			httpmock.RegisterResponder("GET", `=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`, tt.responder)

			got, err := addDay(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
				checkFiles(t, got)
			}
		})
	}
}

// checkFiles verifies that info.json and README.md exist in the exercise's path
func checkFiles(t *testing.T, e *exercise.Exercise) {
	t.Helper()

	_, err := fs.Stat(filepath.Join(e.Path, "info.json"))
	assert.NoError(t, err)

	_, err = fs.Stat(filepath.Join(e.Path, "README.md"))
	assert.NoError(t, err)
}

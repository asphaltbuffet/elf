package aoc

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

var (
	inputDataString = "test input data\ntest input data\n"
	inputDataBytes  = []byte("test input data\ntest input data")
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Helper()

	_ = newTestClient(t)
	require.Equal(t, rClient.BaseURL, "https://test.fake")
	httpmock.ActivateNonDefault(rClient.GetClient())

	return func(t *testing.T) {
		t.Helper()

		httpmock.DeactivateAndReset()
	}
}

func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	httpmock.Reset()

	return func(t *testing.T) {
		t.Helper()

		t.Log("teardown sub-test")
	}
}

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
			name: "all files exist",
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
		{
			name: "add new exercise",
			args: args{2019, 10, "go"},
			want: &exercise.Exercise{
				Year:  2019,
				Day:   10,
				Title: "Test Day One",
				Dir:   "10-testDayOne",
				Path:  filepath.Join("test_exercises", "2019", "10-testDayOne"), // reusing 2015-1 test data, should fix this later
			},
			assertion: assert.NoError,
		},
		{
			name: "missing py implementation",
			args: args{2016, 1, "py"},
			want: &exercise.Exercise{
				Year:  2016,
				Day:   1,
				Title: "Test Day One",
				Dir:   "01-testDayOne",
				Path:  filepath.Join("test_exercises", "2016", "01-testDayOne"),
			},
			assertion: assert.NoError,
		},
		{
			name: "missing year",
			args: args{2020, 1, "py"},
			want: &exercise.Exercise{
				Year:  2020,
				Day:   1,
				Title: "Test Day One",
				Dir:   "01-testDayOne",
				Path:  filepath.Join("test_exercises", "2020", "01-testDayOne"),
			},
			assertion: assert.NoError,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`,
				httpmock.NewStringResponder(http.StatusOK, respBody2015d1),
			)
			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])/input$`,
				httpmock.NewStringResponder(http.StatusOK, inputDataString),
			)

			var err error

			// recreate for each test to keep testing fs clean
			appFs, err = makeTestFs()
			require.NoError(t, err)

			ac, err := GetClient()
			require.NoError(t, err)

			got, err := ac.AddExercise(tt.args.year, tt.args.day, tt.args.language)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)

				// make sure files are there
				checkLanguageDirectoryFiles(t, tt.args.language, got)
				checkExerciseDirectoryFiles(t, got)
			}
		})
	}
}

func goldenValue(t *testing.T, goldenFile string) []byte {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("testdata", goldenFile)) //nolint:gosec // this is test code
	require.NoError(t, err)

	return content
}

func Test_addDay(t *testing.T) {
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
			responder: httpmock.NewStringResponder(http.StatusOK, respBody2015d1),
			want: &exercise.Exercise{
				Year:  2017,
				Day:   1,
				Title: "Test Day One",
				Dir:   "01-testDayOne",
				Path:  filepath.Join("test_exercises", "2017", "01-testDayOne"),
			},
			assertion: assert.NoError,
		},
		{
			name:      "page data not parsable",
			args:      args{year: 2020, day: 1},
			responder: httpmock.NewStringResponder(http.StatusOK, "success getting fake data that isn't what we expect"),
			want:      nil,
			assertion: assert.Error,
			errText:   "getting title for day",
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET", `=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`, tt.responder)

			got, err := addDay(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// checkExerciseDirectoryFiles verifies presence of info.json and README.md
func checkExerciseDirectoryFiles(t *testing.T, e *exercise.Exercise) {
	t.Helper()

	_, err := appFs.Stat(filepath.Join(e.Path, "info.json"))
	assert.NoError(t, err)

	_, err = appFs.Stat(filepath.Join(e.Path, "README.md"))
	assert.NoError(t, err)

	_, err = appFs.Stat(filepath.Join(e.Path, "input.txt"))
	assert.NoError(t, err)
}

// checkLanguageDirectoryFiles verifies presence of exercise.go or __init__.py
func checkLanguageDirectoryFiles(t *testing.T, lang string, e *exercise.Exercise) {
	t.Helper()

	implFiles := map[string]string{
		"go": "exercise.go",
		"py": "__init__.py",
	}

	_, err := appFs.Stat(filepath.Join(e.Path, lang, implFiles[lang]))
	assert.NoError(t, err)
}

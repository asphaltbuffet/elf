package advent

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var NotFoundResponder = httpmock.NewStringResponder(http.StatusNotFound, "404 Not Found")

func setupTestCase(t *testing.T, useTempDir bool) func(t *testing.T) {
	t.Helper()

	cfg, _ = krampus.New()
	cfg.Set("cacheDir", "testdata")
	cfg.Set("exerciseDir", "fs")
	cfg.Set("advent.token", "fake-token")

	if useTempDir {
		cfgDir = t.TempDir()
	} else {
		cfgDir = "testdata"
	}
	cfg.Set("cfgDir", cfgDir)

	exerciseBaseDir = filepath.Join(cfgDir, "exercises")

	rClient.SetBaseURL("https://test.fake")

	httpmock.ActivateNonDefault(rClient.GetClient())

	return func(t *testing.T) {
		t.Helper()

		httpmock.DeactivateAndReset()
		require.NoError(t, os.RemoveAll(exerciseBaseDir))
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

func goldenValue(t *testing.T, goldenFile string) []byte {
	t.Helper()

	content, err := os.ReadFile(goldenFile)
	require.NoError(t, err)

	return content
}

//go:embed testdata/http/input_body
var respBodyInput string

func TestDownload(t *testing.T) {
	type args struct {
		url       string
		lang      string
		overwrite bool
	}

	type goldenFiles struct {
		pageData string
		input    string
	}

	tests := []struct {
		name           string
		args           args
		pageResponder  httpmock.Responder
		inputResponder httpmock.Responder
		golden         goldenFiles
		callCount      int
		assertion      assert.ErrorAssertionFunc
		errText        string
	}{
		{
			name:           "cached data",
			pageResponder:  httpmock.NewStringResponder(http.StatusOK, respBody2015d1),
			inputResponder: httpmock.NewStringResponder(http.StatusOK, respBodyInput),
			args: args{
				url:       "https://adventofcode.com/2015/day/1",
				lang:      "go",
				overwrite: true,
			},
			golden: goldenFiles{
				pageData: filepath.Join("testdata", "golden", "2015-1PuzzleData.golden"),
				input:    filepath.Join("testdata", "golden", "input.golden"),
			},
			callCount: 0,
			assertion: assert.NoError,
		},
		{
			name:           "404 response",
			pageResponder:  NotFoundResponder,
			inputResponder: NotFoundResponder,
			args: args{
				url:       "https://adventofcode.com/2020/day/1",
				lang:      "go",
				overwrite: true,
			},
			callCount: 1,
			assertion: assert.Error,
			errText:   "processing page response",
		},
		{
			name:           "bad url",
			pageResponder:  NotFoundResponder,
			inputResponder: NotFoundResponder,
			args: args{
				url:       "fake/url",
				lang:      "go",
				overwrite: false,
			},
			callCount: 0,
			assertion: assert.Error,
			errText:   "creating exercise from URL",
		},
	}

	teardownTestCase := setupTestCase(t, false)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up testing
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`,
				tt.pageResponder)

			httpmock.RegisterResponder("GET",
				`=~input$`,
				tt.inputResponder)

			// execute function under test
			got, err := Download(tt.args.url, tt.args.lang, tt.args.overwrite)
			t.Log("got", got, "err", err)

			// assert results
			tt.assertion(t, err)
			if err != nil {
				require.ErrorContains(t, err, tt.errText)
			} else {
				// pdWant := goldenValue(t, tt.golden.pageData)
				// inWant := goldenValue(t, tt.golden.input)
				assert.FileExists(t, filepath.Join(got, "input.txt"))

				assert.Equal(t, tt.callCount, httpmock.GetTotalCallCount())
			}
		})
	}
}

func Test_extractTitle(t *testing.T) {
	type args struct {
		page []byte
	}

	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "empty file",
			args: args{
				page: []byte(""),
			},
			want:      "",
			assertion: assert.Error,
		},
		{
			name: "single digit day",
			args: args{
				page: []byte("<h2>--- Day 1: Fake Day Title ---</h2>"),
			},
			want:      "Fake Day Title",
			assertion: assert.NoError,
		},
		{
			name: "two digit day",
			args: args{
				page: []byte("<h2>--- Day 20: Fake Day Title ---</h2>"),
			},
			want:      "Fake Day Title",
			assertion: assert.NoError,
		},
		{
			name: "bad day value",
			args: args{
				page: []byte("<h2>--- Day Two: Fake Day Title ---</h2>"),
			},
			want:      "",
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTitle(tt.args.page)

			tt.assertion(t, err)

			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		year      int
		day       int
		assertion assert.ErrorAssertionFunc
	}{
		{
			name:      "https with valid date",
			url:       "https://adventofcode.com/2015/day/1",
			year:      2015,
			day:       1,
			assertion: assert.NoError,
		},
		{
			name:      "http with valid date",
			url:       "http://adventofcode.com/2015/day/1",
			year:      2015,
			day:       1,
			assertion: assert.NoError,
		},
		{
			name:      "long domain with valid date",
			url:       "https://www.adventofcode.com/2015/day/1",
			year:      2015,
			day:       1,
			assertion: assert.NoError,
		},
		{
			name:      "base url only",
			url:       "https://adventofcode.com",
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
		{
			name:      "incomplete base url",
			url:       "adventofcode.com/2015/day/1",
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
		{
			name:      "no year",
			url:       "https://adventofcode.com/day/1",
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
		{
			name:      "no day",
			url:       "https://adventofcode.com/2015",
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYear, gotDay, err := ParseURL(tt.url)

			tt.assertion(t, err)
			assert.Equal(t, tt.year, gotYear)
			assert.Equal(t, tt.day, gotDay)
		})
	}
}

//go:embed testdata/http/2015-1_resp_body
var respBody2015d1 string

func Test_downloadPuzzlePage(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name          string
		pageResponder httpmock.Responder
		args          args
		golden        string
		assertion     assert.ErrorAssertionFunc
		errText       string
	}{
		{
			name:          "good request for 2015-1",
			pageResponder: httpmock.NewStringResponder(http.StatusOK, respBody2015d1),
			args:          args{year: 2015, day: 1},
			golden:        filepath.Join("testdata", "golden", "2015-1PuzzleData.golden"),
			assertion:     assert.NoError,
		},
		{
			name:          "404 response",
			pageResponder: NotFoundResponder,
			args:          args{year: 2015, day: 1},
			assertion:     assert.Error,
			errText:       "processing page response",
		},
	}

	teardownTestCase := setupTestCase(t, true)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`,
				tt.pageResponder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			got, err := downloadPuzzlePage(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				require.ErrorContains(t, err, tt.errText)
			} else {
				want := goldenValue(t, tt.golden)
				assert.Equal(t, want, got)
			}
		})
	}
}

// func Test_getCachedPuzzlePage(t *testing.T) {
// 	cfgDir = "testdata"

// 	type args struct {
// 		year int
// 		day  int
// 	}

// 	tests := []struct {
// 		name      string
// 		args      args
// 		golden    string
// 		assertion assert.ErrorAssertionFunc
// 	}{
// 		{
// 			name: "cached file exists",
// 			args: args{
// 				year: 2015,
// 				day:  2,
// 			},
// 			golden:    "testdata/golden/2015-02.golden",
// 			assertion: assert.NoError,
// 		},
// 		{
// 			name: "no cached file",
// 			args: args{
// 				year: 2015,
// 				day:  3,
// 			},
// 			assertion: assert.Error,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := getCachedPuzzlePage(tt.args.year, tt.args.day)

// 			tt.assertion(t, err)
// 			if err == nil {
// 				want := goldenValue(t, tt.golden)
// 				assert.Equal(t, want, got)
// 			}
// 		})
// 	}
// }

// func TestExercise_getCachedInput(t *testing.T) {
// 	cfgDir = "testdata"

// 	tests := []struct {
// 		name      string
// 		e         *Exercise
// 		golden    string
// 		assertion assert.ErrorAssertionFunc
// 	}{
// 		{
// 			name:      "cached file exists",
// 			e:         &Exercise{ID: "2015-02"},
// 			golden:    "testdata/golden/input.golden",
// 			assertion: assert.NoError,
// 		},
// 		{
// 			name:      "no cached file",
// 			e:         &Exercise{ID: "2015-03"},
// 			assertion: assert.Error,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.e.getCachedInput()

// 			tt.assertion(t, err)
// 			if err == nil {
// 				want := goldenValue(t, tt.golden)
// 				assert.Equal(t, want, got)
// 			}
// 		})
// 	}
// }

func Test_getExercisePath(t *testing.T) {
	exerciseBaseDir = "testdata/fs"
	// logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	type args struct {
		year int
		day  int
	}
	tests := []struct {
		name     string
		args     args
		wantPath string
		wantOk   bool
	}{
		{
			name: "no year",
			args: args{
				year: 2014,
				day:  1,
			},
			wantPath: "",
			wantOk:   false,
		},
		{
			name: "empty year",
			args: args{
				year: 2015,
				day:  1,
			},
			wantPath: "",
			wantOk:   false,
		},
		{
			name: "full day",
			args: args{
				year: 2017,
				day:  1,
			},
			wantPath: filepath.Join(exerciseBaseDir, "2017", "01-fakeFullDay"),
			wantOk:   true,
		},
		{
			name: "empty day",
			args: args{
				year: 2017,
				day:  2,
			},
			wantPath: filepath.Join(exerciseBaseDir, "2017", "02-fakeEmptyDay"),
			wantOk:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotOk := getExercisePath(tt.args.year, tt.args.day)
			assert.Equal(t, tt.wantPath, gotPath)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestExercise_downloadInput(t *testing.T) {
	tests := []struct {
		name          string
		e             *Exercise
		pageResponder httpmock.Responder
		golden        string
		assertion     assert.ErrorAssertionFunc
		errText       string
	}{
		{
			name:          "new download",
			pageResponder: httpmock.NewStringResponder(http.StatusOK, respBodyInput),
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 1},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			assertion:     assert.NoError,
		},
		{
			name:          "404 response",
			pageResponder: NotFoundResponder,
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 1},
			assertion:     assert.Error,
			errText:       "downloading input data",
		},
	}

	teardownTestCase := setupTestCase(t, true)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET",
				`=~input$`,
				tt.pageResponder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			got, err := tt.e.downloadInput()

			tt.assertion(t, err)
			if err != nil {
				require.ErrorContains(t, err, tt.errText)
			} else {
				want := goldenValue(t, tt.golden)
				assert.Equal(t, want, got)
				// assert.FileExists(t, filepath.Join(cfgDir, "inputs", tt.e.ID))
			}
		})
	}
}

func TestExercise_getInput(t *testing.T) {
	tests := []struct {
		name          string
		e             *Exercise
		pageResponder httpmock.Responder
		golden        string
		callCount     int
		assertion     require.ErrorAssertionFunc
		errText       string
	}{
		{
			name:          "new download",
			pageResponder: httpmock.NewStringResponder(http.StatusOK, respBodyInput),
			e:             &Exercise{ID: "2015-03", Year: 2015, Day: 3},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			callCount:     1,
			assertion:     require.NoError,
		},
		{
			name:          "cached file exists",
			pageResponder: NotFoundResponder,
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 1},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			callCount:     0,
			assertion:     require.NoError,
			errText:       "downloading input data",
		},
		{
			name:          "not cached, 404 response",
			pageResponder: NotFoundResponder,
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 4},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			callCount:     0,
			assertion:     require.NoError,
			errText:       "downloading input data",
		},
	}

	teardownTestCase := setupTestCase(t, true)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			// copy input file to temp dir if it exists
			// _, err := appFs.Stat(filepath.Join("testdata", "inputs", tt.e.ID))
			// if err == nil {
			// 	src, tmpFsErr := appFs.Open(filepath.Join("testdata", "inputs", tt.e.ID))
			// 	require.NoError(t, tmpFsErr, "unable to open testdata input file")
			// 	defer src.Close()

			// 	dst, tmpFsErr := appFs.Create(filepath.Join(cfgDir, "inputs", tt.e.ID))
			// 	require.NoError(t, tmpFsErr, "unable to create temp input file")
			// 	defer dst.Close()

			// 	_, tmpFsErr = io.Copy(dst, src)
			// 	require.NoError(t, tmpFsErr, "unable to copy testdata input file")
			// }

			httpmock.RegisterResponder("GET",
				`=~input$`,
				tt.pageResponder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			got, err := tt.e.getInput()

			tt.assertion(t, err)
			if err != nil {
				require.ErrorContains(t, err, tt.errText)
			} else {
				want := goldenValue(t, tt.golden)
				assert.Equal(t, want, got)
				// assert.FileExists(t, filepath.Join(cfgDir, "inputs", tt.e.ID))
			}
		})
	}
}

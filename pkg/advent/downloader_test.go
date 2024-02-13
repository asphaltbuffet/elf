package advent

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	NotFoundResponder = httpmock.NewStringResponder(http.StatusNotFound, "404 Not Found")

	roBase  afero.Fs
	testFs  afero.Fs
	mockDlr *Downloader
)

// FileExists checks whether a file exists in the given path. It also fails if
// the path points to a directory or there is an error when trying to check the file.
func FileExists(t *testing.T, afs afero.Fs, path string, msgAndArgs ...interface{}) bool {
	t.Helper()

	info, err := afs.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return assert.Fail(t, fmt.Sprintf("unable to find file %q", path), msgAndArgs...)
		}
		return assert.Fail(t, fmt.Sprintf("error when running Fs.Stat(%q): %s", path, err), msgAndArgs...)
	}

	if info.IsDir() {
		return assert.Fail(t, fmt.Sprintf("%q is a directory", path), msgAndArgs...)
	}

	return true
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Helper()

	base := afero.NewBasePathFs(afero.NewOsFs(), "testdata")
	roBase = afero.NewReadOnlyFs(base)

	return func(t *testing.T) {
		t.Helper()

		httpmock.DeactivateAndReset()
	}
}

func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	testFs = afero.NewCopyOnWriteFs(roBase, afero.NewMemMapFs())
	require.NoError(t, testFs.MkdirAll("testCache", 0o755))

	mockDlr = &Downloader{
		appFs:           testFs,
		cacheDir:        "testCache",
		cfgDir:          "./",
		exerciseBaseDir: "exercises",
		exercise:        &Exercise{},
		lang:            "",
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		rClient:         resty.New().SetBaseURL("https://test.fake"),
		token:           "fakeToken",
		url:             "",
	}

	httpmock.ActivateNonDefault(mockDlr.rClient.GetClient())

	httpmock.Reset()

	return func(t *testing.T) {
		t.Helper()

		// t.Log("teardown sub-test")
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

	// type goldenFiles struct {
	// 	pageData string
	// 	input    string
	// }

	tests := []struct {
		name           string
		args           args
		pageResponder  httpmock.Responder
		inputResponder httpmock.Responder
		// golden         goldenFiles
		wantErr error
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
			wantErr: nil,
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
			wantErr: ErrHTTPResponse,
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
			wantErr: ErrInvalidURL,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up testing
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			// set up mocking
			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`,
				tt.pageResponder)

			httpmock.RegisterResponder("GET",
				`=~input$`,
				tt.inputResponder)

			mockDlr.lang = tt.args.lang
			mockDlr.url = tt.args.url

			// execute function under test
			err := mockDlr.Download()

			// assert results
			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				FileExists(t, testFs, filepath.Join(mockDlr.Path(), "input.txt"))
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	type args struct {
		page []byte
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "empty file",
			args: args{
				page: []byte(""),
			},
			want:    "",
			wantErr: ErrInvalidData,
		},
		{
			name: "single digit day",
			args: args{
				page: []byte("<h2>--- Day 1: Fake Day Title ---</h2>"),
			},
			want:    "Fake Day Title",
			wantErr: nil,
		},
		{
			name: "two digit day",
			args: args{
				page: []byte("<h2>--- Day 20: Fake Day Title ---</h2>"),
			},
			want:    "Fake Day Title",
			wantErr: nil,
		},
		{
			name: "bad day value",
			args: args{
				page: []byte("<h2>--- Day Two: Fake Day Title ---</h2>"),
			},
			want:    "",
			wantErr: ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTitle(tt.args.page)

			require.ErrorIs(t, err, tt.wantErr)

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
		wantErr       error
	}{
		{
			name:          "good request for 2015-1",
			pageResponder: httpmock.NewStringResponder(http.StatusOK, respBody2015d1),
			args:          args{year: 2015, day: 1},
			golden:        filepath.Join("testdata", "golden", "2015-1PuzzleData.golden"),
			assertion:     assert.NoError,
			wantErr:       nil,
		},
		{
			name:          "404 response",
			pageResponder: NotFoundResponder,
			args:          args{year: 2015, day: 1},
			assertion:     assert.Error,
			wantErr:       ErrHTTPResponse,
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
				tt.pageResponder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			got, err := mockDlr.downloadPuzzlePage(tt.args.year, tt.args.day)

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
				FileExists(t, testFs, filepath.Join(mockDlr.cacheDir, "pages", makeExerciseID(tt.args.year, tt.args.day)))
			}
		})
	}
}

func Test_getCachedPuzzlePage(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name    string
		args    args
		golden  string
		wantErr error
	}{
		{
			name: "cached file exists",
			args: args{
				year: 2015,
				day:  2,
			},
			golden:  "testdata/golden/2015-02.golden",
			wantErr: nil,
		},
		{
			name: "no cached file",
			args: args{
				year: 2015,
				day:  3,
			},
			wantErr: ErrNotFound,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			got, err := mockDlr.getCachedPuzzlePage(tt.args.year, tt.args.day)

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
			}
		})
	}
}

func TestExercise_getCachedInput(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name    string
		args    args
		golden  string
		wantErr error
	}{
		{
			name:    "cached file exists",
			args:    args{year: 2015, day: 2},
			golden:  "testdata/golden/input.golden",
			wantErr: nil,
		},
		{
			name:    "no cached file",
			args:    args{year: 2015, day: 3},
			golden:  "",
			wantErr: ErrNotFound,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			got, err := mockDlr.getCachedInput(tt.args.year, tt.args.day)

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				want := goldenValue(t, tt.golden)
				assert.Equal(t, want, got)
			}
		})
	}
}

func Test_getExercisePath(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name        string
		args        args
		wantPath    string
		okAssertion assert.BoolAssertionFunc
	}{
		{
			name:        "no year",
			args:        args{year: 2014, day: 1},
			wantPath:    "",
			okAssertion: assert.False,
		},
		{
			name:        "empty year",
			args:        args{year: 2015, day: 1},
			wantPath:    "",
			okAssertion: assert.False,
		},
		{
			name:        "full day",
			args:        args{year: 2017, day: 1},
			wantPath:    filepath.Join("exercises", "2017", "01-fakeFullDay"),
			okAssertion: assert.True,
		},
		{
			name:        "empty day",
			args:        args{year: 2017, day: 2},
			wantPath:    filepath.Join("exercises", "2017", "02-fakeEmptyDay"),
			okAssertion: assert.True,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			// uncomment to view debug logging in test output
			// mockDownloader.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

			gotPath, gotOk := mockDlr.getExercisePath(tt.args.year, tt.args.day)

			require.Equal(t, tt.wantPath, gotPath)
			tt.okAssertion(t, gotOk)
		})
	}
}

func Test_downloadInput(t *testing.T) {
	tests := []struct {
		name          string
		e             *Exercise
		pageResponder httpmock.Responder
		golden        string
		wantErr       error
	}{
		{
			name:          "new download",
			pageResponder: httpmock.NewStringResponder(http.StatusOK, respBodyInput),
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 1},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			wantErr:       nil,
		},
		{
			name:          "404 response",
			pageResponder: NotFoundResponder,
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 1},
			wantErr:       ErrHTTPResponse,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET",
				`=~input$`,
				tt.pageResponder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			mockDlr.exercise = tt.e

			got, err := mockDlr.downloadInput(tt.e.Year, tt.e.Day)

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
				FileExists(t, testFs, filepath.Join(mockDlr.cacheDir, "inputs", makeExerciseID(tt.e.Year, tt.e.Day)))
			}
		})
	}
}

func Test_getInput(t *testing.T) {
	tests := []struct {
		name          string
		e             *Exercise
		pageResponder httpmock.Responder
		golden        string
		callCount     int
		assertion     require.ErrorAssertionFunc
		wantError     error
		statAssertion require.ErrorAssertionFunc
		errText       string
	}{
		{
			name:          "new download",
			pageResponder: httpmock.NewStringResponder(http.StatusOK, respBodyInput),
			e:             &Exercise{ID: "2015-03", Year: 2015, Day: 3},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			assertion:     require.NoError,
			wantError:     nil,
			statAssertion: require.NoError,
		},
		{
			name:          "cached file exists",
			pageResponder: NotFoundResponder,
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 1},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			assertion:     require.NoError,
			wantError:     nil,
			statAssertion: require.NoError,
		},
		{
			name:          "not cached, 404 response",
			pageResponder: NotFoundResponder,
			e:             &Exercise{ID: "2015-01", Year: 2015, Day: 4},
			golden:        filepath.Join("testdata", "golden", "input.golden"),
			assertion:     require.Error,
			statAssertion: require.Error,
			wantError:     ErrHTTPResponse,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET",
				`=~input$`,
				tt.pageResponder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			mockDlr.exercise = tt.e

			got, err := mockDlr.getInput(tt.e.Year, tt.e.Day)

			tt.assertion(t, err)
			if err != nil {
				require.ErrorIs(t, err, tt.wantError)
			} else {
				assert.Equal(t, goldenValue(t, tt.golden), got)
				// _, err = testFs.Stat(filepath.Join(mockDownloader.Path(), "input.txt"))
				// tt.statAssertion(t, err)
			}
		})
	}
}

func Test_makeExercisePath(t *testing.T) {
	type args struct {
		year  int
		day   int
		title string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "happy path",
			args: args{
				year:  2015,
				day:   1,
				title: "Fake Title Day One",
			},
			want: filepath.Join("exercises", "2015", "01-fakeTitleDayOne"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, makeExercisePath("exercises", tt.args.year, tt.args.day, tt.args.title))
		})
	}
}

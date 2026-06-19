package exercise

import (
	_ "embed"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/config"
)

var NotFoundResponder = httpmock.NewStringResponder(http.StatusNotFound, "404 Not Found")

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

			mockDlr.language = tt.args.lang
			mockDlr.url = tt.args.url
			mockDlr.scaffold.inputFileName = "input.txt"
			mockDlr.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

			// execute function under test
			err := mockDlr.Download()

			// assert results
			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				FileExists(t, testFs, filepath.Join(mockDlr.FilePath(), "input.txt"))
			}
		})
	}
}

// TestDownload_existingExercise covers re-downloading an exercise whose directory already exists.
// This path loads info.json instead of assembling a fresh exercise, and must still fetch and write
// the puzzle input rather than writing an empty input.txt.
func TestDownload_existingExercise(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))
	httpmock.RegisterResponder("GET",
		`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`,
		httpmock.NewStringResponder(http.StatusOK, respBody2015d1))
	httpmock.RegisterResponder("GET",
		`=~input$`,
		httpmock.NewStringResponder(http.StatusOK, respBodyInput))

	// 2017-01 exists as a fixture directory with info.json, so Download takes the loadInfo branch.
	mockDlr.language = "go"
	mockDlr.url = "https://adventofcode.com/2017/day/1"
	mockDlr.scaffold.inputFileName = "input.txt"
	mockDlr.scaffold.overwrites = &Overwrites{Input: true}
	mockDlr.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	err := mockDlr.Download()
	require.NoError(t, err)

	got, err := afero.ReadFile(testFs, filepath.Join(mockDlr.FilePath(), "input.txt"))
	require.NoError(t, err)
	assert.Equal(t, respBodyInput, string(got), "input.txt must contain the fetched puzzle input, not be empty")
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

func TestNewDownloader(t *testing.T) {
	type args struct {
		options []func(*Downloader)
		inFile  string
	}

	tests := []struct {
		name      string
		args      args
		want      *Downloader
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "no options",
			args: args{
				options: []func(*Downloader){},
				inFile:  "input.txt",
			},
			want: &Downloader{
				language:        "go",
				exerciseBaseDir: "TEST_exercises",
				cfgDir:          "TEST_cfgdir",
				skipImpl:        false,
				fetcher:         &pageFetcher{cacheDir: "TEST_cache", token: "TEST_token"},
				scaffold:        &exerciseScaffold{inputFileName: "input.txt"},
			},
			assertion: assert.NoError,
		},
		{
			name: "with options",
			args: args{
				options: []func(*Downloader){
					WithDownloadLanguage("py"),
					WithURL("https://fake.url"),
					WithOverwrites(&Overwrites{Input: true}),
					WithSkipImpl(true),
				},
				inFile: "fakeInput.fake",
			},
			want: &Downloader{
				language:        "py",
				exerciseBaseDir: "TEST_exercises",
				cfgDir:          "TEST_cfgdir",
				skipImpl:        true,
				fetcher:         &pageFetcher{cacheDir: "TEST_cache", token: "TEST_token"},
				scaffold:        &exerciseScaffold{inputFileName: "fakeInput.fake"},
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

			t.Setenv("ELF_ADVENT_TOKEN", "TEST_token")
			t.Setenv("ELF_LANGUAGE", "go")

			cfg, cfgErr := config.NewConfig(
				config.WithFs(testFs),
			)
			require.NoError(t, cfgErr)
			cfg.Viper().Set(string(config.AdventDirKey), "TEST_exercises")
			cfg.Viper().Set(string(config.CacheDirKey), "TEST_cache")
			cfg.Viper().Set(string(config.ConfigDirKey), "TEST_cfgdir")
			cfg.Viper().Set(string(config.InputFileKey), tt.args.inFile)

			require.NoError(t, testFs.MkdirAll("TEST_exercises", 0o755))

			got, err := NewDownloader(cfg, tt.args.options...)

			tt.assertion(t, err)
			assert.Equal(t, tt.want.fetcher.cacheDir, got.fetcher.cacheDir)
			assert.Equal(t, tt.want.language, got.language)
			assert.Equal(t, tt.want.scaffold.inputFileName, got.scaffold.inputFileName)
			assert.Equal(t, tt.want.fetcher.token, got.fetcher.token)
			assert.Equal(t, tt.want.exerciseBaseDir, got.exerciseBaseDir)
			assert.Equal(t, tt.want.cfgDir, got.cfgDir)
			assert.Equal(t, tt.want.skipImpl, got.skipImpl)
			assert.NotNil(t, got.logger)
		})
	}
}

func TestDownloader_validate(t *testing.T) {
	tests := []struct {
		name    string
		set     func(*Downloader)
		wantErr error
	}{
		{"all required fields set", func(*Downloader) {}, nil},
		{"language not set", func(d *Downloader) { d.language = "" }, ErrNotConfigured},
		{"client not set", func(d *Downloader) { d.fetcher.rClient = nil }, ErrNotConfigured},
		{"fs not set", func(d *Downloader) { d.appFs = nil }, ErrNotConfigured},
		{"cfg dir not set", func(d *Downloader) { d.cfgDir = "" }, ErrNotConfigured},
		{"cache dir not set", func(d *Downloader) { d.fetcher.cacheDir = "" }, ErrNotConfigured},
		{"base dir not set", func(d *Downloader) { d.exerciseBaseDir = "" }, ErrNotConfigured},
		{"token not set", func(d *Downloader) { d.fetcher.token = "" }, ErrNotConfigured},
	}

	for _, tt := range tests {
		d := &Downloader{
			language:        "fake",
			exerciseBaseDir: "testExercise",
			appFs:           afero.NewMemMapFs(),
			cfgDir:          "TEST_cfgDir",
			skipImpl:        false,
			fetcher: &pageFetcher{
				rClient:  resty.New(),
				token:    "tt.fields.token",
				cacheDir: "TEST_cacheDir",
			},
			scaffold: &exerciseScaffold{
				inputFileName: "tt.fields.inputFileName",
				overwrites:    &Overwrites{},
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			tt.set(d)

			require.ErrorIs(t, d.validate(), tt.wantErr)
		})
	}
}

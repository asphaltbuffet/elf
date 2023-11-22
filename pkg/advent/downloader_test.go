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
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Helper()

	cfgDir = t.TempDir()

	rClient.SetBaseURL("https://test.fake")

	require.Equal(t, "https://test.fake", rClient.BaseURL)
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

	content, err := os.ReadFile(filepath.Join("testdata", goldenFile))
	require.NoError(t, err)

	return content
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
	type args struct {
		url string
	}

	tests := []struct {
		name      string
		args      args
		year      int
		day       int
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "https with valid date",
			args: args{
				url: "https://adventofcode.com/2015/day/1",
			},
			year:      2015,
			day:       1,
			assertion: assert.NoError,
		},
		{
			name: "http with valid date",
			args: args{
				url: "http://adventofcode.com/2015/day/1",
			},
			year:      2015,
			day:       1,
			assertion: assert.NoError,
		},
		{
			name: "long domain with valid date",
			args: args{
				url: "https://www.adventofcode.com/2015/day/1",
			},
			year:      2015,
			day:       1,
			assertion: assert.NoError,
		},
		{
			name: "base url only",
			args: args{
				url: "https://adventofcode.com",
			},
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
		{
			name: "incomplete base url",
			args: args{
				url: "adventofcode.com/2015/day/1",
			},
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
		{
			name: "no year",
			args: args{
				url: "https://adventofcode.com/day/1",
			},
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
		{
			name: "no day",
			args: args{
				url: "https://adventofcode.com/2015",
			},
			year:      0,
			day:       0,
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYear, gotDay, err := ParseURL(tt.args.url)

			tt.assertion(t, err)
			assert.Equal(t, tt.year, gotYear)
			assert.Equal(t, tt.day, gotDay)
		})
	}
}

//go:embed testdata/2015-1_resp_body
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
			golden:        "2015-1PuzzleData.golden",
			assertion:     assert.NoError,
		},
		{
			name:          "404 response",
			pageResponder: httpmock.NewStringResponder(http.StatusNotFound, "404 Not Found"),
			args:          args{year: 2015, day: 1},
			assertion:     assert.Error,
			errText:       "processing page response",
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

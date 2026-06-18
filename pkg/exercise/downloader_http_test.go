package exercise

import (
	_ "embed"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/http/2015-1_resp_body
var respBody2015d1 string

func Test_downloadPage(t *testing.T) {
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

			got, err := mockDlr.fetcher.downloadPage(tt.args.year, tt.args.day)

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
				FileExists(
					t,
					testFs,
					filepath.Join(mockDlr.fetcher.cacheDir, "pages", makeExerciseID(tt.args.year, tt.args.day)),
				)
			}
		})
	}
}

func Test_getCachedPage(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name   string
		args   args
		golden string
		wantOk assert.BoolAssertionFunc
	}{
		{
			name:   "cached file exists",
			args:   args{2015, 2},
			golden: "testdata/golden/2015-02.golden",
			wantOk: assert.True,
		},
		{
			name:   "no cached file",
			args:   args{2015, 3},
			golden: "", // no golden file for failure
			wantOk: assert.False,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			got, gotOk := mockDlr.fetcher.getCachedPage(tt.args.year, tt.args.day)

			tt.wantOk(t, gotOk)
			if gotOk {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
			}
		})
	}
}

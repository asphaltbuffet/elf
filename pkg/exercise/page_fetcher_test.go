package exercise

import (
	_ "embed"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/afero"
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

			got, err := testAdder.fetcher.downloadPage(tt.args.year, tt.args.day)

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
				FileExists(
					t,
					testFs,
					filepath.Join(testAdder.fetcher.cacheDir, "pages", makeExerciseID(tt.args.year, tt.args.day)),
				)
			}
		})
	}
}

func Test_getCachedInput(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name   string
		args   args
		seed   []byte // file contents to seed at the cache path; nil = no file
		wantOk assert.BoolAssertionFunc
	}{
		{
			name:   "cached input exists",
			args:   args{2015, 1},
			seed:   []byte("(())"),
			wantOk: assert.True,
		},
		{
			name:   "no cached file",
			args:   args{2015, 3},
			seed:   nil,
			wantOk: assert.False,
		},
		{
			name:   "empty cached file is a miss",
			args:   args{2015, 2},
			seed:   []byte{},
			wantOk: assert.False,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			if tt.seed != nil {
				inputsDir := filepath.Join(testAdder.fetcher.cacheDir, "inputs")
				require.NoError(t, testFs.MkdirAll(inputsDir, 0o755))
				require.NoError(t, afero.WriteFile(
					testFs,
					filepath.Join(inputsDir, makeExerciseID(tt.args.year, tt.args.day)),
					tt.seed,
					0o600,
				))
			}

			got, gotOk := testAdder.fetcher.getCachedInput(tt.args.year, tt.args.day)

			tt.wantOk(t, gotOk)
			if gotOk {
				assert.Equal(t, tt.seed, got)
			}
		})
	}
}

func Test_downloadInput(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		responder httpmock.Responder
		args      args
		want      []byte
		assertion assert.ErrorAssertionFunc
		wantErr   error
	}{
		{
			name:      "good request for 2015-1",
			responder: httpmock.NewStringResponder(http.StatusOK, "(())\n"),
			args:      args{year: 2015, day: 1},
			want:      []byte("(())"),
			assertion: assert.NoError,
			wantErr:   nil,
		},
		{
			name:      "empty body is rejected",
			responder: httpmock.NewStringResponder(http.StatusOK, "\n"),
			args:      args{year: 2015, day: 1},
			want:      nil,
			assertion: assert.Error,
			wantErr:   ErrEmptyInput,
		},
		{
			name:      "404 response",
			responder: NotFoundResponder,
			args:      args{year: 2015, day: 1},
			want:      nil,
			assertion: assert.Error,
			wantErr:   ErrHTTPResponse,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])/input$`,
				tt.responder)

			httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

			got, err := testAdder.fetcher.downloadInput(tt.args.year, tt.args.day)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)

			if err == nil {
				FileExists(
					t,
					testFs,
					filepath.Join(testAdder.fetcher.cacheDir, "inputs", makeExerciseID(tt.args.year, tt.args.day)),
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

			got, gotOk := testAdder.fetcher.getCachedPage(tt.args.year, tt.args.day)

			tt.wantOk(t, gotOk)
			if gotOk {
				want := goldenValue(t, tt.golden)

				assert.Equal(t, want, got)
			}
		})
	}
}

func Test_newPageFetcher(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		cacheDir string
		wantErr  error
	}{
		{"valid", "fakeToken", "testCache", nil},
		{"empty token", "", "testCache", ErrNotConfigured},
		{"empty cache dir", "fakeToken", "", ErrNotConfigured},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			got, err := newPageFetcher(tt.token, tt.cacheDir, afero.NewMemMapFs(), log)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.NotNil(t, got.rClient)
			assert.Equal(t, tt.token, got.token)
			assert.Equal(t, tt.cacheDir, got.cacheDir)
		})
	}
}

func Test_pageFetcher_sendsUserAgent(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	var gotUA string

	httpmock.RegisterResponder("GET",
		`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])$`,
		func(req *http.Request) (*http.Response, error) {
			gotUA = req.Header.Get("User-Agent")

			return httpmock.NewStringResponse(http.StatusOK, respBody2015d1), nil
		})
	httpmock.RegisterNoResponder(httpmock.NewNotFoundResponder(t.Error))

	_, err := testAdder.fetcher.downloadPage(2015, 1)
	require.NoError(t, err)

	assert.Equal(t, userAgent, gotUA)
}

package aoc

import (
	_ "embed"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

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
			args: args{
				year: 2015,
				day:  1,
			},
			golden:    "2015-1PuzzleData.golden",
			assertion: assert.NoError,
		},
		{
			name:          "404 response",
			pageResponder: httpmock.NewStringResponder(http.StatusNotFound, "404 Not Found"),
			args: args{
				year: 2015,
				day:  1,
			},
			golden:    "",
			assertion: assert.Error,
			errText:   "getting puzzle page",
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
				assert.ErrorContains(t, err, tt.errText)
			} else {
				want := goldenValue(t, tt.golden)
				assert.Equal(t, want, got)
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
		args      args
		responder httpmock.Responder
		want      []byte
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "good response",
			args: args{
				year: 2015,
				day:  1,
			},
			responder: httpmock.NewStringResponder(http.StatusOK, inputDataString),
			want:      inputDataBytes,
			assertion: assert.NoError,
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder("GET",
				`=~^/(201[5-9]|202[012])/day/([1-9]|1[0-9]|2[0-5])/input$`,
				tt.responder)

			got, err := downloadInput(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

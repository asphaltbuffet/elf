package aoc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getCachedPuzzlePage(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		golden    string
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "cached data exists",
			args: args{
				year: 2015,
				day:  1,
			},
			golden:    "2015-1PuzzleData.golden",
			assertion: assert.NoError,
			errText:   "",
		},
		{
			name: "no cached data",
			args: args{
				year: 2016,
				day:  1,
			},
			golden:    "",
			assertion: assert.Error,
			errText:   "reading puzzle page:",
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			got, err := getCachedPuzzlePage(tt.args.year, tt.args.day)

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

func Test_getCachedInput(t *testing.T) {
	type args struct {
		year int
		day  int
	}

	tests := []struct {
		name      string
		args      args
		want      []byte
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "input file exists",
			args: args{
				year: 2015,
				day:  1,
			},
			want:      inputDataBytes,
			assertion: assert.NoError,
		},
		{
			name: "input file not present",
			args: args{
				year: 2015,
				day:  2,
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "reading cached input",
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCachedInput(tt.args.year, tt.args.day)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

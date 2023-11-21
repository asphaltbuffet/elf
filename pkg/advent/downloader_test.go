package advent

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

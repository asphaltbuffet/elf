package advent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

func TestNew(t *testing.T) {
	type args struct {
		id   string
		lang string
	}

	tests := []struct {
		name      string
		args      args
		want      *advent.Exercise
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "first day",
			args: args{id: "2015-01", lang: "go"},
			want: &advent.Exercise{
				ID:       "2015-01",
				Language: "go",
				Year:     2015,
				Day:      1,
			},
			assertion: assert.NoError,
		},
		{
			name: "last day",
			args: args{id: "2015-25", lang: "go"},
			want: &advent.Exercise{
				ID:       "2015-25",
				Language: "go",
				Year:     2015,
				Day:      25,
			},
			assertion: assert.NoError,
		},
		{
			name: "day needs formatting",
			args: args{id: "2023-4", lang: "py"},
			want: &advent.Exercise{
				ID:       "2023-04",
				Language: "py",
				Year:     2023,
				Day:      4,
			},
			assertion: assert.NoError,
		},
		{
			name: "full shorthand",
			args: args{id: "23-4", lang: "py"},
			want: &advent.Exercise{
				ID:       "2023-04",
				Language: "py",
				Year:     2023,
				Day:      4,
			},
			assertion: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := advent.New(tt.args.id, tt.args.lang)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExercise_String(t *testing.T) {
	tests := []struct {
		name string
		e    *advent.Exercise
		want string
	}{
		{"first day in go", &advent.Exercise{ID: "2015-01", Year: 2015, Day: 1, Language: "go"}, "Advent of Code: 2015-01 (Golang)"},
		{"last day in go", &advent.Exercise{ID: "2015-25", Year: 2015, Day: 25, Language: "py"}, "Advent of Code: 2015-25 (Python)"},
		{"invalid language", &advent.Exercise{ID: "2015-01", Year: 2015, Day: 1, Language: "foo"}, "Advent of Code: 2015-01 (INVALID LANGUAGE)"},
		{"empty exercise", &advent.Exercise{}, "Advent of Code: 0000-00 (INVALID LANGUAGE)"},
		{"nil exercise", nil, "Advent of Code: INVALID EXERCISE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.e.String())
		})
	}
}

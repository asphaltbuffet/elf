package advent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func TestExercise_String(t *testing.T) {
	tests := []struct {
		name string
		e    *Exercise
		want string
	}{
		{
			"first day in go",
			&Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "go",
				Year:     2015,
				Day:      1,
				URL:      "",
				Data:     &Data{},
				path:     "",
				runner:   runners.Available["go"]("foo"),
			},
			"Advent of Code 2015, Day 1: Fake Title (Go)",
		},
		{
			"last day in go",
			&Exercise{
				ID:       "2015-25",
				Title:    "Fake Title",
				Language: "py",
				Year:     2015,
				Day:      25,
				URL:      "",
				Data:     &Data{},
				path:     "",
				runner:   runners.Available["py"]("foo"),
			},
			"Advent of Code 2015, Day 25: Fake Title (Python)",
		},
		{
			"invalid language",
			&Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "foo",
				Year:     2015,
				Day:      1,
				URL:      "",
				Data:     &Data{},
				path:     "",
				runner:   nil,
			},
			"Advent of Code 2015, Day 1: Fake Title (INVALID LANGUAGE)",
		},
		{
			"empty exercise",
			&Exercise{},
			"Advent of Code: INVALID EXERCISE",
		},
		{
			"nil exercise",
			nil,
			"Advent of Code: INVALID EXERCISE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.e.String())
		})
	}
}

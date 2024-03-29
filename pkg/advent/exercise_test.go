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
				Path:     "",
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
				Path:     "",
				runner:   runners.Available["py"]("foo"),
			},
			"Advent of Code 2015, Day 25: Fake Title (Python)",
		},
		{
			"invalid language",
			&Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "fake",
				Year:     2015,
				Day:      1,
				URL:      "",
				Data:     &Data{},
				Path:     "",
				runner:   nil,
			},
			"Advent of Code 2015, Day 1: Fake Title (?)",
		},
		{
			"empty exercise",
			&Exercise{},
			"INVALID EXERCISE",
		},
		{
			"no year",
			&Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "fake",
				Year:     0,
				Day:      1,
				URL:      "",
				Data:     &Data{},
				Path:     "",
				runner:   nil,
			},
			"Advent of Code 0, Day 1: Fake Title (?)",
		},
		{
			"no day",
			&Exercise{
				ID:       "2015-01",
				Title:    "Fake Title",
				Language: "fake",
				Year:     2015,
				Day:      0,
				URL:      "",
				Data:     &Data{},
				Path:     "",
				runner:   nil,
			},
			"Advent of Code 2015, Day 0: Fake Title (?)",
		},
		{
			"no title",
			&Exercise{
				ID:       "2015-01",
				Title:    "",
				Language: "fake",
				Year:     2015,
				Day:      1,
				URL:      "",
				Data:     &Data{},
				Path:     "",
				runner:   nil,
			},
			"Advent of Code 2015, Day 1:  (?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.e.String())
		})
	}
}

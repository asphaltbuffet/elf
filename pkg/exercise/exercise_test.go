package exercise

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExercise_String(t *testing.T) {
	tests := []struct {
		name string
		e    *Exercise
		want string
	}{
		{
			name: "valid year",
			e: &Exercise{
				Year:  2015,
				Day:   1,
				Title: "Test Exercise",
				Dir:   "01-testExercise",
				Path:  "/fake/path",
			},
			want: "1 - Test Exercise",
		},
		{
			name: "empty title",
			e: &Exercise{
				Year: 2015,
				Day:  1,
				Dir:  "01-testExercise",
				Path: "/fake/path",
			},
			want: "1 - ",
		},
		{
			name: "year 0",
			e: &Exercise{
				Day:   1,
				Title: "Test Exercise",
				Dir:   "01-testExercise",
				Path:  "/fake/path",
			},
			want: "1 - Test Exercise",
		},
		{
			name: "day 0",
			e: &Exercise{
				Year:  2015,
				Title: "Test Exercise",
				Dir:   "01-testExercise",
				Path:  "/fake/path",
			},
			want: "0 - Test Exercise",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.e.String())
		})
	}
}

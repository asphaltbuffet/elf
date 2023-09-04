package aoc

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/exercise"
)

func TestAOCClient_AddExercise(t *testing.T) {
	type args struct {
		year     int
		day      int
		language string
	}

	tests := []struct {
		name      string
		args      args
		want      *exercise.Exercise
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "already exists, return error",
			args: args{2015, 1, "go"},
			want: &exercise.Exercise{
				Day:  1,
				Name: "Test Day One",
				Dir:  filepath.Join("test_exercises", "2015", "01-testDayOne"),
			},
			assertion: assert.Error,
			errText:   "exercise already exists",
		},
		{
			name: "missing go implementation",
			args: args{2019, 10, "go"},
			want: &exercise.Exercise{
				Day:  10,
				Name: "Test Day One",
				Dir:  filepath.Join("test_exercises", "2019", "10-testDayTen"),
			},
			assertion: assert.Error,
		},
		{
			name: "missing py implementation",
			args: args{2016, 1, "py"},
			want: &exercise.Exercise{
				Day:  1,
				Name: "Test Day One",
				Dir:  filepath.Join("test_exercises", "2016", "01-testDayOne"),
			},
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// recreate for each test to keep testing fs clean
			ac := newTestClient(t)

			got, err := ac.AddExercise(tt.args.year, tt.args.day, tt.args.language)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

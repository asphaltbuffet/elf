package aoc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

func TestAOCClient_RunExercise(t *testing.T) {
	type args struct {
		year int
		day  int
		lang string
	}

	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name:      "exercise doesn't exist",
			args:      args{year: 2020, day: 1, lang: "go"},
			assertion: assert.Error,
			errText:   "getting exercise: no such exercise:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestClient(t)

			err := tc.RunExercise(tt.args.year, tt.args.day, tt.args.lang)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			}
		})
	}
}

func Test_makeMainID(t *testing.T) {
	type args struct {
		part runners.Part
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid 1",
			args: args{part: 1},
			want: "main.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeMainID(tt.args.part)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseMainID(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name     string
		args     args
		wantPart runners.Part
	}{
		{
			name:     "valid 1",
			args:     args{id: "main.1"},
			wantPart: 1,
		},
		{
			name:     "valid 2",
			args:     args{id: "main.2"},
			wantPart: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part := parseMainID(tt.args.id)

			assert.Equal(t, tt.wantPart, part)
		})
	}
}

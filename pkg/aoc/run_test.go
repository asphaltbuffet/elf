package aoc

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/exercise"
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

func TestParseMainIDwPanic(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
	}{
		{"negative part", args{id: "main.-1"}},
		{"too big part", args{id: "main.9001"}},
		{"not a number", args{id: "main.foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() { parseMainID(tt.args.id) })
		})
	}
}

func TestGetRunner(t *testing.T) {
	type args struct {
		e    *exercise.Exercise
		lang string
	}

	tests := []struct {
		name      string
		args      args
		want      runners.Runner
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "valid lang",
			args: args{
				e: &exercise.Exercise{
					Year:  2015,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2015", "01-testDayOne"),
				},
				lang: "go",
			},
			want:      runners.Available["go"](filepath.Join("testdata", "2015", "01-testDayOne")),
			assertion: assert.NoError,
		},
		{
			name: "missing lang",
			args: args{
				e: &exercise.Exercise{
					Year:  2015,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2015", "01-testDayOne"),
				},
				lang: "py",
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "implementation path not found:",
		},
		{
			name: "bad path",
			args: args{
				e: &exercise.Exercise{
					Year:  2015,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  "fake_path",
				},
				lang: "py",
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "getting implementations for exercise:",
		},
	}

	appFs = afero.NewMemMapFs()
	require.NoError(t, appFs.MkdirAll(filepath.Join("testdata", "2015", "01-testDayOne", "go"), 0o755))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRunner(tt.args.e, tt.args.lang)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_handleMainResult(t *testing.T) {
	type args struct {
		r *runners.Result
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid part 1",
			args: args{
				r: &runners.Result{
					TaskID:   "main.1",
					Ok:       true,
					Output:   "asdf",
					Duration: 0.00001,
				},
			},
			want: "Part 1: asdf in 10 µs\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bytes.Buffer

			handleMainResult(&got, tt.args.r)

			assert.Equal(t, tt.want, got.String())
		})
	}
}

package advent

import (
	"bytes"
	"log/slog"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromDir(t *testing.T) {
	var b bytes.Buffer
	tlog := slog.New(slog.NewTextHandler(&b, nil))
	slog.SetDefault(tlog)

	type args struct {
		dir  string
		lang string
	}

	tests := []struct {
		name      string
		args      args
		want      *Exercise
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "valid exercise",
			args: args{dir: "../../testdata/exercises/2015/01-fakeTestDayOne", lang: "go"},
			want: &Exercise{
				ID:       "2015-01",
				Title:    "Fake Test Day One",
				Language: "go",
				Year:     2015,
				Day:      1,
				URL:      "https://fake.fk/2015/day/1",
				Data:     &Data{},
			},
			assertion: assert.NoError,
		},
		{
			name:      "missing exercise directory",
			args:      args{dir: "../../testdata/exercises/2016/01-fakeTestDayOne", lang: "go"},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name:      "missing year directory",
			args:      args{dir: "../../testdata/exercises/2017/01-fakeTestDayOne", lang: "go"},
			want:      nil,
			assertion: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromDir(path.Clean(tt.args.dir), tt.args.lang)

			tt.assertion(t, err)

			if err != nil {
				assert.Nil(t, got)
				assert.NotEmpty(t, b.String())
			} else {
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Title, got.Title)
				assert.Equal(t, tt.want.Language, got.Language)
				assert.Equal(t, tt.want.Year, got.Year)
				assert.Equal(t, tt.want.Day, got.Day)
				assert.Equal(t, tt.want.URL, got.URL)
			}
		})
	}
}

func Test_makeExercisePath(t *testing.T) {
	type args struct {
		year  int
		day   int
		title string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "happy path",
			args: args{
				year:  2015,
				day:   1,
				title: "Fake Title Day One",
			},
			want: filepath.Join("exercises", "2015", "01-fakeTitleDayOne"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, makeExercisePath(tt.args.year, tt.args.day, tt.args.title))
		})
	}
}

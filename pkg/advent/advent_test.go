package advent

import (
	"bytes"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWithOpts(t *testing.T) {
	var b bytes.Buffer
	tlog := slog.New(slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelError}))
	slog.SetDefault(tlog)

	type args struct {
		language string
		opts     []func(*Exercise)
	}

	tests := []struct {
		name      string
		args      args
		want      *Exercise
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "valid exercise",
			args: args{
				language: "go",
				opts:     []func(*Exercise){WithDir("../../testdata/exercises/2015/01-fakeTestDayOne")},
			},
			want: &Exercise{
				ID:       "2015-01",
				Title:    "Fake Test Day One",
				Language: "go",
				Year:     2015,
				Day:      1,
				URL:      "https://fake.fk/2015/day/1",
				Data:     &Data{},
			},
			assertion: require.NoError,
		},
		{
			name: "invalid language",
			args: args{
				language: "fake",
				opts:     []func(*Exercise){WithDir("../../testdata/exercises/2015/01-fakeTestDayOne")},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "no opts",
			args: args{
				language: "go",
				opts:     nil,
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "empty language",
			args: args{
				language: "",
				opts:     []func(*Exercise){WithDir("../../testdata/exercises/2016/01-fakeTestDayOne")},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "missing exercise directory",
			args: args{
				language: "go",
				opts:     []func(*Exercise){WithDir("../../testdata/exercises/2016/01-fakeTestDayOne")},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "missing year directory",
			args: args{
				language: "go",
				opts:     []func(*Exercise){WithDir("../../testdata/exercises/2017/01-fakeTestDayOne")},
			},
			want:      nil,
			assertion: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.language, tt.args.opts...)

			tt.assertion(t, err)

			if err != nil {
				assert.Nil(t, got)
				assert.NotEmpty(t, b.String(), "expected log output at ERROR level")
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

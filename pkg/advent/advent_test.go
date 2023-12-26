package advent

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewWithOpts(t *testing.T) {
	var b bytes.Buffer
	tlog := slog.New(slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelError}))
	slog.SetDefault(tlog)

	type args struct {
		opts []func(*Exercise)
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
				opts: []func(*Exercise){
					WithDir("../../testdata/exercises/2015/01-fakeTestDayOne"),
					WithLanguage("go"),
				},
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
				opts: []func(*Exercise){
					WithDir("../../testdata/exercises/2015/01-fakeTestDayOne"),
					WithLanguage("fake"),
				},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "no opts",
			args: args{
				opts: nil,
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "empty language",
			args: args{
				opts: []func(*Exercise){
					WithDir("../../testdata/exercises/2016/01-fakeTestDayOne"),
					WithLanguage(""),
				},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "missing exercise directory",
			args: args{
				opts: []func(*Exercise){
					WithDir("../../testdata/exercises/2016/01-fakeTestDayOne"),
					WithLanguage("go"),
				},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "missing year directory",
			args: args{
				opts: []func(*Exercise){
					WithDir("../../testdata/exercises/2017/01-fakeTestDayOne"),
					WithLanguage("go"),
				},
			},
			want:      nil,
			assertion: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.opts...)

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

func Test_GetImplementations(t *testing.T) {
	type args struct {
		e *Exercise
	}

	tests := []struct {
		name      string
		args      args
		want      []string
		assertion assert.ErrorAssertionFunc
		wantErr   error
	}{
		{
			name: "two languages",
			args: args{
				&Exercise{
					Path: filepath.Join("testdata", "fs", "2017", "01-fakeFullDay"),
				},
			},
			want:      []string{"go", "py"},
			assertion: assert.NoError,
		},
		{
			name: "one language",
			args: args{
				&Exercise{
					Year:  2017,
					Day:   3,
					Title: "Fake Go Day",
					Path:  filepath.Join("testdata", "fs", "2017", "03-fakeGoDay"),
				},
			},
			want:      []string{"go"},
			assertion: assert.NoError,
		},
		{
			name: "only invalid language",
			args: args{
				&Exercise{
					Year:  2017,
					Day:   4,
					Title: "Fake Lang Day",
					Path:  filepath.Join("testdata", "fs", "2017", "04-fakeLangDay"),
				},
			},
			want:      []string{},
			assertion: assert.NoError,
		},
		{
			name: "valid and invalid languages",
			args: args{
				&Exercise{
					Year:  2017,
					Day:   5,
					Title: "Fake Partial Day",
					Path:  filepath.Join("testdata", "fs", "2017", "05-fakePartialDay"),
				},
			},
			want:      []string{"go"},
			assertion: assert.NoError,
		},
		{
			name: "no languages",
			args: args{
				&Exercise{
					Year:  2017,
					Day:   2,
					Title: "Fake Empty Day",
					Path:  filepath.Join("testdata", "fs", "2017", "02-fakeEmptyDay"),
				},
			},
			want:      []string{},
			assertion: assert.NoError,
		},
		{
			name: "no year",
			args: args{
				&Exercise{
					Year:  2014,
					Day:   14,
					Title: "Fake Missing Year",
					Path:  filepath.Join("testdata", "fs", "2014", "14-fakeMissingYear"),
				},
			},
			want:      nil,
			assertion: assert.Error,
			wantErr:   os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.args.e.GetImplementations()

			require.ErrorIs(t, err, tt.wantErr)
			tt.assertion(t, err)
			if err != nil {
				require.ErrorIs(t, err, tt.wantErr)
				// require.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

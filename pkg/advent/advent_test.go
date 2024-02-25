package advent

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/asphaltbuffet/elf/mocks/krampus"
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
					WithDir("exercises/2017/01-fakeFullDay"),
					WithLanguage("py"),
				},
			},
			want: &Exercise{
				ID:       "2017-01",
				Title:    "Fake Full Day",
				Language: "py",
				Year:     2017,
				Day:      1,
				URL:      "https://fake.fk/2017/day/1",
				Data:     &Data{},
			},
			assertion: require.NoError,
		},
		{
			name: "invalid language",
			args: args{
				opts: []func(*Exercise){
					WithDir("exercises/2017/01-fakeFullDay"),
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
					WithDir("exercises/2016/01-fakeTestDayOne"),
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
					WithDir("exercises/2015/01-fakeTestDayOne"),
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
					WithDir("exercises/2016/01-fakeTestDayOne"),
					WithLanguage("go"),
				},
			},
			want:      nil,
			assertion: require.Error,
		},
	}
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up testing
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			// set up mocks
			mockConfig := mocks.NewMockExerciseConfiguration(t)
			// // mockConfig.EXPECT().GetLanguage().Return(tt.args.lang)
			// mockConfig.EXPECT().GetConfigDir().Return("")
			// mockConfig.EXPECT().GetCacheDir().Return("testCache")
			// mockConfig.EXPECT().GetToken().Return("fakeToken")
			mockConfig.EXPECT().GetFs().Return(testFs)
			mockConfig.EXPECT().GetLogger().Return(slog.New(slog.NewTextHandler(os.Stderr, nil)))

			got, err := New(mockConfig, tt.args.opts...)

			tt.assertion(t, err)

			if err != nil {
				assert.Nil(t, got)
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

func Test_GetImplementations(t *testing.T) {
	type args struct {
		e *Exercise
	}

	tests := []struct {
		name      string
		args      args
		want      []string
		assertion require.ErrorAssertionFunc
		wantErr   error
	}{
		{
			name: "two languages",
			args: args{
				&Exercise{
					Path: filepath.Join("exercises", "2017", "01-fakeFullDay"),
				},
			},
			want:    []string{"go", "py"},
			wantErr: nil,
		},
		{
			name: "one language",
			args: args{
				&Exercise{
					Year:  2017,
					Day:   3,
					Title: "Fake Go Day",
					Path:  filepath.Join("exercises", "2017", "03-fakeGoDay"),
				},
			},
			want:    []string{"go"},
			wantErr: nil,
		},
		{
			name: "no languages",
			args: args{
				&Exercise{
					Year:  2017,
					Day:   2,
					Title: "Fake Empty Day",
					Path:  filepath.Join("exercises", "2017", "02-fakeEmptyDay"),
				},
			},
			want:    []string{},
			wantErr: ErrNoImplementations,
		},
		{
			name: "no year",
			args: args{
				&Exercise{
					Year:  2014,
					Day:   14,
					Title: "Fake Missing Year",
					Path:  filepath.Join("exercises", "2014", "14-fakeMissingYear"),
				},
			},
			want:    nil,
			wantErr: ErrNotFound,
		},
	}
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up testing
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			tt.args.e.appFs = testFs

			got, err := tt.args.e.GetImplementations()

			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

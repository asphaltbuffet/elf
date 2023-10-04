package exercise

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testFs afero.Fs

func TestExercise_GetImplementations(t *testing.T) {
	type args struct {
		e *AdventExercise
	}

	tests := []struct {
		name      string
		args      args
		want      []string
		assertion assert.ErrorAssertionFunc
		errText   string
	}{
		{
			name: "two languages",
			args: args{
				&AdventExercise{
					Year:  2015,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2015", "01-testDayOne"),
				},
			},
			want:      []string{"go", "py"},
			assertion: assert.NoError,
			errText:   "",
		},
		{
			name: "one language",
			args: args{
				&AdventExercise{
					Year:  2016,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2016", "01-testDayOne"),
				},
			},
			want:      []string{"go"},
			assertion: assert.NoError,
			errText:   "",
		},
		{
			name: "only invalid language",
			args: args{
				&AdventExercise{
					Year:  2020,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2020", "01-testDayOne"),
				},
			},
			want:      []string{},
			assertion: assert.NoError,
			errText:   "",
		},
		{
			name: "valid and invalid languages",
			args: args{
				&AdventExercise{
					Year:  2021,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2021", "01-testDayOne"),
				},
			},
			want:      []string{"py"},
			assertion: assert.NoError,
			errText:   "",
		},
		{
			name: "no languages",
			args: args{
				&AdventExercise{
					Year:  2017,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2017", "01-testDayOne"),
				},
			},
			want:      []string{},
			assertion: assert.NoError,
			errText:   "",
		},
		{
			name: "no exercise",
			args: args{
				&AdventExercise{
					Year:  2018,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2018", "01-testDayOne"),
				},
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "file does not exist",
		},
		{
			name: "no year",
			args: args{
				&AdventExercise{
					Year:  2019,
					Day:   1,
					Title: "Test Day One",
					Dir:   "01-testDayOne",
					Path:  filepath.Join("testdata", "2019", "01-testDayOne"),
				},
			},
			want:      nil,
			assertion: assert.Error,
			errText:   "file does not exist",
		},
	}

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.e.GetImplementations(testFs)

			tt.assertion(t, err)
			if err != nil {
				assert.ErrorContains(t, err, tt.errText)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Helper()

	testFs = afero.NewMemMapFs()
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2015", "01-testDayOne", "go"), 0o750))
	// require.NoError(t, afero.WriteFile(testFs, filepath.Join("testdata", "2015", "01-testDayOne", "go", "exercise.go"), []byte("test go impl"), 0o600))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2015", "01-testDayOne", "py"), 0o750))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2016", "01-testDayOne", "go"), 0o750))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2017", "01-testDayOne"), 0o750))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2018"), 0o750))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2020", "01-testDayOne", "fake"), 0o750))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2021", "01-testDayOne", "py"), 0o750))
	require.NoError(t, testFs.MkdirAll(filepath.Join("testdata", "2021", "01-testDayOne", "fake"), 0o750))

	return func(t *testing.T) {
		t.Helper()
	}
}

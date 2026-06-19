package exercise

import (
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExerciseScaffold_write verifies the scaffold lays a finished Exercise out on disk: the input
// file, info.json, and the language template files. The Exercise arrives fully assembled — the
// scaffold invents no data and performs no fetch.
func TestExerciseScaffold_write(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	tests := []struct {
		name      string
		ex        *Exercise
		assertion require.ErrorAssertionFunc
		wantFiles []string
	}{
		{
			name: "writes input, info, and go template",
			ex: &Exercise{
				ID:       "2015-01",
				Title:    "Not Quite Lisp",
				Language: "go",
				Year:     2015,
				Day:      1,
				Path:     filepath.Join("exercises", "2015", "01-notQuiteLisp"),
				Data:     &Data{InputData: "((()))", InputFileName: "input.txt"},
			},
			assertion: require.NoError,
			wantFiles: []string{"input.txt", "info.json", filepath.Join("go", "exercise.go"), "README.md"},
		},
		{
			name: "unknown language errors",
			ex: &Exercise{
				Language: "ruby",
				Path:     filepath.Join("exercises", "2015", "02-other"),
				Data:     &Data{InputData: "x", InputFileName: "input.txt"},
			},
			assertion: require.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teardownSubTest := setupSubTest(t)
			defer teardownSubTest(t)

			s := &exerciseScaffold{
				fs:            testFs,
				inputFileName: "input.txt",
				overwrites:    &Overwrites{},
				logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			_, err := s.write(tt.ex)
			tt.assertion(t, err)

			for _, f := range tt.wantFiles {
				FileExists(t, testFs, filepath.Join(tt.ex.Path, f))
			}
		})
	}
}

// TestExerciseScaffold_writeInputFile_respectsOverwrite verifies the overwrite guard: an existing
// input file is left untouched unless overwrites.Input is set.
func TestExerciseScaffold_writeInputFile_respectsOverwrite(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	teardownSubTest := setupSubTest(t)
	defer teardownSubTest(t)

	exPath := filepath.Join("exercises", "2015", "01-existing")
	require.NoError(t, testFs.MkdirAll(exPath, 0o750))
	fp := filepath.Join(exPath, "input.txt")
	require.NoError(t, afero.WriteFile(testFs, fp, []byte("original"), 0o600))

	ex := &Exercise{Path: exPath, Data: &Data{InputData: "replacement"}}

	s := &exerciseScaffold{
		fs:            testFs,
		inputFileName: "input.txt",
		overwrites:    &Overwrites{Input: false},
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	outcome, err := s.writeInputFile(ex)
	require.NoError(t, err)
	assert.Equal(t, Skipped, outcome, "existing input with overwrite off must report Skipped")

	got, err := afero.ReadFile(testFs, fp)
	require.NoError(t, err)
	assert.Equal(t, "original", string(got), "existing input must be preserved when overwrite is off")
}

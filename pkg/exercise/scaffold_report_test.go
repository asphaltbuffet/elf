package exercise

import (
	"bytes"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutcome_String(t *testing.T) {
	tests := []struct {
		outcome Outcome
		want    string
	}{
		{Added, "added"},
		{Skipped, "skipped"},
		{Replaced, "replaced"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.outcome.String())
		})
	}
}

// TestExerciseScaffold_writeInputFile_outcomes verifies the input writer reports the action it took:
// Added when absent, Replaced when present and overwrite is on.
func TestExerciseScaffold_writeInputFile_outcomes(t *testing.T) {
	tests := []struct {
		name      string
		seed      bool
		overwrite bool
		want      Outcome
	}{
		{name: "absent file is added", seed: false, overwrite: false, want: Added},
		{name: "present file with overwrite is replaced", seed: true, overwrite: true, want: Replaced},
		{name: "present file without overwrite is skipped", seed: true, overwrite: false, want: Skipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			exPath := filepath.Join("exercises", "2015", "01-test")
			require.NoError(t, fs.MkdirAll(exPath, dirPerm))

			fp := filepath.Join(exPath, "input.txt")
			if tt.seed {
				require.NoError(t, afero.WriteFile(fs, fp, []byte("old"), 0o600))
			}

			s := &exerciseScaffold{
				fs:            fs,
				inputFileName: "input.txt",
				overwrites:    &Overwrites{Input: tt.overwrite},
				logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			outcome, err := s.writeInputFile(&Exercise{Path: exPath, Data: &Data{InputData: "new"}})
			require.NoError(t, err)
			assert.Equal(t, tt.want, outcome)
		})
	}
}

// TestExerciseScaffold_write_report verifies write returns a per-file report with relative paths in
// processing order, distinguishing added from skipped files.
func TestExerciseScaffold_write_report(t *testing.T) {
	fs := afero.NewMemMapFs()
	exPath := filepath.Join("exercises", "2015", "01-test")
	require.NoError(t, fs.MkdirAll(exPath, dirPerm))

	// Pre-seed README so it reports Skipped while everything else is Added.
	require.NoError(t, afero.WriteFile(fs, filepath.Join(exPath, "README.md"), []byte("hi"), 0o600))

	s := &exerciseScaffold{
		fs:            fs,
		inputFileName: "input.txt",
		overwrites:    &Overwrites{},
		logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	report, err := s.write(&Exercise{Path: exPath, Language: "go", Data: &Data{InputData: "x"}})
	require.NoError(t, err)

	want := Report{
		{Path: "input.txt", Outcome: Added},
		{Path: "info.json", Outcome: Added},
		{Path: "README.md", Outcome: Skipped},
		{Path: filepath.Join("go", "exercise.go"), Outcome: Added},
	}
	assert.Equal(t, want, report)
}

func TestRenderReport(t *testing.T) {
	buf := &bytes.Buffer{}

	RenderReport(buf, Report{
		{Path: "input.txt", Outcome: Added},
		{Path: "info.json", Outcome: Skipped},
		{Path: filepath.Join("go", "exercise.go"), Outcome: Replaced},
	})

	out := buf.String()
	assert.Contains(t, out, "input.txt")
	assert.Contains(t, out, "added")
	assert.Contains(t, out, "info.json")
	assert.Contains(t, out, "skipped")
	assert.Contains(t, out, filepath.Join("go", "exercise.go"))
	assert.Contains(t, out, "replaced")
}

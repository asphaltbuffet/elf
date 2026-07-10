package exercise

import (
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/protocol"
)

func TestExercise_declaredParts(t *testing.T) {
	t.Run("problem declares only part one", func(t *testing.T) {
		ex := &Exercise{Kind: KindProblem}
		assert.Equal(t, []protocol.Part{protocol.PartOne}, ex.declaredParts())
	})

	t.Run("puzzle declares part one and two", func(t *testing.T) {
		ex := &Exercise{Kind: KindPuzzle}
		assert.Equal(t, []protocol.Part{protocol.PartOne, protocol.PartTwo}, ex.declaredParts())
	})

	t.Run("empty kind defaults to two parts", func(t *testing.T) {
		ex := &Exercise{}
		assert.Equal(t, []protocol.Part{protocol.PartOne, protocol.PartTwo}, ex.declaredParts())
	})
}

func TestMakeProblemID(t *testing.T) {
	assert.Equal(t, "euler-42", makeProblemID(42))
	assert.Equal(t, "euler-100", makeProblemID(100))
	assert.Equal(t, "euler-7", makeProblemID(7))
}

func TestNewProblemFromSource(t *testing.T) {
	ex := newProblemFromSource(problemSource{
		baseDir:  "/work",
		language: "go",
		title:    "Multiples of 3 or 5",
		number:   1,
	})

	assert.Equal(t, KindProblem, ex.Kind)
	assert.Equal(t, 1, ex.Number)
	assert.Equal(t, "euler-1", ex.ID)
	assert.Equal(t, "go", ex.Language)
	assert.Equal(t, filepath.Join("/work", "1"), ex.Path)
	assert.Empty(t, ex.Data.InputData)
	assert.Empty(t, ex.Data.InputFileName)
	assert.Nil(t, ex.Data.TestCases.Two)
	assert.Zero(t, ex.Year)
	assert.Zero(t, ex.Day)
	assert.Empty(t, ex.URL)
}

func TestProblemAdder_Add(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg, err := config.NewConfig(config.WithFs(fs))
	require.NoError(t, err)

	adder, err := NewProblemAdder(cfg,
		WithProblemNumber(42),
		WithProblemLanguage("go"),
		WithProblemTitle("Test Problem"),
	)
	require.NoError(t, err)

	require.NoError(t, adder.Add())

	assert.Contains(t, adder.FilePath(), filepath.Join("euler", "42"))
	assert.NotEmpty(t, adder.Report())

	// info.json exists and reloads as a Problem
	ex := &Exercise{Path: adder.FilePath()}
	require.NoError(t, ex.loadInfo(fs, slog.New(slog.NewTextHandler(io.Discard, nil))))
	assert.Equal(t, KindProblem, ex.Kind)
	assert.Equal(t, 42, ex.Number)
}

func TestNewProblemAdder_ValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		opts         []func(*ProblemAdder)
		wantErr      error
		emptyCfgLang bool
	}{
		{
			name: "empty language",
			opts: []func(*ProblemAdder){
				WithProblemNumber(42),
				WithProblemTitle("Test Problem"),
			},
			wantErr:      ErrEmptyLanguage,
			emptyCfgLang: true,
		},
		{
			name: "zero number",
			opts: []func(*ProblemAdder){
				WithProblemLanguage("go"),
				WithProblemNumber(0),
				WithProblemTitle("Test Problem"),
			},
			wantErr: ErrInvalidData,
		},
		{
			name: "negative number",
			opts: []func(*ProblemAdder){
				WithProblemLanguage("go"),
				WithProblemNumber(-1),
				WithProblemTitle("Test Problem"),
			},
			wantErr: ErrInvalidData,
		},
		{
			name: "empty title",
			opts: []func(*ProblemAdder){
				WithProblemLanguage("go"),
				WithProblemNumber(42),
			},
			wantErr: ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
			require.NoError(t, err)

			if tt.emptyCfgLang {
				// Drive the language default empty so the "no language" guard is
				// reachable: omitting the -l flag alone now falls back to the
				// config default (see ADR-0018 language fallback), so the guard
				// only fires when the config default is itself empty.
				cfg.SetValue(config.LanguageKey, "")
			}

			adder, err := NewProblemAdder(cfg, tt.opts...)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, adder)
		})
	}
}

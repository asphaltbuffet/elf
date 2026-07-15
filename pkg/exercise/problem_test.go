package exercise

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><h2>Test Problem</h2></body></html>`))
	}))
	defer srv.Close()

	fs := afero.NewMemMapFs()
	cfg, err := config.NewConfig(config.WithFs(fs))
	require.NoError(t, err)

	adder, err := NewProblemAdder(cfg,
		WithProblemNumber(42),
		WithProblemLanguage("go"),
		WithProblemFetcher(newTestProblemFetcher(srv.URL)),
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

func TestProblemAdder_Add_fetchesTitle(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><h2>Coded Triangle Numbers</h2></body></html>`))
	}))
	defer srv.Close()

	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	p, err := NewProblemAdder(cfg,
		WithProblemNumber(42), WithProblemLanguage("go"),
		WithProblemFetcher(newTestProblemFetcher(srv.URL)))
	require.NoError(t, err)

	require.NoError(t, p.Add())
	assert.False(t, p.TitlePlaceholdered())

	ex := &Exercise{Path: p.FilePath()}
	require.NoError(t, ex.loadInfo(cfg.GetFs(), cfg.GetLogger()))
	assert.Equal(t, "Coded Triangle Numbers", ex.Title)
}

func TestProblemAdder_Add_placeholderOnTransientFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	p, err := NewProblemAdder(cfg,
		WithProblemNumber(42), WithProblemLanguage("go"),
		WithProblemFetcher(newTestProblemFetcher(srv.URL)))
	require.NoError(t, err)

	require.NoError(t, p.Add()) // degrades, does not fail
	assert.True(t, p.TitlePlaceholdered())

	ex := &Exercise{Path: p.FilePath()}
	require.NoError(t, ex.loadInfo(cfg.GetFs(), cfg.GetLogger()))
	assert.Equal(t, placeholderTitle, ex.Title)
}

func TestProblemAdder_Add_badNumberHardFails(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><p>no problem here</p></body></html>`))
	}))
	defer srv.Close()

	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	p, err := NewProblemAdder(cfg,
		WithProblemNumber(99999), WithProblemLanguage("go"),
		WithProblemFetcher(newTestProblemFetcher(srv.URL)))
	require.NoError(t, err)

	err = p.Add()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidData)

	// nothing scaffolded: the exercise dir must not exist. Add() returns before
	// setting p.path, so check the directory it would have used.
	wantPath := filepath.Join(cfg.GetEulerDir(), "99999")
	ok, statErr := afero.DirExists(cfg.GetFs(), wantPath)
	require.NoError(t, statErr)
	assert.False(t, ok)
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
			},
			wantErr:      ErrEmptyLanguage,
			emptyCfgLang: true,
		},
		{
			name: "zero number",
			opts: []func(*ProblemAdder){
				WithProblemLanguage("go"),
				WithProblemNumber(0),
			},
			wantErr: ErrInvalidData,
		},
		{
			name: "negative number",
			opts: []func(*ProblemAdder){
				WithProblemLanguage("go"),
				WithProblemNumber(-1),
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

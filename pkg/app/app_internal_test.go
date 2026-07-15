package app

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// TestApp_Add_ErrNotConfigured verifies App.Add surfaces the Adder's construction
// error when required configuration (e.g. the advent token) is missing.
func TestApp_Add_ErrNotConfigured(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := New(cfg)

	_, _, err = a.Add("https://adventofcode.com/2015/day/1", "go", nil)
	require.Error(t, err)
}

// TestApp_AddProblem_wiring verifies App.AddProblem wires exercise.NewProblemAdder
// with the derived-title signature. The happy path (title fetch) is covered
// deterministically in pkg/exercise; here we assert the wiring via a
// network-free constructor-error path.
//
// An empty lang argument does NOT reach NewProblemAdder as an empty language:
// exercise.WithProblemLanguage("") is a documented no-op (it only overrides
// when non-empty), so ProblemAdder falls back to cfg.GetLanguage(), whose
// default is "go" (see pkg/config/defaults.go) — never empty in this path. So
// exercising ErrEmptyLanguage from pkg/app is not reachable through AddProblem's
// public arguments; that guard is covered directly in pkg/exercise. Instead we
// drive number<=0, which NewProblemAdder rejects with ErrInvalidData before any
// fetch — equally deterministic and network-free, and still proves AddProblem
// propagates a construction error under the new arity.
func TestApp_AddProblem_wiring(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := New(cfg)

	report, path, placeholdered, err := a.AddProblem(0, "go") // invalid problem number

	require.Error(t, err)
	require.ErrorIs(t, err, exercise.ErrInvalidData)
	assert.Nil(t, report)
	assert.Empty(t, path)
	assert.False(t, placeholdered)
}

// TestApp_Analyze_ErrNoData verifies App.Analyze surfaces the Analyzer's error
// when the target directory has no benchmark data to graph.
func TestApp_Analyze_ErrNoData(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := New(cfg)

	_, err = a.Analyze("does/not/exist", "")
	require.Error(t, err)
}

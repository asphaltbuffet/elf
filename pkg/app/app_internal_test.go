package app

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/pkg/config"
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

// TestApp_AddProblem verifies App.AddProblem wires exercise.NewProblemAdder
// and returns a report and file path for a scaffolded Project Euler problem.
func TestApp_AddProblem(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := New(cfg)

	report, path, err := a.AddProblem(42, "go", "Test Problem")

	require.NoError(t, err)
	assert.Contains(t, path, filepath.Join("euler", "42"))
	assert.NotEmpty(t, report)
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

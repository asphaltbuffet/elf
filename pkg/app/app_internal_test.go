package app

import (
	"testing"

	"github.com/spf13/afero"
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

// TestApp_Analyze_ErrNoData verifies App.Analyze surfaces the Analyzer's error
// when the target directory has no benchmark data to graph.
func TestApp_Analyze_ErrNoData(t *testing.T) {
	cfg, err := config.NewConfig(config.WithFs(afero.NewMemMapFs()))
	require.NoError(t, err)

	a := New(cfg)

	err = a.Analyze("does/not/exist", "")
	require.Error(t, err)
}

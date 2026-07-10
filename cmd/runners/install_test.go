package runners_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/asphaltbuffet/elf/cmd/runners"
)

func TestGetRunnersCmd(t *testing.T) {
	runners.ResetForTest()
	cmd := runners.GetRunnersCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "runners", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestGetRunnersCmd_HasInstallSubcommand(t *testing.T) {
	runners.ResetForTest()
	cmd := runners.GetRunnersCmd()
	sub, _, err := cmd.Find([]string{"install"})
	require.NoError(t, err)
	assert.Equal(t, "install", sub.Use)
}

func TestGetRunnersCmd_HasListSubcommand(t *testing.T) {
	runners.ResetForTest()
	cmd := runners.GetRunnersCmd()
	sub, _, err := cmd.Find([]string{"list"})
	require.NoError(t, err)
	assert.Equal(t, "list", sub.Use)
}

func TestRunInstallCmd_WritesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir) // on Linux, UserConfigDir uses this

	// Reset command singleton
	runners.ResetForTest()

	cmd := runners.GetRunnersCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	installCmd, _, findErr := cmd.Find([]string{"install"})
	require.NoError(t, findErr)
	installCmd.SetOut(&out)

	err := installCmd.RunE(installCmd, nil)
	require.NoError(t, err)

	runnersDir := filepath.Join(tmpDir, "elf", "runners")
	assert.FileExists(t, filepath.Join(runnersDir, "python.tmpl"))
	assert.FileExists(t, filepath.Join(runnersDir, "go.tmpl"))
	assert.FileExists(t, filepath.Join(runnersDir, "bash.tmpl"))
	assert.FileExists(t, filepath.Join(runnersDir, "f77.tmpl"))
	assert.FileExists(t, filepath.Join(runnersDir, "lua.tmpl"))
	assert.FileExists(t, filepath.Join(runnersDir, "csharp.tmpl"))
	assert.Contains(t, out.String(), "[[runner]]")
	assert.Contains(t, out.String(), "YOUR_MODULE")
	assert.Contains(t, out.String(), "gfortran")
	assert.Contains(t, out.String(), "dkjson")
	assert.Contains(t, out.String(), "dotnet")
}

// TestRunInstallCmd_GoBlockIsBuildable guards the fix for #178: the shipped Go
// runner block must set wrapper_ext and wrapper_subdir so the rendered
// package-main harness gets a .go extension and its own subdir (otherwise it
// clashes with the solution package in the same lang dir and go build fails).
func TestRunInstallCmd_GoBlockIsBuildable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	runners.ResetForTest()

	cmd := runners.GetRunnersCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	installCmd, _, findErr := cmd.Find([]string{"install"})
	require.NoError(t, findErr)
	installCmd.SetOut(&out)

	require.NoError(t, installCmd.RunE(installCmd, nil))

	cfg := out.String()
	assert.Contains(t, cfg, `wrapper_ext = ".go"`,
		"go block must set wrapper_ext so the harness is written as .go source")
	assert.Contains(t, cfg, `wrapper_subdir = "cmd"`,
		"go block must set wrapper_subdir so package main does not clash with the solution package")
}

func TestRunInstallCmd_SkipsExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Pre-create the runners dir and one file
	runnersDir := filepath.Join(tmpDir, "elf", "runners")
	require.NoError(t, os.MkdirAll(runnersDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(runnersDir, "python.tmpl"), []byte("existing"), 0o644))

	runners.ResetForTest()
	cmd := runners.GetRunnersCmd()
	var out bytes.Buffer

	installCmd, _, findErr := cmd.Find([]string{"install"})
	require.NoError(t, findErr)
	installCmd.SetOut(&out)

	err := installCmd.RunE(installCmd, nil)
	require.NoError(t, err)

	// python.tmpl should be skipped (still has old content)
	content, readErr := os.ReadFile(filepath.Join(runnersDir, "python.tmpl"))
	require.NoError(t, readErr)
	assert.Equal(t, "existing", string(content))

	assert.Contains(t, out.String(), "Skipping")

	// go.tmpl did not pre-exist, so it should have been written
	assert.FileExists(t, filepath.Join(runnersDir, "go.tmpl"))
}

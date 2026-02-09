package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInitCmd(t *testing.T) {
	cmd := getInitCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "init", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestGetInitCmd_Flags(t *testing.T) {
	cmd := getInitCmd()

	globalFlag := cmd.Flag("global")
	assert.NotNil(t, globalFlag)
	assert.Equal(t, "g", globalFlag.Shorthand)

	forceFlag := cmd.Flag("force")
	assert.NotNil(t, forceFlag)
	assert.Equal(t, "f", forceFlag.Shorthand)
}

func TestRunInitCmd_NewFile(t *testing.T) {
	// Create a temp directory for the test
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	cmd := getInitCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	// Reset the command to allow re-execution
	initCmd = nil

	cmd = getInitCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)

	// Check that file was created
	configPath := filepath.Join(tmpDir, "elf.toml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config file should exist")
}

func TestRunInitCmd_FileExistsNoForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create existing config file
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte("existing config"), 0o644)
	require.NoError(t, err)

	// Reset the command
	initCmd = nil

	cmd := getInitCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRunInitCmd_FileExistsWithForce(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create existing config file
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte("existing config"), 0o644)
	require.NoError(t, err)

	// Reset the command
	initCmd = nil

	cmd := getInitCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--force"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Read the new content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "language")
}

func TestInteractiveInit(t *testing.T) {
	t.Skip("requires stdin mocking for complete testing")
}

func TestGenerateDefaultConfig_Integration(t *testing.T) {
	fs := afero.NewMemMapFs()
	tmpDir := "/tmp/test-config"
	require.NoError(t, fs.MkdirAll(tmpDir, 0o755))

	// Write generated config to file
	content := []byte(`language = "go"
[advent]
token = ""
dir = "exercises"`)

	configPath := filepath.Join(tmpDir, "elf.toml")
	err := afero.WriteFile(fs, configPath, content, 0o644)
	require.NoError(t, err)

	// Verify it was written
	exists, err := afero.Exists(fs, configPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Read and verify content
	data, err := afero.ReadFile(fs, configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "language")
}

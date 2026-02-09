package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUpdateTokenCmd(t *testing.T) {
	cmd := getUpdateTokenCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "update-token", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestGetUpdateTokenCmd_Flags(t *testing.T) {
	cmd := getUpdateTokenCmd()

	tokenFlag := cmd.Flag("token")
	assert.NotNil(t, tokenFlag)
	assert.Equal(t, "t", tokenFlag.Shorthand)

	configFileFlag := cmd.Flag("config-file")
	assert.NotNil(t, configFileFlag)
	assert.Equal(t, "c", configFileFlag.Shorthand)
}

func TestRunUpdateTokenCmd_WithTokenFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create initial config
	configContent := `language = "go"

[advent]
token = "old-token"
dir = "exercises"
`
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Reset command
	updateTokenCmd = nil

	cmd := getUpdateTokenCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--token", "new-updated-token"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify the file was updated
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "new-updated-token")

	// Verify output message
	assert.Contains(t, out.String(), "Token updated")
}

func TestRunUpdateTokenCmd_UpdatesExistingToken(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create config with existing token
	configContent := `language = "go"

[advent]
token = "original-token"
dir = "exercises"
`
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Reset command
	updateTokenCmd = nil

	cmd := getUpdateTokenCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--token", "replaced-token"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify the file was updated
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "replaced-token")
	assert.NotContains(t, string(content), "original-token")
}

func TestRunUpdateTokenCmd_PreservesOtherSettings(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create config with multiple settings
	configContent := `language = "python"

[advent]
token = "old-token"
dir = "my-exercises"
`
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Reset command
	updateTokenCmd = nil

	cmd := getUpdateTokenCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--token", "new-token"})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify other settings are preserved
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "python")
	assert.Contains(t, string(content), "my-exercises")
}

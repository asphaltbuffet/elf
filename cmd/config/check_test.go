package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCheckCmd(t *testing.T) {
	cmd := getCheckCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "check", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestGetCheckCmd_Flags(t *testing.T) {
	cmd := getCheckCmd()

	configFileFlag := cmd.Flag("config-file")
	assert.NotNil(t, configFileFlag)
	assert.Equal(t, "c", configFileFlag.Shorthand)
}

func TestRunCheckCmd_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create a config file
	configContent := `language = "python"

[advent]
token = "test-token-value-here"
dir = "my-exercises"
`
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Reset command
	checkCmd = nil

	cmd := getCheckCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Configuration Status")
	assert.Contains(t, output, "python")
	assert.Contains(t, output, "my-exercises")
	// Token should be masked
	assert.Contains(t, output, "test...here")
}

func TestRunCheckCmd_ValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create config with placeholder token
	configContent := `language = "go"

[advent]
token = "default-placeholder"
dir = "exercises"
`
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Reset command
	checkCmd = nil

	cmd := getCheckCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Validation")
	// Should show validation error for placeholder token
	assert.Contains(t, output, "placeholder")
}

func TestRunCheckCmd_ShowsSettings(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create a config file
	configContent := `language = "rust"

[advent]
token = "my-special-token123"
dir = "special-exercises"
`
	configPath := filepath.Join(tmpDir, "elf.toml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	// Reset command
	checkCmd = nil

	cmd := getCheckCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "rust")
	assert.Contains(t, output, "special-exercises")
	assert.Contains(t, output, "Current Settings")
}

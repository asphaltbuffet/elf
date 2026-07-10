package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const dirPerm = 0o755

// WriteConfig writes the current configuration to the specified path.
// If path is empty, it writes to the default config directory.
// Creates parent directories as needed.
func (c *Config) WriteConfig(path string) error {
	if path == "" {
		path = filepath.Join(c.GetConfigDir(), DefaultConfigFileBase+"."+DefaultConfigExt)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := c.fs.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	return c.viper.WriteConfigAs(path)
}

// SafeWriteConfig writes the configuration only if the file doesn't exist.
// Returns an error if the file already exists.
func (c *Config) SafeWriteConfig(path string) error {
	if path == "" {
		path = filepath.Join(c.GetConfigDir(), DefaultConfigFileBase+"."+DefaultConfigExt)
	}

	// Check if file exists
	if _, err := c.fs.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	return c.WriteConfig(path)
}

// SetToken updates the advent token value in the configuration.
func (c *Config) SetToken(token string) {
	c.viper.Set(string(AdventTokenKey), token)
}

// SetValue sets a configuration value by key.
func (c *Config) SetValue(key Key, value any) {
	c.viper.Set(string(key), value)
}

// GetAllSettings returns all configuration settings as a map.
func (c Config) GetAllSettings() map[string]any {
	return c.viper.AllSettings()
}

// GenerateDefaultConfig returns a default TOML configuration template string.
func GenerateDefaultConfig() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "~/.config/elf"
	} else {
		configDir = filepath.Join(configDir, "elf")
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = "~/.cache/elf"
	} else {
		cacheDir = filepath.Join(cacheDir, "elf")
	}

	return fmt.Sprintf(`# elf configuration file
# See: https://github.com/asphaltbuffet/elf

# Default implementation language for solutions
language = %q

# Directory for configuration files
config-dir = %q

# Directory for cached data
cache-dir = %q

# Default input file name
input-file = %q

[advent]
# Advent of Code authentication token
# Get this from your browser cookies at adventofcode.com (session cookie)
token = ""

# Directory for Advent of Code exercises
dir = %q

[euler]
# Directory for Project Euler problems (sibling of the AoC exercise directory)
dir = %q
`, defaults[LanguageKey], configDir, cacheDir, defaults[InputFileKey], defaults[AdventDirKey], defaults[EulerDirKey])
}

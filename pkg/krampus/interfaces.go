package krampus

import (
	"log/slog"

	"github.com/spf13/afero"
)

// ConfigurationReader is an interface that defines methods for reading configuration.
type ConfigurationReader interface {
	// GetConfigFileUsed returns the configuration file used.
	GetConfigFileUsed() string
}

// ExerciseConfiguration represents the interface for exercise configuration.
type ExerciseConfiguration interface {
	// GetBaseDir returns the base directory.
	GetBaseDir() string

	// GetFs returns the file system.
	GetFs() afero.Fs

	// GetLanguage returns the language.
	GetLanguage() string

	// GetLogger returns the logger.
	GetLogger() *slog.Logger
}

// DownloadConfiguration is an interface that extends the ExerciseConfiguration interface.
// It represents the configuration for downloading exercises.
type DownloadConfiguration interface {
	ExerciseConfiguration

	// GetCacheDir returns the directory where downloaded exercises are cached.
	GetCacheDir() string

	// GetConfigDir returns the configuration directory.
	GetConfigDir() string

	// GetInputFilename returns the input filename.
	GetInputFilename() string

	// GetToken returns the authentication token for downloading exercises.
	GetToken() string
}

// ConfigurationWriter is an interface for writing configuration files.
type ConfigurationWriter interface {
	// WriteConfig writes the configuration to the specified path.
	WriteConfig(path string) error

	// SafeWriteConfig writes the configuration only if the file doesn't exist.
	SafeWriteConfig(path string) error

	// SetToken updates the advent token value.
	SetToken(token string)

	// SetValue sets a configuration value by key.
	SetValue(key ConfigKey, value any)

	// GetAllSettings returns all configuration settings as a map.
	GetAllSettings() map[string]any
}

// ConfigurationValidator is an interface for validating configuration.
type ConfigurationValidator interface {
	// ValidateToken checks if the advent token is properly configured.
	ValidateToken() error

	// ValidateDirectories checks if configured directories exist.
	ValidateDirectories() []error

	// Validate performs full validation of the configuration.
	Validate() []error
}

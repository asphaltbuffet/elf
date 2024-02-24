package krampus

import (
	"log/slog"

	"github.com/spf13/afero"
)

// ConfigurationReader is an interface that defines methods for reading configuration.
type ConfigurationReader interface {
	// GetConfigFileUsed returns the configuration file used.
	GetConfigFileUsed() string

	// GetFs returns the file system.
	GetFs() afero.Fs
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

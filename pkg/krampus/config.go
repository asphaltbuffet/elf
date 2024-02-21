package krampus

import (
	"log/slog"

	"github.com/spf13/afero"
)

type ConfigurationReader interface {
	GetConfigFileUsed() string
	GetFs() afero.Fs
}

type ExerciseConfiguration interface {
	GetLanguage() string
	GetConfigDir() string
	GetLogger() *slog.Logger
	GetFs() afero.Fs
}

type DownloadConfiguration interface {
	GetToken() string
	GetCacheDir() string
	ExerciseConfiguration
}

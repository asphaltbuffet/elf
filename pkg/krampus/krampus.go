package krampus

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
)

var cfg *viper.Viper

func New() (*viper.Viper, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = viper.New()

	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.StampMilli,
		}),
	))

	cfg.SetEnvPrefix("elf")

	_ = cfg.BindEnv("advent.token", "ELF_ADVENT_TOKEN")
	cfg.SetDefault("advent.token", "")

	cfg.SetDefault("advent.user", "")
	cfg.SetDefault("advent.dir", "exercises")
	cfg.SetDefault("euler.dir", "problems")

	_ = cfg.BindEnv("language", "ELF_LANGUAGE")
	cfg.SetDefault("language", "go")

	configDir, err := os.UserConfigDir()
	if err != nil {
		slog.Error("get default user config dir", "error", tint.Err(err))
		return nil, err
	}
	cfg.SetDefault("config-dir", configDir)

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Error("get default user cache dir", "error", tint.Err(err))
		return nil, err
	}

	cfg.SetDefault("cache-dir", cacheDir)

	cfg.SetConfigName("elf.toml")
	cfg.SetConfigType("toml")

	userCfg, err := os.UserConfigDir()
	if err == nil {
		cfg.AddConfigPath(userCfg)
	}

	cfg.AddConfigPath(".")
	cfg.AddConfigPath("$HOME/.config/elf")

	err = cfg.ReadInConfig()
	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// only return error if it's not a missing config file
			slog.Error("failed to read config file", "error", err, "config", cfg.ConfigFileUsed())
			return nil, err
		}

		slog.Warn("no config file found")
	} else {
		slog.Debug("starting with config file", "config", cfg.ConfigFileUsed())
	}

	return cfg, nil
}

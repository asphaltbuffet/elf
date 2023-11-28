// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// application build information set by the linker.
var (
	Version string
)

var (
	rootCmd *cobra.Command
	cfg     = viper.New()
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := GetRootCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}

// GetRootCommand returns the root command for the CLI.
func GetRootCommand() *cobra.Command {
	if rootCmd == nil {
		rootCmd = &cobra.Command{
			Use:     "elf [command]",
			Version: Version,
			Short:   "elf is a programming challenge helper application",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				appFs := afero.NewOsFs()

				return initialize(appFs)
			},
		}
	}

	rootCmd.AddCommand(GetSolveCmd())
	rootCmd.AddCommand(GetTestCmd())
	rootCmd.AddCommand(GetDownloadCmd())
	// rootCmd.AddCommand(GetBenchmarkCmd())

	return rootCmd
}

func initialize(fs afero.Fs) error {
	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.StampMilli,
		}),
	))

	viper.SetEnvPrefix("elf")

	_ = viper.BindEnv("advent.token", "ELF_ADVENT_TOKEN")
	viper.SetDefault("advent.token", "")

	viper.SetDefault("advent.user", "")
	viper.SetDefault("advent.dir", "exercises")
	viper.SetDefault("euler.dir", "problems")

	_ = viper.BindEnv("language", "ELF_LANGUAGE")
	viper.SetDefault("language", "go")

	viper.SetFs(fs)

	viper.SetConfigName("elf.toml")
	viper.SetConfigType("toml")

	userCfg, err := os.UserConfigDir()
	if err == nil {
		viper.AddConfigPath(userCfg)
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/elf")

	err = viper.ReadInConfig()
	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			// only return error if it's not a missing config file
			slog.Error("failed to read config file", "error", err, "config", cfg.ConfigFileUsed())
			return err
		}

		slog.Warn("no config file found")
	} else {
		slog.Debug("starting elf", "version", Version, "config", viper.ConfigFileUsed())
	}

	return nil
}

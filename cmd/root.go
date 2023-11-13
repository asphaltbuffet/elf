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
			Level:      slog.LevelDebug,
			TimeFormat: time.StampMilli,
		}),
	))

	cfg = viper.New()
	cfg.SetDefault("advent.token", "")
	cfg.SetDefault("advent.user", "")
	cfg.SetDefault("advent.dir", "exercises")
	cfg.SetDefault("euler.dir", "problems")
	cfg.SetDefault("language", "go")

	cfg.SetFs(fs)
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
			return err
		}

		slog.Warn("no config file found")
	} else {
		slog.Info("starting elf", "version", Version, "config", cfg.ConfigFileUsed())
	}

	return nil
}

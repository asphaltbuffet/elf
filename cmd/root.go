// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// application build information set by the linker
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
	cobra.CheckErr(GetRootCommand().Execute())
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
	// rootCmd.AddCommand(GetAddCmd())
	// rootCmd.AddCommand(GetBenchmarkCmd())
	// rootCmd.AddCommand(GetShowCmd())
	// rootCmd.AddCommand(GetInfoCmd())

	return rootCmd
}

func initialize(fs afero.Fs) error {
	fmt.Println("initializing...")

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

	if err := cfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// only return error if it's not a missing config file
			return err
		}
	}

	return nil
}

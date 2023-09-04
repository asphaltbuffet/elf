// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/aoc"
)

// application build information set by the linker
var (
	version string
	commit  string
	date    string
)

var (
	rootCmd *cobra.Command
	yearArg string
	dayArg  int
	langArg string

	acc *aoc.AOCClient
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v, d string) {
	version, date = v, d

	cobra.CheckErr(GetRootCommand().Execute())
}

// GetRootCommand returns the root command for the CLI.
func GetRootCommand() *cobra.Command {
	if rootCmd == nil {
		rootCmd = &cobra.Command{
			Use:               "elf [command]",
			Version:           fmt.Sprintf("%s (%s) %s", version, commit, date),
			Short:             "elf is an Advent of Code helper application",
			Long:              `TODO: add a long description`,
			PersistentPreRunE: initialize,
		}
	}

	// TODO: should these be flags or positional args?
	rootCmd.PersistentFlags().StringVarP(&yearArg, "year", "y", "", "exercise year")
	rootCmd.PersistentFlags().IntVarP(&dayArg, "day", "d", 0, "exercise day")
	rootCmd.PersistentFlags().StringVarP(&langArg, "lang", "L", "", "implementation language")

	rootCmd.AddCommand(GetAddCmd())
	rootCmd.AddCommand(GetBenchmarkCmd())
	rootCmd.AddCommand(GetGraphCmd())
	// TODO: add init command
	rootCmd.AddCommand(GetRunCmd())
	rootCmd.AddCommand(GetVisualizeCmd())

	return rootCmd
}

func initialize(cmd *cobra.Command, args []string) error {
	// TODO: check if the user has initialized the application, set up defaults/load config
	var err error

	acc, err = aoc.NewAOCClient()
	if err != nil {
		return fmt.Errorf("unable to start the application: %w", err)
	}

	return nil
}

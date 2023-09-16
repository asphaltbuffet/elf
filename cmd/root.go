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
	yearArg int
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

	rootCmd.PersistentFlags().IntVarP(&yearArg, "year", "y", 0, "exercise year")
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
	if !haveValidYearFlag() {
		return fmt.Errorf("invalid year: %d", yearArg)
	}

	if !haveValidDayFlag() {
		return fmt.Errorf("invalid day: %d", dayArg)
	}

	// TODO: check if the user has initialized the application, set up defaults/load config
	var err error

	acc, err = aoc.NewAOCClient()
	if err != nil {
		return fmt.Errorf("unable to start the application: %w", err)
	}

	return nil
}

func haveValidYearFlag() bool {
	return yearArg >= MinYear && yearArg <= MaxYear
}

func haveValidDayFlag() bool {
	return dayArg >= 1 && dayArg <= 25
}

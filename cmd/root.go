package cmd

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/cmd/analyze"
	"github.com/asphaltbuffet/elf/cmd/benchmark"
	"github.com/asphaltbuffet/elf/cmd/config"
	"github.com/asphaltbuffet/elf/cmd/download"
	"github.com/asphaltbuffet/elf/cmd/man"
	runnerspkg "github.com/asphaltbuffet/elf/cmd/runners"
	"github.com/asphaltbuffet/elf/cmd/solve"
	"github.com/asphaltbuffet/elf/cmd/test"
	versioncmd "github.com/asphaltbuffet/elf/cmd/version"
	"github.com/asphaltbuffet/elf/cmd/visualize"
)

var rootCmd *cobra.Command

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// err := GetRootCommand().Execute()
	// if err != nil {
	if err := fang.Execute(context.Background(), GetRootCommand(), fang.WithVersion(versioncmd.Version)); err != nil {
		os.Exit(1)
	}
}

// GetRootCommand returns the root command for the CLI.
func GetRootCommand() *cobra.Command {
	var cfgFile string
	if rootCmd == nil {
		rootCmd = &cobra.Command{
			Use:   "elf [command]",
			Short: "elf is a programming challenge helper application",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cmd.Help()
			},
		}

		rootCmd.Flags().StringVarP(&cfgFile, "config-file", "c", "", "configuration file")

		rootCmd.AddCommand(analyze.GetAnalyzeCmd())
		rootCmd.AddCommand(benchmark.GetBenchmarkCmd())
		rootCmd.AddCommand(config.GetConfigCmd())
		rootCmd.AddCommand(download.GetDownloadCmd())
		rootCmd.AddCommand(man.NewManCmd())
		rootCmd.AddCommand(runnerspkg.GetRunnersCmd())
		rootCmd.AddCommand(solve.GetSolveCmd())
		rootCmd.AddCommand(test.GetTestCmd())
		rootCmd.AddCommand(visualize.GetVisualizeCmd())
	}

	return rootCmd
}

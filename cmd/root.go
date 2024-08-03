// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/cmd/analyze"
	"github.com/asphaltbuffet/elf/cmd/benchmark"
	"github.com/asphaltbuffet/elf/cmd/download"
	"github.com/asphaltbuffet/elf/cmd/export"
	"github.com/asphaltbuffet/elf/cmd/man"
	"github.com/asphaltbuffet/elf/cmd/solve"
	"github.com/asphaltbuffet/elf/cmd/test"
	versionCmd "github.com/asphaltbuffet/elf/cmd/version"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var rootCmd *cobra.Command

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
	var cfgFile string
	var cfg *krampus.Config

	if rootCmd == nil {
		rootCmd = &cobra.Command{
			Use:   "elf [command]",
			Short: "elf is a programming challenge helper application",
			Run: func(cmd *cobra.Command, _ []string) {
				var err error
				cfg, err = krampus.NewConfig(krampus.WithFile(cfgFile))
				if err != nil {
					cmd.PrintErr(err)
				}

				cmd.Println("config file:", cfg.GetConfigFileUsed())
				cmd.Println("language:", cfg.GetLanguage())
				cmd.Println("token:", cfg.GetToken())
			},
		}

		rootCmd.Flags().StringVarP(&cfgFile, "config-file", "c", "", "configuration file")

		rootCmd.AddCommand(analyze.GetAnalyzeCmd())
		rootCmd.AddCommand(benchmark.GetBenchmarkCmd())
		rootCmd.AddCommand(download.GetDownloadCmd())
		rootCmd.AddCommand(export.NewCommand(cfg))
		rootCmd.AddCommand(man.NewManCmd())
		rootCmd.AddCommand(solve.GetSolveCmd())
		rootCmd.AddCommand(test.GetTestCmd())
		rootCmd.AddCommand(versionCmd.NewVersionCmd())
	}

	return rootCmd
}

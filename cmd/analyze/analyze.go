// Package analyze is the analyze subcommand.
package analyze

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	advent "github.com/asphaltbuffet/elf/pkg/advent/analyze"
	"github.com/asphaltbuffet/elf/pkg/analysis"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var (
	analyzeCmd *cobra.Command

	outFile string
)

func GetAnalyzeCmd() *cobra.Command {
	if analyzeCmd == nil {
		analyzeCmd = &cobra.Command{
			Use:     "analyze",
			Aliases: []string{"a", "analyse"},
			Args:    cobra.ExactArgs(1),
			Short:   "analysis of run-time metrics",
			RunE:    runAnalyzeCmd,
			Example: "elf analyze ~/advent/2015/01-exercise/",
		}

		analyzeCmd.Flags().StringVarP(&outFile, "graph", "g", "./run-times.png", "graph output file")
	}

	return analyzeCmd
}

func runAnalyzeCmd(cmd *cobra.Command, args []string) error {
	var aa analysis.Analyzer

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := krampus.NewConfig(krampus.WithFile(cf))
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("analysis dir: %w", err)
	}

	out, err := filepath.Abs(outFile)
	if err != nil {
		return fmt.Errorf("output file: %w", err)
	}

	aa, err = advent.NewAnalyzer(cfg, advent.WithDirectory(dir), advent.WithOutput(out))
	if err != nil {
		return fmt.Errorf("creating grapher: %w", err)
	}

	return aa.Graph()
}

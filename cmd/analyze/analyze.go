// Package analyze is the analyze subcommand.
package analyze

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	analyzer "github.com/asphaltbuffet/elf/pkg/analyze"
	"github.com/asphaltbuffet/elf/pkg/config"
)

// Analyzer is the interface for benchmark analysis.
type Analyzer interface {
	Graph() error
}

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
	var aa Analyzer

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := config.NewConfig(config.WithFile(cf))
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

	aa, err = analyzer.NewAnalyzer(cfg, analyzer.WithDirectory(dir), analyzer.WithOutput(out))
	if err != nil {
		return fmt.Errorf("creating grapher: %w", err)
	}

	return aa.Graph()
}

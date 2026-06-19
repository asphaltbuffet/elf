// Package analyze is the analyze subcommand.
package analyze

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	analyzer "github.com/asphaltbuffet/elf/pkg/analyze"
	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

// Analyzer is the interface for benchmark analysis.
type Analyzer interface {
	Graph() error
}

var (
	analyzeCmd *cobra.Command

	outFile string

	// Factory variables for testing.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
	makeAnalyzer = func(logger *slog.Logger, dir, out string) (Analyzer, error) {
		return analyzer.NewAnalyzer(logger, analyzer.WithDirectory(dir), analyzer.WithOutput(out))
	}
)

// GetAnalyzeCmd returns the cobra command for benchmark analysis.
func GetAnalyzeCmd() *cobra.Command {
	if analyzeCmd == nil {
		analyzeCmd = &cobra.Command{
			Use:     "analyze",
			Aliases: []string{"a", "analyse"},
			Args:    cobra.ExactArgs(1),
			Short:   "graph run-time benchmark data",
			Long: "Graph benchmark data for a target directory. If the target is a single " +
				"exercise, languages are compared in a box plot; if it is a year, days are " +
				"compared in a line graph. The graph is written to <target>/run-times.png " +
				"unless -g is given.",
			RunE:    runAnalyzeCmd,
			Example: "elf analyze ~/advent/2015/01-exercise/\nelf analyze ~/advent/2015/",
		}

		analyzeCmd.Flags().StringVarP(&outFile, "graph", "g", "",
			"graph output path (default: <target>/run-times.png)")
	}

	return analyzeCmd
}

func runAnalyzeCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	appPkg.RegisterRunners(cfg)

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("analysis dir: %w", err)
	}

	out := ""
	if outFile != "" {
		out, err = filepath.Abs(outFile)
		if err != nil {
			return fmt.Errorf("output file: %w", err)
		}
	}

	aa, err := makeAnalyzer(cfg.GetLogger(), dir, out)
	if err != nil {
		return fmt.Errorf("creating grapher: %w", err)
	}

	return aa.Graph()
}

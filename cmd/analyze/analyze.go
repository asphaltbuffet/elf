package analyze

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	advent "github.com/asphaltbuffet/elf/pkg/advent/analyze"
	"github.com/asphaltbuffet/elf/pkg/analysis"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var (
	analyzeCmd *cobra.Command

	outFile   string
	graphType string
	byYear    bool
	byDay     bool
	compare   bool
)

func GetAnalyzeCmd() *cobra.Command {
	if analyzeCmd == nil {
		analyzeCmd = &cobra.Command{
			Use:     "analyze [flags] path/to/directory",
			Aliases: []string{"a", "analyse"},
			Args:    cobra.ExactArgs(1),
			Short:   "analysis of run-time metrics",
			RunE:    runAnalyzeCmd,
		}

		analyzeCmd.Flags().StringVarP(&outFile, "graph", "g", "./run-times.png", "graph output file")
		analyzeCmd.Flags().StringVarP(&graphType, "type", "t", "line", "type of output graph")

		analyzeCmd.Flags().BoolVarP(&byYear, "year", "y", true, "generate analysis by each year")
		analyzeCmd.Flags().BoolVarP(&byDay, "day", "d", false, "generate separate analysis for each day")
		analyzeCmd.Flags().BoolVarP(&compare, "compare", "c", false, "compare run-time metrics")
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
		return fmt.Errorf("output file: %w", err)
	}

	aa, err = advent.NewAnalyzer(cfg, advent.WithDirectory(dir))
	if err != nil {
		return fmt.Errorf("creating grapher: %w", err)
	}

	switch {
	case outFile != "":
		return aa.Graph(analysis.StringToGraphType(graphType))

	case byYear:
		return aa.Stats()

	case byDay:
		return aa.Stats()

	case compare:
		return aa.Stats()

	default:
		return errors.New("no analysis type")
	}
}

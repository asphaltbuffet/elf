package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

var (
	graphCmd *cobra.Command

	outfile string
)

type Grapher interface {
	NewGraph(string) (*Grapher, error)
	Graph(string) error
}

func GetGraphCmd() *cobra.Command {
	if graphCmd == nil {
		graphCmd = &cobra.Command{
			Use:   "graph <path> [-o <output>]",
			Args:  cobra.ExactArgs(1),
			Short: "generate run-time graph",
			RunE:  runGraphCmd,
		}
	}

	graphCmd.Flags().StringVarP(&outfile, "output", "o", "./run-times.png", "file to write output to")

	return graphCmd
}

func runGraphCmd(cmd *cobra.Command, args []string) error {
	cwd, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("getting absolute path for output file: %w", err)
	}

	cmd.Println("searching for benchmark files in", cwd)

	g, err := advent.NewGraph(cwd)
	if err != nil {
		return fmt.Errorf("creating grapher: %w", err)
	}

	return g.Graph(outfile)
}

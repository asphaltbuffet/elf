package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

var (
	benchmarkCmd *cobra.Command
	iterations   int
)

const benchmarkExample = `elf benchmark`

func GetBenchmarkCmd() *cobra.Command {
	if benchmarkCmd == nil {
		benchmarkCmd = &cobra.Command{
			Use:     "benchmark (-n <iterations>)",
			Aliases: []string{"bench", "b"},
			Example: benchmarkExample,
			Args:    cobra.NoArgs,
			Short:   "benchmark all implementations for the challenge",
			RunE:    runBenchmarkCmd,
		}

		benchmarkCmd.Flags().IntVarP(&iterations, "num", "n", 1, "number of iterations")
	}

	return solveCmd
}

type Benchmarker interface {
	Benchmark(int) error
	String() string
}

func runBenchmarkCmd(cmd *cobra.Command, _ []string) error {
	var (
		ex  Benchmarker
		err error
	)

	dir, err := os.Getwd()
	if err != nil {
		slog.Error("getting current directory", tint.Err(err))
		return err
	}

	slog.Debug("benchmarking exercise", slog.Group("exercise", "dir", dir))

	ex, err = advent.New(advent.WithDir(dir))
	if err != nil {
		slog.Error("creating exercise", tint.Err(err))
		return err
	}

	if solveErr := ex.Benchmark(iterations); solveErr != nil {
		slog.Error("solving exercise", tint.Err(solveErr))
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

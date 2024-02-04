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

		benchmarkCmd.Flags().IntVarP(&iterations, "num", "n", 3, "number of iterations")
	}

	return benchmarkCmd
}

type Benchmarker interface {
	Benchmark(int) ([]advent.TaskResult, error)
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

	ex, err = advent.New(advent.WithDir(dir), advent.WithLanguage("go")) // TODO: make language configurable
	if err != nil {
		slog.Error("creating exercise", tint.Err(err))
		return err
	}

	_, err = ex.Benchmark(iterations)
	if err != nil {
		slog.Error("solving exercise", tint.Err(err))
		cmd.PrintErrln("Failed to solve: ", err)
	}

	// for _, result := range results {
	// 	r := result
	// 	cmd.Printf("%+v\n", r)
	// }

	return nil
}

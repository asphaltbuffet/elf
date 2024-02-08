package cmd

import (
	"log/slog"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

var (
	benchmarkCmd *cobra.Command
	iterations   int
)

const benchmarkExample = `
elf benchmark --num=5 /path/to/exercise
elf benchmark /path/to/exercise`

func GetBenchmarkCmd() *cobra.Command {
	if benchmarkCmd == nil {
		benchmarkCmd = &cobra.Command{
			Use:     "benchmark [path/to/exercise]",
			Aliases: []string{"bench", "b"},
			Example: benchmarkExample,
			Args:    cobra.ExactArgs(1),
			Short:   "benchmark all implementations for the challenge",
			RunE:    runBenchmarkCmd,
		}

		benchmarkCmd.Flags().IntVarP(&iterations, "num", "n", 10, "number of iterations")
	}

	return benchmarkCmd
}

type Benchmarker interface {
	Benchmark(int) ([]advent.TaskResult, error)
	String() string
}

func runBenchmarkCmd(cmd *cobra.Command, args []string) error {
	var (
		ex  Benchmarker
		err error
	)

	dir, err := filepath.Abs(args[0])
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

	return nil
}

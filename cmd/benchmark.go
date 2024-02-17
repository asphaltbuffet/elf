package cmd

import (
	"log/slog"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/krampus"
	"github.com/asphaltbuffet/elf/pkg/tasks"
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
	Benchmark(afero.Fs, int) ([]tasks.Result, error)
	String() string
}

func runBenchmarkCmd(cmd *cobra.Command, args []string) error {
	var (
		ex  Benchmarker
		err error
	)

	cfg, err := krampus.NewConfig()
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		slog.Error("getting current directory", tint.Err(err))
		return err
	}

	slog.Debug("benchmarking exercise", slog.Group("exercise", "dir", dir))

	// TODO: language shouldn't be required for benchmarking
	ex, err = advent.New(&cfg, advent.WithDir(dir), advent.WithLanguage("go"))
	if err != nil {
		slog.Error("creating exercise", tint.Err(err))
		return err
	}

	_, err = ex.Benchmark(cfg.GetFs(), iterations)
	if err != nil {
		slog.Error("solving exercise", tint.Err(err))
		cmd.PrintErrln("Failed to solve: ", err)
	}

	return nil
}

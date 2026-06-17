// Package benchmark is the benchmark subcommand.
package benchmark

import (
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

var (
	benchmarkCmd *cobra.Command
	iterations   int

	// Factory variables for testing.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
	makeBenchmarker = func(cfg config.ExerciseConfiguration, dir string) (Benchmarker, error) {
		return exercise.NewBenchmarker(cfg, exercise.WithExerciseDir(dir))
	}
)

// DefaultIterations is the default number of benchmark runs per implementation.
const DefaultIterations = 10

const benchmarkExample = `
elf benchmark --num=5 /path/to/exercise
elf benchmark /path/to/exercise`

// GetBenchmarkCmd returns the cobra command for benchmarking exercise implementations.
func GetBenchmarkCmd() *cobra.Command {
	if benchmarkCmd == nil {
		benchmarkCmd = &cobra.Command{
			Use:     "benchmark",
			Aliases: []string{"b"},
			Example: benchmarkExample,
			Args:    cobra.ExactArgs(1),
			Short:   "benchmark all implementations for the challenge",
			RunE:    runBenchmarkCmd,
		}

		benchmarkCmd.Flags().IntVarP(&iterations, "num", "n", DefaultIterations, "number of iterations")
		benchmarkCmd.Flags().StringP("config-file", "c", "", "configuration file")
	}

	return benchmarkCmd
}

// Benchmarker is the interface for running benchmark tasks on an exercise.
type Benchmarker interface {
	Benchmark(afero.Fs, int) ([]tasks.Result, error)
	String() string
}

func runBenchmarkCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	ex, err := makeBenchmarker(&cfg, dir)
	if err != nil {
		return err
	}

	_, err = ex.Benchmark(cfg.GetFs(), iterations)
	if err != nil {
		cmd.PrintErrln("benchmark failed:", err)
	}

	// return nil regardless of failure; this wasn't necessarily user error and
	// we don't need to print the error message twice
	return nil
}

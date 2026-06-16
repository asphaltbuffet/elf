// Package benchmark is the benchmark subcommand.
package benchmark

import (
	"context"
	"io"
	"log/slog"
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
	makeBenchmarker = func(lang, dir string, fs afero.Fs, logger *slog.Logger) (Benchmarker, error) {
		ex, err := exercise.Load(dir, lang, "", fs, logger)
		if err != nil {
			return nil, err
		}

		return exercise.NewBenchmarker(ex), nil
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
	Benchmark(
		ctx context.Context,
		fs afero.Fs,
		logger *slog.Logger,
		w io.Writer,
		cb func(tasks.Result),
		iterations int,
	) ([]tasks.Result, error)
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

	lang := cfg.GetLanguage()

	ex, err := makeBenchmarker(lang, dir, cfg.GetFs(), cfg.GetLogger())
	if err != nil {
		return err
	}

	_, err = ex.Benchmark(cmd.Context(), cfg.GetFs(), cfg.GetLogger(), cmd.OutOrStdout(), nil, iterations)
	if err != nil {
		cmd.PrintErrln("benchmark failed:", err)
	}

	// return nil regardless of failure; this wasn't necessarily user error and
	// we don't need to print the error message twice
	return nil
}

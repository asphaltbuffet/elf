// Package benchmark is the benchmark subcommand.
package benchmark

import (
	"path/filepath"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

var (
	benchmarkCmd *cobra.Command
	iterations   int

	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
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

func runBenchmarkCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	appPkg.RegisterRunners(cfg)

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	a := appPkg.New(cfg)

	_, benchErr := a.Benchmark(cmd.Context(), dir, cfg.GetLanguage(), cmd.OutOrStdout(), nil, iterations)
	if benchErr != nil {
		cmd.PrintErrln("benchmark failed:", benchErr)
	}

	// return nil regardless of failure; this wasn't necessarily user error and
	// we don't need to print the error message twice
	return nil
}

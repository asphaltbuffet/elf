// Package benchmark is the benchmark subcommand.
package benchmark

import (
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/internal/render"
	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

var (
	benchmarkCmd *cobra.Command
	iterations   int
	plainFlag    bool
	jsonFlag     bool
	timeoutFlag  time.Duration

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
		benchmarkCmd.Flags().
			DurationVarP(&timeoutFlag, "timeout", "t", 0, "per-task timeout (0 or negative = no timeout; omit to use config default)")
		benchmarkCmd.Flags().BoolVar(&plainFlag, "plain", false, "disable live output (plain renderer)")
		benchmarkCmd.Flags().BoolVar(&jsonFlag, "json", false, "emit machine-readable JSON output")
		benchmarkCmd.MarkFlagsMutuallyExclusive("plain", "json")
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

	if cmd.Flags().Changed("timeout") {
		a.SetTaskTimeout(timeoutFlag)
	}

	h := render.Header{Language: cfg.GetLanguage()}
	_, benchErr := render.Run(cmd.Context(), cmd.OutOrStdout(), h, plainFlag, jsonFlag,
		func(cb func(tasks.Event)) ([]tasks.Result, error) {
			return a.Benchmark(cmd.Context(), dir, cfg.GetLanguage(), cb, iterations)
		})
	if benchErr != nil {
		cmd.PrintErrln("benchmark failed:", benchErr)
	}

	// return nil regardless of failure; this wasn't necessarily user error and
	// we don't need to print the error message twice
	return nil
}

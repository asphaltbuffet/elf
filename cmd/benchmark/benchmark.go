package benchmark

import (
	"path/filepath"

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

const DefaultIterations = 10

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

		benchmarkCmd.Flags().IntVarP(&iterations, "num", "n", DefaultIterations, "number of iterations")
		benchmarkCmd.Flags().StringP("config-file", "c", "", "configuration file")
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

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := krampus.NewConfig(krampus.WithFile(cf))
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	ex, err = advent.NewBenchmarker(&cfg, advent.WithExerciseDir(dir))
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

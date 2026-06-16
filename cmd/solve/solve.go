// Package solve is the solve subcommand.
package solve

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

var (
	solveCmd *cobra.Command
	language string
	input    string
	noTest   bool

	// Factory variables for testing.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
	makeChallenge = func(lang, dir, inputFile string, fs afero.Fs, logger *slog.Logger) (Challenge, error) {
		return exercise.Load(dir, lang, inputFile, fs, logger)
	}
)

const exampleText = `
  elf solve --lang=go --no-test
  elf solve --lang=py
  elf solve # using default language from config`

// GetSolveCmd returns the cobra command for solving an exercise.
func GetSolveCmd() *cobra.Command {
	if solveCmd == nil {
		solveCmd = &cobra.Command{
			Use:     "solve [--lang=<language>] [--no-test] path/to/exercise",
			Aliases: []string{"s"},
			Example: exampleText,
			Args:    cobra.ExactArgs(1),
			Short:   "solve a challenge",
			RunE:    runSolveCmd,
		}

		solveCmd.Flags().BoolVarP(&noTest, "no-test", "X", false, "skip tests")
		solveCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")

		solveCmd.Flags().StringP("config-file", "c", "", "configuration file")
		solveCmd.Flags().StringVarP(&input, "input-file", "i", "", "override input file")
	}

	return solveCmd
}

// Challenge is the interface for solving an exercise challenge.
type Challenge interface {
	Solve(
		ctx context.Context,
		fs afero.Fs,
		logger *slog.Logger,
		runner runners.Runner,
		w io.Writer,
		cb func(tasks.Result),
		skipTests bool,
	) ([]tasks.Result, error)
	String() string
}

func runSolveCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	if language == "" {
		language = cfg.GetLanguage()
	}

	if input == "" {
		input = cfg.GetInputFilename()
	}

	logger := cfg.GetLogger()
	logger.Debug("solving exercise", slog.Group("exercise", "dir", dir, "language", language))

	ch, err := makeChallenge(language, dir, filepath.Clean(input), cfg.GetFs(), logger)
	if err != nil {
		return err
	}

	rc, ok := runners.Available[language]
	if !ok {
		return exercise.ErrNoRunner
	}

	_, solveErr := ch.Solve(cmd.Context(), cfg.GetFs(), logger, rc(dir), cmd.OutOrStdout(), nil, noTest)
	if solveErr != nil {
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

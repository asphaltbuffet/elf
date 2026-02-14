// Package solve is the solve subcommand.
package solve

import (
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/config"
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
	makeChallenge = func(cfg config.ExerciseConfiguration, lang, dir, inputFile string) (Challenge, error) {
		return advent.New(cfg,
			advent.WithLanguage(lang),
			advent.WithDir(dir),
			advent.WithInputFile(inputFile))
	}
)

const exampleText = `
  elf solve --lang=go --no-test
  elf solve --lang=py
  elf solve # using default language from config`

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

type Challenge interface {
	Solve(bool) ([]tasks.Result, error)
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

	cfg.GetLogger().Debug("solving exercise", slog.Group("exercise", "dir", dir, "language", language))

	ch, err := makeChallenge(&cfg, language, dir, filepath.Clean(input))
	if err != nil {
		return err
	}

	_, solveErr := ch.Solve(noTest)
	if solveErr != nil {
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

package cmd

import (
	"log/slog"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var (
	solveCmd *cobra.Command
	language string
	noTest   bool
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
	}

	return solveCmd
}

type Challenge interface {
	Solve(bool) ([]advent.TaskResult, error)
	String() string
}

func runSolveCmd(cmd *cobra.Command, args []string) error {
	var (
		ch  Challenge
		err error
	)

	cfg, err := krampus.NewConfig()

	dir, err := filepath.Abs(args[0])
	if err != nil {
		slog.Error("getting current directory", tint.Err(err))
		return err
	}

	slog.Debug("solving exercise", slog.Group("exercise", "dir", dir, "language", language))

	if language == "" {
		language = cfg.GetLanguage()
	}

	ch, err = advent.New(&cfg, advent.WithLanguage(language), advent.WithDir(dir))
	if err != nil {
		slog.Error("creating exercise", tint.Err(err))
		return err
	}

	_, solveErr := ch.Solve(noTest)
	if solveErr != nil {
		slog.Error("solving exercise", tint.Err(solveErr))
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

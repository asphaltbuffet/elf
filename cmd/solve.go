package cmd

import (
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/krampus"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

var (
	solveCmd *cobra.Command
	language string
	input    string
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

	solveCmd.Flags().StringP("config-file", "c", "", "configuration file")
	solveCmd.Flags().StringVarP(&input, "input-file", "i", "", "override input file")

	return solveCmd
}

type Challenge interface {
	Solve(bool) ([]tasks.Result, error)
	String() string
}

func runSolveCmd(cmd *cobra.Command, args []string) error {
	var (
		ch  Challenge
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

	if language == "" {
		language = cfg.GetLanguage()
	}

	if input == "" {
		input = cfg.GetInputFilename()
	}

	cfg.GetLogger().Debug("solving exercise", slog.Group("exercise", "dir", dir, "language", language))

	ch, err = advent.New(&cfg,
		advent.WithLanguage(language),
		advent.WithDir(dir),
		advent.WithInputFile(filepath.Clean(input)))
	if err != nil {
		return err
	}

	_, solveErr := ch.Solve(noTest)
	if solveErr != nil {
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

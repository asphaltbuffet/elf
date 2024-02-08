package cmd

import (
	"log/slog"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

var testCmd *cobra.Command

type ChallengeTester interface {
	Test() ([]advent.TaskResult, error)
	String() string
}

const exampleTestText = `
elf test /path/to/exercise --lang=go
elf test /path/to/exercise`

func GetTestCmd() *cobra.Command {
	if testCmd == nil {
		testCmd = &cobra.Command{
			Use:     "test FILEPATH",
			Aliases: []string{"t"},
			Example: exampleTestText,
			Args:    cobra.ExactArgs(1),
			Short:   "test a challenge",
			RunE:    runTestCmd,
		}

		testCmd.Flags().StringVarP(&language, "lang", "l", "", "implementation language")
	}

	return testCmd
}

func runTestCmd(cmd *cobra.Command, args []string) error {
	var (
		ch  ChallengeTester
		err error
	)

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	if language == "" {
		language = cfg.GetString("language")
	}

	ch, err = advent.New(advent.WithLanguage(language), advent.WithDir(dir))
	if err != nil {
		return err
	}

	slog.Debug("testing exercise", slog.Any("challenge", ch))

	_, err = ch.Test()
	if err != nil {
		slog.Error("testing exercise", tint.Err(err))
		cmd.Printf("Failed to run tests: %v\n", err)
	}

	return nil
}

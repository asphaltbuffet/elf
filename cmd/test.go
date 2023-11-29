package cmd

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

var testCmd *cobra.Command

type ChallengeTester interface {
	Test() error
	String() string
}

const exampleTestText = `  elf test --lang=go
    elf test --lang=py
    elf test # using default language from config`

func GetTestCmd() *cobra.Command {
	if testCmd == nil {
		testCmd = &cobra.Command{
			Use:     "test [--lang=<language>]",
			Aliases: []string{"t"},
			Example: exampleTestText,
			Args:    cobra.NoArgs,
			Short:   "test a challenge",
			RunE:    runTestCmd,
		}

		testCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")
	}

	return testCmd
}

func runTestCmd(cmd *cobra.Command, _ []string) error {
	var (
		ch  ChallengeTester
		err error
	)

	dir, err := os.Getwd()
	if err != nil {
		slog.Error("getting current directory", tint.Err(err))
		return err
	}

	slog.Debug("testing exercise", slog.Any("challenge", ch))

	ch, err = advent.New(language, advent.WithDir(dir))
	if err != nil {
		slog.Error("loading exercise", tint.Err(err))
		return err
	}

	if testErr := ch.Test(); testErr != nil {
		slog.Error("testing exercise", tint.Err(testErr))
		cmd.Println("Failed to run tests: ", testErr)
	}

	return nil
}

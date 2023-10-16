package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/euler"
)

var testCmd *cobra.Command

type ChallengeTester interface {
	Test() error
	String() string
	SetLanguage(string)
}

const exampleTestText = `  elf test a2015-01 --lang=go
    elf test e1 --lang=py

  If no language is given, the default language is used: 

    elf test a2015-01`

func GetTestCmd() *cobra.Command {
	if testCmd == nil {
		testCmd = &cobra.Command{
			Use:     "test <id> [--lang=<language>]",
			Aliases: []string{"t"},
			Example: exampleTestText,
			Args:    cobra.ExactArgs(1),
			Short:   "test a challenge",
			RunE:    runTestCmd,
		}

		testCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")
	}

	return testCmd
}

func runTestCmd(cmd *cobra.Command, args []string) error {
	var (
		ch  ChallengeTester
		err error
	)

	key, id := args[0][0], args[0][1:]

	switch key {
	case adventKey:
		ch, err = advent.New(id, language)
		if err != nil {
			return err
		}

	case eulerKey:
		eulerID, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("invalid project euler ID: %s is not a number", id)
		}

		ch = euler.New(eulerID, language)

	default:
		return fmt.Errorf("invalid ID specified")
	}

	ch.SetLanguage(language)

	if testErr := ch.Test(); testErr != nil {
		fmt.Println("Failed to run tests: ", testErr)
	}

	return nil
}

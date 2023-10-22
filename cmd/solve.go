package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/euler"
)

var (
	solveCmd *cobra.Command
	language string
	noTest   bool
)

const (
	adventKey = 'a'
	eulerKey  = 'e'
)

const exampleText = `  elf solve a2015-01 --lang=go --no-test
    elf solve e1 --lang=py

  If no language is given, the default language is used: 

    elf solve a2015-01`

func GetSolveCmd() *cobra.Command {
	if solveCmd == nil {
		solveCmd = &cobra.Command{
			Use:     "solve <id> [--lang=<language>] [--no-test]",
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
	Solve() error
	String() string
}

func runSolveCmd(cmd *cobra.Command, args []string) error {
	var (
		ch  Challenge
		err error
	)

	key, id := args[0][0], args[0][1:]
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	slog.Info("solving advent exercise", "dir", dir)

	switch key {
	case adventKey:
		ch, err = advent.NewFromDir(dir, language)
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
		return fmt.Errorf("no ID specified")
	}

	if solveErr := ch.Solve(); solveErr != nil {
		fmt.Println("Failed to solve: ", solveErr)
	}

	return nil
}

package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/aoc"
)

var (
	addCmd *cobra.Command

	MaxYear int = aoc.MaxYear()
)

func GetAddCmd() *cobra.Command {
	if addCmd == nil {
		addCmd = &cobra.Command{
			Use:   "add [-y|--year] [-d|--day] [-L|--language]",
			Args:  cobra.NoArgs,
			Short: "add a new exercise",
			RunE: func(cmd *cobra.Command, args []string) error {
				_, err := acc.AddExercise(yearArg, dayArg, langArg)

				return err
			},
		}
	}

	return addCmd
}

func validYearCompletionArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	out := []string{}

	for _, y := range aoc.ValidYears() {
		out = append(out, strconv.Itoa(y))
	}

	return out, cobra.ShellCompDirectiveNoFileComp
}

package cmd

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	addCmd *cobra.Command

	MaxYear int = getMaxYear()
)

const MinYear int = 2015

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
	y := getValidYears()

	return y, cobra.ShellCompDirectiveNoFileComp
}

func getValidYears() []string {
	var years []string

	for i := MinYear; i <= MaxYear; i++ {
		years = append(years, strconv.Itoa(i))
	}

	return years
}

func getMaxYear() int {
	maxYear := time.Now().Year()

	if time.Now().Month() != time.December {
		maxYear--
	}

	return maxYear
}

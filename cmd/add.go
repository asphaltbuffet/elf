package cmd

import (
	"fmt"
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
			Use:               "add year day [language]",
			Args:  cobra.NoArgs,
			Args:              cobra.MatchAll(cobra.ExactArgs(1), validateAddInput),
			Short:             "add a new exercise",
			RunE: func(cmd *cobra.Command, args []string) error {
				if next {
					return fmt.Errorf("not implemented")
				}

				year, _ := strconv.Atoi(args[0])

				_, err := acc.AddExercise(year, dayArg, langArg)

				return err
			},
		}
	}

	return addCmd
}

func validateAddInput(cmd *cobra.Command, args []string) error {
	// we are assured there is only one arg [see cobra.ExactArgs(1)]
	y, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid year: %s", args[0])
	}

	if y < MinYear {
		return fmt.Errorf("year is out of range: %s < %d", args[0], MinYear)
	}

	if y > MaxYear {
		return fmt.Errorf("year is out of range: %s > %d", args[0], MaxYear)
	}

	if !next && dayArg < 1 || dayArg > 25 {
		return fmt.Errorf("day is out of range: %d: 1 <= day <= 25", dayArg)
	}

	return nil
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

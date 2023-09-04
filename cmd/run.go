package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/runners"
)

var (
	runCmd   *cobra.Command
	testOnly bool
	noTest   bool
)

func GetRunCmd() *cobra.Command {
	if runCmd == nil {
		runCmd = &cobra.Command{
			Use:               "run year day language [flags]",
			ValidArgsFunction: validYearCompletionArgs,
			Args:              cobra.MatchAll(cobra.ExactArgs(3), validateRunInput),
			Short:             "run an exercise",
			RunE: func(cmd *cobra.Command, args []string) error {
				return RunRunCmd(args)
			},
		}
	}

	runCmd.Flags().BoolVarP(&testOnly, "test-only", "t", false, "only run test inputs")
	runCmd.Flags().BoolVarP(&noTest, "no-test", "x", false, "do not run test inputs")
	runCmd.MarkFlagsMutuallyExclusive("test-only", "no-test")

	return runCmd
}

func validateRunInput(cmd *cobra.Command, args []string) error {
	// we are assured there are 3 args
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

	d, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid day: %s", args[1])
	}

	if !next && (d < 1 || d > 25) {
		return fmt.Errorf("day is out of range: %d: 1 <= day <= 25", d)
	}

	if len(args) == 4 {
		if _, ok := runners.Available[args[2]]; !ok {
			return fmt.Errorf("invalid language: %s", args[2])
		}
	}

	return nil
}

func RunRunCmd(args []string) error {
	y, _ := strconv.Atoi(args[0])
	d, _ := strconv.Atoi(args[1])
	lang := args[2]

	err := acc.RunExercise(y, d, lang)
	if err != nil {
		return fmt.Errorf("run exercise: %w", err)
	}

	return nil
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	runCmd   *cobra.Command
	testOnly bool
	noTest   bool
)

func GetRunCmd() *cobra.Command {
	if runCmd == nil {
		runCmd = &cobra.Command{
			Use:   "run [-y|--year] [-d|--day] [-L|language] [-t|--test-only] [-x|--no-test]",
			Args:  cobra.NoArgs,
			Short: "run an exercise",
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

func RunRunCmd(args []string) error {
	err := acc.RunExercise(yearArg, dayArg, langArg)
	if err != nil {
		return fmt.Errorf("run exercise: %w", err)
	}

	return nil
}

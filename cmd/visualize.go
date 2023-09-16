package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var visualizeCmd *cobra.Command

func GetVisualizeCmd() *cobra.Command {
	if visualizeCmd == nil {
		visualizeCmd = &cobra.Command{
			Use:               "visualize year day [language]",
			ValidArgsFunction: validYearCompletionArgs,
			Args:              cobra.NoArgs,
			Short:             "visualize an exercise",
			RunE: func(cmd *cobra.Command, args []string) error {
				return RunVisualizeCmd(args)
			},
		}
	}

	return visualizeCmd
}

func RunVisualizeCmd(args []string) error {
	return fmt.Errorf("not implemented")
}

// func runVisualize(runner runners.Runner, exerciseInputString string) error {
// 	id := "vis"

// 	// directory the runner is run in, which is the exercise directory
// 	r, err := runner.Run(&runners.Task{
// 		TaskID:    id,
// 		Part:      runners.Visualize,
// 		Input:     exerciseInputString,
// 		OutputDir: ".",
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	bold.Print("Visualization: ") //nolint:errcheck,gosec // printing to stdout

// 	var status, followUpText string

// 	if !r.Ok {
// 		status = incompleteLabel
// 		followUpText = fmt.Sprintf(" saying %q", r.Output)
// 	} else {
// 		status = passLabel
// 	}

// 	if followUpText == "" {
// 		followUpText = fmt.Sprintf(" in %s", humanize.SI(r.Duration, "s"))
// 	}

// 	fmt.Print(status)
// 	dimmed.Println(followUpText) //nolint:errcheck,gosec // printing to stdout

// 	return nil
// }

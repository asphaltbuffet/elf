// Package runners provides the "elf runners" command and its subcommands.
package runners

import "github.com/spf13/cobra"

var runnersCmd *cobra.Command

// GetRunnersCmd returns the runners parent command.
func GetRunnersCmd() *cobra.Command {
	if runnersCmd == nil {
		runnersCmd = &cobra.Command{
			Use:               "runners",
			Short:             "Manage elf runner plugins",
			Long:              "Manage runner plugins that execute exercise solutions in different languages.",
			Args:              cobra.NoArgs,
			ValidArgsFunction: cobra.NoFileCompletions,
		}

		runnersCmd.AddCommand(getInstallCmd())
		runnersCmd.AddCommand(getListCmd())
	}

	return runnersCmd
}

// ResetForTest resets command singletons. For use in tests only.
func ResetForTest() {
	runnersCmd = nil
	installCmd = nil
	listCmd = nil
}

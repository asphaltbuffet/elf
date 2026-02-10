package config

import (
	"github.com/spf13/cobra"
)

var configCmd *cobra.Command

// GetConfigCmd returns the config command.
func GetConfigCmd() *cobra.Command {
	if configCmd == nil {
		configCmd = &cobra.Command{
			Use:               "config",
			Short:             "Manage elf configuration",
			Long:              "Manage elf configuration files and settings.",
			Args:              cobra.NoArgs,
			ValidArgsFunction: cobra.NoFileCompletions,
		}

		configCmd.AddCommand(getInitCmd())
		configCmd.AddCommand(getCheckCmd())
		configCmd.AddCommand(getUpdateTokenCmd())
	}

	return configCmd
}

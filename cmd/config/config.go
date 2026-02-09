package config

import (
	"github.com/spf13/cobra"
)

var configCmd *cobra.Command

// GetConfigCmd returns the config command.
func GetConfigCmd() *cobra.Command {
	if configCmd == nil {
		configCmd = &cobra.Command{
			Use:   "config",
			Short: "Manage elf configuration",
			Long: `Manage elf configuration files and settings.

Available subcommands:
  init          Create a new configuration file
  check         Display and validate current configuration
  update-token  Update the Advent of Code authentication token`,
		}

		configCmd.AddCommand(getInitCmd())
		configCmd.AddCommand(getCheckCmd())
		configCmd.AddCommand(getUpdateTokenCmd())
	}

	return configCmd
}

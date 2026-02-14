package config

import (
	"fmt"

	"github.com/spf13/cobra"

	elfcfg "github.com/asphaltbuffet/elf/pkg/config"
)

var checkCmd *cobra.Command

const longDescr = `Display current elf configuration settings and validate them.

Shows all configuration values including:
  - Config file location
  - Language setting
  - Directory paths
  - Token status (masked for security)

Also validates:
  - Token is set and not placeholder
  - Configured directories exist`

func getCheckCmd() *cobra.Command {
	if checkCmd == nil {
		checkCmd = &cobra.Command{
			Use:               "check",
			Short:             "Display and validate current configuration",
			Long:              longDescr,
			RunE:              runCheckCmd,
			Args:              cobra.NoArgs,
			ValidArgsFunction: cobra.NoFileCompletions,
			Example: `elf config check
elf config check -c /path/to/elf.toml`,
		}

		checkCmd.Flags().StringP("config-file", "c", "", "configuration file to check")
	}

	return checkCmd
}

func runCheckCmd(cmd *cobra.Command, _ []string) error {
	cfgFile, _ := cmd.Flags().GetString("config-file")

	cfg, err := elfcfg.NewConfig(elfcfg.WithFile(cfgFile))
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Display header
	cmd.Println("Configuration Status")
	cmd.Println("====================")
	cmd.Println()

	// Config file
	configFile := cfg.GetConfigFileUsed()
	if configFile == "" {
		configFile = "(no config file found, using defaults)"
	}

	cmd.Printf("Config file: %s\n", configFile)
	cmd.Println()

	// Current settings
	cmd.Println("Current Settings")
	cmd.Println("----------------")
	cmd.Printf("  Language:      %s\n", cfg.GetLanguage())
	cmd.Printf("  Exercise dir:  %s\n", cfg.GetBaseDir())
	cmd.Printf("  Config dir:    %s\n", cfg.GetConfigDir())
	cmd.Printf("  Cache dir:     %s\n", cfg.GetCacheDir())
	cmd.Printf("  Input file:    %s\n", cfg.GetInputFilename())
	cmd.Printf("  Advent token:  %s\n", elfcfg.MaskToken(cfg.GetToken()))
	cmd.Println()

	// Validation
	cmd.Println("Validation")
	cmd.Println("----------")

	errs := cfg.Validate()
	if len(errs) == 0 {
		cmd.Println("  All checks passed!")
	} else {
		for _, e := range errs {
			cmd.Printf("  - %s\n", e.Error())
		}
	}

	return nil
}

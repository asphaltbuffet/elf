package runners

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	elfcfg "github.com/asphaltbuffet/elf/pkg/config"
)

var listCmd *cobra.Command

func getListCmd() *cobra.Command {
	if listCmd == nil {
		listCmd = &cobra.Command{
			Use:               "list",
			Short:             "List configured runners",
			Long:              "Lists all runners configured in elf.toml and whether their template files exist on disk.",
			Args:              cobra.NoArgs,
			ValidArgsFunction: cobra.NoFileCompletions,
			RunE:              runListCmd,
			Example:           "elf runners list",
		}

		listCmd.Flags().StringP("config-file", "c", "", "configuration file to read")
	}

	return listCmd
}

func runListCmd(cmd *cobra.Command, _ []string) error {
	cfgFile, _ := cmd.Flags().GetString("config-file")

	cfg, err := elfcfg.NewConfig(elfcfg.WithFile(cfgFile))
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	descs := cfg.GetRunners()

	if len(descs) == 0 {
		cmd.Println("No runners configured.")
		cmd.Println("Run 'elf runners install' to install built-in runner templates.")
		return nil
	}

	cmd.Printf("%-8s  %-16s  %-7s  %s\n", "KEY", "NAME", "STATUS", "TEMPLATE PATH")
	cmd.Println("--------  ----------------  -------  -----------------------------------")

	for _, d := range descs {
		status := "ok"
		templatePath := d.Prepare.TemplatePath

		if templatePath == "" {
			status = "no tmpl"
		} else if _, statErr := os.Stat(templatePath); statErr != nil {
			status = "missing"
		}

		cmd.Printf("%-8s  %-16s  %-7s  %s\n", d.Key, d.Name, status, templatePath)
	}

	return nil
}

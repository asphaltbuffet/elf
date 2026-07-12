// Package encrypt is the encrypt subcommand.
package encrypt

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

var (
	encryptCmd *cobra.Command

	// makeConfig is the sanctioned config seam (not a domain seam); tests stub it.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

// GetEncryptCmd returns the cobra command for encrypting an exercise's Solution Set.
func GetEncryptCmd() *cobra.Command {
	if encryptCmd == nil {
		encryptCmd = &cobra.Command{
			Use:   "encrypt path/to/exercise",
			Args:  cobra.ExactArgs(1),
			Short: "encrypt an exercise's solution files with age",
			Long: "Encrypt an exercise's Solution Set (info.json and language " +
				"subdirectories) to per-file .age siblings, using the SSH public keys in " +
				"encrypt.recipients. Plaintext is left in place; gitignore it and commit the " +
				".age files. Run after solving (build artifacts are removed by elf's cleanup).",
			Example: "elf encrypt euler/42",
			RunE:    runEncryptCmd,
		}

		encryptCmd.Flags().StringP("config-file", "c", "", "configuration file")
	}

	return encryptCmd
}

func runEncryptCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	if err = appPkg.RegisterRunners(cfg); err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("exercise dir: %w", err)
	}

	report, err := appPkg.New(cfg).Encrypt(dir)
	if err != nil {
		return err
	}

	for _, e := range report {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", e.Outcome, e.Path)
	}

	return nil
}

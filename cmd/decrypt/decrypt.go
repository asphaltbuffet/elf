// Package decrypt is the decrypt subcommand.
package decrypt

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

var (
	decryptCmd *cobra.Command

	identity string
	force    bool

	// makeConfig is the sanctioned config seam (not a domain seam); tests stub it.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

// GetDecryptCmd returns the cobra command for decrypting an exercise's Solution Set.
func GetDecryptCmd() *cobra.Command {
	if decryptCmd == nil {
		decryptCmd = &cobra.Command{
			Use:   "decrypt path/to/exercise",
			Args:  cobra.ExactArgs(1),
			Short: "decrypt an exercise's .age solution files",
			Long: "Decrypt an exercise's .age files back to plaintext using an on-disk SSH " +
				"private key (default ~/.ssh/id_ed25519). The .age files are kept. Existing " +
				"plaintext is skipped unless --force. Only unencrypted (no-passphrase) keys are " +
				"supported.",
			Example: "elf decrypt euler/42\nelf decrypt euler/42 -i ~/.ssh/id_work -f",
			RunE:    runDecryptCmd,
		}

		decryptCmd.Flags().StringP("config-file", "c", "", "configuration file")
		decryptCmd.Flags().
			StringVarP(&identity, "identity", "i", "", "SSH private key file (default ~/.ssh/id_ed25519)")
		decryptCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing plaintext")
	}

	return decryptCmd
}

func runDecryptCmd(cmd *cobra.Command, args []string) error {
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

	report, err := appPkg.New(cfg).Decrypt(dir, identity, force)
	if err != nil {
		return err
	}

	for _, e := range report {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", e.Outcome, e.Path)
	}

	return nil
}

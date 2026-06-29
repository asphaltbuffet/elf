// Package download is the download subcommand.
package download

import (
	"fmt"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

var (
	downloadCmd *cobra.Command
	language    string
	forceInput  bool

	// makeConfig is the sanctioned config seam tests stub to avoid touching the real
	// filesystem/env; the domain operation itself is exercised in pkg/app (see ADR-0005).
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

const exampleDownloadText = `elf download https://example.com --lang=go
elf download https://example.com --force --lang=py
elf download https://example.com`

// GetDownloadCmd returns the cobra command for downloading a challenge from a URL.
func GetDownloadCmd() *cobra.Command {
	if downloadCmd == nil {
		downloadCmd = &cobra.Command{
			Use:     "download",
			Aliases: []string{"d"},
			Example: exampleDownloadText,
			Args:    cobra.ExactArgs(1),
			Short:   "download challenge info from url",
			RunE:    runDownloadCmd,
		}

		downloadCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")
		downloadCmd.Flags().BoolVarP(&forceInput, "force-input", "I", false, "overwrite existing input file")

		downloadCmd.Flags().StringP("config-file", "c", "", "configuration file")
	}

	return downloadCmd
}

func runDownloadCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	appPkg.RegisterRunners(cfg)

	forced := &exercise.Overwrites{
		Input: forceInput,
	}

	report, path, err := appPkg.New(cfg).Add(args[0], language, forced)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, path)
	exercise.RenderReport(out, report)

	return nil
}

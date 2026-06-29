// Package download is the download subcommand.
package download

import (
	"fmt"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// Adder makes a challenge exercise exist in the workspace.
type Adder interface {
	Add() error
	FilePath() string
	Report() exercise.Report
}

var (
	downloadCmd *cobra.Command
	language    string
	forceInput  bool

	// Factory variables for testing.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
	makeAdder = func(cfg config.Config, url, lang string, forced *exercise.Overwrites) (Adder, error) {
		return exercise.NewAdder(cfg,
			exercise.WithURL(url),
			exercise.WithLanguage(lang),
			exercise.WithOverwrites(forced),
		)
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

	chdl, err := makeAdder(cfg, args[0], language, forced)
	if err != nil {
		return fmt.Errorf("creating adder: %w", err)
	}

	if err = chdl.Add(); err != nil {
		return fmt.Errorf("adding challenge: %w", err)
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, chdl.FilePath())
	exercise.RenderReport(out, chdl.Report())

	return nil
}

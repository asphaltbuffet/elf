// Package download is the download subcommand.
package download

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

// Downloader is an interface for downloading challenges.
type Downloader interface {
	Download() error
	FilePath() string
}

var (
	downloadCmd *cobra.Command
	language    string
	forceInput  bool

	// Factory variables for testing.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
	makeDownloader = func(cfg config.DownloadConfiguration, url, lang string, forced *exercise.Overwrites) (Downloader, error) {
		return exercise.NewDownloader(cfg,
			exercise.WithURL(url),
			exercise.WithDownloadLanguage(lang),
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

	forced := &exercise.Overwrites{
		Input: forceInput,
	}

	chdl, err := makeDownloader(&cfg, args[0], language, forced)
	if err != nil {
		return fmt.Errorf("creating downloader: %w", err)
	}

	if err = chdl.Download(); err != nil {
		return fmt.Errorf("downloading challenge: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), chdl.FilePath())

	return nil
}

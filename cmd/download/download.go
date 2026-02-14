// Package download is the download subcommand.
package download

import (
	"fmt"
	"strings"

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
)

const exampleDownloadText = `elf download https://example.com --lang=go
elf download https://example.com --force --lang=py
elf download https://example.com`

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

// // https://adventofcode.com/2022/day/1
// reAdvent := `^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})$`

func runDownloadCmd(cmd *cobra.Command, args []string) error {
	var err error
	var chdl Downloader

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := config.NewConfig(config.WithFile(cf))
	if err != nil {
		return err
	}

	forced := &exercise.Overwrites{
		Input: forceInput,
	}

	switch {
	case strings.Contains(args[0], "adventofcode.com/"):
		chdl, err = exercise.NewDownloader(&cfg,
			exercise.WithURL(args[0]),
			exercise.WithDownloadLanguage(language),
			exercise.WithOverwrites(forced),
		)
		if err != nil {
			return fmt.Errorf("downloading advent challenge: %w", err)
		}

	default:
		return fmt.Errorf("unsupported URL: %s", args[0])
	}

	err = chdl.Download()
	if err != nil {
		return fmt.Errorf("downloading challenge: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), chdl.FilePath())

	return nil
}

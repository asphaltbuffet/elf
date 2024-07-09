package download

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

// Downloader is an interface for downloading challenges.
type Downloader interface {
	Download() error
}

var (
	downloadCmd *cobra.Command
	skipLang    bool
	language    string
	forceAll    bool
	forceInfo   bool
	forceReadme bool
	forceInput  bool
)

const exampleDownloadText = `  elf download https://example.com --lang=go
    elf download https://example.com --force --lang=py

  If no language is given, the default language is used:

    elf download https://example.com`

func GetDownloadCmd() *cobra.Command {
	if downloadCmd == nil {
		downloadCmd = &cobra.Command{
			Use:     "download [flags] url",
			Aliases: []string{"d"},
			Example: exampleDownloadText,
			Args:    cobra.ExactArgs(1),
			Short:   "download a challenge",
			RunE:    runDownloadCmd,
		}

		downloadCmd.Flags().BoolVarP(&skipLang, "skip-lang", "L", false, "skip creating implementation files")
		downloadCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")
		downloadCmd.Flags().BoolVarP(&forceAll, "force-all", "A", false, "overwrite existing files")
		downloadCmd.Flags().BoolVarP(&forceInfo, "force-info", "N", false, "overwrite existing info file")
		downloadCmd.Flags().BoolVarP(&forceReadme, "force-readme", "R", false, "overwrite existing README file")
		downloadCmd.Flags().BoolVarP(&forceInput, "force-input", "I", false, "overwrite existing input file")

		downloadCmd.Flags().StringP("config-file", "c", "", "configuration file")
	}

	return downloadCmd
}

// // https://adventofcode.com/2022/day/1
// reAdvent := `^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})$`
// // https://projecteuler.net/problem=1
// reEuler := `^https?://(www\.)?projecteuler\.net/problem=(?P<num>\d{1,3})$`

func runDownloadCmd(cmd *cobra.Command, args []string) error {
	var err error
	var chdl Downloader

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := krampus.NewConfig(krampus.WithFile(cf))
	if err != nil {
		return err
	}

	forced := &advent.Overwrites{
		// All:    forceAll,
		// Info:   forceInfo,
		// Readme: forceReadme,
		Input: forceInput,
	}

	switch {
	case strings.Contains(args[0], "adventofcode.com/"):
		chdl, err = advent.NewDownloader(&cfg,
			advent.WithURL(args[0]),
			advent.WithDownloadLanguage(language),
			advent.WithOverwrites(forced),
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

	cmd.Printf("New challenge created in: %s", chdl)

	return nil
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var (
	downloadCmd *cobra.Command
	noLang      bool
	dlAll       bool
	forceAll    bool
	dlInfo      bool
	forceInfo   bool
	dlReadme    bool
	forceReadme bool
	dlInput     bool
	forceInput  bool
)

const exampleDownloadText = `  elf download https://example.com --lang=go
    elf download https://example.com --force --lang=py

  If no language is given, the default language is used: 

    elf download https://example.com`

func GetDownloadCmd() *cobra.Command {
	if downloadCmd == nil {
		downloadCmd = &cobra.Command{
			Use:     "download <url> [-A | -inrL | -[a]INR] [--no-lang | --lang <string>]",
			Aliases: []string{"d"},
			Example: exampleDownloadText,
			Args:    cobra.ExactArgs(1),
			Short:   "download a challenge",
			RunE:    runDownloadCmd,
		}

		// downloadCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")
		downloadCmd.Flags().BoolVarP(&noLang, "no-lang", "L", false, "do not create language directory")
		// downloadCmd.MarkFlagsMutuallyExclusive("lang", "no-lang")

		downloadCmd.Flags().BoolVarP(&dlAll, "all", "a", false, "download/create all missing files")
		downloadCmd.Flags().BoolVarP(&forceAll, "force-all", "A", false, "download/create all files; overwrite existing")
		downloadCmd.MarkFlagsMutuallyExclusive("all", "force-all")

		downloadCmd.Flags().BoolVarP(&dlInfo, "info", "n", false, "create info file, if missing")
		downloadCmd.Flags().BoolVarP(&forceInfo, "force-info", "N", false, "create info file, overwrite existing")
		downloadCmd.MarkFlagsMutuallyExclusive("info", "force-info")

		downloadCmd.Flags().BoolVarP(&dlReadme, "readme", "r", false, "create README file, if missing")
		downloadCmd.Flags().BoolVarP(&forceReadme, "force-readme", "R", false, "create README file, overwrite existing")
		downloadCmd.MarkFlagsMutuallyExclusive("readme", "force-readme")

		downloadCmd.Flags().BoolVarP(&dlInput, "input", "i", false, "create input file, if missing")
		downloadCmd.Flags().BoolVarP(&forceInput, "force-input", "I", false, "download input file; overwrite existing")
		downloadCmd.MarkFlagsMutuallyExclusive("input", "force-input")
	}

	return downloadCmd
}

type Downloader interface {
	Download() error
	Path() string
}

// // https://adventofcode.com/2022/day/1
// reAdvent := `^https?://(www\.)?adventofcode\.com/(?P<year>\d{4})/day/(?P<day>\d{1,2})$`
// // https://projecteuler.net/problem=1
// reEuler := `^https?://(www\.)?projecteuler\.net/problem=(?P<num>\d{1,3})$`

func runDownloadCmd(cmd *cobra.Command, args []string) error {
	var err error
	var chdl Downloader

	cfg, err := krampus.NewConfig()
	if err != nil {
		return err
	}

	switch {
	case strings.Contains(args[0], "adventofcode.com/"):
		chdl, err = advent.NewDownloader(&cfg, args[0], language)
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

	cmd.Printf("New challenge created in: %s", chdl.Path())

	return nil
}

// Package download is the download subcommand.
package download

import (
	"fmt"
	"strconv"

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

const exampleDownloadText = `elf download https://adventofcode.com/2015/day/1 --lang=go
elf download 2015 1 --lang=go
elf download 2015 1 --force-input`

// GetDownloadCmd returns the cobra command for downloading a challenge from a URL.
func GetDownloadCmd() *cobra.Command {
	if downloadCmd == nil {
		downloadCmd = &cobra.Command{
			Use:     "download (<url> | <year> <day>)",
			Aliases: []string{"d"},
			Example: exampleDownloadText,
			Args:    cobra.RangeArgs(1, maxDownload),
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
	target, err := resolveTarget(args)
	if err != nil {
		return err
	}

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	appPkg.RegisterRunners(cfg)

	forced := &exercise.Overwrites{
		Input: forceInput,
	}

	report, path, err := appPkg.New(cfg).Add(target, language, forced)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, path)
	exercise.RenderReport(out, report)

	return nil
}

const (
	minAoCYear  = 2015 // Advent of Code's first year
	minAoCDay   = 1
	maxAoCDay   = 25 // AoC runs Dec 1–25
	maxDownload = 2  // max positional args: <url> or <year> <day>
)

// resolveTarget turns the command's positional args into a puzzle URL. One arg
// is treated as a URL and passed through unchanged (ParseURL validates it
// downstream). Two args are parsed as <year> <day>, range-checked against AoC's
// real bounds, and assembled into the canonical URL. Assembly errors name the
// offending value so the user sees a clear message instead of a failed request.
func resolveTarget(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	if len(args) != 2 { //nolint:mnd // 2 == the <year> <day> form; cobra RangeArgs guards this
		return "", fmt.Errorf("expected a URL or <year> <day>, got %d arguments", len(args))
	}

	year, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("year %q is not a number", args[0])
	}

	day, err := strconv.Atoi(args[1])
	if err != nil {
		return "", fmt.Errorf("day %q is not a number", args[1])
	}

	if year < minAoCYear {
		return "", fmt.Errorf("year %d is before Advent of Code started (%d)", year, minAoCYear)
	}

	if day < minAoCDay || day > maxAoCDay {
		return "", fmt.Errorf("day %d is out of range (%d–%d)", day, minAoCDay, maxAoCDay)
	}

	return fmt.Sprintf("https://adventofcode.com/%d/day/%d", year, day), nil
}

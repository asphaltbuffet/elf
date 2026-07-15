// Package add is the add subcommand, with per-kind subcommands (aoc, euler).
package add

import (
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
)

const (
	eulerArgCount = 1

	minAoCYear = 2015 // Advent of Code's first year
	minAoCDay  = 1
	maxAoCDay  = 25 // AoC runs Dec 1–25
	maxAoCArgs = 2  // max positional args: <url> or <year> <day>
)

var (
	addCmd     *cobra.Command
	language   string
	forceInput bool

	// makeConfig is the sanctioned config seam (not a domain seam); tests stub it.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

// GetAddCmd returns the `add` command with per-kind subcommands.
func GetAddCmd() *cobra.Command {
	if addCmd == nil {
		addCmd = &cobra.Command{
			Use:   "add",
			Short: "add a challenge exercise to the workspace",
		}
		addCmd.PersistentFlags().StringP("config-file", "c", "", "configuration file")

		addCmd.AddCommand(aocCmd(), eulerCmd())
	}

	return addCmd
}

func aocCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "aoc (<url> | <year> <day>)",
		Short: "add an Advent of Code puzzle from a URL",
		Args:  cobra.RangeArgs(1, maxAoCArgs),
		RunE:  runAoCCmd,
	}
	c.Flags().StringVarP(&language, "lang", "l", "", "solution language")
	c.Flags().BoolVarP(&forceInput, "force-input", "I", false, "overwrite existing input file")

	return c
}

func eulerCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "euler <number>",
		Short: "add a Project Euler problem",
		Args:  cobra.ExactArgs(eulerArgCount),
		RunE:  runEulerCmd,
	}
	c.Flags().StringVarP(&language, "lang", "l", "", "solution language")

	return c
}

func runEulerCmd(cmd *cobra.Command, args []string) error {
	number, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("problem number %q is not a number", args[0])
	}

	if number <= 0 {
		return fmt.Errorf("problem number must be positive, got %d", number)
	}

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	if err = appPkg.RegisterRunners(cfg); err != nil {
		return err
	}

	report, path, placeholdered, err := appPkg.New(cfg).AddProblem(number, language)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, path)
	exercise.RenderReport(out, report)

	printTitleWarning(out, placeholdered)

	return nil
}

// printTitleWarning writes a warning to w when the title fetch failed and
// AddProblem fell back to a placeholder title ("Untitled" in info.json). It is
// a no-op when placeholdered is false.
func printTitleWarning(w io.Writer, placeholdered bool) {
	if !placeholdered {
		return
	}

	// placeholderTitleEcho mirrors pkg/exercise's placeholderTitle for display
	// only; it is not imported to avoid exporting a domain constant across the
	// package boundary just for a warning string.
	const placeholderTitleEcho = "Untitled"

	_, _ = fmt.Fprintf(w,
		"warning: could not fetch title from projecteuler.net; wrote %q — edit info.json to fix\n",
		placeholderTitleEcho)
}

func runAoCCmd(cmd *cobra.Command, args []string) error {
	target, err := resolveAoCTarget(args)
	if err != nil {
		return err
	}

	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	if err = appPkg.RegisterRunners(cfg); err != nil {
		return err
	}

	forced := &exercise.Overwrites{Input: forceInput}

	report, path, err := appPkg.New(cfg).Add(target, language, forced)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, path)
	exercise.RenderReport(out, report)

	return nil
}

// resolveAoCTarget turns the command's positional args into a puzzle URL. One arg
// is treated as a URL and passed through unchanged (ParseURL validates it
// downstream). Two args are parsed as <year> <day>, range-checked against AoC's
// real bounds, and assembled into the canonical URL. Assembly errors name the
// offending value so the user sees a clear message instead of a failed request.
func resolveAoCTarget(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	if len(args) != maxAoCArgs {
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

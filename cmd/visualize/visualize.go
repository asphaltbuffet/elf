// Package visualize is the visualize subcommand.
package visualize

import (
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/internal/render"
	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

var (
	visualizeCmd *cobra.Command
	language     string
	outdir       string
	plainFlag    bool
	jsonFlag     bool
	timeoutFlag  time.Duration

	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

const exampleText = `
  elf visualize --lang=go path/to/exercise
  elf visualize --lang=go --outdir=/tmp path/to/exercise`

// GetVisualizeCmd returns the cobra command for visualizing an exercise.
func GetVisualizeCmd() *cobra.Command {
	if visualizeCmd == nil {
		visualizeCmd = &cobra.Command{
			Use:     "visualize [--lang=<language>] [--outdir=<dir>] path/to/exercise",
			Aliases: []string{"vis", "v"},
			Example: exampleText,
			Args:    cobra.ExactArgs(1),
			Short:   "visualize a challenge solution",
			RunE:    runVisualizeCmd,
		}

		visualizeCmd.Flags().StringVarP(&language, "lang", "l", "", "implementation language")
		visualizeCmd.Flags().StringVarP(
			&outdir, "outdir", "o", "",
			"output directory (default: the exercise directory)")

		visualizeCmd.Flags().StringP("config-file", "c", "", "configuration file")
		visualizeCmd.Flags().
			DurationVarP(&timeoutFlag, "timeout", "t", 0, "per-task timeout (0 or negative = no timeout; omit to use config default)")
		visualizeCmd.Flags().BoolVar(&plainFlag, "plain", false, "disable live output (plain renderer)")
		visualizeCmd.Flags().BoolVar(&jsonFlag, "json", false, "emit machine-readable JSON output")
		visualizeCmd.MarkFlagsMutuallyExclusive("plain", "json")
	}

	return visualizeCmd
}

func runVisualizeCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	appPkg.RegisterRunners(cfg)

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	if language == "" {
		language = cfg.GetLanguage()
	}

	outdir, err = resolveOutdir(outdir, dir)
	if err != nil {
		return err
	}

	a := appPkg.New(cfg)

	if cmd.Flags().Changed("timeout") {
		a.SetTaskTimeout(timeoutFlag)
	}

	h := render.Header{Language: language}
	_, visErr := render.Run(cmd.Context(), cmd.OutOrStdout(), h, plainFlag, jsonFlag,
		func(cb func(tasks.Event)) ([]tasks.Result, error) {
			return a.Visualize(cmd.Context(), dir, language, outdir, cb)
		})
	if visErr != nil {
		cfg.GetLogger().Error("visualizing exercise", tint.Err(visErr))
		cmd.Printf("Failed to visualize: %v\n", visErr)
	}

	return nil
}

// resolveOutdir returns the directory visualization artifacts are written to.
// When the user did not pass --outdir, it defaults to the exercise's own
// directory (dir, already absolute) rather than the current working directory
// (see ADR-0015). An explicit --outdir is resolved to an absolute path.
func resolveOutdir(outdir, dir string) (string, error) {
	if outdir == "" {
		return dir, nil
	}

	return filepath.Abs(outdir)
}

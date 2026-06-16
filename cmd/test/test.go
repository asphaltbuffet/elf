// Package test is the test subcommand.
package test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/config"
	"github.com/asphaltbuffet/elf/pkg/exercise"
	"github.com/asphaltbuffet/elf/pkg/runners"
	"github.com/asphaltbuffet/elf/pkg/tasks"
)

var (
	testCmd  *cobra.Command
	language string

	// Factory variables for testing.
	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
	makeChallengeTester = func(lang, dir string, fs afero.Fs, logger *slog.Logger) (ChallengeTester, error) {
		return exercise.Load(dir, lang, "", fs, logger)
	}
)

// ChallengeTester is the interface for running tests against an exercise challenge.
type ChallengeTester interface {
	Test(
		ctx context.Context,
		logger *slog.Logger,
		runner runners.Runner,
		w io.Writer,
		cb func(tasks.Result),
	) ([]tasks.Result, error)
	String() string
}

const exampleTestText = `
elf test /path/to/exercise --lang=go
elf test /path/to/exercise`

// GetTestCmd returns the cobra command for testing an exercise.
func GetTestCmd() *cobra.Command {
	if testCmd == nil {
		testCmd = &cobra.Command{
			Use:     "test FILEPATH",
			Aliases: []string{"t"},
			Example: exampleTestText,
			Args:    cobra.ExactArgs(1),
			Short:   "test a challenge",
			RunE:    runTestCmd,
		}

		testCmd.Flags().StringVarP(&language, "lang", "l", "", "implementation language")
		testCmd.Flags().StringP("config-file", "c", "", "configuration file")
	}

	return testCmd
}

func runTestCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	if language == "" {
		language = cfg.GetLanguage()
	}

	logger := cfg.GetLogger()

	ch, err := makeChallengeTester(language, dir, cfg.GetFs(), logger)
	if err != nil {
		return err
	}

	logger.Debug("testing exercise", slog.Any("challenge", ch))

	rc, ok := runners.Available[language]
	if !ok {
		return exercise.ErrNoRunner
	}

	_, err = ch.Test(cmd.Context(), logger, rc(dir), cmd.OutOrStdout(), nil)
	if err != nil {
		logger.Error("testing exercise", tint.Err(err))
		cmd.Printf("Failed to run tests: %v\n", err)
	}

	return nil
}

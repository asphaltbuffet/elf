// Package test is the test subcommand.
package test

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
	testCmd     *cobra.Command
	language    string
	plainFlag   bool
	jsonFlag    bool
	timeoutFlag time.Duration

	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

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
		testCmd.Flags().
			DurationVarP(&timeoutFlag, "timeout", "t", 0, "per-task timeout (0 or negative = no timeout; omit to use config default)")
		testCmd.Flags().BoolVar(&plainFlag, "plain", false, "disable live output (plain renderer)")
		testCmd.Flags().BoolVar(&jsonFlag, "json", false, "emit machine-readable JSON output")
		testCmd.MarkFlagsMutuallyExclusive("plain", "json")
	}

	return testCmd
}

func runTestCmd(cmd *cobra.Command, args []string) error {
	cf, _ := cmd.Flags().GetString("config-file")

	cfg, err := makeConfig(cf)
	if err != nil {
		return err
	}

	if err = appPkg.RegisterRunners(cfg); err != nil {
		return err
	}

	dir, err := filepath.Abs(args[0])
	if err != nil {
		return err
	}

	if language == "" {
		language = cfg.GetLanguage()
	}

	a := appPkg.New(cfg)

	if cmd.Flags().Changed("timeout") {
		a.SetTaskTimeout(timeoutFlag)
	}

	h := render.Header{Language: language}
	_, testErr := render.Run(cmd.Context(), cmd.OutOrStdout(), h, plainFlag, jsonFlag,
		func(cb func(tasks.Event)) ([]tasks.Result, error) {
			return a.Test(cmd.Context(), dir, language, "", cb)
		})
	if testErr != nil {
		cfg.GetLogger().Error("testing exercise", tint.Err(testErr))
		cmd.Printf("Failed to run tests: %v\n", testErr)
	}

	return nil
}

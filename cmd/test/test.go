// Package test is the test subcommand.
package test

import (
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

var (
	testCmd     *cobra.Command
	language    string
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
	}

	return testCmd
}

func runTestCmd(cmd *cobra.Command, args []string) error {
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

	a := appPkg.New(cfg)

	if cmd.Flags().Changed("timeout") {
		a.SetTaskTimeout(timeoutFlag)
	}

	_, testErr := a.Test(cmd.Context(), dir, language, "", cmd.OutOrStdout(), nil)
	if testErr != nil {
		cfg.GetLogger().Error("testing exercise", tint.Err(testErr))
		cmd.Printf("Failed to run tests: %v\n", testErr)
	}

	return nil
}

// Package solve is the solve subcommand.
package solve

import (
	"path/filepath"

	"github.com/spf13/cobra"

	appPkg "github.com/asphaltbuffet/elf/pkg/app"
	"github.com/asphaltbuffet/elf/pkg/config"
)

var (
	solveCmd *cobra.Command
	language string
	input    string
	noTest   bool

	makeConfig = func(cf string) (config.Config, error) {
		return config.NewConfig(config.WithFile(cf))
	}
)

const exampleText = `
  elf solve --lang=go --no-test
  elf solve --lang=py
  elf solve # using default language from config`

// GetSolveCmd returns the cobra command for solving an exercise.
func GetSolveCmd() *cobra.Command {
	if solveCmd == nil {
		solveCmd = &cobra.Command{
			Use:     "solve [--lang=<language>] [--no-test] path/to/exercise",
			Aliases: []string{"s"},
			Example: exampleText,
			Args:    cobra.ExactArgs(1),
			Short:   "solve a challenge",
			RunE:    runSolveCmd,
		}

		solveCmd.Flags().BoolVarP(&noTest, "no-test", "X", false, "skip tests")
		solveCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")

		solveCmd.Flags().StringP("config-file", "c", "", "configuration file")
		solveCmd.Flags().StringVarP(&input, "input-file", "i", "", "override input file")
	}

	return solveCmd
}

func runSolveCmd(cmd *cobra.Command, args []string) error {
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

	if input == "" {
		input = cfg.GetInputFilename()
	}

	a := appPkg.New(cfg)

	_, solveErr := a.Solve(cmd.Context(), dir, language, filepath.Clean(input), cmd.OutOrStdout(), nil, noTest)
	if solveErr != nil {
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

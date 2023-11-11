package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/advent"
)

var (
	solveCmd *cobra.Command
	language string
	noTest   bool
)

const exampleText = `  elf solve --lang=go --no-test
  elf solve --lang=py
  elf solve # using default language from config`

func GetSolveCmd() *cobra.Command {
	if solveCmd == nil {
		solveCmd = &cobra.Command{
			Use:     "solve <id> [--lang=<language>] [--no-test]",
			Aliases: []string{"s"},
			Example: exampleText,
			Args:    cobra.NoArgs,
			Short:   "solve a challenge",
			RunE:    runSolveCmd,
		}

		solveCmd.Flags().BoolVarP(&noTest, "no-test", "X", false, "skip tests")
		solveCmd.Flags().StringVarP(&language, "lang", "l", "", "solution language")
	}

	return solveCmd
}

type Challenge interface {
	Solve() error
	String() string
}

type Info struct {
	ChallengeType string `json:"type"`
}

func runSolveCmd(cmd *cobra.Command, _ []string) error {
	var (
		ch  Challenge
		err error
	)

	dir, err := os.Getwd()
	if err != nil {
		slog.Error("getting current directory", tint.Err(err))
		return err
	}

	info, err := ReadInfo(dir)
	if err != nil {
		slog.Error("getting exercise info", tint.Err(err))
		return err
	}

	slog.Debug("solving exercise", slog.Group("exercise", "dir", dir, "language", language, "type", info.ChallengeType))

	ch, err = advent.NewFromDir(dir, language)
	if err != nil {
		slog.Error("creating exercise", tint.Err(err))
		return err
	}

	if solveErr := ch.Solve(); solveErr != nil {
		slog.Error("solving exercise", tint.Err(solveErr))
		cmd.PrintErrln("Failed to solve: ", solveErr)
	}

	return nil
}

func ReadInfo(dir string) (*Info, error) {
	fn := filepath.Join(dir, "info.json")

	data, err := os.ReadFile(path.Clean(fn))
	if err != nil {
		slog.Debug("failed to read info", tint.Err(err))
		return nil, fmt.Errorf("read info file %q: %w", fn, err)
	}

	d := &Info{}

	err = json.Unmarshal(data, d)
	if err != nil {
		slog.Debug("failed to unmarshall info", tint.Err(err))
		return nil, fmt.Errorf("unmarshal info file %s: %w", fn, err)
	}

	return d, nil
}

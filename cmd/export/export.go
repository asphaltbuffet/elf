package export

import (
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/asphaltbuffet/elf/pkg/krampus"
)

var (
	year   int
	format string

	config krampus.ExerciseConfiguration
)

func NewCommand(cfg krampus.ExerciseConfiguration) *cobra.Command {
	config = cfg

	cmd := &cobra.Command{
		Use:   "export [flags]",
		Short: "export challenge data",
		Args:  cobra.NoArgs,
		RunE:  runExportCmd,
	}

	cmd.Flags().StringVarP(&format, "format", "F", "text", "output format")
	cmd.Flags().IntVarP(&year, "year", "y", time.Now().Year(), "restrict to specific year")

	return cmd
}

func runExportCmd(_ *cobra.Command, _ []string) error {
	if config == nil {
		return errors.New("no configuration provided")
	}

	return nil
}

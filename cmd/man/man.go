// Package man is the man subcommand.
package man

import (
	"fmt"
	"os"

	mango "github.com/muesli/mango-cobra"
	"github.com/muesli/roff"
	"github.com/spf13/cobra"
)

// NewManCmd returns the cobra command for generating man pages.
func NewManCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "man",
		Short:                 "Generate command line manpages",
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Hidden:                true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mp, err := mango.NewManPage(1, cmd.Root())
			if err != nil {
				return err
			}

			_, err = fmt.Fprint(os.Stdout, mp.Build(roff.NewDocument()))

			return err
		},
	}

	return cmd
}

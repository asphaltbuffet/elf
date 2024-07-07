package man

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

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
			if err := doc.GenManTree(cmd.Root(), nil, "manpages/"); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

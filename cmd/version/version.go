package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprint(cmd.OutOrStdout(), version)
		},
	}

	return cmd
}

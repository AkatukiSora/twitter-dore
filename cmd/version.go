package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = ""
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, _ []string) {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), versionString()); err != nil {
				// Printing errors shouldn't propagate; Cobra will report them if needed.
				cmd.PrintErrln(err)
			}
		},
	}
}

func versionString() string {
	if commit != "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return version
}

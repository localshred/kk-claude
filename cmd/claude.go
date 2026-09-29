// Package cmd implements the kk-claude command.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the root command and returns a process exit code.
func Execute() int {
	root := &cobra.Command{
		Use:           "kk-claude",
		Short:         "Pick/resume a saved Claude session, or start a named new one",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "kk-claude:", err)
		return 1
	}
	return 0
}

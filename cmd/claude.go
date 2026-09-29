// Package cmd implements the kk-claude command.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/localshred/kk-claude/internal/session"
	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/koolkit/pkg/pick"
	"github.com/spf13/cobra"
)

// Execute runs the root command and returns a process exit code.
func Execute() int {
	if err := newRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "kk-claude:", err)
		return 1
	}
	return 0
}

func newApp(out *cobra.Command) session.App {
	return session.App{
		Out:    out.OutOrStdout(),
		Launch: session.ExecLauncher,
		PickOne: func(items []store.Session, label func(store.Session) string) (store.Session, bool, error) {
			return pick.One(items, label)
		},
		PickMany: func(items []store.Session, label func(store.Session) string) ([]store.Session, error) {
			return pick.Many(items, label)
		},
		Prompt: prompt,
		Now:    time.Now,
	}
}

func newRoot() *cobra.Command {
	_ = store.Migrate()
	root := &cobra.Command{
		Use:           "kk-claude",
		Short:         "Pick/resume a saved Claude session, or start a named new one",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cwd, _ := os.Getwd()
			return newApp(cmd).Pick(cwd)
		},
	}
	root.AddCommand(
		&cobra.Command{
			Use: "list", Short: "Print saved sessions in alphabetical order",
			RunE: func(cmd *cobra.Command, _ []string) error { return newApp(cmd).List() },
		},
		&cobra.Command{
			Use: "scratch", Short: "Launch a throwaway session, unsaved",
			RunE: func(cmd *cobra.Command, _ []string) error { return newApp(cmd).Scratch() },
		},
		&cobra.Command{
			Use: "keep <name> [session-id]", Short: "Save the current session under <name>",
			Args: cobra.RangeArgs(1, 2),
			RunE: func(cmd *cobra.Command, args []string) error {
				cwd, _ := os.Getwd()
				return newApp(cmd).Keep(args, cwd)
			},
		},
		&cobra.Command{
			Use: "prune", Short: "Auto-remove dead sessions, then multi-select to prune more",
			RunE: func(cmd *cobra.Command, _ []string) error { return newApp(cmd).Prune() },
		},
	)
	return root
}

func prompt(msg string) string {
	fmt.Print(msg + ": ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line)
}

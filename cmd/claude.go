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

var (
	flagHigh  bool
	flagLow   bool
	flagMid   bool
	flagModel string
)

// appFrom builds an App with the resolved --model/--effort flags applied.
func appFrom(cmd *cobra.Command) (session.App, error) {
	model, err := session.ModelFor(flagModel)
	if err != nil {
		return session.App{}, err
	}
	app := newApp(cmd)
	app.Model = model
	app.Effort = effortFrom()
	return app, nil
}

// effortFrom resolves the --low/--mid/--high flags to a claude effort level,
// highest wins; empty when none set.
func effortFrom() string {
	switch {
	case flagHigh:
		return "high"
	case flagMid:
		return "medium"
	case flagLow:
		return "low"
	default:
		return ""
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
			app, err := appFrom(cmd)
			if err != nil {
				return err
			}
			cwd, _ := os.Getwd()
			return app.Pick(cwd)
		},
	}
	root.PersistentFlags().StringVarP(&flagModel, "model", "m", "", "Model: s (sonnet 5.5) or o (opus 5.5 1m)")
	root.PersistentFlags().BoolVar(&flagLow, "low", false, "Low effort for the session")
	root.PersistentFlags().BoolVar(&flagMid, "mid", false, "Medium effort for the session")
	root.PersistentFlags().BoolVar(&flagHigh, "high", false, "High effort for the session")
	root.AddCommand(
		&cobra.Command{
			Use: "list", Short: "Print saved sessions in alphabetical order",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := appFrom(cmd)
				if err != nil {
					return err
				}
				return app.List()
			},
		},
		&cobra.Command{
			Use: "scratch", Short: "Launch a throwaway session, unsaved (low effort sonnet by default)",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := appFrom(cmd)
				if err != nil {
					return err
				}
				return app.Scratch()
			},
		},
		&cobra.Command{
			Use: "review <url>", Short: "Scratch session seeded with /spiff-review <url> (high effort opus by default)",
			Args: cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := appFrom(cmd)
				if err != nil {
					return err
				}
				return app.Review(args[0])
			},
		},
		&cobra.Command{
			Use: "keep <name> [session-id]", Short: "Save the current session under <name>",
			Args: cobra.RangeArgs(1, 2),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := appFrom(cmd)
				if err != nil {
					return err
				}
				cwd, _ := os.Getwd()
				return app.Keep(args, cwd)
			},
		},
		&cobra.Command{
			Use: "prune", Short: "Auto-remove dead sessions, then multi-select to prune more",
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := appFrom(cmd)
				if err != nil {
					return err
				}
				return app.Prune()
			},
		},
	)
	return root
}

func prompt(msg string) string {
	fmt.Print(msg + ": ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(line)
}

package session

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/koolkit/pkg/ui"
)

// ExecLauncher runs claude in dir with inherited stdio (claude is an
// interactive TUI, so its output must not be captured).
func ExecLauncher(dir string, args ...string) error {
	c := exec.Command("claude", args...)
	c.Dir = dir
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

// Scratch launches a throwaway, unsaved claude session in the current dir.
func (a App) Scratch() error {
	fmt.Fprintln(a.Out, "🚀 Starting scratch session...")
	return a.Launch("")
}

func (a App) launchNew(s store.Session) error {
	fmt.Fprintf(a.Out, "🚀 Starting %s...\n", ui.Cyan(s.Name))
	return a.Launch(s.Cwd, "--session-id", s.ID, "--name", s.Name)
}

func (a App) launchResume(s store.Session) error {
	fmt.Fprintf(a.Out, "🚀 Resuming %s...\n", ui.Cyan(s.Name))
	dir := s.Cwd
	if dir == "" {
		dir, _ = os.Getwd()
	}
	return a.Launch(dir, "--resume", s.ID)
}

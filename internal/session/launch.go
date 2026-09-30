package session

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/koolkit/pkg/ui"
)

// Fully-qualified model names for the -m alias.
const (
	modelOpus   = "us.anthropic.claude-opus-5-5[1m]"
	modelSonnet = "us.anthropic.claude-sonnet-5-5"
)

// ExecLauncher runs claude in dir with inherited stdio (claude is an
// interactive TUI, so its output must not be captured).
func ExecLauncher(dir string, args ...string) error {
	c := exec.Command("claude", args...)
	c.Dir = dir
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

// ModelFor maps the -m alias to a fully-qualified model name. An empty alias
// yields an empty name (no --model flag); an unknown alias is an error.
func ModelFor(alias string) (string, error) {
	switch alias {
	case "":
		return "", nil
	case "o":
		return modelOpus, nil
	case "s":
		return modelSonnet, nil
	default:
		return "", fmt.Errorf("unknown model %q (use s for sonnet or o for opus)", alias)
	}
}

// Review launches an unsaved session seeded with the /spiff-review prompt for
// url. It defaults to high effort on opus unless overridden by explicit flags.
func (a App) Review(url string) error {
	fmt.Fprintln(a.Out, "🚀 Starting review session...")
	args := append(a.extra(modelOpus, "high"), "/spiff-review "+url)
	return a.Launch("", args...)
}

// Scratch launches a throwaway, unsaved claude session in the current dir. It
// defaults to low effort on sonnet unless overridden by explicit flags.
func (a App) Scratch() error {
	fmt.Fprintln(a.Out, "🚀 Starting scratch session...")
	return a.Launch("", a.extra(modelSonnet, "low")...)
}

// extra returns the trailing --model/--effort flags for a launch, preferring
// the explicit App values and falling back to the per-command defaults.
func (a App) extra(defModel, defEffort string) []string {
	model, effort := a.Model, a.Effort
	if model == "" {
		model = defModel
	}
	if effort == "" {
		effort = defEffort
	}
	var args []string
	if model != "" {
		args = append(args, "--model", model)
	}
	if effort != "" {
		args = append(args, "--effort", effort)
	}
	return args
}

func (a App) launchNew(s store.Session) error {
	fmt.Fprintf(a.Out, "🚀 Starting %s...\n", ui.Cyan(s.Name))
	args := append([]string{"--session-id", s.ID, "--name", s.Name}, a.extra("", "")...)
	return a.Launch(s.Cwd, args...)
}

func (a App) launchResume(s store.Session) error {
	fmt.Fprintf(a.Out, "🚀 Resuming %s...\n", ui.Cyan(s.Name))
	dir := s.Cwd
	if dir == "" {
		dir, _ = os.Getwd()
	}
	args := append([]string{"--resume", s.ID}, a.extra("", "")...)
	return a.Launch(dir, args...)
}

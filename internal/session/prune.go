package session

import (
	"fmt"

	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/kk-claude/internal/transcript"
	"github.com/localshred/koolkit/pkg/ui"
)

// Prune auto-removes sessions whose transcript is gone (reporting them), then
// offers a multi-select over the survivors to prune more.
func (a App) Prune() error {
	sessions, err := store.Read()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Fprintln(a.Out, "📭 No sessions to prune.")
		return nil
	}
	if !transcript.ProjectsAvailable() {
		fmt.Fprintln(a.Out, "⚠️  Claude projects dir not found; skipping prune to avoid data loss.")
		return nil
	}
	var alive, dead []store.Session
	for _, s := range sessions {
		if transcript.Alive(s.ID) {
			alive = append(alive, s)
		} else {
			dead = append(dead, s)
		}
	}
	if len(dead) > 0 {
		fmt.Fprintln(a.Out, "🧹 Auto-pruned (transcript gone):")
		for _, s := range dead {
			fmt.Fprintf(a.Out, "   %s\n", ui.Dim(s.Name))
		}
	} else {
		fmt.Fprintln(a.Out, "✅ No dead sessions to auto-prune.")
	}
	if err := store.Write(alive); err != nil {
		return err
	}
	if len(alive) == 0 {
		fmt.Fprintln(a.Out, "📭 No sessions remain.")
		return nil
	}
	picked, err := a.PickMany(alive, func(s store.Session) string {
		return fmt.Sprintf("%s  %s", s.Name, s.Cwd)
	})
	if err != nil {
		return err
	}
	if len(picked) == 0 {
		fmt.Fprintln(a.Out, "👍 Kept all.")
		return nil
	}
	drop := map[string]bool{}
	for _, s := range picked {
		drop[s.ID] = true
	}
	var kept []store.Session
	for _, s := range alive {
		if !drop[s.ID] {
			kept = append(kept, s)
		}
	}
	if err := store.Write(kept); err != nil {
		return err
	}
	fmt.Fprintln(a.Out, "🧹 Pruned:")
	for _, s := range alive {
		if drop[s.ID] {
			fmt.Fprintf(a.Out, "   %s\n", ui.Dim(s.Name))
		}
	}
	return nil
}

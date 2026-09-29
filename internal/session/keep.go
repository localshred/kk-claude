package session

import (
	"fmt"

	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/kk-claude/internal/transcript"
	"github.com/localshred/koolkit/pkg/ui"
)

// Keep saves a session under args[0]. The id is args[1] when given, else the
// current session id inferred from cwd's newest transcript. Used to promote a
// scratch session once it is worth keeping.
func (a App) Keep(args []string, cwd string) error {
	var name string
	if len(args) > 0 {
		name = store.SanitizeName(args[0])
	}
	if name == "" {
		return fmt.Errorf("usage: kk-claude keep <name> [session-id]")
	}
	id := ""
	if len(args) > 1 {
		id = args[1]
	}
	if id == "" {
		id = transcript.CurrentSessionID(cwd)
	}
	if id == "" {
		return fmt.Errorf("could not determine the current session id")
	}
	sessions, err := store.Read()
	if err != nil {
		return err
	}
	if err := store.Write(store.Upsert(sessions, store.Session{Name: name, ID: id, Cwd: cwd})); err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "💾 Saved %s → %s\n", ui.Cyan(name), ui.Dim(id))
	return nil
}

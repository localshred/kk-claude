package session

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/localshred/kk-claude/internal/store"
)

const newMarker = "__NEW__"

// Pick fuzzy-selects a saved session to resume, or the leading New-session
// entry to start a named new session.
func (a App) Pick(cwd string) error {
	sessions, err := store.Read()
	if err != nil {
		return err
	}
	entries := append([]store.Session{{Name: "✨ New session", ID: newMarker}}, sessions...)
	sel, ok, err := a.PickOne(entries, func(s store.Session) string {
		if s.ID == newMarker {
			return s.Name
		}
		return fmt.Sprintf("%s  %s", s.Name, s.Cwd)
	})
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(a.Out, "❌ Nothing picked.")
		return nil
	}
	if sel.ID == newMarker {
		return a.newSession(cwd)
	}
	return a.launchResume(sel)
}

func (a App) newSession(cwd string) error {
	name := store.SanitizeName(a.Prompt("Session name"))
	if name == "" {
		fmt.Fprintln(a.Out, "❌ Cancelled (empty name).")
		return nil
	}
	sessions, err := store.Read()
	if err != nil {
		return err
	}
	for _, e := range sessions {
		if e.Name == name {
			return a.resumeOrReplace(e, cwd)
		}
	}
	s := store.Session{Name: name, ID: genID(), Cwd: cwd}
	if err := store.Write(store.Upsert(sessions, s)); err != nil {
		return err
	}
	return a.launchNew(s)
}

func (a App) resumeOrReplace(existing store.Session, cwd string) error {
	switch strings.ToLower(strings.TrimSpace(a.Prompt("Name exists — [r]eopen / [o]verwrite / [c]ancel"))) {
	case "r", "reopen":
		return a.launchResume(existing)
	case "o", "overwrite":
		s := store.Session{Name: existing.Name, ID: genID(), Cwd: cwd}
		sessions, err := store.Read()
		if err != nil {
			return err
		}
		if err := store.Write(store.Upsert(sessions, s)); err != nil {
			return err
		}
		return a.launchNew(s)
	default:
		fmt.Fprintln(a.Out, "❌ Cancelled.")
		return nil
	}
}

// genID returns a random RFC-4122-ish session id using only the standard
// library (no third-party uuid dependency).
func genID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Package store persists the curated Claude session list as
// `<name> <id> <cwd>` lines under the koolkit data dir.
package store

import (
	"os"
	"path/filepath"
	"strings"
)

// Session is one saved Claude session: a display name, the Claude session
// id, and the directory it was started in (may be empty).
type Session struct {
	Name string
	ID   string
	Cwd  string
}

// DataDir is koolkit's claude state dir, honoring XDG_DATA_HOME.
func DataDir() string {
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, _ := os.UserHomeDir()
		data = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(data, "koolkit", "claude")
}

// Path is the session list file.
func Path() string { return filepath.Join(DataDir(), "sessions.txt") }

// Read parses the session file, skipping blank and malformed lines. A
// missing file yields an empty slice.
func Read() ([]Session, error) {
	data, err := os.ReadFile(Path())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Session
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		s := Session{Name: fields[0], ID: fields[1]}
		if len(fields) >= 3 {
			s.Cwd = fields[2]
		}
		out = append(out, s)
	}
	return out, nil
}

// SanitizeName trims and collapses internal whitespace to hyphens so the
// name stays parseable as a single field.
func SanitizeName(name string) string {
	return strings.Join(strings.Fields(name), "-")
}

// Upsert removes any session sharing s.Name, then appends s.
func Upsert(sessions []Session, s Session) []Session {
	out := make([]Session, 0, len(sessions)+1)
	for _, e := range sessions {
		if e.Name != s.Name {
			out = append(out, e)
		}
	}
	return append(out, s)
}

// Write persists sessions as `<name> <id> <cwd>` lines (cwd omitted when
// blank), creating the data dir if needed.
func Write(sessions []Session) error {
	if err := os.MkdirAll(DataDir(), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, s := range sessions {
		parts := []string{s.Name, s.ID}
		if s.Cwd != "" {
			parts = append(parts, s.Cwd)
		}
		b.WriteString(strings.Join(parts, " "))
		b.WriteByte('\n')
	}
	return os.WriteFile(Path(), []byte(b.String()), 0o644)
}

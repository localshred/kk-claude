// Package transcript inspects Claude Code's per-project transcript files
// (`~/.claude/projects/<encoded-cwd>/<id>.jsonl`) to derive session ages,
// liveness, and the current session id.
package transcript

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var reNonAlnum = regexp.MustCompile(`[^A-Za-z0-9]`)

// Ages holds how long ago a session's transcript was created and last
// modified. OK is false when the transcript is missing.
type Ages struct {
	Created time.Duration
	Updated time.Duration
	OK      bool
}

// AgesOf returns creation/modification ages relative to now, OK=false when
// the transcript is gone.
func AgesOf(id string, now time.Time) Ages {
	path, ok := File(id)
	if !ok {
		return Ages{}
	}
	fi, err := os.Stat(path)
	if err != nil {
		return Ages{}
	}
	mod := fi.ModTime()
	return Ages{Created: now.Sub(mod), Updated: now.Sub(mod), OK: true}
}

// Alive reports whether a transcript exists for id.
func Alive(id string) bool {
	_, ok := File(id)
	return ok
}

// CurrentSessionID is the id of the newest transcript in the project dir for
// cwd, or "" when none exists. The active session's transcript is written
// continuously, so it sorts newest.
func CurrentSessionID(cwd string) string {
	dir := filepath.Join(ProjectsDir(), EncodeCwd(cwd))
	matches, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	var newest string
	var newestMod time.Time
	for _, m := range matches {
		fi, err := os.Stat(m)
		if err != nil {
			continue
		}
		if newest == "" || fi.ModTime().After(newestMod) {
			newest, newestMod = m, fi.ModTime()
		}
	}
	if newest == "" {
		return ""
	}
	return trimJSONL(filepath.Base(newest))
}

// EncodeCwd maps a path to Claude Code's project-dir name. Each
// non-alphanumeric character becomes a hyphen (runs are NOT collapsed),
// matching Claude Code's project-dir encoding (e.g. /Users/bj/.claude →
// -Users-bj--claude).
func EncodeCwd(cwd string) string { return reNonAlnum.ReplaceAllString(cwd, "-") }

// File returns the path to a session's `<id>.jsonl` under any project dir.
func File(id string) (string, bool) {
	matches, _ := filepath.Glob(filepath.Join(ProjectsDir(), "*", id+".jsonl"))
	if len(matches) == 0 {
		return "", false
	}
	return matches[0], true
}

// ProjectsAvailable reports whether Claude Code's projects dir exists.
// When it does not, transcript liveness cannot be judged and callers must
// not treat every session as dead.
func ProjectsAvailable() bool {
	fi, err := os.Stat(ProjectsDir())
	return err == nil && fi.IsDir()
}

// ProjectsDir is where Claude Code stores per-project transcripts.
func ProjectsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "projects")
}

func trimJSONL(base string) string {
	return strings.TrimSuffix(base, ".jsonl")
}

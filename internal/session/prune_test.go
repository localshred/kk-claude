package session

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/localshred/kk-claude/internal/store"
)

func liveSession(t *testing.T, name, id string) store.Session {
	t.Helper()
	proj := filepath.Join(os.Getenv("HOME"), ".claude", "projects", "p")
	os.MkdirAll(proj, 0o755)
	os.WriteFile(filepath.Join(proj, id+".jsonl"), []byte("{}"), 0o644)
	return store.Session{Name: name, ID: id}
}

func TestPruneAutoRemovesDead(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	alive := liveSession(t, "alive", "id-live")
	store.Write([]store.Session{alive, {Name: "dead", ID: "id-dead"}})
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		PickMany: func([]store.Session, func(store.Session) string) ([]store.Session, error) { return nil, nil },
	}
	if err := app.Prune(); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Read()
	if len(got) != 1 || got[0].Name != "alive" {
		t.Fatalf("dead should be auto-pruned, got %+v", got)
	}
}

func TestPruneMultiSelectRemovesChosen(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	a := liveSession(t, "keep", "id-keep")
	b := liveSession(t, "drop", "id-drop")
	store.Write([]store.Session{a, b})
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		PickMany: func(items []store.Session, _ func(store.Session) string) ([]store.Session, error) {
			for _, s := range items {
				if s.Name == "drop" {
					return []store.Session{s}, nil
				}
			}
			return nil, nil
		},
	}
	if err := app.Prune(); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Read()
	if len(got) != 1 || got[0].Name != "keep" {
		t.Fatalf("selected survivor should be pruned, got %+v", got)
	}
}

func TestPruneEmptyStore(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	app := App{Out: &out, Now: time.Now,
		PickMany: func([]store.Session, func(store.Session) string) ([]store.Session, error) {
			t.Fatal("must not pick on empty store")
			return nil, nil
		}}
	if err := app.Prune(); err != nil {
		t.Fatal(err)
	}
}

func TestPruneNoDeadReportsClean(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	a := liveSession(t, "alive1", "id-a1")
	b := liveSession(t, "alive2", "id-a2")
	store.Write([]store.Session{a, b})
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		PickMany: func(_ []store.Session, _ func(store.Session) string) ([]store.Session, error) { return nil, nil },
	}
	if err := app.Prune(); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "✅ No dead sessions to auto-prune.") {
		t.Fatalf("expected no-dead message, got: %s", output)
	}
}

func TestPruneSkipsWhenNoProjectsDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	a := store.Session{Name: "keepme", ID: "id-1"}
	b := store.Session{Name: "keepme2", ID: "id-2"}
	store.Write([]store.Session{a, b})
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		PickMany: func([]store.Session, func(store.Session) string) ([]store.Session, error) {
			t.Fatal("must not pick when projects dir is absent")
			return nil, nil
		},
	}
	if err := app.Prune(); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Read()
	if len(got) != 2 {
		t.Fatalf("no projects dir must leave store untouched, got %+v", got)
	}
	if !strings.Contains(out.String(), "skipping prune") {
		t.Fatalf("expected skip warning, got: %s", out.String())
	}
}

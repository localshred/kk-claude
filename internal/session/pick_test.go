package session

import (
	"bytes"
	"testing"
	"time"

	"github.com/localshred/kk-claude/internal/store"
)

func TestPickResumesSelected(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	store.Write([]store.Session{{Name: "proj", ID: "id-1", Cwd: "/tmp/proj"}})
	var c capture
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		Launch: func(dir string, args ...string) error { c.dir, c.args = dir, args; return nil },
		PickOne: func(items []store.Session, _ func(store.Session) string) (store.Session, bool, error) {
			for _, s := range items {
				if s.Name == "proj" {
					return s, true, nil
				}
			}
			return store.Session{}, false, nil
		},
	}
	if err := app.Pick("/tmp/cwd"); err != nil {
		t.Fatal(err)
	}
	if c.args[0] != "--resume" || c.args[1] != "id-1" {
		t.Fatalf("should resume selected, got %v", c.args)
	}
}

func TestPickScratchEntryLaunchesScratch(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	store.Write([]store.Session{{Name: "proj", ID: "id-1", Cwd: "/tmp/proj"}})
	var c capture
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		Launch: func(dir string, args ...string) error { c.dir, c.args = dir, args; return nil },
		PickOne: func(items []store.Session, _ func(store.Session) string) (store.Session, bool, error) {
			for _, s := range items {
				if s.ID == scratchMarker {
					return s, true, nil
				}
			}
			return store.Session{}, false, nil
		},
	}
	if err := app.Pick("/tmp/cwd"); err != nil {
		t.Fatal(err)
	}
	want := []string{"--model", modelSonnet, "--effort", "low"}
	if c.dir != "" || !equal(c.args, want) {
		t.Fatalf("scratch entry should launch scratch, got dir=%q args=%v", c.dir, c.args)
	}
}

func TestPickCancelIsClean(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	store.Write([]store.Session{{Name: "proj", ID: "id-1"}})
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		Launch: func(string, ...string) error { t.Fatal("must not launch on cancel"); return nil },
		PickOne: func([]store.Session, func(store.Session) string) (store.Session, bool, error) {
			return store.Session{}, false, nil
		},
	}
	if err := app.Pick("/tmp/cwd"); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Nothing picked")) {
		t.Fatalf("cancel output = %q", out.String())
	}
}

func TestPickNewSessionPromptsAndLaunches(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var c capture
	var out bytes.Buffer
	app := App{
		Out: &out, Now: time.Now,
		Launch: func(dir string, args ...string) error { c.dir, c.args = dir, args; return nil },
		Prompt: func(string) string { return "fresh proj" },
		PickOne: func(items []store.Session, _ func(store.Session) string) (store.Session, bool, error) {
			return items[0], true, nil
		},
	}
	if err := app.Pick("/tmp/new"); err != nil {
		t.Fatal(err)
	}
	if c.args[0] != "--session-id" || c.dir != "/tmp/new" {
		t.Fatalf("new session argv=%v dir=%q", c.args, c.dir)
	}
	got, _ := store.Read()
	if len(got) != 1 || got[0].Name != "fresh-proj" {
		t.Fatalf("new session not saved sanitized: %+v", got)
	}
}

func TestPickNewSessionOverwriteRegeneratesID(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	store.Write([]store.Session{{Name: "dup", ID: "old-id", Cwd: "/tmp/old"}})
	var c capture
	var out bytes.Buffer
	prompts := []string{"dup", "o"}
	i := 0
	app := App{
		Out: &out, Now: time.Now,
		Launch: func(dir string, args ...string) error { c.dir, c.args = dir, args; return nil },
		Prompt: func(string) string { p := prompts[i]; i++; return p },
		PickOne: func(items []store.Session, _ func(store.Session) string) (store.Session, bool, error) {
			return items[0], true, nil
		},
	}
	if err := app.Pick("/tmp/new"); err != nil {
		t.Fatal(err)
	}
	if c.args[0] != "--session-id" || c.args[1] == "old-id" {
		t.Fatalf("overwrite should regenerate id, got %v", c.args)
	}
}

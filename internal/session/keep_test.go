package session

import (
	"bytes"
	"testing"
	"time"

	"github.com/localshred/kk-claude/internal/store"
)

func TestKeepWithExplicitID(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	app := App{Out: &out, Now: time.Now}
	if err := app.Keep([]string{"myproj", "id-42"}, "/tmp/here"); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Read()
	if len(got) != 1 || got[0].Name != "myproj" || got[0].ID != "id-42" || got[0].Cwd != "/tmp/here" {
		t.Fatalf("kept = %+v", got)
	}
}

func TestKeepBlankNameErrors(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	app := App{Out: &out, Now: time.Now}
	if err := app.Keep([]string{"   "}, "/tmp"); err == nil {
		t.Fatal("blank name should error")
	}
}

func TestKeepNoIDNoTranscriptErrors(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	app := App{Out: &out, Now: time.Now}
	if err := app.Keep([]string{"name"}, "/tmp/no-transcript"); err == nil {
		t.Fatal("no id and no current transcript should error")
	}
}

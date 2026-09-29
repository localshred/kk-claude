package store

import (
	"os"
	"path/filepath"
	"testing"
)

func withTempData(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
}

func TestWriteReadRoundTrip(t *testing.T) {
	withTempData(t)
	in := []Session{
		{Name: "alpha", ID: "id-a", Cwd: "/tmp/a"},
		{Name: "beta", ID: "id-b", Cwd: ""},
	}
	if err := Write(in); err != nil {
		t.Fatal(err)
	}
	got, err := Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "alpha" || got[0].Cwd != "/tmp/a" {
		t.Fatalf("read = %+v", got)
	}
	if got[1].Name != "beta" || got[1].Cwd != "" {
		t.Fatalf("cwd-less entry = %+v", got[1])
	}
}

func TestReadSkipsMalformedAndBlank(t *testing.T) {
	withTempData(t)
	if err := os.MkdirAll(DataDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(Path(), []byte("good id-1 /tmp\n\nonlyname\n  \nname2 id-2\n"), 0o644)
	got, err := Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 valid lines, got %+v", got)
	}
}

func TestReadMissingFileIsEmpty(t *testing.T) {
	withTempData(t)
	got, err := Read()
	if err != nil || len(got) != 0 {
		t.Fatalf("missing file: got %+v err %v", got, err)
	}
}

func TestUpsertReplacesByName(t *testing.T) {
	in := []Session{{Name: "a", ID: "1"}, {Name: "b", ID: "2"}}
	out := Upsert(in, Session{Name: "a", ID: "9"})
	if len(out) != 2 || out[len(out)-1].ID != "9" {
		t.Fatalf("upsert = %+v", out)
	}
}

func TestSanitizeNameCollapsesWhitespace(t *testing.T) {
	if got := SanitizeName("  my  session\tname "); got != "my-session-name" {
		t.Fatalf("sanitize = %q", got)
	}
}

func TestPathUnderKoolkitData(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dir)
	if want := filepath.Join(dir, "koolkit", "claude", "sessions.txt"); Path() != want {
		t.Fatalf("Path = %q, want %q", Path(), want)
	}
}

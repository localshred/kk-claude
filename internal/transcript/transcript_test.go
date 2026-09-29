package transcript

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncodeCwd(t *testing.T) {
	if got := EncodeCwd("/Users/bj/code/src"); got != "-Users-bj-code-src" {
		t.Fatalf("encode = %q", got)
	}
}

func TestFileAndAliveFindTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	proj := filepath.Join(home, ".claude", "projects", "-tmp-x")
	os.MkdirAll(proj, 0o755)
	os.WriteFile(filepath.Join(proj, "sess-1.jsonl"), []byte("{}\n"), 0o644)

	if _, ok := File("sess-1"); !ok {
		t.Fatal("File should locate the transcript")
	}
	if !Alive("sess-1") {
		t.Fatal("Alive should be true")
	}
	if Alive("missing") {
		t.Fatal("Alive should be false for a missing id")
	}
}

func TestAgesOfMissingIsNotOK(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if AgesOf("nope", time.Now()).OK {
		t.Fatal("missing transcript must yield OK=false")
	}
}

func TestCurrentSessionIDPicksNewest(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cwd := "/tmp/proj"
	proj := filepath.Join(home, ".claude", "projects", EncodeCwd(cwd))
	os.MkdirAll(proj, 0o755)
	older := filepath.Join(proj, "old.jsonl")
	newer := filepath.Join(proj, "new.jsonl")
	os.WriteFile(older, []byte("{}"), 0o644)
	os.WriteFile(newer, []byte("{}"), 0o644)
	old := time.Now().Add(-time.Hour)
	os.Chtimes(older, old, old)

	if got := CurrentSessionID(cwd); got != "new" {
		t.Fatalf("current = %q, want \"new\"", got)
	}
	if got := CurrentSessionID("/nonexistent"); got != "" {
		t.Fatalf("no transcripts should give empty, got %q", got)
	}
}

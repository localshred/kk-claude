package session

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/koolkit/pkg/ui"
)

func testApp(out *bytes.Buffer) App {
	return App{
		Out:    out,
		Launch: func(dir string, args ...string) error { return nil },
		Now:    func() time.Time { return time.Unix(1000000, 0) },
	}
}

func TestListEmpty(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	if err := testApp(&out).List(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("No saved sessions")) {
		t.Fatalf("empty list output = %q", out.String())
	}
}

func TestListRendersNamesAndDashForMissingTranscript(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	store.Write([]store.Session{{Name: "zeta", ID: "id-z", Cwd: "/tmp/z"}})
	var out bytes.Buffer
	if err := testApp(&out).List(); err != nil {
		t.Fatal(err)
	}
	plain := ui.StripANSI(out.String())
	if !bytes.Contains([]byte(plain), []byte("zeta")) {
		t.Fatalf("missing session name: %q", plain)
	}
	if !bytes.Contains([]byte(plain), []byte("—")) {
		t.Fatalf("missing transcript should render em dash: %q", plain)
	}
}

func TestListSortsByName(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	proj := filepath.Join(os.Getenv("HOME"), ".claude", "projects", "p")
	os.MkdirAll(proj, 0o755)
	os.WriteFile(filepath.Join(proj, "id-b.jsonl"), []byte("{}"), 0o644)
	store.Write([]store.Session{
		{Name: "beta", ID: "id-b"},
		{Name: "alpha", ID: "id-a"},
	})
	var out bytes.Buffer
	testApp(&out).List()
	plain := ui.StripANSI(out.String())
	if bytes.Index([]byte(plain), []byte("alpha")) > bytes.Index([]byte(plain), []byte("beta")) {
		t.Fatalf("alpha should sort before beta: %q", plain)
	}
}

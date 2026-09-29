package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestListSubcommandEndToEnd(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	os.WriteFile(filepath.Join(home, ".claude-sessions.txt"), []byte("legacyproj id-1 /tmp/l\n"), 0o644)

	root := newRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte("legacyproj")) {
		t.Fatalf("list after migration should show legacy session: %q", out.String())
	}
}

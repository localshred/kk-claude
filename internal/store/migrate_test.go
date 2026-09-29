package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateCopiesLegacyWhenStoreAbsent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	legacy := filepath.Join(home, ".claude-sessions.txt")
	os.WriteFile(legacy, []byte("proj id-1 /tmp/proj\n"), 0o644)

	if err := Migrate(); err != nil {
		t.Fatal(err)
	}
	got, _ := Read()
	if len(got) != 1 || got[0].Name != "proj" {
		t.Fatalf("migrated store = %+v", got)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatal("legacy file must be left in place")
	}
}

func TestMigrateNoopWhenStoreExists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	os.WriteFile(filepath.Join(home, ".claude-sessions.txt"), []byte("legacy id-x\n"), 0o644)
	if err := Write([]Session{{Name: "current", ID: "id-c"}}); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(); err != nil {
		t.Fatal(err)
	}
	got, _ := Read()
	if len(got) != 1 || got[0].Name != "current" {
		t.Fatalf("migrate must not clobber existing store, got %+v", got)
	}
}

func TestMigrateNoopWhenNoLegacy(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if err := Migrate(); err != nil {
		t.Fatalf("migrate with no legacy should be a no-op, got %v", err)
	}
	if got, _ := Read(); len(got) != 0 {
		t.Fatalf("store should stay empty, got %+v", got)
	}
}

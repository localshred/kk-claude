package session

import (
	"bytes"
	"testing"
	"time"

	"github.com/localshred/kk-claude/internal/store"
)

type capture struct {
	dir  string
	args []string
}

func appWithCapture(c *capture, out *bytes.Buffer) App {
	return App{
		Out: out,
		Now: func() time.Time { return time.Unix(0, 0) },
		Launch: func(dir string, args ...string) error {
			c.dir, c.args = dir, args
			return nil
		},
	}
}

func TestLaunchNewBuildsArgv(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	if err := app.launchNew(store.Session{Name: "proj", ID: "id-1", Cwd: "/tmp/proj"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"--session-id", "id-1", "--name", "proj"}
	if c.dir != "/tmp/proj" || !equal(c.args, want) {
		t.Fatalf("launchNew dir=%q args=%v", c.dir, c.args)
	}
}

func TestLaunchResumeBuildsArgv(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	if err := app.launchResume(store.Session{Name: "p", ID: "id-2", Cwd: "/tmp/p"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"--resume", "id-2"}
	if c.dir != "/tmp/p" || !equal(c.args, want) {
		t.Fatalf("launchResume dir=%q args=%v", c.dir, c.args)
	}
}

func TestScratchLaunchesBareClaude(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	if err := app.Scratch(); err != nil {
		t.Fatal(err)
	}
	if len(c.args) != 0 {
		t.Fatalf("scratch should pass no args, got %v", c.args)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

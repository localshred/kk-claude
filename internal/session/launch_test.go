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

func TestScratchDefaultsLowSonnet(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	if err := app.Scratch(); err != nil {
		t.Fatal(err)
	}
	want := []string{"--model", modelSonnet, "--effort", "low"}
	if c.dir != "" || !equal(c.args, want) {
		t.Fatalf("scratch dir=%q args=%v, want %v", c.dir, c.args, want)
	}
}

func TestScratchExplicitFlagsOverrideDefault(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	app.Model, app.Effort = modelOpus, "high"
	if err := app.Scratch(); err != nil {
		t.Fatal(err)
	}
	want := []string{"--model", modelOpus, "--effort", "high"}
	if !equal(c.args, want) {
		t.Fatalf("scratch override args=%v, want %v", c.args, want)
	}
}

func TestReviewDefaultsHighOpusWithPrompt(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	if err := app.Review("https://x/pull/1"); err != nil {
		t.Fatal(err)
	}
	want := []string{"--model", modelOpus, "--effort", "high", "/spiff-review https://x/pull/1"}
	if c.dir != "" || !equal(c.args, want) {
		t.Fatalf("review dir=%q args=%v, want %v", c.dir, c.args, want)
	}
}

func TestLaunchNewThreadsExplicitFlags(t *testing.T) {
	var c capture
	var out bytes.Buffer
	app := appWithCapture(&c, &out)
	app.Model, app.Effort = modelSonnet, "medium"
	if err := app.launchNew(store.Session{Name: "p", ID: "id", Cwd: "/tmp/p"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"--session-id", "id", "--name", "p", "--model", modelSonnet, "--effort", "medium"}
	if !equal(c.args, want) {
		t.Fatalf("launchNew args=%v, want %v", c.args, want)
	}
}

func TestModelFor(t *testing.T) {
	cases := map[string]string{"": "", "s": modelSonnet, "o": modelOpus}
	for in, want := range cases {
		got, err := ModelFor(in)
		if err != nil || got != want {
			t.Errorf("ModelFor(%q) = (%q, %v), want %q", in, got, err, want)
		}
	}
	if _, err := ModelFor("x"); err == nil {
		t.Error("ModelFor(\"x\") should error")
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

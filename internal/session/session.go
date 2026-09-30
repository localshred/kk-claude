// Package session orchestrates the kk-claude subcommands: it reads the
// store, scans transcripts, drives the picker, and launches Claude. The
// side-effecting seams (picker, launcher, prompt, clock) are injected via
// App so the handlers are testable with fakes.
package session

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/localshred/kk-claude/internal/store"
	"github.com/localshred/kk-claude/internal/transcript"
	"github.com/localshred/koolkit/pkg/ui"
)

// Launcher runs the claude binary in dir with inherited stdio.
type Launcher func(dir string, args ...string) error

// App holds the injectable seams the command handlers depend on. Model and
// Effort are the resolved --model/--effort values (empty when unset), applied
// to launched sessions.
type App struct {
	Out      io.Writer
	Launch   Launcher
	PickOne  func([]store.Session, func(store.Session) string) (store.Session, bool, error)
	PickMany func([]store.Session, func(store.Session) string) ([]store.Session, error)
	Prompt   func(string) string
	Now      func() time.Time
	Model    string
	Effort   string
}

// List prints saved sessions alphabetically with live start/update ages.
func (a App) List() error {
	sessions, err := store.Read()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Fprintln(a.Out, "📭 No saved sessions.")
		return nil
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].Name < sessions[j].Name })
	rows := [][]string{{ui.Dim("NAME"), ui.Dim("STARTED"), ui.Dim("UPDATED"), ui.Dim("CWD"), ui.Dim("ID")}}
	for _, s := range sessions {
		ages := transcript.AgesOf(s.ID, a.Now())
		started, updated := ui.Dim("—"), ui.Dim("—")
		if ages.OK {
			started = ui.Dim(ui.Humanize(ages.Created))
			updated = ui.AgeCell(ages.Updated)
		}
		cwd := "—"
		if s.Cwd != "" {
			cwd = s.Cwd
		}
		rows = append(rows, []string{ui.Cyan(s.Name), started, updated, ui.Dim(cwd), ui.Dim(s.ID)})
	}
	ui.PrintTable(a.Out, rows)
	return nil
}

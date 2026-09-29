# kk-claude usage

`kk-claude` is invoked through `kk` as `kk claude [command]`. It can also be
run directly as `kk-claude [command]` if the binary is on `PATH` (useful for
local development).

## Commands

### `kk claude` (no subcommand)

Fuzzy-picks from a list that starts with a leading "New session" entry
followed by all saved sessions (name and cwd shown per row):

```
kk claude
```

- Picking "New session" prompts for a name (via stdin), sanitizes it
  (whitespace collapsed to hyphens), and:
  - If no saved session has that name: generates a new random session id,
    saves `<name> <id> <cwd>` to the store, and launches
    `claude --session-id <id> --name <name>` in the current directory.
  - If a saved session already has that name: prompts
    `Name exists — [r]eopen / [o]verwrite / [c]ancel`.
    - `r` / `reopen` resumes the existing session (see Resuming below).
    - `o` / `overwrite` generates a new id, replaces the stored entry for
      that name, and launches a new session under it.
    - anything else cancels.
- Picking a saved session resumes it (see Resuming below).
- Cancelling the picker (e.g. Esc) prints `Nothing picked.` and exits
  cleanly.

### `kk claude list`

Prints saved sessions in alphabetical order by name, as a table with
columns `NAME`, `STARTED`, `UPDATED`, `CWD`, `ID`:

```
kk claude list
```

`STARTED`/`UPDATED` are derived from the mtime of the session's transcript
file (see Session state below) and show as humanized durations (e.g.
`3h ago`). Both show `—` when no transcript is found for that session id.

### `kk claude scratch`

Launches Claude Code directly with no session id, name, or store entry —
nothing is recorded, so there is nothing to resume later:

```
kk claude scratch
```

### `kk claude keep <name> [session-id]`

Saves a session under `<name>` in the current working directory's store
entry:

```
kk claude keep my-feature
kk claude keep my-feature 1b2e4b8a-...
```

- `<name>` is required and is sanitized the same way as in the picker.
- `[session-id]` is optional. When omitted, `kk-claude` looks at
  `~/.claude/projects/<encoded-cwd>/*.jsonl` for the current directory and
  picks the most recently modified transcript, treating its filename (minus
  `.jsonl`) as the current session id. This is how a scratch session gets
  promoted to a saved one — it does not need to have been started via
  `kk-claude`.
- If no session id can be determined (no matching transcript), the command
  fails with `could not determine the current session id`.
- Saving under a name that already exists overwrites that entry (`Upsert`
  removes any existing entry with the same name before appending the new
  one).

### `kk claude prune`

Removes dead sessions (those whose transcript file no longer exists), then
offers a multi-select over the remaining sessions to remove more:

```
kk claude prune
```

- If Claude Code's projects directory (`~/.claude/projects`) does not
  exist, prune is skipped entirely to avoid mistaking "we can't check
  liveness" for "everything is dead."
- Sessions with no matching transcript are auto-removed and listed under
  "Auto-pruned."
- Survivors are then offered as a multi-select list (name and cwd per row);
  anything selected there is also removed and listed under "Pruned."
- If nothing is selected in the multi-select, existing sessions are kept
  ("Kept all").

## Session state

Saved sessions live in a plain text file at:

```
$XDG_DATA_HOME/koolkit/sessions/claude.txt
```

or, when `XDG_DATA_HOME` is unset:

```
~/.local/share/koolkit/sessions/claude.txt
```

Each line is space-separated: `<name> <id> <cwd>` (the trailing `<cwd>` is
omitted for lines with no recorded directory). Writes go through a
temp-file-plus-rename so a crash mid-write cannot corrupt the file.

On first run, if this file does not yet exist but a legacy file at
`~/.claude-sessions.txt` does, `kk-claude` copies the legacy file into the
new location once. It never overwrites an existing store and never touches
the legacy file.

## How resuming works

A saved session is resumed by directory and Claude Code's own session id.
`kk-claude` does not read or replay transcript contents itself — it defers
entirely to `claude --resume <id>`:

- The launch directory is the session's saved `cwd`; if that is empty, the
  current working directory is used instead.
- The command run is `claude --resume <id>`, with stdio inherited so
  Claude Code's interactive TUI works normally.

Liveness (`list`'s ages, `prune`'s dead-session detection) is inferred by
looking for a `<id>.jsonl` file anywhere under
`~/.claude/projects/*/`. The project subdirectory name itself follows
Claude Code's own encoding of a working directory: every character that is
not a letter or digit becomes a hyphen (runs of separators are not
collapsed), so `/Users/bj/.claude` becomes `-Users-bj--claude`.

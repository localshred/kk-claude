# kk-claude

`kk-claude` is a [koolkit](https://github.com/localshred/koolkit) plugin for
managing [Claude Code](https://claude.com/claude-code) sessions. It keeps a
small named list of sessions you want to come back to, lets you fuzzy-pick
one to resume, and can launch a throwaway session when you don't want to
save anything.

It works by reading Claude Code's own transcript files under
`~/.claude/projects/` to check whether a saved session is still alive and to
report its start/update times.

## Install

Requires [`kk`](https://github.com/localshred/koolkit) (the koolkit plugin
manager):

```
kk install claude
```

## Usage

Run with no arguments to fuzzy-pick a saved session to resume, or start a
new named one:

```
kk claude
```

Other subcommands:

```
kk claude list             # print saved sessions, alphabetically
kk claude scratch          # launch a throwaway, unsaved session
kk claude keep <name>      # save the current session under <name>
kk claude prune            # remove dead sessions, then multi-select more
```

See [docs/usage.md](docs/usage.md) for the full command reference,
including flags, where session state lives, and how resume works.

## Requirements

`kk-claude` declares the following runtime dependency (checked by
`kk doctor`):

- `claude` — the Claude Code CLI, used to start and resume sessions

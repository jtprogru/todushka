# todushka

Terminal Things 3: a keyboard-first TUI todo manager with the Things 3 workflow (Inbox / Today / Upcoming / Anytime / Someday + Areas + Projects + tags + repeats) and an embedded local store. Pure Go, no CGO, single static binary.

## Install

```sh
go install github.com/jtprogru/todushka/cmd/todushka@latest
```

Or build from source:

```sh
git clone https://github.com/jtprogru/todushka
cd todushka
task build           # -> bin/todushka
```

## Quick start

Launch the TUI:

```sh
todushka
```

Add a task without opening the TUI:

```sh
todushka add "buy milk" --tag shop --deadline 2026-06-01
todushka today
todushka complete a3fx7q
```

Export and import the local database as JSON:

```sh
todushka export > backup.json
todushka import --json backup.json --yes
```

## Data location

`todushka` stores its data following the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).

| Path                                                | Contents                          |
|-----------------------------------------------------|-----------------------------------|
| `$XDG_DATA_HOME/todushka/db` (or `~/.local/share/…`) | bbolt database                    |
| `$XDG_STATE_HOME/todushka/log` (or `~/.local/state/…`) | rotating error log               |

## Keybindings

Documented in the TUI (`?` for context-sensitive help). Default global hotkeys:

| Key            | Action                       |
|----------------|------------------------------|
| `1` .. `6`     | Switch system list (Inbox/Today/Upcoming/Anytime/Someday/Logbook) |
| `j` / `↓`      | Move cursor down             |
| `k` / `↑`      | Move cursor up               |
| `n`            | Open Quick Entry             |
| `Enter`        | Open editor for the selected task |
| `c`            | Complete the selected task   |
| `?`            | Toggle help                  |
| `q` / `Ctrl+C` | Quit                         |

## CLI reference

```sh
todushka add "<title>" [--project <name>] [--area <name>] [--tag <name>] [--start YYYY-MM-DD] [--deadline YYYY-MM-DD] [--someday]
todushka today [--json]
todushka complete <short-id>
todushka area    add|list|delete [--force]
todushka project add|list|delete [--area <name>] [--deadline YYYY-MM-DD] [--auto-close]
todushka export  [--json <path>]
todushka import  --json <path> --yes
```

Run `todushka` without arguments to open the TUI. Press `?` in the TUI for context-sensitive help.

## Themes

The TUI ships with Catppuccin palettes. Pick one via the `TODUSHKA_THEME` environment variable:

| Value           | Palette                      |
|-----------------|------------------------------|
| (unset, dark)   | Catppuccin Macchiato (default) |
| `light`, `latte`| Catppuccin Latte             |
| `NO_COLOR=1`    | Monochrome (bold/underline)  |

## Limitations (v1)

- **No sync** — data lives in a single local bbolt file.
- **No SIGTERM cleanup** — the CLI relies on bbolt's transactional safety; abrupt termination during a long-running `import` or `export` cannot corrupt the database, but no graceful "finishing current op" handler is wired. Avoid killing the process mid-import.
- **No TUI screens for Areas / Projects** — manage them via the `area` / `project` CLI subcommands. Tasks created in the TUI go to Inbox unless edited later.
- **No repeat-rule UI** — recurring tasks are editable via JSON import or programmatic API. The TUI editor covers Title / Notes / dates / Tags / Someday.
- **Logbook retention is unlimited.**

## Status

v1 — local-only, no sync, no calendar integration. See `.spec/features/todo-tui/` for the full feature plan.

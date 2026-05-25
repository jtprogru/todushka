# TUI Shell — Requirements

**Status:** Draft
**Author:** spec-driven-dev (Claude)
**Date:** 2026-05-25

## Overview

Расширение TUI до production-уровня "shell": (a) full-screen rendering — TUI занимает всю высоту/ширину терминала без пустого пространства; (b) header показывает `(1) Inbox [24]`-style сегменты с keybind-prefix, label и count; (c) footer в zellij-style — mode chip слева + контекстные key hints, разделённые `│`; (d) YAML-конфиг с env-var и CLI-flag overrides, auto-create при первом запуске. Реализация изолирована в `internal/tui` + extends `internal/config`. Storage / app / domain слои не затрагиваются. Backward-compat: при default config поведение визуально совместимо с v0.3.0 (с улучшенным header/footer).

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| Body Height | Доступная высота для содержимого: `m.height - viewHeader_height - viewFooter_height`. | `internal/tui` (вычисляется в `View()`) |
| List Count | Количество задач в конкретном `listKind`, кэшированное в `Model.listCounts`. Обновляется через `Repository` calls в Cmd. | `internal/tui` (`Model.listCounts map[listKind]int`) |
| Header Segment | Один из 6 visual блоков в header'е, формата `(N) Name [Count]`. N — keybind digit (1-6); active segment — inverted background. | `internal/tui` (новая функция `renderHeaderSegment`) |
| Compact Header | Mode рендера header'а при `m.width < 80`: `(N)I[Count]` где I = первая буква label. | `internal/tui` (флаг в `viewHeader`) |
| Shell Mode | Семантическое состояние TUI: `HELP`, `EDITOR`, `CONFIRM`, `FILTER`, `SELECT`, `NORMAL`. Detected from Model state (priority order). | `internal/tui` (новая функция `currentMode`) |
| Mode Chip | Визуальный indicator текущего Mode в footer: coloured background + mode label (e.g. `-- FILTER --`). | `internal/tui` (новая функция `renderModeChip`) |
| Key Hint | Один пункт в footer формата `key: action`. Per-mode hint sets — см. REQ-3.x. | `internal/tui` (новая функция `modeKeyHints`) |
| App Config | Структура `Config` с настройками: theme, dual_pane_min_width, list_pane_share, bulk_confirm_threshold, notes_max_lines. | `internal/config` (новый тип `AppConfig`) |
| Config Precedence | Порядок resolve: CLI flag `--config` > env vars `TODUSHKA_*` > config file > code defaults. | `internal/config` (loader logic) |

## User Stories

- Как **пользователь с большим терминалом (40+ rows)**, я хочу чтобы TUI занимал всю доступную высоту, а не оставлял пустоту снизу — это создаёт ощущение "полноценного приложения".
- Как **новый пользователь**, я хочу видеть в header'е какие клавиши переключают списки и сколько задач в каждом — без необходимости открывать help.
- Как **пользователь zellij/tmux**, я хочу узнаваемый mode-aware footer с подсказками текущего контекста (filter / select / normal) — снижает порог входа.
- Как **пользователь с предпочтениями** (порог dual-pane, тема, etc.), я хочу настраивать поведение через `~/.config/todushka/config.yaml` без перекомпиляции — стандарт для CLI-tools.
- Как **тестировщик / power-user**, я хочу указывать альтернативный config через `--config <path>` для experimentation без перезаписи defaults.

## Requirements

### Group 1 — Full-screen rendering

**REQ-1.1** WHEN the TUI receives `tea.WindowSizeMsg`, the system SHALL store BOTH `msg.Width` in `Model.width` AND `msg.Height` in `Model.height`.

**REQ-1.2** WHEN `m.height >= 10` AND `m.width >= 40` AND `m.screen != screenEditor`, the system SHALL render `View()` such that `lipgloss.Height(View()) == m.height` (full-screen body clamp via `lipgloss.NewStyle().Height(bodyH).Render(body)`).

**REQ-1.3** WHEN `m.height < 10` OR `m.width < 40` (very small terminal) OR `m.height == 0` (no WindowSizeMsg yet), the system SHALL render in legacy mode (no height clamp) — preserves backward compat for tests and tiny terminals.

**REQ-1.4** WHEN `m.screen == screenEditor`, the system SHALL render editor at its existing `editorWidth()` size without applying full-screen clamp (editor controls own dimensions).

### Group 2 — Header counts and indicators

**REQ-2.1** WHEN rendering the header AND `m.width >= 80`, the system SHALL render each of the 6 list segments as `(N) Name [Count]` where:
- `N` is the digit keybinding (1 for Inbox, 2 for Today, …, 6 for Logbook),
- `Name` is the list label (`Inbox`, `Today`, etc.),
- `Count` is the cached `m.listCounts[listKind]` value.

**REQ-2.2** WHEN rendering a non-active Header Segment AND not in compact mode, the system SHALL style `(N)` with `theme.Accent` foreground, `Name` with `theme.Subtext`, and `[Count]` with `theme.Dim`.

**REQ-2.3** WHEN rendering the **active** Header Segment (matching `m.activeList`), the system SHALL apply `theme.Header` style (inverted background + bold) to the entire segment, overriding individual sub-styles.

**REQ-2.4** WHEN `m.width < 80` (compact mode), the system SHALL render each segment as `(N)I[Count]` where `I` is the first letter of the list name (I, T, U, A, S, L).

**REQ-2.5** WHEN `m.tasks` is updated via `tasksLoadedMsg`, the system SHALL trigger a Cmd that fetches counts for all 6 lists via `Service.ListInbox/ListToday/.../ListLogbook` AND emits `countsLoadedMsg{counts map[listKind]int}`.

**REQ-2.6** WHEN `countsLoadedMsg` is received, the system SHALL store the counts in `Model.listCounts`, replacing any previous values.

**REQ-2.7** WHEN rendering a Header Segment AND the corresponding list count is not yet populated in `m.listCounts`, the system SHALL display `[?]` instead of `[Count]` (initial state before first refresh).

**REQ-2.8** WHEN a list is empty, the system SHALL display `[0]` (not hide it).

### Group 3 — Status line (zellij-style)

**REQ-3.1** WHEN rendering the footer, the system SHALL produce output of the form `<mode-chip> │ <hint1> │ <hint2> │ … │ <status-message?>` where pipes are visually rendered as ` │ ` separator.

**REQ-3.2** WHEN computing current Shell Mode, the system SHALL apply the following priority order (first match wins):
1. `screen == screenHelp` → `HELP`
2. `screen == screenEditor` → `EDITOR`
3. `confirm != nil` → `CONFIRM`
4. `filtering == true` → `FILTER`
5. `len(selected) > 0` → `SELECT`
6. otherwise → `NORMAL`

**REQ-3.3** WHEN rendering the Mode Chip, the system SHALL style it as `-- <MODE> --` with `theme.Header` background (accent color) and bold text.

**REQ-3.4** WHEN current mode is `NORMAL`, the system SHALL display these key hints: `/: filter`, `space: select`, `n: quick`, `↵: edit`, `c: complete`, `?: help`, `q: quit`.

**REQ-3.5** WHEN current mode is `FILTER`, the system SHALL display the filter input `Filter: <query>_` followed by hints: `↵: save`, `esc: cancel`.

**REQ-3.6** WHEN current mode is `SELECT`, the system SHALL display hints: `c/x/d/p: bulk`, `*: all`, `esc: clear`, `Selected: <N>`.

**REQ-3.7** WHEN current mode is `CONFIRM`, the system SHALL display hints: `y: yes`, `any: cancel`.

**REQ-3.8** WHEN current mode is `EDITOR`, the system SHALL display hints: `Tab: next field`, `Shift+Tab: prev`, `Ctrl+S: save`, `esc: cancel`.

**REQ-3.9** WHEN current mode is `HELP`, the system SHALL display hint: `?: close`.

**REQ-3.10** WHEN `m.statusMsg` is non-empty, the system SHALL append the status message at the right side of the footer styled with `theme.StatusError` (red) for errors or `theme.StatusInfo` (green) for info — preserving current behavior.

### Group 4 — Configuration

**REQ-4.1** WHEN the application starts, the system SHALL resolve the config file path with the following precedence:
1. `--config <path>` CLI flag (root cobra command, persistent flag),
2. `TODUSHKA_CONFIG` env var,
3. Default: `$XDG_CONFIG_HOME/todushka/config.yaml` (fallback to `$HOME/.config/todushka/config.yaml`).

**REQ-4.2** WHEN the resolved config file does NOT exist, the system SHALL create the parent directory (mode 0750) AND write a YAML file with defaults AND inline comments documenting each setting.

**REQ-4.3** WHEN the resolved config file exists AND is valid YAML, the system SHALL parse it into `AppConfig` struct with these fields and defaults:
- `theme: "macchiato"` (string; values: macchiato | latte | mono | <empty=auto>)
- `dual_pane_min_width: 100` (int)
- `list_pane_share: 0.45` (float)
- `bulk_confirm_threshold: 5` (int)
- `notes_max_lines: 8` (int)

**REQ-4.4** WHEN the config file contains unknown fields, the system SHALL ignore them silently (forward-compat for future settings).

**REQ-4.5** WHEN the config file contains an invalid value (e.g. `theme: unknown`, `notes_max_lines: -5`), the system SHALL log a warning to `LogPath()` AND use the default value for that field.

**REQ-4.6** WHEN any of the following env vars is set, the system SHALL use its value (parsed to the field's type) instead of the config-file value:
- `TODUSHKA_THEME` → `theme`
- `TODUSHKA_DUAL_PANE_MIN_WIDTH` → `dual_pane_min_width`
- `TODUSHKA_LIST_PANE_SHARE` → `list_pane_share`
- `TODUSHKA_BULK_CONFIRM_THRESHOLD` → `bulk_confirm_threshold`
- `TODUSHKA_NOTES_MAX_LINES` → `notes_max_lines`

**REQ-4.7** WHEN an env var contains an invalid value (e.g. `TODUSHKA_NOTES_MAX_LINES=abc`), the system SHALL log a warning AND fall back to the next precedence level (config file or default).

**REQ-4.8** WHEN constructing the TUI Model, the system SHALL inject the resolved `AppConfig` AND use its values for:
- `dualPaneMinWidth` constant in `details.go` (replace `const` with `m.config.DualPaneMinWidth`)
- `listPaneShare` constant
- `bulkConfirmThreshold` constant in `bulk.go`
- `detailsNotesMaxLines` constant
- `theme` selected via `SelectTheme` using `config.Theme` (instead of only `TODUSHKA_THEME` env)

### Group 5 — Backward compatibility

**REQ-5.1** WHEN `AppConfig` is zero-valued (defaults), the system SHALL produce visually identical output to v0.3.0 for the same tasks except for new header/footer chrome.

**REQ-5.2** WHEN existing tests run without explicit config override, the system SHALL use defaults — no test signature changes required (NewModel signature may add `cfg AppConfig` parameter; tests can pass `AppConfig{}` zero-value or use a `NewModelWithDefaults` helper).

**REQ-5.3** WHEN feature complete, the existing 117 TUI tests SHALL continue to pass (with adjustments only where viewHeader/viewFooter assertions become obsolete due to new format — updates to those assertions are acceptable).

## Topological Order

```
Group 4 (Config)         — foundation; values are needed by Groups 1, 2, 3 thresholds
Group 1 (Full-screen)    — depends on m.height + clamp logic; independent of header/footer content
Group 2 (Header counts)  — depends on Group 4 (theme), independent of Group 1 height calc
Group 3 (Status line)    — depends on Group 4 (theme), Group 2 (mode detection)
Group 5 (Backward compat) — cross-cutting; verified after all other groups
```

Reason: Config provides the configurable thresholds (dual-pane, bulk confirm, notes lines) and theme. Full-screen and header/footer can develop in parallel after Config is in place.

## Conflict Priority

**REQ-1.2 (height clamp) vs REQ-1.4 (editor own dimensions):**
Not a conflict — explicit branches by `m.screen`. REQ-1.4 takes priority for editor screen.

**REQ-2.5 (counts refresh on tasksLoadedMsg) vs REQ-4.x (counts could be Cmd-blocking):**
Not a conflict — counts refresh runs as `tea.Cmd` (async), non-blocking. Initial state is `[?]` per REQ-2.7.

**REQ-3.2 (mode priority order):**
Internal priority resolution — first match wins. Defined explicitly to prevent ambiguity.

**REQ-4.6 (env > config) vs REQ-4.8 (config injected into Model):**
Not a conflict — env override happens during config load (before Model construction).

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Should config file format also support TOML / JSON? | YAML is recommended (REQ-4.x), but design may want to abstract via `Loader` interface to enable future TOML. v1 = YAML only. | REQ-4.3 |
| Mode chip width: fixed (always 10 chars padded) or variable (length of mode name + ` -- ` decorations)? | Affects header alignment. Fixed gives stable layout; variable is more compact. | REQ-3.3 |
| Counts refresh Cmd: one batch (single Cmd doing 6 ListXxx calls sequentially) or `tea.Batch(6 cmds)`? | Sequential simpler (one nameCacheLoadedMsg style); batch parallel. bbolt locking serializes anyway — sequential is simpler. | REQ-2.5 |
| Auto-create config file: write actual values or template with all keys commented out (user uncomments to override)? | Commented template hints "you can change this"; actual values look like committed choices. | REQ-4.2 |
| `--config` flag visibility: hidden from help (advanced) or visible? | Stand для CLI tools is visible. | REQ-4.1 |

## Verification Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test (race)| `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

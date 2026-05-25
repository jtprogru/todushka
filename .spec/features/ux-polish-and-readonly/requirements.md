# UX Polish (v0.5.0) — Requirements

**Status:** Draft
**Author:** spec-driven-dev (Claude)
**Date:** 2026-05-25

## Scope clarification

Per Phase 1 explore + user decision: фича разделена.

- **v0.5.0 `ux-polish`** (этот документ) — items 1, 2, 3, 5: theme auto-detect, refresh after edit, Anytime в editor, section borders.
- **v0.6.0 `readonly-mode`** (отдельный pipeline) — item 4: read-only режим для second instance.

Branch name `ux-polish-and-readonly` остаётся для исторической ясности pipeline'а; commits ясно укажут v0.5.0 scope.

## Overview

Четыре когезивных TUI-улучшения:

1. **Theme auto-detect** — `config.theme = "auto"` или `"system"` определяет OS dark/light setting; macOS через `defaults read -g AppleInterfaceStyle`, Linux через gsettings, остальное → fallback dark.
2. **Instant refresh after edit** — editorSavedMsg inline-splices обновлённую task в `m.tasks` перед async refresh; пользователь видит изменения немедленно.
3. **Anytime toggle в editor** — Someday checkbox заменён на 2-state `When: [Anytime] / [Someday]` radio-style; Space переключает.
4. **Section borders** — тонкие `─` separator'ы между header/body и body/footer для визуального разделения.

Реализация изолирована в `internal/tui/`, `internal/config/`. Storage/app/domain слои не затрагиваются.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| Auto Theme Detection | Платформо-специфичная процедура определения dark/light setting OS. Триггерится когда `config.Theme ∈ {"auto", "system"}`. | `internal/tui/detect_dark_*.go` (новые platform files) |
| Detected Theme | Результат Auto Detection: `"macchiato"` (dark), `"latte"` (light), либо `"macchiato"` если detection failed. | `internal/tui` (helper function) |
| Edit Splice | Операция inline-замены обновлённой задачи в `m.tasks` по ID без обращения к storage. | `internal/tui/app.go` editorSavedMsg case |
| When Toggle | 2-state radio в editor: Anytime ↔ Someday. Space переключает. Internal mapping: Anytime → `task.Someday = false`; Someday → `task.Someday = true`. | `internal/tui/editor.go` (`when` field) |
| Section Separator | Тонкий `─` line glyph, повторённый на всю ширину терминала, разделяющий header / body / footer. Styled через `theme.Help` foreground. | `internal/tui` (новый helper `renderSeparator`) |

## User Stories

- Как **пользователь, переключающий OS dark/light**, я хочу чтобы todushka следовал OS-настройке автоматически — без правки config или env var.
- Как **пользователь, редактирующий задачу**, я хочу видеть обновлённый title/notes сразу после `Ctrl+S` — без необходимости нажимать Tab для refresh.
- Как **пользователь, работающий с Anytime-bucket**, я хочу явный toggle в editor — симметрично Someday — чтобы понимать какая когорта у задачи.
- Как **пользователь, смотрящий на UI**, я хочу видеть визуально различные блоки header / body / footer — без сливающихся текстов.

## Requirements

### Group 1 — Theme auto-detect

**REQ-1.1** WHEN `AppConfig.Theme == "auto"` OR `AppConfig.Theme == "system"`, the system SHALL trigger Auto Theme Detection at TUI initialization AND use the result as the active theme.

**REQ-1.2** WHEN Auto Theme Detection runs on Darwin (macOS), the system SHALL execute `defaults read -g AppleInterfaceStyle` with a 500ms timeout. If output equals `"Dark"` → use macchiato; if command exits non-zero or returns other output → use latte.

**REQ-1.3** WHEN Auto Theme Detection runs on Linux, the system SHALL execute `gsettings get org.gnome.desktop.interface color-scheme` with a 500ms timeout. If output contains `"dark"` (case-insensitive) → use macchiato; if `"light"` → use latte; on error → fallback to macchiato.

**REQ-1.4** WHEN Auto Theme Detection runs on any other platform (Windows, BSD, etc.), the system SHALL skip detection AND use macchiato by default.

**REQ-1.5** WHEN `AppConfig.Validate` receives `Theme == "auto"` or `"system"`, the system SHALL treat both as valid values (no warning emitted).

**REQ-1.6** WHEN `NO_COLOR` env is set, the system SHALL use monochrome theme regardless of `AppConfig.Theme` value — NO_COLOR takes absolute precedence (REQ-1.1 не отменяет).

**REQ-1.7** WHEN Auto Theme Detection completes (success or failure), the system SHALL NOT block TUI startup — execution time is bounded by the 500ms timeout per REQ-1.2/1.3.

### Group 2 — Instant refresh after edit

**REQ-2.1** WHEN `editorSavedMsg{updated: t}` is processed, the system SHALL find the task with `t.ID` in `m.tasks` AND replace it in-place (preserving slice index), producing immediate visual update on the next View() render.

**REQ-2.2** WHEN `editorSavedMsg` is processed, the system SHALL also fire a `tea.Batch` Cmd containing `loadCurrentList` AND `fetchListCounts` for async refresh of sort order, header counts, and name caches.

**REQ-2.3** WHEN `editorSavedMsg` is processed AND `t.ID` is not found in `m.tasks` (e.g., a different list is active), the system SHALL skip the inline splice silently AND rely on the async `loadCurrentList` to surface the change later.

**REQ-2.4** WHEN the inline splice succeeds, the resulting `m.tasks` slice SHALL have the same length as before (no insertion / no deletion).

### Group 3 — Anytime toggle в editor

**REQ-3.1** WHEN the editor renders the `When` section, the system SHALL display 2 radio-style lines:
```
[•] Anytime
[ ] Someday
```
or, if Someday is currently selected:
```
[ ] Anytime
[•] Someday
```

**REQ-3.2** WHEN the `When` field is focused (Tab order: after Tags), the system SHALL highlight it via `theme.Selected` (matching current Someday focus visual).

**REQ-3.3** WHEN `Space` is pressed while the `When` field is focused, the system SHALL toggle between Anytime and Someday selection.

**REQ-3.4** WHEN the editor saves via `Ctrl+S`, the system SHALL map `When == Anytime` to `task.Someday = false` AND `When == Someday` to `task.Someday = true`.

**REQ-3.5** WHEN the editor renders `When == Anytime` AND the task has `AreaID == nil AND ProjectID == nil`, the system SHALL display an inline hint:
```
(will appear in Inbox without Area/Project)
```
in `theme.Dim` (subtext) style below the `[•] Anytime / [ ] Someday` block.

**REQ-3.6** WHEN existing tests assert on editor Someday behavior (e.g., `TestTUI_EditorTabCyclesFields`), the system SHALL preserve Tab navigation (field count unchanged: title, notes, start, deadline, tags, when).

### Group 4 — Section borders

**REQ-4.1** WHEN `View()` renders in full-screen mode (`m.height >= 10 AND m.width >= 40`), the system SHALL insert a Section Separator immediately below `viewHeader()` (i.e., between header and body).

**REQ-4.2** WHEN `View()` renders in full-screen mode, the system SHALL insert a Section Separator immediately above `viewFooter()` (i.e., between body and footer).

**REQ-4.3** WHEN rendering a Section Separator, the system SHALL produce a line of `─` characters with length equal to `m.width`, styled via `theme.Help.Render(...)` (subtext foreground, no background).

**REQ-4.4** WHEN `View()` renders in legacy mode (`m.height < 10 OR m.width < 40`), the system SHALL omit Section Separators (preserve current legacy behavior).

**REQ-4.5** WHEN `m.screen == screenEditor`, the system SHALL omit Section Separators (editor controls its own rendering).

**REQ-4.6** WHEN Section Separators are present, the full-screen height calculation SHALL allocate exactly 1 row per separator (so `bodyH = m.height - viewHeader_h - 1 - 1 - viewFooter_h`).

### Group 5 — Backward compatibility

**REQ-5.1** WHEN existing test fixtures use `config.Defaults()`, the system SHALL preserve current visual behavior for all 174 tests (with adjustments only for the obsolete Someday checkbox label).

**REQ-5.2** WHEN `AppConfig.Theme` is set explicitly (e.g., `"macchiato"`), Auto Theme Detection SHALL be skipped (REQ-1.1 only triggers for `auto`/`system`).

**REQ-5.3** WHEN editor field count is queried (e.g., for Tab cycling), the system SHALL report 6 fields (title, notes, start, deadline, tags, when) — same as before (Someday field renamed to when, not removed).

## Topological Order

```
Group 1 (Theme auto)        — independent foundation; new platform-specific files
Group 4 (Section borders)   — independent; pure visual change in View()
Group 2 (Refresh after edit) — independent; Update case change
Group 3 (Anytime toggle)    — depends on existing editor structure; rename field
Group 5 (Backward compat)   — cross-cutting; verified after all
```

Reason: 4 groups are mostly independent and can develop in parallel. Group 3 has slight dependency on existing editor file structure but doesn't block anything.

## Conflict Priority

**REQ-1.6 (NO_COLOR overrides) vs REQ-1.1 (auto detection):**
NO_COLOR имеет absolute precedence. Если установлен, не запускаем detection вообще — экономия времени старта.

**REQ-4.1/4.2 (separators in full-screen) vs REQ-4.4 (no separators in legacy):**
Не конфликт — explicit branches на `m.height >= 10 AND m.width >= 40`. Только один путь активен на каждый render.

**REQ-2.1 (inline splice) vs REQ-2.2 (async loadCurrentList):**
Не конфликт — оба happen at once. Inline splice даёт immediate update; async loadCurrentList дополняет refresh sort/counts.

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Auto detection caching: stick result в memory на весь session vs re-detect при resize или refresh? | Re-detect добавляет latency; sticky может расходиться с runtime OS change. | REQ-1.1 |
| Linux detection: только `gsettings` или включить `kreadconfig5` (KDE) и `$GTK_THEME` parse? | Покрытие vs complexity. | REQ-1.3 |
| Refresh after edit: должен ли `nameCacheLoadedMsg` дозагрузить tags для нового task'а синхронно перед next render? | Inline splice использует name cache; если новые tags не в cache — fallback short-ID на 1 frame. Acceptable? | REQ-2.1 |
| Section separator color: единый `theme.Help` для обоих или разные (header/footer accent vs body subtext)? | Visual consistency vs hierarchy hint. | REQ-4.3 |
| Editor When toggle: дополнительные options (Today / Specific Date) или строго `Anytime/Someday`? | Покрытие vs scope. | REQ-3.1 |

## Verification Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test (race)| `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

# Dual-Pane Layout for TUI — Requirements

**Status:** Draft
**Author:** spec-driven-dev (Claude)
**Date:** 2026-05-25

## Overview

При ширине терминала ≥ 100 колонок и `screen == screenList` (без editor / help / confirm modal) TUI рендерит горизонтальный split: слева — список задач (текущий viewList с полным функционалом filter/select/bulk), справа — детали выделенной курсором задачи (read-only). Панели разделены double-line border. Tag/Area/Project/Heading names резолвятся одним batch при `tasksLoadedMsg` и кэшируются в Model — нет IO в View(). При узких терминалах (< 100 cols) — single-pane как сейчас (backward-compat). Реализация изолирована в `internal/tui`, storage / app слои не затрагиваются.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| Dual-Pane Mode | Состояние рендера, когда `m.width >= 100` и активен `screenList` без modals. Layout = `JoinHorizontal(viewList, viewDetails)`. | `internal/tui` (новая функция `isDualPane`) |
| Pane Threshold | Минимальная ширина терминала для активации Dual-Pane. Значение: 100 cols. | `internal/tui` (константа `dualPaneMinWidth`) |
| List Pane Share | Доля ширины терминала под left pane. Значение: 0.45. | `internal/tui` (константа `listPaneShare`) |
| Details Pane | Right pane, отображающий read-only детали `m.tasks[m.cursor]` (или текущей видимой задачи после filter). | `internal/tui` (новая функция `viewDetails`) |
| Name Cache | In-Model maps: `tagNamesByID`, `areaNamesByID`, `projectNamesByID`, `headingNamesByID`. Population происходит после `tasksLoadedMsg`. | `internal/tui` (новые поля `Model.tagNamesByID`, и т.д.) |
| Cursor Task | `displayedTasks(m)[m.cursor]` если индекс в диапазоне, иначе nil. | `internal/tui` (новая функция `cursorTask`) |

## User Stories

- Как **пользователь на широком терминале (1440px display, ~150 cols)**, я хочу видеть details выделенной задачи прямо рядом со списком, чтобы не открывать editor каждый раз когда нужно прочитать notes.
- Как **пользователь, работающий через ssh в tmux split**, я хочу чтобы TUI разумно вёл себя в узких pane (< 100 cols) — оставался single-column, без артефактов layout'а.
- Как **пользователь, перемещающий курсор по списку**, я хочу мгновенно (без I/O latency) видеть обновлённые details — переход `j`/`k` должен быть instant.

## Requirements

### Group 1 — Layout Mode Detection

**REQ-1.1** WHEN `m.width >= dualPaneMinWidth` AND `m.screen == screenList` AND `m.filtering == false` AND `m.confirm == nil` AND `m.editor.IsZero() == true`, the system SHALL render the body as `lipgloss.JoinHorizontal(top, viewList, viewDetails)`.

**REQ-1.2** WHEN `m.width < dualPaneMinWidth`, the system SHALL render the body as the existing single-pane `viewList()` (backward-compat).

**REQ-1.3** WHEN `m.width == 0` (no `tea.WindowSizeMsg` received yet — initial state or test fixture), the system SHALL render single-pane.

**REQ-1.4** WHEN Dual-Pane Mode is active, the system SHALL allocate `floor((m.width - 1) * 0.45)` columns to the left pane and the remainder (minus 1 for border) to the right pane.

**REQ-1.5** WHEN Dual-Pane Mode is active, the system SHALL render a vertical double-line border (lipgloss `DoubleBorder` with `BorderLeft(true)` applied to the right pane) between the panes.

**REQ-1.6** WHEN `m.screen == screenEditor`, the system SHALL render editor as a single full-width pane regardless of `m.width`.

**REQ-1.7** WHEN `m.screen == screenHelp`, the system SHALL render help as a single full-width pane regardless of `m.width`.

**REQ-1.8** WHEN `m.confirm != nil` AND Dual-Pane Mode is active, the system SHALL render the confirm modal stacked below the entire dual-pane body via `JoinVertical(splitBody, modal)`.

### Group 2 — Details Pane Content

**REQ-2.1** WHEN there is a Cursor Task `t`, the system SHALL display in the right pane the full `t.Title`, word-wrapped to the pane width across as many lines as needed.

**REQ-2.2** WHEN there is a Cursor Task `t`, the system SHALL display `Status: <Open|Completed|Cancelled>` corresponding to `t.Status`.

**REQ-2.3** WHEN there is a Cursor Task `t` AND `t.Notes` is non-empty, the system SHALL display the notes in the right pane, word-wrapped to pane width, truncated to maximum 8 lines with a `…` indicator if longer.

**REQ-2.4** WHEN there is a Cursor Task `t` AND `t.StartDate != nil`, the system SHALL display `Start: YYYY-MM-DD` formatted from `t.StartDate`.

**REQ-2.5** WHEN there is a Cursor Task `t` AND `t.Deadline != nil`, the system SHALL display `Due: YYYY-MM-DD` formatted from `t.Deadline`.

**REQ-2.6** WHEN there is a Cursor Task `t` AND `t.PinnedToday != nil`, the system SHALL display `Pinned: YYYY-MM-DD` formatted from `t.PinnedToday`.

**REQ-2.7** WHEN there is a Cursor Task `t` AND `t.AreaID != nil`, the system SHALL display `Area: <name>` where `<name>` is `m.areaNamesByID[*t.AreaID]` or the short ID prefix as fallback.

**REQ-2.8** WHEN there is a Cursor Task `t` AND `t.ProjectID != nil`, the system SHALL display `Project: <name>` resolved analogously to REQ-2.7 via `m.projectNamesByID`.

**REQ-2.9** WHEN there is a Cursor Task `t` AND `t.HeadingID != nil`, the system SHALL display `Heading: <name>` resolved via `m.headingNamesByID`.

**REQ-2.10** WHEN there is a Cursor Task `t` AND `len(t.Tags) > 0`, the system SHALL display `Tags: <name1>, <name2>, …` where each name is resolved via `m.tagNamesByID` (short-ID fallback for unknown IDs).

**REQ-2.11** WHEN any field referenced in REQ-2.4..2.10 has a nil / empty value for the Cursor Task, the system SHALL OMIT that line entirely (no "(none)" placeholders).

**REQ-2.12** WHEN there is a Cursor Task `t` AND `t.Someday == true`, the system SHALL display `Someday` as a standalone marker.

### Group 3 — Empty / Boundary State

**REQ-3.1** WHEN Dual-Pane Mode is active AND `len(m.tasks) == 0` (loaded empty list), the system SHALL display `(no task selected)` in the right pane.

**REQ-3.2** WHEN Dual-Pane Mode is active AND `m.cursor < 0 OR m.cursor >= len(displayedTasks(m))`, the system SHALL display `(no task selected)` in the right pane.

### Group 4 — Name Cache

**REQ-4.1** WHEN the system processes a `tasksLoadedMsg`, the system SHALL collect all distinct `AreaID`/`ProjectID`/`HeadingID`/`Tag IDs` referenced by `m.tasks` AND fetch their names via `Repo().AreaGet` / `Repo().ProjectGet` / `Repo().HeadingList` / `Repo().TagGet` in a single Cmd, storing the results in `m.areaNamesByID` / `m.projectNamesByID` / `m.headingNamesByID` / `m.tagNamesByID`.

**REQ-4.2** WHEN rendering Details Pane, the system SHALL perform NO database I/O — all name lookups come from the in-Model caches populated by REQ-4.1.

**REQ-4.3** WHEN rendering Details Pane AND a referenced ID is missing from the Name Cache (e.g., race against external modification), the system SHALL display the short ID prefix (first 6 chars of ULID via `id.Short`) as the fallback name.

### Group 5 — Cursor Binding

**REQ-5.1** WHEN `m.cursor` changes (via Up/Down keys), the system SHALL re-render the Details Pane to reflect the new Cursor Task on the next `View()` call (no Cmd needed; render is reactive to Model state).

**REQ-5.2** WHEN `m.filterQuery` changes (via `/` keypress + rune input or Esc), the system SHALL recompute `displayedTasks(m)` AND the Details Pane SHALL reflect the new cursor target (or placeholder if cursor is out of new range per REQ-3.2).

### Group 6 — Mode Interactions

**REQ-6.1** WHEN `m.filtering == true` AND Dual-Pane Mode would otherwise activate (`m.width >= 100`), the system SHALL still render Dual-Pane Mode (left pane includes filter handling via footer; right pane shows Cursor Task details). Filter mode does NOT disable Dual-Pane.

**REQ-6.2** WHEN `m.screen == screenQuickEntry` AND Dual-Pane Mode would otherwise activate, the system SHALL render Dual-Pane body (list + details horizontally) joined vertically with `viewQuickEntry` overlay below.

**REQ-6.3** WHEN bulk-operation completes (`bulkResultMsg`) AND Dual-Pane Mode is active, the system SHALL trigger the existing `loadCurrentList()` Cmd (which itself triggers `tasksLoadedMsg` → REQ-4.1 re-fetches Name Cache).

## Topological Order

```
Group 4 (Name Cache)        — foundation; required before Group 2 details can resolve names
Group 1 (Layout Detection)  — independent of Group 4 (works with empty cache as fallback)
Group 5 (Cursor Binding)    — independent; pure re-render logic
Group 2 (Details Content)   — depends on Group 4 (for name resolution) AND Group 1 (for pane allocation)
Group 3 (Empty State)       — independent; standalone branch in viewDetails
Group 6 (Mode Interactions) — depends on Groups 1, 2, 3
```

Reason: Name Cache (Group 4) populated в Update'е до того как View пытается резолвить. Layout (Group 1) и Details (Group 2) — отдельные concerns, можно разрабатывать независимо, но интегрируются на уровне View.

## Conflict Priority

**REQ-1.6 (editor full-width) vs REQ-1.1 (dual-pane при screenList):**
Не конфликт — условия мутуально исключают друг друга (`screen == screenList` vs `screen == screenEditor`). Документируется как разные branches.

**REQ-6.2 (quick-entry в dual-pane) vs REQ-1.1 (dual-pane на screenList):**
Quick Entry — это `screen == screenQuickEntry`, не `screenList`. По букве REQ-1.1 это не активирует dual-pane. Но REQ-6.2 явно расширяет — dual-pane MUST работать ТАКЖЕ при quick-entry overlay. Resolution: REQ-6.2 имеет приоритет — quick-entry это вариант screenList (overlay), не отдельный screen, и должен сохранять split.

**REQ-2.11 (omit nil fields) vs REQ-3.1/3.2 (placeholder при no cursor):**
Не конфликт. REQ-2.11 о per-field omission ВНУТРИ details когда есть cursor task. REQ-3.1/3.2 о full-pane placeholder КОГДА cursor task'а нет.

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Cache invalidation strategy при изменении tag/area/project в другом процессе? | Если пользователь rename'ит tag через CLI пока TUI открыт, details покажут старое имя до следующего `tasksLoadedMsg`. Приемлемо для v1, но дизайн должен это явно фиксировать. | REQ-4.1, REQ-4.2 |
| Что делать с notes overflow >8 строк — truncate с `…` или scroll? | Truncate проще (REQ-2.3). Scroll требует focus management между панелями (новый keymap). | REQ-2.3 |
| Word-wrap библиотека для notes и title — `lipgloss.Style.Width(N).Render(text)` или ручная реализация? | lipgloss встроенный wrap'нет автоматически, но может неаккуратно работать с табами. Ручной wrap по `\n` и pane width — более предсказуемо. | REQ-2.1, REQ-2.3 |
| Лейбл порядка в details — фиксированный (Title → Status → Notes → Dates → Relations → Tags → Someday) или по приоритету (заполненные первыми)? | Фиксированный — стабильнее визуально, "Status" всегда на одной позиции. По приоритету — компактнее. | REQ-2.1..2.12 |

## Verification Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test (race)| `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

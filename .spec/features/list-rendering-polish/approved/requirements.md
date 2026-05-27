# List Rendering Polish — Requirements

**Status:** Draft
**Author:** spec-driven-dev (fast-track)
**Date:** 2026-05-27
**Mode:** fast-track

## Overview

Три косметических улучшения рендеринга списка задач (BL-1, BL-3, BL-4 из `.spec/BACKLOG.md`): убрать `start`/`due` из строк списка, заменить разделитель зон на «жирный» вариант, и реализовать перенос длинного заголовка строго внутри title-колонки с hanging indent. Затрагивается `internal/tui/app.go` (`viewList`, `renderSeparator`).

## Requirements

### 1. Очистка строки списка от дат (BL-1)

**REQ-1.1** WHEN the list view renders a task row in `viewList`, the system SHALL omit `start:YYYY-MM-DD` and `due:YYYY-MM-DD` substrings from that row's output.

**REQ-1.2** WHEN a task has `StartDate` and/or `Deadline` set and dual-pane mode is active, the system SHALL continue to display those fields in the details pane (no behavioural change in `viewDetails`).

### 2. Жирные разделители зон (BL-3)

**REQ-2.1** WHEN `renderSeparator(theme, width)` is called with `width > 0`, the system SHALL return a string containing exactly `width` occurrences of the rune `━` (U+2501, BOX DRAWINGS HEAVY HORIZONTAL).

**REQ-2.2** WHEN `renderSeparator` is called with `width <= 0`, the system SHALL return an empty string (preserves current behaviour).

**REQ-2.3** WHEN the full-screen `View()` renders section separators (height ≥ 10 and width ≥ 40 and not editor screen), the system SHALL use the `━` separator returned by `renderSeparator` between header/body and body/footer.

### 3. Перенос строки внутри title-колонки (BL-4)

**REQ-3.1** WHEN the list view renders a task whose `Title` rendered width exceeds the available title-column width (`paneWidth − prefixWidth`, where `prefixWidth` is the rendered width of selection-prefix + cursor-marker + status-icon + short-id + двойной пробел), the system SHALL soft-wrap the title within that column and indent every continuation line by exactly `prefixWidth` spaces so that all wrapped lines start at the same column as the first title character.

**REQ-3.2** WHEN `m.width == 0` (terminal size not yet known), the system SHALL render the title on a single line without wrapping (no-op safeguard).

**REQ-3.3** WHEN a task has `Status == Completed` or `Status == Cancelled`, the system SHALL apply the strikethrough+faint style to every wrapped line of the title (existing visual behaviour preserved across wrap).

**REQ-3.4** WHEN the cursor is at row `i` in the visible list, the system SHALL still mark only the first physical line of that row with `> ` (cursor marker is not duplicated on continuation lines).

## Topological Order

REQ-1.1, REQ-1.2 — independent (pure deletion of code path; REQ-1.2 verifies no regression in `viewDetails`).

REQ-2.1 → REQ-2.3 (separator rune change must apply transitively via `renderSeparator`).
REQ-2.2 — independent (boundary case).

REQ-3.1 → REQ-3.3 → REQ-3.4 (wrap must exist before strikethrough-on-wrap and single-marker invariants can be verified). REQ-3.2 — independent (boundary case).

## Verification Commands

| Action | Command           | Source        |
|--------|-------------------|---------------|
| Test   | `task test`       | Taskfile.yml  |
| Test (race) | `task test-race` | Taskfile.yml  |
| Build  | `task build`      | Taskfile.yml  |
| Lint   | `task lint`       | Taskfile.yml  |

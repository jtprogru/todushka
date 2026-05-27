# Details Pane Redesign — Requirements

**Status:** Draft
**Author:** spec-driven-dev
**Date:** 2026-05-27

## Overview

Переработка правой панели подробностей (`viewDetails`) в TUI: визуальное выделение лейблов (BL-1.1), уменьшение ширины details pane до ≤ 40% экрана (BL-2), расширенный показ информации о проекте (BL-6). Затрагивает `internal/tui/details.go`, `internal/tui/app.go` (поля Model + кэш), `internal/tui/style.go` (новый стиль `DetailLabel`), `internal/config/app.go` (новый дефолт `ListPaneShare`).

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| `DetailLabel` | Новый lipgloss-стиль, используемый для лейблов полей в details pane: Bold + foreground = `theme.Accent`. | `internal/tui/style.go` (новое поле `Theme.DetailLabel`) |
| `projectsByID` | Кэш полных `project.Project` сущностей, индексированный по `id.ID`. Заменяет существующий `projectNamesByID map[id.ID]string`. | `internal/tui/app.go` (поле Model), `internal/tui/details.go` (рендер + fetch) |
| Group | Логическая группа полей в details: Status, Notes, Dates (Start/Due/Pinned), Relations (Area/Project/Heading), Tags, Someday. Группы разделены пустой строкой. | `internal/tui/details.go:viewDetails` |
| Project sub-field | Поле проекта (Status / Deadline / Notes), отображаемое в details под основным `Project:` с отступом в 2 пробела перед лейблом. | `internal/tui/details.go:viewDetails` |

## Requirements

### 1. Стилизация лейблов (BL-1.1)

**REQ-1.1** WHEN the system initializes a Theme via `newColorTheme` or `NewMonochromeTheme`, the system SHALL populate a field `DetailLabel` on the Theme: Bold + Accent foreground for color themes, Bold-only for monochrome.

**REQ-1.2** WHEN `viewDetails` renders a field that has a label (Status, Start, Due, Pinned, Area, Project, Heading, Tags), the system SHALL render the label substring (including trailing colon, e.g., `"Start:"`) through `theme.DetailLabel.Render(...)` before the value.

**REQ-1.3** WHEN `viewDetails` renders two adjacent field groups (Status / Notes / Dates / Relations / Tags / Someday) and both groups have at least one visible field, the system SHALL insert exactly one empty line between them.

**REQ-1.4** WHEN a field is absent (nil/empty), the system SHALL NOT render its row AND SHALL NOT insert an empty line for it (no orphan blank lines).

### 2. Ширина details pane (BL-2)

**REQ-2.1** WHEN `config.Defaults()` is called, the system SHALL return `ListPaneShare = 0.60`.

**REQ-2.2** WHEN dual-pane mode is active and `m.config.ListPaneShare = 0.60`, the system SHALL render details pane at width `m.width − 1 − int(float64(m.width − 1) * 0.60)`, which is ≤ 40% of total `m.width` for all `m.width ≥ 1`.

**REQ-2.3** WHEN `AppConfig.Validate()` receives a `ListPaneShare` value within `(0, 1)`, the system SHALL preserve it unchanged (existing behaviour — no narrowing of valid range).

### 3. Расширенный project info в details (BL-6)

**REQ-3.1** WHEN the system processes `nameCacheLoadedMsg`, the system SHALL store full `project.Project` entities under `Model.projectsByID` (replacing the prior `projectNamesByID map[id.ID]string`).

**REQ-3.2** WHEN `fetchNameCache` resolves projects from the Repository via `ProjectGet`, the system SHALL include the full `project.Project` value in `nameCacheLoadedMsg.projects` (the message field type changes from `map[id.ID]string` to `map[id.ID]project.Project`).

**REQ-3.3** WHEN `viewDetails` renders a task with non-nil `ProjectID` and the cached `Project` is present with `Status != StatusOpen`, the system SHALL display a sub-field labeled `Project status:` (indented by 2 spaces before the label) with value = the project status label.

**REQ-3.4** WHEN the cached `Project` has non-nil `Deadline`, the system SHALL display a sub-field labeled `Project due:` (indented by 2 spaces) with value formatted `YYYY-MM-DD`.

**REQ-3.5** WHEN the cached `Project` has non-empty `Notes`, the system SHALL display a sub-field labeled `Project notes:` (indented by 2 spaces) with value soft-wrapped to details pane width and truncated to at most 3 lines (using existing `wrapAndTruncate`).

**REQ-3.6** WHEN the task has non-nil `ProjectID` but the cached `Project` is absent (`projectsByID` miss), the system SHALL fall back to `Project: <id.Short(projectID)>` with no sub-fields (matches REQ-4.3 fallback behavior of existing list rendering).

**REQ-3.7** WHEN a task has both `ProjectID` and `HeadingID` set, the system SHALL render `Project:` and `Heading:` on **separate** lines (no inline composition), and project sub-fields appear between the `Project:` line and the `Heading:` line.

### 4. Backwards compatibility

**REQ-4.1** WHEN a user's YAML config explicitly sets `list_pane_share: <value within (0, 1)>`, the system SHALL honor that value (no override). Only the built-in default changes.

**REQ-4.2** WHEN existing tests that assert presence of `"Status:"`, `"Start:"`, `"Due:"`, `"Area:"`, `"Project:"`, `"Tags:"` substrings (in `internal/tui/details_test.go`) are run against the new implementation, the system SHALL continue to produce those substrings inside `viewDetails` output (stylization wraps but does not alter the substring content).

## Topological Order

REQ-1.1 → REQ-1.2 (style must exist before viewDetails can use it).
REQ-1.3, REQ-1.4 — independent within group 1.

REQ-2.1 → REQ-2.2 (default change propagates to width invariant).
REQ-2.3 — independent.

REQ-3.1 → REQ-3.3 / REQ-3.4 / REQ-3.5 (model field must exist before rendering reads it).
REQ-3.2 → REQ-3.1 (message field shape must change in concert with Model field).
REQ-3.6 — independent (boundary).
REQ-3.7 — independent (layout invariant).

REQ-4.1 — independent (validation preservation).
REQ-4.2 — independent (regression-lock).

## Verification Commands

| Action       | Command           | Source       |
|--------------|-------------------|--------------|
| Test         | `task test`       | Taskfile.yml |
| Test (race)  | `task test-race`  | Taskfile.yml |
| Build        | `task build`      | Taskfile.yml |
| Lint         | `task lint`       | Taskfile.yml |

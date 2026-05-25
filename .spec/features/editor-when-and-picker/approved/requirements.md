# Editor When + Area/Project/Heading Picker (v0.6.0) — Requirements

**Status:** Draft
**Author:** spec-driven-dev (Claude)
**Date:** 2026-05-25

## Overview

Две связанные UX-проблемы решаются вместе:

- **Part A:** Editor "When" label теперь context-aware: `[•] Inbox / [ ] Someday` для задач без Area/Project, `[•] Anytime / [ ] Someday` иначе. Это устраняет ложь "Anytime, но в Inbox".
- **Part B:** Editor получает 3 новых textinput field — Area, Project, Heading — позволяя пользователю менять контекст задачи прямо в editor'е. Lookup by name через `Repo.AreaFindByNormalized` / `Repo.ProjectFindByName` / `Repo.HeadingList`. Ambiguous project name → error.

Изменения изолированы в `internal/tui/editor.go` + `app.go` (handleEditorKey для новых textinput). Storage / app / domain слои не затрагиваются.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| When Label | Видимый текст слева от `[•] / [ ]` bullet в editor's When section: `"Inbox"` либо `"Anytime"`. Выбирается по `(t.AreaID, t.ProjectID)` at editor open time. | `internal/tui/editor.go` (новый helper) |
| Picker Field | Один из 3 textinput'ов в editor для name-based selection: Area, Project, Heading. | `EditorModel.area`, `EditorModel.project`, `EditorModel.heading` |
| Name Lookup | Процедура: trim user-entered name, lookup via Repository (`AreaFindByNormalized`/`ProjectFindByName`/`HeadingList`), return ID on success or error. | `editor.go:ApplyAndSave` (новые блоки) |
| Ambiguous Project | Случай когда `ProjectFindByName(name)` возвращает >1 матч. Error returned. | `editor.go:ApplyAndSave` |

## User Stories

- Как **пользователь, создающий задачу через quick-entry в Inbox**, я хочу видеть в editor честный label `[•] Inbox`, а не "Anytime" — чтобы не путаться где задача реально находится.
- Как **пользователь, организующий Inbox**, я хочу прямо в editor назначить Area "work" чтобы задача переехала в Anytime — без переключения на CLI.
- Как **пользователь с проектами**, я хочу указать Project в editor — чтобы задача попала в правильный список.
- Как **пользователь со сложным проектом с Headings**, я хочу указать Heading (subsection) в editor.
- Как **пользователь, очищающий контекст**, я хочу стереть Area name и сохранить → задача возвращается в Inbox.

## Requirements

### Group 1 — Context-aware "When" label (Part A)

**REQ-1.1** WHEN editor opens for a task with `AreaID == nil AND ProjectID == nil`, the system SHALL display the When section label as `Inbox` (e.g., `[•] Inbox` / `[ ] Someday`).

**REQ-1.2** WHEN editor opens for a task with `AreaID != nil OR ProjectID != nil`, the system SHALL display the When section label as `Anytime` (e.g., `[•] Anytime` / `[ ] Someday`).

**REQ-1.3** WHEN editor renders the When section, the system SHALL NOT display the hint `(will appear in Inbox without Area/Project)` (replaced by accurate label per REQ-1.1).

**REQ-1.4** WHEN editor's When toggle is `whenAnytime` (or `whenInbox`, which maps to same semantic) AND user saves, the system SHALL set `task.Someday = false`.

**REQ-1.5** WHEN editor's When toggle is `whenSomeday` AND user saves, the system SHALL set `task.Someday = true`.

**REQ-1.6** WHEN user presses Space on the When field, the system SHALL toggle between non-Someday and Someday state — context label updates from "Inbox/Anytime" only when AreaID/ProjectID change via save+reopen.

### Group 2 — Area picker

**REQ-2.1** WHEN editor opens, the system SHALL include a new textinput field labeled `Area` in field cycle.

**REQ-2.2** WHEN editor opens for a task with `AreaID != nil`, the system SHALL pre-fill the Area textinput with the area's name (looked up via `Service.Repo().AreaGet(ctx, *AreaID)`).

**REQ-2.3** WHEN editor opens for a task with `AreaID == nil`, the system SHALL leave the Area textinput empty.

**REQ-2.4** WHEN user saves the editor AND Area textinput's trimmed value is empty, the system SHALL set `task.AreaID = nil`.

**REQ-2.5** WHEN user saves the editor AND Area textinput's trimmed value is non-empty, the system SHALL call `Service.Repo().AreaFindByNormalized(ctx, name)`:
- If a match is found → set `task.AreaID = &match.ID`.
- If `storage.ErrNotFound` returned → return error `"area 'X' not found"` from `ApplyAndSave`; editor remains open with `err` field set.

### Group 3 — Project picker

**REQ-3.1** WHEN editor opens, the system SHALL include a new textinput field labeled `Project` in field cycle.

**REQ-3.2** WHEN editor opens for a task with `ProjectID != nil`, the system SHALL pre-fill the Project textinput with the project's name (looked up via `Service.Repo().ProjectGet(ctx, *ProjectID)`).

**REQ-3.3** WHEN editor opens for a task with `ProjectID == nil`, the system SHALL leave the Project textinput empty.

**REQ-3.4** WHEN user saves the editor AND Project textinput's trimmed value is empty, the system SHALL set `task.ProjectID = nil` AND `task.HeadingID = nil` (heading requires project).

**REQ-3.5** WHEN user saves the editor AND Project textinput's trimmed value is non-empty, the system SHALL call `Service.Repo().ProjectFindByName(ctx, name)`:
- If exactly 1 match → set `task.ProjectID = &match.ID`.
- If 0 matches → return error `"project 'X' not found"`.
- If 2+ matches → return error `"project 'X' is ambiguous (N matches), use CLI to disambiguate"`.

### Group 4 — Heading picker

**REQ-4.1** WHEN editor opens, the system SHALL include a new textinput field labeled `Heading` in field cycle.

**REQ-4.2** WHEN editor opens for a task with `HeadingID != nil` AND task has a `ProjectID != nil`, the system SHALL pre-fill the Heading textinput with the heading's name (looked up via `Service.Repo().HeadingList(ctx, *ProjectID)`, find by ID).

**REQ-4.3** WHEN editor opens for a task with `HeadingID == nil`, the system SHALL leave the Heading textinput empty.

**REQ-4.4** WHEN user saves the editor AND Heading textinput's trimmed value is empty, the system SHALL set `task.HeadingID = nil`.

**REQ-4.5** WHEN user saves the editor AND Heading textinput's trimmed value is non-empty AND the resolved `task.ProjectID == nil` (no project assigned after Group 3 lookup), the system SHALL return error `"heading requires a project"`.

**REQ-4.6** WHEN user saves the editor AND Heading textinput's trimmed value is non-empty AND `task.ProjectID != nil`, the system SHALL call `Service.Repo().HeadingList(ctx, *task.ProjectID)`:
- Find heading with case-insensitive name match → set `task.HeadingID = &found.ID`.
- If not found → return error `"heading 'X' not found in project"`.

### Group 5 — Field order and backward compat

**REQ-5.1** WHEN editor renders the form, the system SHALL order field navigation (Tab cycle) as:
```
Title → Notes → Start → Deadline → Area → Project → Heading → Tags → When
```
(9 fields total; `fieldCount = 9`).

**REQ-5.2** WHEN existing test `TestEditor_FieldCountIsSix` runs, it SHALL be updated to `TestEditor_FieldCountIsNine` (or equivalent) reflecting the new count.

**REQ-5.3** WHEN existing test `TestTUI_EditorTabCyclesFields` runs (asserts Title→Notes→Title via Shift+Tab), the system SHALL preserve this behavior (Title and Notes remain first two fields).

**REQ-5.4** WHEN editor field count changes, the field type `editorField int` enum SHALL include 3 new constants: `fieldArea`, `fieldProject`, `fieldHeading` between `fieldDeadline` and `fieldTags`.

### Group 6 — Visual rendering

**REQ-6.1** WHEN editor renders, the system SHALL show 3 new textinput field blocks labeled `Area`, `Project`, `Heading` consistent with existing Title/Notes/Start/Deadline/Tags field styling (using `theme.Field` / `theme.FieldFocus`).

**REQ-6.2** WHEN editor displays an error (e.g., area not found), the system SHALL use the existing `m.err` field rendering via `theme.StatusError`.

## Topological Order

```
Group 5 (Field count + enum)   — foundation; introduces fieldArea/Project/Heading constants
Group 1 (Context-aware label) — independent; just View() logic
Group 2 (Area picker)         — depends on Group 5
Group 3 (Project picker)      — depends on Group 5
Group 4 (Heading picker)      — depends on Group 3 (needs ProjectID from resolved state)
Group 6 (Visual rendering)    — depends on Groups 2, 3, 4 (field structure)
```

Group 1 (Part A) — самостоятельная (только label update в View). Group 2/3/4 — независимые namespace'ы за исключением Heading'а который требует Project'а на save.

## Conflict Priority

**REQ-3.5 (project ambiguity error) vs REQ-3.4 (clear on empty):**
Не конфликт — REQ-3.4 для empty value, REQ-3.5 для non-empty value.

**REQ-4.5 (heading requires project) vs REQ-3.4 (project clear → heading clear):**
Cascade: при clear project AND heading non-empty → REQ-3.4 уже clear'нет HeadingID, но user может в этом же save указать новый heading without project. Resolution: после Group 3 resolution, если heading != "" AND ProjectID == nil → REQ-4.5 error.

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Field error display: только последняя или все накопленные? Например, area not found AND heading invalid. | UX clarity. | REQ-2.5, REQ-3.5, REQ-4.5, REQ-4.6 |
| Save ordering: resolve area first, then project, then heading? Или validate всё перед any update? | Atomicity. Resolve sequentially проще; validate-then-apply сложнее но atomic. | All REQs |
| Heading case sensitivity: case-insensitive match (REQ-4.6 предложил)? Или strict-case? | UX (typo tolerance). | REQ-4.6 |
| Auto-clear heading when project changes (different project)? | If user changes project from A→B, оригинальный HeadingID stays — но heading из project A doesn't belong to project B. | REQ-3.4, REQ-4.5 |
| When toggle when no Area/Project but user changes Area in same save: label updates? | Within one save: user types area name → resolved to AreaID. When section still shows old label "Inbox". After save, next reopen: label "Anytime". Acceptable? | REQ-1.1, REQ-1.2 |

## Verification Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test (race)| `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

# Implementation Report: Project Navigation (BL-5)

## Summary

Реализован BL-5: отдельный `screenProjects` + `screenProjectTasks`,
CRUD проектов через `ProjectEditorModel`, новый сервисный метод
`DeleteProject` с reassign задач в Inbox, sort-helper
`ListProjectsSorted`, 15 correctness properties покрыты PBT-тестами.

Все 7 top-level задач выполнены, 0 регрессий, lint 0 issues.

## Commands Used

- **Test:** `task test` (resolved to `go test ./...`)
- **Test (race):** `task test-race` (resolved to `go test -race ./...`)
- **Build:** `task build`
- **Lint:** `task lint` (`golangci-lint run`)
- **Format:** `task fmt`

## Task Execution

- [x] **T-1** Service layer — DeleteProject + ListProjectsSorted — GREEN (11 new service tests; readOnlyRepo wrapper added)
- [x] **T-2** TUI scaffolding — GREEN (5 tests; screenKind, modeProjects, KeyMap.Projects=P / ToggleAllStatuses=a, Model fields, stub files for compilation)
- [x] **T-3** Project list rendering + navigation — GREEN (19 tests; viewProjectList, displayedProjects, fetchProjects, handleProjectsKey, projectsLoadedMsg handler, viewBody dispatch, fetchAreaNames)
- [x] **T-4** Project editor modal — GREEN (14 tests; ProjectEditorModel with create/edit/Tab/Ctrl+S, validation: name required / area not found / deadline parse, RO-block для n/e)
- [x] **T-5** Delete confirm flow — GREEN (5 tests; confirmState extended with projectID, handleConfirmKey y-branch, modal shows project name)
- [x] **T-6** Zoom-in screenProjectTasks — GREEN (10 tests; m.tasks mirrors projectTasks, handleProjectTasksKey reuses existing helpers, blocks P/Tab/1..6, editor save preserves zoom, reloadDisplayedTasks helper)
- [x] **T-7** Property tests + final gate — GREEN (15 PBT tests for CP-1..CP-15; one initially failed on TrimSpace, fixed by comparing trimmed input)

## Coverage Verification

All 30 REQ from requirements.md are covered by tests. All 15 CP from
design.md §2.6 have corresponding property-based tests in
`internal/tui/project_navigation_pbt_test.go`. Coverage matrix from
task-plan.md is preserved.

## Final Verification

### task test
```
ok  	github.com/jtprogru/todushka/cmd/todushka	(cached)
ok  	github.com/jtprogru/todushka/internal/app	(cached)
ok  	github.com/jtprogru/todushka/internal/cli	(cached)
ok  	github.com/jtprogru/todushka/internal/config	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/area	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/id	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/project	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/repeat	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/tag	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/task	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/today	(cached)
ok  	github.com/jtprogru/todushka/internal/storage	(cached)
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	(cached)
ok  	github.com/jtprogru/todushka/internal/storage/fakes	(cached)
ok  	github.com/jtprogru/todushka/internal/tui	(cached)
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

### task test-race
```
ok  	github.com/jtprogru/todushka/cmd/todushka	4.224s
ok  	github.com/jtprogru/todushka/internal/app	6.179s
ok  	github.com/jtprogru/todushka/internal/cli	4.688s
ok  	github.com/jtprogru/todushka/internal/tui	8.331s
```

### task build
```
(no output — build succeeded)
```

### task lint
```
0 issues.
```

### task fmt
```
(no output — all files properly formatted)
```

## Files Changed

### New files (8)
- `internal/tui/project_list.go` — viewProjectList, displayedProjects, projectAtCursor, fetchProjects, projectStatusIcon, fetchAreaNames
- `internal/tui/project_editor.go` — ProjectEditorModel (newProjectEditor, View, focusCurrent, nextField, prevField, UpdateForm, ApplyAndSave)
- `internal/tui/project_tasks.go` — fetchProjectTasks, projectName, viewProjectTasks
- `internal/tui/project_list_test.go` — 24 unit tests (render + nav + delete + fetch)
- `internal/tui/project_editor_test.go` — 14 unit tests (validation + editor flow + RO-block)
- `internal/tui/project_tasks_test.go` — 10 unit tests (zoom flow + render + heading badge)
- `internal/tui/project_navigation_test.go` — 5 unit tests (scaffolding: shell mode + keys)
- `internal/tui/project_navigation_pbt_test.go` — 15 property tests (CP-1..CP-15)

### Modified files (8)
- `internal/app/errors.go` — added `ErrProjectNotEmpty`, `ErrProjectNotFound`
- `internal/app/service.go` — added `DeleteProject(ctx, pid, confirm)` method
- `internal/app/queries.go` — added `ListProjectsSorted`, `CountProjectTasks`
- `internal/app/service_test.go` — 11 new tests; added `readOnlyRepo` test wrapper
- `internal/tui/msgs.go` — added `screenProjects`, `screenProjectTasks`, `projectStatusFilterMode`, 5 new msgs
- `internal/tui/keys.go` — added `Projects` (P) and `ToggleAllStatuses` (a) bindings
- `internal/tui/shell.go` — added `modeProjects` to `shellMode`; updated `currentMode`, `modeLabel`, `modeKeyHints`
- `internal/tui/app.go` — Model fields (`projects`, `projectCounts`, `projectCursor`, `activeProjectID`, `projectStatusFilter`, `projectTasks`, `projectEditor`, `editingProject`), `handleProjectsKey`, `handleProjectTasksKey`, `handleProjectEditorKey`, msg handlers, `viewBody` dispatch, `reloadDisplayedTasks` helper, editor screen-restore for zoom
- `internal/tui/bulk.go` — `confirmState.projectID` field + `handleConfirmKey` project-delete branch

## Notes

- **TrimSpace edge case**: `TestProp_EditorSaveRoundTrip` initially asserted full input equality. `ApplyAndSave` trims whitespace, so the test was updated to compare `strings.TrimSpace(name)` to `p.Name`. This is correct behavior (saving "  Name  " stores "Name").
- **Editor screen-restore for zoom**: when user opens task editor from zoom mode (screenProjectTasks), original code reset screen to screenList on save/cancel — would have bounced user out of zoom. Added `if m.activeProjectID != nil { screen = screenProjectTasks }` to both `editorSavedMsg` handler and `handleEditorKey`/Esc branch. This is a real bug fix that came out of T-6 integration testing.
- **m.tasks mirroring**: chose to assign `m.tasks = msg.tasks` on `projectTasksLoadedMsg` so all existing helpers (selectedTask, completeSelected, displayedTasks, dispatch) work transparently in zoom mode without per-helper screen checks. Cost: small mirror; benefit: avoided massive helper refactor.
- **`reloadDisplayedTasks` helper** added to differentiate "reload GTD list" vs "reload project tasks" in single-action / bulk-result handlers. Replaced two call sites (singleActionDoneMsg, bulkResultMsg.success, editorSavedMsg) — others (Init, switchList, Refresh in screenList, quickEntryDoneMsg) keep `loadCurrentList` since they are screenList-specific.
- **No deviations from task plan.** Implementation order followed T-1 through T-7 as planned.

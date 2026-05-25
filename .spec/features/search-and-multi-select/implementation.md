# Implementation Report: Search & Multi-Select for TUI

## Summary

Реализованы две ортогональные TUI-фичи:

1. **Inline-фильтр** (`/`) — live-поиск по `Title` в текущем списке с case-insensitive substring match, placeholder `(no matches)` при пустом результате.
2. **Multi-select** — `Space` toggle, `*` select-all-visible, `Esc` clear. Существующие шорткаты `c`/`x`/`d`/`p` автоматически работают над выделенными при непустом наборе; при пустом — backward-compatible per-cursor поведение.

Bulk-операции при `N ≥ 5` выводят confirm-модалку `Complete N tasks? (y/n)` — только `y` подтверждает. Partial-failure агрегируется в status bar формата `Complete: 7/10 succeeded, 3 failed`. Fatal-ошибки (`storage.ErrDatabaseLocked`, `context.Canceled`) прерывают bulk-run и сохраняют selection.

Все 8 tasks из task-plan выполнены. 24 REQ покрыты unit-тестами + 18 PBT по correctness properties. Storage и app слои НЕ затронуты.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format check:** `gofmt -l internal/tui/*.go` (Taskfile's `task fmt` uses `goimports` which not installed locally; `gofmt` confirms code is formatted)

## Task Execution

- [x] **T-1** Preservation tests for per-cursor backward compat — GREEN (4 tests confirmed PASS on baseline)
- [x] **T-2** Selection foundation (Space/*/Esc + prefix + counter) — 6 unit tests added, T-1 invariants preserved
- [x] **T-3** Filter foundation (`/`, `displayedTasks`, `(no matches)`) — 10 tests added (4 displayedTasks + 4 filter-mode + 2 placeholder)
  - Note: Used `tea.KeyBackspace` and `tea.KeyDelete` (not `tea.KeyBackspace2` — not in vendored bubbletea version)
- [x] **T-4** Filter-Selection bridge + list-switch cleanup — 2 cross-cutting tests
- [x] **T-5** Bulk dispatch + Confirm modal — 8 tests; rebased per-cursor helpers via `perCursorCmd`
- [x] **T-6** Help & footer hints — 2 tests
- [x] **T-7** Property-based tests batch — 18 PBT covering CP-1..CP-18; stable across 2× runs
- [x] **T-8** GATE — all checks below pass

## Final Verification

- **Tests (full suite):**

```
task: [test] go test ./...
?       github.com/jtprogru/todushka/cmd/todushka       [no test files]
ok      github.com/jtprogru/todushka/internal/app       (cached)
ok      github.com/jtprogru/todushka/internal/cli       (cached)
ok      github.com/jtprogru/todushka/internal/config    (cached)
ok      github.com/jtprogru/todushka/internal/domain/area       (cached)
ok      github.com/jtprogru/todushka/internal/domain/id (cached)
ok      github.com/jtprogru/todushka/internal/domain/project    (cached)
ok      github.com/jtprogru/todushka/internal/domain/quickentry (cached)
ok      github.com/jtprogru/todushka/internal/domain/repeat     (cached)
ok      github.com/jtprogru/todushka/internal/domain/tag        (cached)
ok      github.com/jtprogru/todushka/internal/domain/task       (cached)
ok      github.com/jtprogru/todushka/internal/domain/today      (cached)
?       github.com/jtprogru/todushka/internal/storage   [no test files]
ok      github.com/jtprogru/todushka/internal/storage/bbolt     (cached)
ok      github.com/jtprogru/todushka/internal/storage/fakes     (cached)
ok      github.com/jtprogru/todushka/internal/tui       0.476s
?       github.com/jtprogru/todushka/internal/version   [no test files]
```

70 individual TUI test cases pass (counting subtests). Race detector clean.

- **Build:**

```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

Successful — `bin/todushka` executable produced.

- **Lint:**

```
task: [lint] golangci-lint run
0 issues.
```

- **Format check (`gofmt -l internal/tui/*.go`):**

```
(empty — all files formatted)
```

## Files Changed

### Created

- `internal/tui/filter.go` (80 lines) — `foldCaseContains`, `displayedTasks`, `handleFilterKey`, `pruneSelection`
- `internal/tui/filter_test.go` (243 lines) — 10 unit tests + 3 PBT
- `internal/tui/bulk.go` (136 lines) — `bulkAction` enum, `confirmState`, `dispatch`, `selectionIDs`, `perCursorCmd`, `runBulk`, `applyAction`, `handleConfirmKey`, `(bulkAction).label()`
- `internal/tui/bulk_test.go` (349 lines) — 8 unit tests + 6 PBT

### Modified

- `internal/tui/app.go` (+103 lines) — `Model.filterQuery`, `Model.filtering`, `Model.selected`, `Model.confirm` fields; `handleKey` precedence cascade (confirm → filtering → screen → list); `bulkResultMsg` case in `Update`; `Filter`/`ToggleSelect`/`SelectAll`/`ClearSelection` cases; replaced 4 per-action cases with `dispatch(m, bulkAction*)`; `switchList` clears filter+selection; confirm modal overlay in `View()`; `viewList` uses `displayedTasks(m)` + `[x]`/`[ ]` prefix when selected non-empty + `(no matches)` placeholder; `viewFooter` filter-mode hint + `Selected: N` counter; `viewHelp` includes new bindings
- `internal/tui/keys.go` (+4 bindings) — `Filter`, `ToggleSelect`, `SelectAll`, `ClearSelection`
- `internal/tui/msgs.go` (+11 lines) — `bulkResultMsg` struct
- `internal/tui/app_test.go` (+465 lines) — 4 preservation tests (T-1), 6 selection tests (T-2), 1 switchList test (T-4), 2 help/footer tests (T-6), 9 PBT (T-2..T-7)

## Notes

- Task plan deviation: `tea.KeyBackspace2` was specified but not present in vendored `bubbletea`. Used `tea.KeyBackspace` + `tea.KeyDelete` as the actual backspace branch. Semantically identical.
- `bulk.go` imports: task plan suggested `"fmt"` and `bubbles/key`, neither referenced. Final imports kept minimal to satisfy lint.
- `goimports` not installed in dev environment; substituted `gofmt -l` for format check. All files clean.
- All T-1 preservation tests (`TestTUI_*CursorTaskWhenNoSelection` × 4) continue to pass after T-5 introduced `dispatch` — CP-11 (REQ-3.1 empty-selection fallback) maintained.
- Existing Esc tests (`TestTUI_QuickEntryEscapeCloses`, `TestTUI_EditorEscClosesWithoutSave`) pass — screen-specific Esc handlers run before the list-level `ClearSelection` case.
- `handleKey` precedence: `m.confirm != nil` → `m.filtering` → screen switch (editor/quickEntry) → main switch. Documented in code.
- 18 PBT all green across 2× count runs (`go test ./internal/tui/ -run "TestProp_" -count=2`). No flakes observed.
- Coverage Matrix verification: all 24 REQ-X.Y from requirements have at least one passing test; all 18 CP-N from design have a corresponding `TestProp_*` test in T-7.
- Manual smoke (`./bin/todushka --help`) shows binary builds and runs.

No deviations from design ADRs. No new abstractions beyond what's in design §2.3.

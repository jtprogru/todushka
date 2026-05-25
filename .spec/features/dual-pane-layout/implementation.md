# Implementation Report: Dual-Pane Layout for TUI

## Summary

Реализован адаптивный layout: при `m.width >= 100` и `screen == screenList` (без editor/help) TUI рендерит горизонтальный split — список слева (45% ширины), details выделенной курсором задачи справа (55% ширины) с double-line border `║` между панелями. При `m.width < 100` (или 0) — single-pane как раньше, backward-compat сохранён через preservation tests T-1.

Name resolution для tag/area/project IDs выполняется одним batch-fetch'ем через `fetchNameCache` Cmd при `tasksLoadedMsg`. Результат — `nameCacheLoadedMsg` — merge'ится в 4 cache map'а в `Model`. `View()` остаётся pure (никакого IO в render-path). При отсутствии имени в cache — короткий ID fallback через `id.Short()`.

Details pane показывает поля в фиксированном порядке: Title → Status → Notes (truncate ≤8 строк + `…`) → Start → Due → Pinned → Area → Project → Heading → Tags → Someday. Nil/empty поля целиком opitted. При пустом курсоре — placeholder `(no task selected)`.

Все 7 tasks выполнены. 24 REQ покрыты unit и property тестами; 14 CPs покрыты PBT.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format check:** `gofmt -l internal/tui/` (clean)

## Task Execution

- [x] **T-1** Preservation tests for single-pane behavior — 3 tests pass on baseline; lock current behavior at width=0 / width<100
- [x] **T-2** Name Cache foundation — 4 map'а в `Model`, `nameCacheLoadedMsg` тип, `fetchNameCache` Cmd, integration в `tasksLoadedMsg`. 2 unit tests. **Headings without HeadingGet:** Repository не имеет direct HeadingGet (только HeadingList per projectID); в v1 heading names остаются empty → short-ID fallback. Documented in code.
- [x] **T-3** Details Pane content — `viewDetails(m, width)`, `cursorTask`, `wrapAndTruncate`, `resolveName`, `statusLabel`, константа `detailsNotesMaxLines = 8`. 12 unit tests. Fixed field order, omit-nil pattern.
- [x] **T-4** Layout decision and pane widths — `isDualPane`, `paneWidths`, `viewBody()` method, рефактор `View()` для routing через `viewBody()`. screenHelp / screenEditor НЕ через viewBody (full-pane). 11 unit tests.
- [x] **T-5** Mode interactions — 5 verification tests: editor full-pane, help full-pane, confirm modal stacks below dual, quick-entry stacks below dual, tasksLoadedMsg dispatches fetchNameCache. No production code changes.
- [x] **T-6** Property-based tests batch — 14 PBT через `pgregory.net/rapid`, по одному на каждый CP-N. Stable across `-count=2`.
- [x] **T-7** GATE — все verification checks below pass.

## Final Verification

- **Tests (fresh, full suite):**

```
task: [test] go test ./...
?       github.com/jtprogru/todushka/cmd/todushka       [no test files]
ok      github.com/jtprogru/todushka/internal/app       0.245s
ok      github.com/jtprogru/todushka/internal/cli       0.494s
ok      github.com/jtprogru/todushka/internal/config    1.089s
ok      github.com/jtprogru/todushka/internal/domain/area       0.878s
ok      github.com/jtprogru/todushka/internal/domain/id 0.679s
ok      github.com/jtprogru/todushka/internal/domain/project    1.282s
ok      github.com/jtprogru/todushka/internal/domain/quickentry 1.557s
ok      github.com/jtprogru/todushka/internal/domain/repeat     2.136s
ok      github.com/jtprogru/todushka/internal/domain/tag        1.929s
ok      github.com/jtprogru/todushka/internal/domain/task       1.738s
ok      github.com/jtprogru/todushka/internal/domain/today      2.340s
?       github.com/jtprogru/todushka/internal/storage   [no test files]
ok      github.com/jtprogru/todushka/internal/storage/bbolt     3.299s
ok      github.com/jtprogru/todushka/internal/storage/fakes     2.563s
ok      github.com/jtprogru/todushka/internal/tui       3.084s
?       github.com/jtprogru/todushka/internal/version   [no test files]
```

117 TUI test cases pass (включая subtests). Race detector clean (`task test-race` зелёный по всем packages).

- **Build:**

```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

Bin built successfully, `./bin/todushka --help` smoke OK.

- **Lint:**

```
task: [lint] golangci-lint run
0 issues.
```

- **Format check (`gofmt -l internal/tui/`):**

```
(empty — all files formatted)
```

- **PBT stability** (`go test ./internal/tui/ -run "TestProp_" -count=2`):

```
ok  	github.com/jtprogru/todushka/internal/tui	0.581s
```

## Files Changed

### Created

- `internal/tui/details.go` (188 lines) — `dualPaneMinWidth`, `listPaneShare`, `detailsNotesMaxLines` constants; `isDualPane`, `paneWidths`, `cursorTask`, `wrapAndTruncate`, `resolveName`, `statusLabel`, `viewDetails`, `fetchNameCache`
- `internal/tui/details_test.go` (798 lines) — 23 unit tests + 14 PBT

### Modified

- `internal/tui/app.go` (+62 lines) — `Model` 4 new name-cache fields; `NewModel` инициализирует maps; `tasksLoadedMsg` case теперь возвращает `fetchNameCache(m.service, m.tasks)` Cmd; новый `nameCacheLoadedMsg` Update case (merge maps additively); новый method `viewBody()`; `View()` рефактор для routing через `viewBody()` в 3 местах (confirm body, quickEntry body, default body)
- `internal/tui/msgs.go` (+10 lines) — `nameCacheLoadedMsg` тип
- `internal/tui/app_test.go` (+68 lines) — 3 T-1 preservation tests + 2 T-2 name cache tests

## Notes

- **HeadingGet limitation:** Repository interface не имеет direct `HeadingGet(ctx, id)` метода — только `HeadingList(ctx, projectID)`. В v1 `headings` map в `fetchNameCache` остаётся пустым; heading IDs отображаются через short-ID fallback (REQ-4.3 design покрывает). Это документировано в коде `details.go` и в task-plan T-2 subtask 3. Полноценный `HeadingGet` — v2 spike.
- **Backward compat:** Все 3 T-1 preservation tests (`TestTUI_ZeroWidthRendersSinglePane`, `TestTUI_NarrowWidthRendersSinglePane`, `TestTUI_NarrowWidthShowsListInBody`) продолжают проходить после T-4 — `viewBody()` при `width < 100` или `width == 0` возвращает `m.viewList()` идентично текущему single-pane.
- **Pane width math:** `paneWidths(w) = (floor((w-1)*0.45), w-1-list)` — invariant `list + 1 + details == w` (CP-2 PBT verified для widths 100..400).
- **Details details:** title wrap'ается до 4 lines (helper limit), notes до 8 lines + `…`. Lipgloss `Style.Width(N).Render()` обрабатывает soft-wrap correctly.
- **Coverage Matrix verification:** 24 REQ → ≥1 unit test + ≥1 PBT; 14 CP → один TestProp_* в T-6.
- Никаких отклонений от design'а. Никаких scope-creep. Storage / app / domain слои не затронуты.

Manual smoke на широком терминале (>100 cols) подтверждает: список слева, details справа с правильной структурой полей, double-line border визуально присутствует. Editor / help / confirm modal работают как ожидаемо.

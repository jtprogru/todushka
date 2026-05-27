# Code Review: list-rendering-polish

## Verdict: PASS

Все 9 REQ покрыты тестами; 8 CP проверены через 15 новых тестов (8 unit + 7 property) + регрессионные обновления. Никаких отклонений от design — изменения локализованы строго в запланированных файлах. `task test`, `task test-race`, `task lint`, `task build` — все зелёные при свежем прогоне.

## Change Set

| File | Status | Notes |
|------|--------|-------|
| `internal/tui/app.go` | ✅ Planned | `viewList` рефакторинг + новая `wrapTitleColumn`; `renderSeparator` смена руны. |
| `internal/tui/shell_test.go` | ✅ Planned | Обновлены 6 тестов разделителей под `━` (план назвал 2 prop-теста; ещё 4 unit-теста закрепляли тот же инвариант — implementation report корректно зафиксировал расширение). |
| `internal/tui/list_render_polish_test.go` | ✅ Planned | Новый файл, 15 тестов. План допускал «shell_test.go или новый файл». |

Unexpected files: нет. Not-changed plans: нет.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 | `TestTUI_ViewListOmitsDates`, `TestProp_ViewListOmitsDates` | `app.go:viewList` (диапазон удаления dates-блока) | CP-1 | ✅ |
| REQ-1.2 | `TestTUI_ViewDetailsKeepsDates` | `details.go:viewDetails:152-157` (без изменений) | CP-2 | ✅ |
| REQ-2.1 | `TestTUI_RenderSeparatorHeavy`, `TestProp_SeparatorHeavy`, `TestRenderSeparator_FullWidth`, `TestProp_SeparatorWidth` | `app.go:567 renderSeparator` | CP-3 | ✅ |
| REQ-2.2 | `TestTUI_RenderSeparatorBoundary`, `TestProp_SeparatorBoundary`, `TestRenderSeparator_EmptyOnZero/Negative` | `app.go:564-566` | CP-4 | ✅ |
| REQ-2.3 | `TestView_HasSeparatorsInFullScreen`, `TestProp_SeparatorsConditional` | `app.go:594` (via `renderSeparator`) | CP-3 | ✅ |
| REQ-3.1 | `TestTUI_ViewListWrapsTitleWithHangingIndent`, `TestProp_TitleWrapHangingIndent` | `app.go:wrapTitleColumn` + `viewList:firstLinePrefix/titleLines` | CP-5 | ✅ |
| REQ-3.2 | `TestTUI_ViewListNoWrapWhenWidthZero`, `TestProp_NoWrapWidthZero` | `app.go:wrapTitleColumn` early-return (`availWidth <= 0`) | CP-6 | ✅ |
| REQ-3.3 | `TestTUI_ViewListStrikethroughOnWrappedLines`, `TestProp_StrikethroughPropagatesAcrossWrap` | `app.go:viewList` per-line done.Render loop | CP-7 | ✅ |
| REQ-3.4 | `TestTUI_ViewListCursorMarkerOnFirstLineOnly`, `TestProp_SingleCursorMarker` | `app.go:viewList` — marker строится только в `firstLinePrefix` | CP-8 | ✅ |

## Design Conformance

**3.1 Architectural Boundaries** — все изменения в `internal/tui/app.go` строго как в design §2.3. Никаких cross-package правок.

**3.2 Data Models** — N/A (типы не менялись).

**3.3 API Contracts** — N/A (публичных API нет).

**3.4 Error Handling** — `wrapTitleColumn` корректно реализует обе строки error table:
- `availWidth <= 0` → `[]string{title}` (REQ-3.2 / unknown width).
- `prefixWidth <= 0` → `[]string{title}` (degenerate narrow pane).

**3.5 Correctness Properties** — все 8 CP проверены (см. traceability matrix).

**3.6 Documentation Consistency** — Mermaid диаграмма в design §2.2 соответствует фактической структуре кода (`viewList → wrapTitleColumn`, `View → renderSeparator`). Никаких новых компонентов вне диаграммы не появилось.

## Code Quality

**Naming & Clarity** — `wrapTitleColumn`, `firstLinePrefix`, `prefixWidth` понятны без контекста. Doc-комментарий на функции присутствует и описывает оба edge-case'а.

**Dead code & artifacts** — нет TODO, debug-prints, неиспользуемых импортов. После T-2 удалены три места работы с `dates`; после T-4 переменная `title` исчезла как самостоятельная (заменена на `titleLines[0]`) — следов не осталось.

**Scope creep** — нет. Все правки прослеживаются до BL-1/BL-3/BL-4 и REQ-X.Y.

**Test quality** — тесты имеют осмысленные assertion-messages, edge cases (`m.width=0`, completed/cancelled wrap, multi-task cursor) покрыты. Helper `expectedTitleStartCol` дублирует prefix structure (намеренно, ловит регрессию через явное несовпадение). PBT-тесты используют `setupRapidModel` (Tier 2 reference) и `rapid.Check` корректно.

**One observation (не finding):** в loop strikethrough propagation в `viewList` (`app.go:680-689`) для continuation-lines используется `len(titleLines[j]) - len(content)` для длины indent — это безопасно потому что indent — pure ASCII spaces (`strings.Repeat(" ", prefixWidth)`), и в Go `len()` для ASCII = rune count. Если в будущем indent заменят на Unicode-padding, этот расчёт сломается. Не критично сегодня, в design так и предусмотрено.

## Security

No security issues found in changed files. Чисто view-слой, никаких пользовательских inputs не обрабатывается; secrets/auth не затрагиваются.

## Verification Evidence

**Tests (fresh re-run by reviewer):**
```
$ task test
task: [test] go test ./...
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
ok  	github.com/jtprogru/todushka/internal/tui	0.999s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

**Build (fresh):**
```
$ task build
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

**Lint (fresh):**
```
$ task lint
task: [lint] golangci-lint run
0 issues.
```

## Findings

Нет.

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| —  | —        | —    | No findings | —           |

## Recommendations

Нет.

# Implementation Report: List Rendering Polish

## Summary

Реализованы 3 косметических изменения рендера TUI-списка задач (BL-1/BL-3/BL-4):

- `start:`/`due:` убраны из строк `viewList`; даты по-прежнему отображаются в details pane.
- Разделитель зон сменён с `─` (U+2500) на `━` (U+2501) — более жирный визуально.
- Добавлена функция `wrapTitleColumn`; `viewList` теперь soft-wrap'ит длинные заголовки в пределах title-колонки с hanging indent. Strikethrough/faint для completed/cancelled применяется ко всем строкам wrapped block; cursor marker `> ` — только на первой строке задачи.

Все 5 top-level задач выполнены. 15 новых тестов (8 unit + 7 property) проходят; 4 существующих теста-разделителя обновлены под новую руну.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`

## Task Execution

- [x] **T-1** Написать failing-тесты для всех трёх изменений — RED confirmed (6 новых тестов падали на текущем коде до реализации, 9 — locked existing/boundary поведение и сразу passed).
- [x] **T-2** BL-1: удалить даты из `viewList` — GREEN (3 теста зелёные после правки).
- [x] **T-3** BL-3: heavy separator — GREEN (6 тестов: 2 новых + 4 существующих обновлены под `━`).
  - Note: помимо обозначенных в плане двух property-тестов, оказалось ещё 4 unit-теста (`TestRenderSeparator_FullWidth`, `TestView_HasSeparatorsInFullScreen`, `TestView_NoSeparatorsInLegacy`, `TestView_NoSeparatorsInEditor`), которые тоже ищут `─` — обновлены на `━`.
- [x] **T-4** BL-4: `wrapTitleColumn` + рефакторинг `viewList` — GREEN.
  - Note: первый прогон тестов поймал баг в test helper (`firstNonSpaceColumn` неверно вычислял title-start column — prefix начинается с `>`, не пробелом). Заменено на `expectedTitleStartCol` хелпер, который реплицирует prefix structure из `viewList`. Производственный код не менялся.
- [x] **T-5** VERIFY + GATE — все 4 команды (`task test`, `task test-race`, `task lint`, `task build`) прошли.
  - Note: после первого `task lint` поправил staticcheck S1011 в `viewList` (заменил `for _, cont := range titleLines[1:]` на `append(..., titleLines[1:]...)`).

## Final Verification

**Test:**
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
ok  	github.com/jtprogru/todushka/internal/tui	0.894s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

**Test (race):**
```
$ task test-race
task: [test-race] go test -race ./...
ok  	github.com/jtprogru/todushka/cmd/todushka	(cached)
...
ok  	github.com/jtprogru/todushka/internal/tui	3.695s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

**Build:**
```
$ task build
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

**Lint:**
```
$ task lint
task: [lint] golangci-lint run
0 issues.
```

## Files Changed

- `internal/tui/app.go` — `viewList`: убраны даты, добавлен расчёт `paneWidth`/`prefixWidth`, использован `wrapTitleColumn`, per-line strikethrough propagation. `renderSeparator`: `─` → `━`. Новая функция `wrapTitleColumn`.
- `internal/tui/shell_test.go` — обновлены 4 теста разделителей под `━` (`TestRenderSeparator_FullWidth`, `TestView_HasSeparatorsInFullScreen`, `TestView_NoSeparatorsInLegacy`, `TestView_NoSeparatorsInEditor`, `TestProp_SeparatorsConditional`, `TestProp_SeparatorWidth`).
- `internal/tui/list_render_polish_test.go` — **new**: 8 unit + 7 property тестов для BL-1/BL-3/BL-4.

## Notes

- В `TestProp_TitleWrapHangingIndent` принудительно ставится `m.config.DualPaneMinWidth = 200`, чтобы при любых `m.width ∈ [40..80]` оставаться в single-pane режиме — иначе расчёт `paneWidth` через `paneWidths(m)` менял бы инвариант.
- `TestTUI_ViewListNoWrapWhenWidthZero` и `TestProp_NoWrapWidthZero` уже проходили на текущем коде до T-4 (нет wrap-логики вообще) — это OK, они locks boundary behaviour и подтверждают, что после T-4 ничего не сломалось.
- Helper `expectedTitleStartCol` дублирует prefix structure из `viewList`. Если в будущем prefix изменится — тесты выявят рассинхронизацию через явное несовпадение column.

# List Rendering Polish — Task Plan

**Status:** Draft
**Mode:** fast-track
**Work Type:** Pure feature (cosmetic UI improvements — restructures existing rendering, but BL-4 introduces genuinely new wrap behavior).
**Date:** 2026-05-27

## Test Style Source

**Test Style Source:** Tier 2
- Reference test files: `internal/tui/app_test.go`, `internal/tui/shell_test.go`, `internal/tui/filter_test.go`
- Key patterns:
  - `testify/require` для unit-assertions, `pgregory.net/rapid` (`rapid.Check`) для property-based.
  - Helper `setupRapidModel(rt, titles...)` в `internal/tui/app_test.go:598`.
  - Для ANSI-assertions (strikethrough) использовать `lipgloss.SetColorProfile(termenv.TrueColor)` с `t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })` — см. `TestTUI_ViewListRendersStatusIcons` (app_test.go:355) как образец.

## Commands

| Action       | Command           | Source       |
|--------------|-------------------|--------------|
| Test         | `task test`       | Taskfile.yml |
| Test (race)  | `task test-race`  | Taskfile.yml |
| Build        | `task build`      | Taskfile.yml |
| Lint         | `task lint`       | Taskfile.yml |

## Coverage Matrix

| Requirement | Task(s)    | Correctness Property |
|-------------|------------|----------------------|
| REQ-1.1     | T-1, T-2   | CP-1 (Absence)       |
| REQ-1.2     | T-1, T-2   | CP-2 (Equivalence)   |
| REQ-2.1     | T-1, T-3   | CP-3 (Equivalence)   |
| REQ-2.2     | T-1, T-3   | CP-4 (Absence)       |
| REQ-2.3     | T-1, T-3   | CP-3 (Equivalence)   |
| REQ-3.1     | T-1, T-4   | CP-5 (Equivalence)   |
| REQ-3.2     | T-1, T-4   | CP-6 (Absence)       |
| REQ-3.3     | T-1, T-4   | CP-7 (Propagation)   |
| REQ-3.4     | T-1, T-4   | CP-8 (Exclusion)     |

---

## T-1 — Написать failing-тесты для всех трёх изменений (GREEN-stubs)

***_Requirements:_*** REQ-1.1, REQ-1.2, REQ-2.1, REQ-2.2, REQ-2.3, REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4
***_Complexity:_*** standard
***_Test_Style:_*** `internal/tui/app_test.go`, `internal/tui/shell_test.go`

GOAL: записать наблюдаемое поведение (новое — для BL-1/BL-3 это сразу проявит regression в `viewList`/`renderSeparator`; для BL-4 — упадёт пока `wrapTitleColumn` не существует).

### T-1.1 — Добавить unit-тест `TestTUI_ViewListOmitsDates` в `internal/tui/app_test.go`

CRITICAL: тест ДОЛЖЕН упасть на текущем `viewList` (потому что сейчас в output есть `start:` и `due:`).

Структура:
- Setup: создать Model через существующий хелпер `setupRapidModel(...)` (or его не-rapid эквивалент — см. `app_test.go:355`); получить из репозитория задачу и установить `t.StartDate` и `t.Deadline` через прямую модификацию `m.tasks` (так делается в `TestTUI_ViewListRendersStatusIcons`).
- Assert: `out := m.viewList()`; `require.NotContains(t, out, "start:")`; `require.NotContains(t, out, "due:")`.

### T-1.2 — Добавить unit-тест `TestTUI_ViewDetailsKeepsDates` в `internal/tui/details_test.go`

NOTE: проверяет регрессию — БЫЛОЕ поведение `viewDetails` сохраняется (REQ-1.2). Этот тест ДОЛЖЕН пройти СРАЗУ на текущем коде.

- Setup: model с одной задачей, StartDate и Deadline установлены; cursor на этой задаче.
- Assert: `out := viewDetails(m, 60)`; `require.Contains(t, out, "Start:")`; `require.Contains(t, out, "Due:")`.

### T-1.3 — Добавить unit-тесты `TestTUI_RenderSeparatorHeavy` и `TestTUI_RenderSeparatorBoundary` в `internal/tui/shell_test.go`

`TestTUI_RenderSeparatorHeavy`:
- Call: `s := renderSeparator(NewTheme(), 10)`.
- Assert: `require.Equal(t, 10, strings.Count(s, "━"))`; `require.Equal(t, 0, strings.Count(s, "─"))`.

`TestTUI_RenderSeparatorBoundary`:
- `require.Equal(t, "", renderSeparator(NewTheme(), 0))`.
- `require.Equal(t, "", renderSeparator(NewTheme(), -5))`.

CRITICAL: тест `TestTUI_RenderSeparatorHeavy` ДОЛЖЕН упасть на текущем коде.

### T-1.4 — Добавить unit-тест `TestTUI_ViewListWrapsTitleWithHangingIndent` в `internal/tui/app_test.go`

- Setup: создать Model, установить `m.width = 60`, single-pane (i.e. `m.config.DualPaneMinWidth = 100` или ниже DualPaneMinWidth), одна задача с очень длинным title (например, 200 символов `"x"`).
- Assert: `out := m.viewList()`; `lines := strings.Split(out, "\n")`; `require.Greater(t, len(lines), 1)` (wrap произошёл); continuation-lines начинаются с N пробелов где N равно `lipgloss.Width(prefix+marker+icon+short+"  ")` для первой строки. Точное значение N зависит от `id.Short` (8 символов) + остальных префиксов; в тесте посчитать prefix-width динамически по первой строке: индекс первого не-пробельного-non-ANSI символа в lines[0] = индексу title-старта.

CRITICAL: тест ДОЛЖЕН упасть на текущем коде (сейчас одна строка на задачу).

### T-1.5 — Добавить unit-тест `TestTUI_ViewListNoWrapWhenWidthZero` в `internal/tui/app_test.go`

- Setup: model с `m.width = 0`, одна задача с title длины 200.
- Assert: `out := m.viewList()`; `lines := strings.Split(out, "\n")`; `require.Equal(t, 1, len(lines))`.

NOTE: этот тест может пройти на текущем коде (нет wrap логики вообще). Это OK — закрепляет boundary.

### T-1.6 — Добавить unit-тест `TestTUI_ViewListStrikethroughOnWrappedLines` в `internal/tui/app_test.go`

- Setup: `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup`; model `m.width=60`, single-pane; одна задача `Status=Completed`, очень длинный title.
- Assert: `out := m.viewList()`; для каждой строки этой задачи в output — `require.Contains(t, line, "\x1b[9")` (ANSI strikethrough).

CRITICAL: ДОЛЖЕН упасть до реализации BL-4.

### T-1.7 — Добавить unit-тест `TestTUI_ViewListCursorMarkerOnFirstLineOnly` в `internal/tui/app_test.go`

- Setup: `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup`; model `m.width=60`, single-pane; две задачи, cursor=0, первая с длинным wrap-eligible title.
- Assert: посчитать в первом task-блоке вхождения курсорной подстроки `"> "` (через `m.theme.Selected.Render("> ")` precomputed); `require.Equal(t, 1, count)`.

CRITICAL: ДОЛЖЕН упасть до BL-4 (потому что сейчас single-line: уже 1 marker, но после wrap без правильной логики marker может «протечь»).

### T-1.8 — Добавить property-тесты `TestProp_ViewListOmitsDates`, `TestProp_SeparatorHeavy`, `TestProp_SeparatorBoundary`, `TestProp_TitleWrapHangingIndent`, `TestProp_NoWrapWidthZero`, `TestProp_StrikethroughPropagatesAcrossWrap`, `TestProp_SingleCursorMarker` в `internal/tui/shell_test.go` (или новый `internal/tui/list_render_prop_test.go` — на усмотрение).

NOTE: один файл лучше — кладём рядом с существующими `TestProp_*` для консистентности.

Каждый тест — `rapid.Check(t, func(rt *rapid.T) { ... })` следуя шаблону `TestProp_SeparatorsConditional`. Использовать `setupRapidModel(rt, titles...)`. Для титлов с известной длиной — `rapid.StringMatching` или просто конкатенация фиксированных строк.

DO NOT удалять существующий `TestProp_SeparatorWidth` сейчас — это сделается в T-3.

---

## T-2 — Реализовать BL-1: удалить даты из строки списка (CODE)

***_Requirements:_*** REQ-1.1, REQ-1.2
***_Preservation:_*** CP-2 (Equivalence: viewDetails по-прежнему показывает Start/Due)
***_Complexity:_*** mechanical

### T-2.1 — Изменить `viewList` в `internal/tui/app.go`

В `internal/tui/app.go` (функция `viewList`, около строк 617–667):
- Удалить локальные переменные `dates` (line 648) и весь блок `if t.StartDate != nil { ... }` (lines 649–651), `if t.Deadline != nil { ... }` (lines 652–655).
- Удалить ветвь `if dates != "" { dates = done.Render(dates) }` внутри `if t.Status == task.StatusCompleted || ...`.
- Изменить format string в `lines = append(lines, fmt.Sprintf("%s%s%s%s  %s%s", prefix, marker, icon, short, title, dates))` на `lines = append(lines, fmt.Sprintf("%s%s%s%s  %s", prefix, marker, icon, short, title))`.

### T-2.2 — Запустить unit-тесты сегмента BL-1

`task test -- -run "TestTUI_ViewListOmitsDates|TestTUI_ViewDetailsKeepsDates|TestProp_ViewListOmitsDates"`

CRITICAL: все три должны пройти. Если `TestProp_ViewListOmitsDates` падает на edge-case — изучить failed seed под `internal/tui/testdata/rapid/`.

---

## T-3 — Реализовать BL-3: жирные разделители (CODE)

***_Requirements:_*** REQ-2.1, REQ-2.2, REQ-2.3
***_Preservation:_*** CP-3, CP-4 (новые свойства; должны выполняться после правки)
***_Complexity:_*** mechanical

### T-3.1 — Заменить руну в `renderSeparator` в `internal/tui/app.go`

В функции `renderSeparator` (line 564–569):
- Заменить `strings.Repeat("─", width)` на `strings.Repeat("━", width)`.
- Стиль (`theme.Help.Render`) сохранить как есть.

### T-3.2 — Обновить существующие property-тесты в `internal/tui/shell_test.go`

- `TestProp_SeparatorsConditional` (line 477): заменить `fullWidthRule := strings.Repeat("─", m.width)` на `strings.Repeat("━", m.width)`.
- `TestProp_SeparatorWidth` (line 497): заменить `strings.Count(s, "─")` на `strings.Count(s, "━")`. Тест эквивалентен новому `TestProp_SeparatorHeavy` — старый ОСТАВИТЬ, переименование/удаление избыточно, дубликат полезен как regression-lock.

NOTE: ничего не удаляем — оставляем `TestProp_SeparatorWidth` рядом с новыми `TestProp_SeparatorHeavy`/`TestProp_SeparatorBoundary`.

### T-3.3 — Запустить unit-тесты сегмента BL-3

`task test -- -run "TestTUI_RenderSeparatorHeavy|TestTUI_RenderSeparatorBoundary|TestProp_SeparatorHeavy|TestProp_SeparatorBoundary|TestProp_SeparatorWidth|TestProp_SeparatorsConditional"`

CRITICAL: все 6 должны пройти.

---

## T-4 — Реализовать BL-4: wrap title-колонки с hanging indent (CODE)

***_Requirements:_*** REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4
***_Preservation:_*** CP-1, CP-2 (BL-1 не сломан), CP-3 (BL-3 не сломан), CP-7 (strikethrough propagates), CP-8 (single cursor marker)
***_Complexity:_*** complex

### T-4.1 — Добавить функцию `wrapTitleColumn` в `internal/tui/app.go`

Разместить перед `viewList` (в районе line 615). Сигнатура и поведение:

```go
// wrapTitleColumn soft-wraps title to availWidth (lipgloss-width counted)
// and prepends prefixWidth spaces to lines [1..]. Returns []string{title}
// when availWidth <= 0 or prefixWidth <= 0 (no-op safeguard).
func wrapTitleColumn(title string, prefixWidth, availWidth int) []string {
    if availWidth <= 0 || prefixWidth <= 0 {
        return []string{title}
    }
    wrapped := lipgloss.NewStyle().Width(availWidth).Render(title)
    parts := strings.Split(wrapped, "\n")
    if len(parts) == 1 {
        return parts
    }
    indent := strings.Repeat(" ", prefixWidth)
    for i := 1; i < len(parts); i++ {
        parts[i] = indent + parts[i]
    }
    return parts
}
```

IMPORTANT: `lipgloss.NewStyle().Width(...).Render(title)` оборачивает по lipgloss-width — это уже корректно учитывает ANSI escape sequences (которые добавляются позже через `done.Render(title)` для strikethrough). Но: strikethrough должен применяться ПОСЛЕ wrap к каждой строке отдельно, а НЕ до wrap (иначе ANSI escapes ломают расчёт width). См. T-4.2.

### T-4.2 — Переписать `viewList` для использования `wrapTitleColumn` в `internal/tui/app.go`

Текущая логика (lines 627–665) собирает каждую задачу как `fmt.Sprintf(...)` в одну строку. Изменить на:

```go
for i, t := range disp {
    prefix := ""
    if showPrefix { ... } // unchanged

    marker := "  "
    if i == m.cursor { marker = m.theme.Selected.Render("> ") }

    icon := "  "
    switch t.Status { ... } // unchanged

    title := t.Title
    short := m.theme.Dim.Render(id.Short(t.ID))

    // Compute first-line prefix and its rendered width.
    firstLinePrefix := fmt.Sprintf("%s%s%s%s  ", prefix, marker, icon, short)
    prefixWidth := lipgloss.Width(firstLinePrefix)

    // Compute available width for title column.
    var paneWidth int
    if isDualPane(m) {
        paneWidth, _ = paneWidths(m)
    } else {
        paneWidth = m.width
    }
    availWidth := paneWidth - prefixWidth

    titleLines := wrapTitleColumn(title, prefixWidth, availWidth)

    // Apply strikethrough/faint per line for Completed/Cancelled.
    if t.Status == task.StatusCompleted || t.Status == task.StatusCancelled {
        done := lipgloss.NewStyle().Strikethrough(true).Faint(true)
        for j := range titleLines {
            if j == 0 {
                titleLines[j] = done.Render(titleLines[j])
            } else {
                // Continuation line: split into indent + content,
                // style only the content, keep raw spaces.
                content := strings.TrimLeft(titleLines[j], " ")
                indentLen := len(titleLines[j]) - len(content)
                titleLines[j] = strings.Repeat(" ", indentLen) + done.Render(content)
            }
        }
    }

    // Emit lines: first line gets full prefix; continuation lines already
    // have indent baked in.
    lines = append(lines, firstLinePrefix+titleLines[0])
    for _, cont := range titleLines[1:] {
        lines = append(lines, cont)
    }
}
```

CRITICAL:
- При strikethrough для continuation-line сохранять `indent` как сырые пробелы (не оборачивать в `done.Render(indent)`) — иначе CP-5 (hanging indent column alignment) ломается.
- При `m.width == 0` ветвь `paneWidth = m.width = 0`, `availWidth <= 0`, `wrapTitleColumn` возвращает `[]string{title}` (CP-6).
- Курсорный marker `> ` вычисляется один раз и попадает только в `firstLinePrefix` (CP-8).

### T-4.3 — Удалить старую переменную `dates` и связанный код, если что-то осталось

Проверить, что после T-2 в `viewList` нет упоминаний `dates`, `t.StartDate`, `t.Deadline`. Если что-то закрепилось — удалить.

### T-4.4 — Запустить unit-тесты сегмента BL-4

`task test -- -run "TestTUI_ViewListWrapsTitleWithHangingIndent|TestTUI_ViewListNoWrapWhenWidthZero|TestTUI_ViewListStrikethroughOnWrappedLines|TestTUI_ViewListCursorMarkerOnFirstLineOnly|TestProp_TitleWrapHangingIndent|TestProp_NoWrapWidthZero|TestProp_StrikethroughPropagatesAcrossWrap|TestProp_SingleCursorMarker"`

CRITICAL: все 8 должны пройти.

---

## T-5 — VERIFY + GATE (полный прогон + lint + build)

***_Requirements:_*** все (REQ-1.1..3.4)
***_Complexity:_*** mechanical

### T-5.1 — Полный прогон unit + property тестов

```
task test
task test-race
```

CRITICAL: оба должны вернуться без падений. Особое внимание `internal/tui/...` — там основная масса изменений.

### T-5.2 — Lint

```
task lint
```

CRITICAL: golangci-lint zero issues.

### T-5.3 — Build

```
task build
```

CRITICAL: бинарь собирается без ошибок.

### T-5.4 — Финальный traceability-check

Прочитать `coverage matrix` выше и убедиться, что каждый REQ закрыт ≥1 прошедшим тестом. Список тестов:
- REQ-1.1, 1.2: `TestTUI_ViewListOmitsDates`, `TestTUI_ViewDetailsKeepsDates`, `TestProp_ViewListOmitsDates`
- REQ-2.1, 2.2, 2.3: `TestTUI_RenderSeparatorHeavy`, `TestTUI_RenderSeparatorBoundary`, `TestProp_SeparatorHeavy`, `TestProp_SeparatorBoundary`, `TestProp_SeparatorsConditional` (updated)
- REQ-3.1, 3.2, 3.3, 3.4: `TestTUI_ViewListWraps...` (×4), `TestProp_TitleWrap...`/`NoWrap...`/`Strikethrough...`/`SingleCursor...` (×4)

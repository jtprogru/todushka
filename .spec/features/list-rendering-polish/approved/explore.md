---
feature: list-rendering-polish
phase: explore
mode: fast-track
backlog_items: [BL-1, BL-3, BL-4]
---

# Exploration: List Rendering Polish

## Intent

Три косметических улучшения рендеринга списка задач, накопившихся в `.spec/BACKLOG.md`. Список (`viewList`) сейчас перегружен `start:YYYY-MM-DD due:YYYY-MM-DD` хвостом справа от заголовка, разделители зон тонкие и сливаются с фоном, а длинные заголовки задач переносятся от левого края экрана — это ломает вертикальное выравнивание колонок icon/short. Цель — убрать даты из строки списка (они уже есть в details pane), сделать разделители визуально заметнее и реализовать перенос строки строго внутри title-колонки с hanging indent.

## Investigation

- `internal/tui/app.go:617-666` — `viewList` собирает строку через `fmt.Sprintf("%s%s%s%s  %s%s", prefix, marker, icon, short, title, dates)`. Поля `start`/`due` приклеиваются к `dates` (lines 648-655) и идут после title без какого-либо контроля ширины.
- `internal/tui/app.go:562-569` — `renderSeparator(theme, width)` рендерит `strings.Repeat("─", width)` через `theme.Help` (italic subtext). Других разделителей в TUI нет.
- `internal/tui/details.go:140-181` — `viewDetails` уже показывает `Start:`/`Due:` отдельными строками с двойным пробелом перед значением. Удаление дат из списка не теряет информацию — она остаётся в правой панели.
- `internal/tui/details.go:103-113` — `wrapAndTruncate(text, width, maxLines)` использует `lipgloss.NewStyle().Width(width).Render(text)` для soft-wrap. Применима как утилита для wrapping title-колонки в списке, но потребуется добавить hanging indent для второй+ строк.
- Тесты, которые сейчас завязаны на текущее поведение:
  - `internal/tui/shell_test.go:477` `TestProp_SeparatorsConditional` — ищет `strings.Repeat("─", m.width)` в `View()`. При смене символа разделителя требуется обновить.
  - `internal/tui/shell_test.go:497` `TestProp_SeparatorWidth` — считает руны `─` в выводе `renderSeparator`. Аналогично.
  - `internal/tui/app_test.go:355-373` `TestTUI_ViewListRendersStatusIcons` — проверяет наличие `✓ `/`✗ ` иконок и `\x1b[9...` (strikethrough). Дат не касается, должен пройти без правок.
  - Прочие тесты `viewList` (`filter_test.go:119`, `app_test.go:439/448/658`) проверяют title/short — даты не утверждают.

## Build Tooling

- **Orchestrator:** Taskfile.yml
- **Test:** `task test` (alias `go test ./...`), `task test-race` для race detector
- **Build:** `task build`
- **Lint:** `task lint` (golangci-lint)
- **Source:** `Taskfile.yml`

## Recommended Direction

Один fast-track цикл, без новых файлов:

1. **BL-1 (удаление дат из списка):** в `viewList` удалить блок построения `dates` (`app.go:648-655`) и заменить format string на `"%s%s%s%s  %s"` без `dates`. Тривиально.
2. **BL-3 (жирнее разделители):** заменить руну `─` (`U+2500 BOX DRAWINGS LIGHT HORIZONTAL`) на `━` (`U+2501 BOX DRAWINGS HEAVY HORIZONTAL`) внутри `renderSeparator`. Опционально — добавить `.Bold(true)` к стилю. Обновить два теста.
3. **BL-4 (перенос внутри title-колонки):** добавить новую утилиту `renderTitleColumn(title, prefix, width)` (или inline в `viewList`), которая:
   - считает «фиксированную» ширину префикса = `lipgloss.Width(prefix + marker + icon + short + "  ")`,
   - доступная ширина title = `paneWidth - prefixWidth`,
   - оборачивает title через `lipgloss.NewStyle().Width(availTitleWidth).Render(title)`,
   - для строк со 2-й и далее добавляет отступ из пробелов длиной `prefixWidth` (hanging indent).
   
   `paneWidth` берётся из `paneWidths(m)` для dual-pane или из `m.width` для single-pane. Если `m.width == 0` (initial state) — wrap пропускается, рендер как сейчас.

## Scope Boundaries

- **Must-have (v1):** BL-1, BL-3, BL-4 как описано выше.
- **Deferred (v2):** настройка thicker-separator через config (символ как опция), wrap-policy: truncate-with-ellipsis vs full-wrap (сейчас выбираем full-wrap).
- **Needs spike:** none — все три пункта локализованы в `viewList` и `renderSeparator`.

## Constraints & Risks

- **Hanging indent + selection marker (`> `):** marker рендерится через `m.theme.Selected.Render("> ")` (ANSI-стилизован). При hanging indent indent рассчитывается через `lipgloss.Width()`, что корректно учитывает ANSI. Проверяемое инвариант — все строки в выводе wrap-блока должны иметь одинаковую `lipgloss.Width()` отступа до title.
- **Filter highlights / strikethrough:** completed/cancelled задачи стилизуются через `lipgloss.NewStyle().Strikethrough(true).Faint(true).Render(title)` ДО forming. После wrap нужно убедиться, что strikethrough применяется к каждой строке wrapped block корректно (lipgloss handles это через style.Width().Render()). Проверим в тесте.
- **`m.width == 0` initial state:** до первого `tea.WindowSizeMsg` ширина = 0. Тогда title-wrap должен no-op (отдать title как есть).
- **Property test `TestProp_SeparatorWidth`:** считает `─` руны. После смены на `━` обновим на новую руну.

## Assumptions & Open Questions

- [ASSUMPTION: пользователь согласен на потерю дат из строки списка — они остаются в details pane (BL-1.1 позже сделает их там жирными)].
- [ASSUMPTION: `━` (HEAVY HORIZONTAL) выглядит «жирнее» приемлемо без дополнительного `.Bold(true)`. Если визуально мало — добавим bold во время implementation. Решение об этом можно отложить до visual review].
- [ASSUMPTION: wrap-policy = full-wrap (показывать все строки). Альтернатива «truncate с …» осознанно отложена — Things 3 ведёт себя как full-wrap].

**Open questions:** None identified.

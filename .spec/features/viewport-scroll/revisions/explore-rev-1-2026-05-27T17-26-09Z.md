# Exploration: Viewport Scroll (BL-7)

## Intent

Когда количество задач/проектов больше чем доступная высота экрана,
курсор по `j/k` уходит «вниз за обрез» — пользователь не видит, где
он находится. Нужно реализовать viewport со scroll-offset, который
следует за курсором (edge-follow + 3-строчный buffer, vim-like).
Затрагивает три render-функции: `viewList` (screenList),
`viewProjectList` (screenProjects), `viewProjectTasks`
(screenProjectTasks).

## Root Cause

`internal/tui/app.go:638-695` — `viewList` итерирует **весь** `disp`
slice без учёта доступной высоты. `View()` в строках 597-603 считает
`bodyH` и обрезает overflow через `lipgloss.MaxHeight(bodyH)` — но
**не сдвигает** содержимое в зависимости от позиции курсора. Когда
`m.cursor >= bodyH`, render отрисовывает курсор, но lipgloss
обрезает строки cверху-вниз, и `> task N` оказывается ниже
видимой области.

Аналогично:
- `internal/tui/project_list.go:viewProjectList` — итерирует
  `disp := displayedProjects(m)` целиком.
- `internal/tui/project_tasks.go:viewProjectTasks` — итерирует
  `disp := displayedTasks(m)` целиком.

Это не регрессия от BL-5 — баг существовал в `viewList` с самого
начала dual-pane work (BL-2 наследовал старую логику). BL-5 просто
унаследовал паттерн в новые render-функции.

## Build Tooling

- **Orchestrator:** Taskfile.yml
- **Test:** `task test` / `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`

## Recommended Direction

Добавить scroll-offset в Model и адаптировать рендер:

1. **Model fields:** `m.scrollOffset int` (для `m.cursor` в screenList
   и screenProjectTasks — общий, т.к. m.tasks mirrors projectTasks
   в zoom) и `m.projectScrollOffset int` (для `m.projectCursor` в
   screenProjects).

2. **Helper** в новом файле `internal/tui/viewport.go`:
   ```go
   func ensureCursorVisible(cursor, scrollOff, visibleCount, scrolloff int) int
   ```
   Возвращает новый scrollOff:
   - Если `cursor < scrollOff + scrolloff` → `scrollOff = max(0, cursor - scrolloff)`
   - Если `cursor >= scrollOff + visibleCount - scrolloff` → `scrollOff = cursor - visibleCount + scrolloff + 1`
   - Clamp в `[0, max(0, totalCount - visibleCount)]`

3. **Helper** `m.visibleRows() int` — возвращает `bodyH` = `m.height -
   headerH - footerH - 2`. Headers/footers рендерятся для замера высоты
   (стоимость пренебрежимая на keypress).

4. **j/k handlers** (3 места: screenList, screenProjects,
   screenProjectTasks) после изменения `cursor` вызывают
   `ensureCursorVisible` и обновляют scroll-offset.

5. **Reset scrollOffset** при: `switchList`, входе в screenProjects,
   выходе из zoom, переключении filter, любом действии что меняет
   `m.tasks`/`m.projects` (включая `tasksLoadedMsg`,
   `projectsLoadedMsg`, `projectTasksLoadedMsg`). Минимальная
   стратегия: при reload — clamp offset, не сбрасывать в 0.

6. **Render** — viewList / viewProjectList / viewProjectTasks слайсят
   `disp[scrollOffset : scrollOffset + visibleCount]` (с границами).

**Scrolloff const** = `3` per выбор пользователя.

**Логические vs визуальные строки:** scroll-математика работает в
терминах **логических задач/проектов** (одна задача = 1 unit),
независимо от того что `wrapTitleColumn` может развернуть длинный
title на несколько screen-строк. Это типичное поведение и понятная
mental model. lipgloss-clamp в `View()` остаётся как safety net для
extreme wrap-overflow на последнем визуальном blockе.

## Scope Boundaries

### Must-have (v1)
- Scroll-offset для всех трёх view'ов (viewList, viewProjectList,
  viewProjectTasks).
- Edge-follow + scrolloff=3.
- Reset/clamp при перезагрузке списка.
- Tests: unit (scroll math) + integration (cursor над пределом виден).

### Deferred (v2)
- PageUp/PageDown bindings (большие прыжки).
- Home/End (G/gg vim-style).
- Visual scrollbar indicator (например, `▲▼` маркеры справа).
- Mouse-wheel scrolling.
- Учёт wrapped title height в scroll-математике (logical-row подход
  достаточен для v1).

### Needs spike
- (нет) — bug fix с известным reproduction и понятной механикой.

## Assumptions & Open Questions

- [ASSUMPTION: scrolloff=3 — фиксированная константа, не config.
  Если пользователь захочет настройку → отдельный pipeline.]
- [ASSUMPTION: scroll работает на logical-row level (1 task =
  1 unit), не на screen-line level. Wrap длинного title не
  учитывается в scroll math — мы принимаем небольшой visual jitter
  на хвосте overflow при множественных длинных titles подряд.]
- [ASSUMPTION: одно поле `m.scrollOffset` для tasks работает и в
  screenList, и в screenProjectTasks — потому что m.tasks mirrors
  m.projectTasks в zoom. При zoom/zoom-out reset offset = 0.]

**Open questions:** None identified.

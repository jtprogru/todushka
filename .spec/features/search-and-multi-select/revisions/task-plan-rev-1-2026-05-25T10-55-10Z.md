# Search & Multi-Select for TUI — Task Plan

## Preamble

### Work Type Classification

**Pure feature** — новая функциональность с partial-preservation: код keymap'а и per-action хелперов модифицируется, но публичный контракт (поведение `c`/`x`/`d`/`p` при пустом selection) сохраняется. Преservation-тесты на T-1 фиксируют этот контракт перед началом изменений.

### Test Style Source

**Tier 2** — adjacent tests
- **Reference unit tests:** `internal/tui/app_test.go` (testify `require`, table-driven, fixture `newTestModel(t)` поверх `fakes.New()` + `fixedClock{}`, прямой `Update(tea.KeyMsg{...})` без bubbletea-runtime, `cmd()` вызывается и проверяется тип `tea.Msg`).
- **Reference property tests:** `internal/domain/repeat/*_test.go` (pattern `rapid.Check(t, func(t *rapid.T) { ... })` с генераторами через `rapid.SliceOfN`, `rapid.IntRange`, и т.п.).
- **Key patterns to follow:**
  - Naming: `TestXxx` (unit), `prop_Xxx` (property) в обычных `func TestXxx(t *testing.T)`-оболочках.
  - Assertions: `require.X(t, ...)` (не `assert`).
  - Cmd verification: `msg := cmd(); require.IsType(t, expectedMsg, msg)`.
  - Table-driven: `name`-keyed cases в `[]struct{...}`.

### Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test race  | `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

### Coverage Matrix

| Requirement | Task(s) | Correctness Property |
|-------------|---------|----------------------|
| REQ-1.1 (`/` enters Filter Mode) | T-3 | CP-2 (Propagation: Filter state transitions) |
| REQ-1.2 (live substring match) | T-3, T-7 | CP-1 (Equivalence: Filter result) |
| REQ-1.3 (Enter saves query) | T-3 | CP-2 |
| REQ-1.4 (Esc clears query) | T-3 | CP-2 |
| REQ-1.5 (`(no matches)` placeholder) | T-3 | CP-3 (Propagation: Empty visible placeholder) |
| REQ-1.6 (list switch clears filter) | T-4 | CP-4 (Equivalence: switchList resets) |
| REQ-1.7 (whitespace = empty) | T-3 | CP-1 |
| REQ-2.1 (`Space` toggle) | T-2 | CP-5 (Round-trip: Space involutive) |
| REQ-2.2 (`[x]`/`[ ]` prefix when selected) | T-2 | CP-6 (Exclusion: Prefix iff non-empty) |
| REQ-2.3 (no prefix when empty) | T-2 | CP-6 |
| REQ-2.4 (`*` selects visible) | T-2 | CP-7 (Equivalence: Star selects all visible) |
| REQ-2.5 (Esc clears selection) | T-2 | CP-8 (Equivalence: Esc clears) |
| REQ-2.6 (filter hides → drop) | T-4 | CP-9 (Absence: Selection ⊆ visible) |
| REQ-2.7 (list switch clears selection) | T-4 | CP-4 |
| REQ-2.8 (`Selected: N` counter) | T-2 | CP-10 (Propagation: Counter in status) |
| REQ-3.1 (empty → per-cursor) | T-1 (preservation), T-5 | CP-11 (Equivalence: Empty action ≡ cursor) |
| REQ-3.2 (1≤N<5 no confirm) | T-5 | CP-12 (Exclusion: Threshold gate) |
| REQ-3.3 (N≥5 confirm) | T-5 | CP-12 |
| REQ-3.4 (only `y` confirms) | T-5 | CP-13 (Exclusion: Only y) |
| REQ-3.5 (aggregate partial failure) | T-5 | CP-14 (Equivalence: Aggregate math) |
| REQ-3.6 (clear after bulk success) | T-5 | CP-15 (Equivalence: Success clears) |
| REQ-3.7 (fatal preserves selection) | T-5 | CP-16 (Absence: Fatal preserves) |
| REQ-4.1 (help includes new keys) | T-6 | CP-17 (Propagation: Help) |
| REQ-4.2 (footer hints) | T-6 | CP-18 (Propagation: Footer) |

Каждое REQ покрыто ≥1 task'ом. Каждое CP покрыто ≥1 property-test в T-7.

---

## Task Order

```
T-1 GREEN (preservation: per-cursor backward compat)
  → T-2 CODE (Selection foundation + tests)
    → T-3 CODE (Filter foundation + tests)
      → T-4 CODE (Filter-Selection bridge + list-switch cleanup + tests)
        → T-5 CODE (Bulk dispatch + Confirm modal + tests)
          → T-6 CODE (Help & footer hints + tests)
            → T-7 GREEN (Property-based tests batch)
              → T-8 GATE (Checkpoint)
```

---

## Task: T-1 — Write preservation tests for per-cursor backward compatibility

*_Requirements: REQ-3.1_*
*_Test_Style: Tier 2 (`internal/tui/app_test.go`)_*
*_Complexity: standard_*

GOAL: Зафиксировать текущее поведение `c`/`x`/`d`/`p` (over cursor task) до того как будет введён bulk-dispatcher. Эти тесты должны проходить **до**, **во время** и **после** реализации фичи.

IMPORTANT: Эти тесты не описывают новую функциональность — они описывают **существующий** контракт, который мы НЕ должны сломать. После T-5 они должны продолжать проходить с `len(selected) == 0`.

DO NOT: Изменять production-код в этой задаче.

Subtasks:
- [ ] 1. В `internal/tui/app_test.go` добавить `TestTUI_CompleteCursorTaskWhenNoSelection`: создать Model через `newTestModel(t)`, через service создать 2 task'и (`svc.AddTask(...)`), вызвать `m.Init()`, выполнить `cmd := m.loadCurrentList(); m.tasks = (msg of cmd).tasks`, поставить `m.cursor = 0`, вызвать `m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})`, выполнить возвращённый Cmd, перезагрузить список и проверить что первый task имеет `Status == StatusCompleted`. — `task test`
- [ ] 2. В `internal/tui/app_test.go` добавить `TestTUI_CancelCursorTaskWhenNoSelection` по той же схеме с клавишей `x` и проверкой `Status == StatusCancelled`. — `task test`
- [ ] 3. В `internal/tui/app_test.go` добавить `TestTUI_DeleteCursorTaskWhenNoSelection` с клавишей `d` и проверкой через `svc.Repo().TaskGet(...)` что возвращается `storage.ErrNotFound` (или `task.DeletedAt != nil` если soft-delete). — `task test`
- [ ] 4. В `internal/tui/app_test.go` добавить `TestTUI_PinCursorTaskWhenNoSelection` с клавишей `p` и проверкой `PinnedToday != nil`. — `task test`
- [ ] 5. Запустить `task test` — все четыре теста должны пройти на текущей кодовой базе (зелёный baseline). Зафиксировать коммитом `test(tui): lock per-cursor action behavior before multi-select rewrite`.

After all subtasks: Run `task lint` to confirm no style errors.

---

## Task: T-2 — Implement Selection foundation

*_Requirements: REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4, REQ-2.5, REQ-2.8_*
*_Preservation: CP-11 (empty selection → per-cursor behavior unchanged), все REQ из T-1_*
*_Test_Style: Tier 2 (`internal/tui/app_test.go`)_*
*_Complexity: standard_*

GOAL: Добавить `Model.selected map[id.ID]struct{}` и связанную state-машину: `Space` toggle, `*` select-all-visible, `Esc` clear, render `[x] `/`[ ] ` префикса только при непустом selection, `Selected: N` в footer.

CRITICAL: На этом этапе `c`/`x`/`d`/`p` ещё НЕ должны работать над `selected` — только cursor. Bulk dispatch — отдельный T-5. Тесты T-1 должны продолжать проходить.

IMPORTANT: После каждого subtask запускать `task test`. Если падает что-то из T-1 — немедленно остановиться и разобраться.

Subtasks:
- [ ] 1. В `internal/tui/app.go` добавить поле `selected map[id.ID]struct{}` в struct `Model` (после `width int`). В `NewModel` инициализировать `selected: make(map[id.ID]struct{})`. — `task test`
- [ ] 2. В `internal/tui/keys.go` добавить четыре новых binding'а в `KeyMap`: `ToggleSelect` (key `" "` / "space"), `SelectAll` (key `*`), `ClearSelection` (key `esc`). В `DefaultKeyMap()` инициализировать соответствующими `key.NewBinding(...)` с help-описаниями. — `task test`
- [ ] 3. В `internal/tui/app_test.go` добавить `TestSelection_SpaceToggle`: создать Model + 1 task в списке, поставить cursor на 0, нажать Space → проверить `len(m.selected) == 1`; нажать Space ещё раз → проверить `len(m.selected) == 0`. — `task test` (должен упасть, реализации ещё нет)
- [ ] 4. В `internal/tui/app.go` в `handleKey` (после блока списка-keys, перед `Up`/`Down`) добавить case `key.Matches(msg, m.keys.ToggleSelect)`: если `m.selectedTask() != nil`, toggle `m.selected[sel.ID]`. — `task test` (Space-тест должен теперь пройти)
- [ ] 5. В `internal/tui/app_test.go` добавить `TestSelection_StarSelectsAllVisible`: 3 task'и в списке, нажать `*` → проверить `len(m.selected) == 3` и что все IDs из `m.tasks` присутствуют в `m.selected`. — `task test`
- [ ] 6. В `internal/tui/app.go` добавить case `key.Matches(msg, m.keys.SelectAll)`: для каждой `t` в `m.tasks` (на этом шаге `displayedTasks` ещё нет — используем `m.tasks` напрямую; T-3 обновит на `displayedTasks(m)`) добавить `t.ID` в `m.selected`. — `task test`
- [ ] 7. В `internal/tui/app_test.go` добавить `TestSelection_EscClearsSelection`: с непустым `selected`, нажать Esc → проверить `len(m.selected) == 0`. — `task test`
- [ ] 8. В `internal/tui/app.go` в `handleKey` добавить case `key.Matches(msg, m.keys.ClearSelection) && len(m.selected) > 0` (ставить **до** общего Esc handler'а quickEntry/editor): очистить `m.selected` через `m.selected = make(map[id.ID]struct{})`. — `task test`
- [ ] 9. В `internal/tui/app_test.go` добавить `TestSelection_PrefixHiddenWhenEmpty`: с `len(selected) == 0`, вызвать `m.viewList()` → проверить что результат НЕ содержит `[x]` или `[ ]`. И `TestSelection_PrefixVisibleWhenNonEmpty`: выделить 1 task, проверить что `viewList()` содержит `[x] ` (для выделенной) и `[ ] ` (для остальных). — `task test`
- [ ] 10. В `internal/tui/app.go` в `viewList` (строки 398-421) изменить cycle: если `len(m.selected) > 0`, перед `marker := "  "` блоком добавить prefix `[x] ` или `[ ] ` в зависимости от `_, ok := m.selected[t.ID]`. — `task test`
- [ ] 11. В `internal/tui/app_test.go` добавить `TestSelection_StatusBarCounter`: с `len(selected) == 3`, вызвать `m.viewFooter()` → проверить что результат содержит подстроку `Selected: 3`. — `task test`
- [ ] 12. В `internal/tui/app.go` в `viewFooter` добавить блок: если `len(m.selected) > 0`, в `right` добавить `m.theme.Selected.Render(fmt.Sprintf("Selected: %d", len(m.selected)))` через separator. — `task test`

After all subtasks: Run `task test-race && task lint`. Все тесты из T-1 ДОЛЖНЫ продолжать проходить (CP-11 preservation).

---

## Task: T-3 — Implement Filter foundation

*_Requirements: REQ-1.1, REQ-1.2, REQ-1.3, REQ-1.4, REQ-1.5, REQ-1.7_*
*_Preservation: все CP из T-2, CP-11, тесты T-1_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Добавить Filter Mode: `/` входит, printable runes расширяют `filterQuery`, `Esc` чистит и выходит, `Enter` сохраняет и выходит. `displayedTasks(m)` возвращает substring-фильтрованный slice. `viewList` отображает `(no matches)` если фильтр активен и результат пуст. `*` теперь работает над `displayedTasks(m)`.

IMPORTANT: На этом этапе Filter-Selection bridge (REQ-2.6) ещё НЕ реализуется — это T-4. Если изменение фильтра скрывает выделенную task'у, она остаётся в `selected` до T-4.

Subtasks:
- [ ] 1. В `internal/tui/app.go` добавить поля `filterQuery string` и `filtering bool` в `Model` (рядом с `selected`). — `task test`
- [ ] 2. В `internal/tui/keys.go` добавить `Filter key.Binding` в `KeyMap`. В `DefaultKeyMap()`: `Filter: key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))`. — `task test`
- [ ] 3. Создать новый файл `internal/tui/filter.go` с функцией `foldCaseContains(haystack, needle string) bool`. Реализация: `strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))`. `strings.ToLower` использует Unicode-aware lower-case в Go. — `task test`
- [ ] 4. В `internal/tui/filter.go` добавить функцию `displayedTasks(m Model) []task.Task`: если `strings.TrimSpace(m.filterQuery) == ""`, вернуть `m.tasks`; иначе вернуть новый slice с теми `t`, где `foldCaseContains(t.Title, strings.TrimSpace(m.filterQuery)) == true`, сохраняя порядок. — `task test`
- [ ] 5. Создать новый файл `internal/tui/filter_test.go` с тестом `TestDisplayedTasks_EmptyQuery`: создать Model с 3 tasks, `filterQuery = ""` → `displayedTasks(m)` равен `m.tasks`. Добавить `TestDisplayedTasks_WhitespaceQuery`: `filterQuery = "   "` → результат равен `m.tasks` (REQ-1.7). Добавить `TestDisplayedTasks_SubstringMatch`: tasks с titles ["Купить молоко", "Помыть машину", "Сходить в магазин"], query "ом" → 2 совпадения. Добавить `TestDisplayedTasks_CaseInsensitive`: query "купить" → matches "Купить молоко". — `task test`
- [ ] 6. В `internal/tui/filter_test.go` добавить `TestFilter_SlashEntersMode`: нажать `/` → `m.filtering == true`, `m.filterQuery == ""`. Печатать руны 'к','у','п' через `Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'к'}})` и т.д. → `m.filterQuery == "куп"`. — `task test` (должен упасть, реализации нет)
- [ ] 7. В `internal/tui/filter.go` добавить функцию `handleFilterKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd)`: обрабатывает Esc (filterQuery="", filtering=false), Enter (filtering=false), Backspace (отрезать последнюю руну от filterQuery), KeyRunes (append к filterQuery). — `task test`
- [ ] 8. В `internal/tui/app.go` в `handleKey` ДО общего `switch` блока добавить early-return: `if m.filtering { return handleFilterKey(m, msg) }`. И добавить новый case в общем switch: `case key.Matches(msg, m.keys.Filter)`: `m.filtering = true; m.filterQuery = ""; return m, nil`. — `task test`
- [ ] 9. В `internal/tui/filter_test.go` добавить `TestFilter_EnterPreservesQuery`: войти в filter mode, ввести "куп", нажать Enter → `m.filtering == false`, `m.filterQuery == "куп"`. И `TestFilter_EscClearsQuery`: то же самое но Esc → `m.filtering == false`, `m.filterQuery == ""`. — `task test`
- [ ] 10. В `internal/tui/app.go` в `viewList` (строка 399 — проверка `len(m.tasks) == 0`) заменить на `disp := displayedTasks(m); if len(disp) == 0 { ... if m.filterQuery != "" { return "(no matches)" }; return "(no tasks)" }`. И заменить цикл `for i, t := range m.tasks` на `for i, t := range disp`. — `task test`
- [ ] 11. В `internal/tui/filter_test.go` добавить `TestFilter_NoMatchesPlaceholder`: tasks с titles, filterQuery без совпадений → `m.viewList()` содержит `(no matches)`. — `task test`
- [ ] 12. В `internal/tui/app.go` в `viewFooter` добавить блок: если `m.filtering`, заменить `hints` на `"Filter: " + m.filterQuery + "_  Enter=save  Esc=cancel"`. — `task test`
- [ ] 13. В `internal/tui/app.go` в case для `SelectAll` (T-2 subtask 6) заменить `for _, t := range m.tasks` на `for _, t := range displayedTasks(m)`. — `task test`

After all subtasks: Run `task test-race && task lint`. Все тесты из T-1 и T-2 продолжают проходить.

---

## Task: T-4 — Implement Filter-Selection bridge and list-switch cleanup

*_Requirements: REQ-1.6, REQ-2.6, REQ-2.7_*
*_Preservation: все CP из T-2, T-3, CP-11_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Связать filter и selection: при изменении filterQuery выделенные task'и, выпавшие из видимости, удаляются из `selected` (CP-9). При смене активного списка (Tab/Shift+Tab/1-6) оба state очищаются (CP-4).

Subtasks:
- [ ] 1. В `internal/tui/filter_test.go` добавить `TestFilter_ChangeDropsHiddenFromSelection`: создать 3 task'и с титлами ["A молоко", "B хлеб", "C молоко"], войти в filter mode, выделить task с ID `tasks[0]` и `tasks[2]` (оба содержат "молоко"). Затем ввести в filter "хлеб" — `displayedTasks(m)` теперь только task `tasks[1]`. Проверить что `len(m.selected) == 0` (оба ID из selected были скрыты фильтром). — `task test` (должен упасть)
- [ ] 2. В `internal/tui/filter.go` в `handleFilterKey` после любого изменения `filterQuery` (KeyRunes, Backspace) добавить вызов хелпера `pruneSelection(&m)`. Определить функцию `pruneSelection(m *Model)`: построить set видимых ID через `displayedTasks(*m)`, затем для каждого ID в `m.selected` если его нет в visible — удалить. — `task test`
- [ ] 3. В `internal/tui/app_test.go` добавить `TestSwitchList_ClearsFilterAndSelection`: создать Model в screenList=listInbox, установить `m.filterQuery="x"`, `m.filtering=true`, `m.selected={id1, id2}`. Вызвать `m.Update(tea.KeyMsg{Type: tea.KeyTab})`. Проверить `m.filterQuery == ""`, `m.filtering == false`, `len(m.selected) == 0`. Повторить для Shift+Tab и для клавиш '1'-'6'. — `task test`
- [ ] 4. В `internal/tui/app.go` в функции `switchList(l listKind)` (строка 258) перед `m.activeList = l` добавить три строки: `m.filterQuery = ""`, `m.filtering = false`, `m.selected = make(map[id.ID]struct{})`. — `task test`

After all subtasks: Run `task test-race && task lint`. T-1, T-2, T-3 тесты продолжают проходить.

---

## Task: T-5 — Implement Bulk dispatch and Confirm modal

*_Requirements: REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4, REQ-3.5, REQ-3.6, REQ-3.7_*
*_Preservation: CP-11 (T-1 тесты должны продолжать проходить — empty selection = per-cursor)_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Заменить `completeSelected`/`cancelSelected`/`deleteSelected`/`pinSelected` хелперы на единый `dispatch(m, action) (Model, tea.Cmd)`. Логика: пустой `selected` → fallthrough на cursor; `1 ≤ N < 5` → `runBulk` Cmd сразу; `N ≥ 5` → `m.confirm = &confirmState{...}`, рендер модалки, ожидание `y` или другой клавиши.

CRITICAL: T-1 preservation тесты ДОЛЖНЫ продолжать проходить — они проверяют контракт `empty selected → cursor`.

IMPORTANT: `runBulk` запускается через `tea.Cmd` (асинхронно), возвращает `bulkResultMsg`. Обработка в `Update` отдельным `case`.

Subtasks:
- [ ] 1. В `internal/tui/msgs.go` добавить тип `bulkResultMsg struct { action bulkAction; succeeded int; failed int; lastErr error; fatal bool }`. — `task test`
- [ ] 2. Создать новый файл `internal/tui/bulk.go` с типом `bulkAction int` и константами `bulkActionComplete bulkAction = iota; bulkActionCancel; bulkActionDelete; bulkActionPin`. Добавить `const bulkConfirmThreshold = 5`. Добавить тип `confirmState struct { action bulkAction; ids []id.ID }`. — `task test`
- [ ] 3. В `internal/tui/app.go` добавить поле `confirm *confirmState` в `Model` (рядом с `selected`). В `NewModel` поле остаётся `nil` по умолчанию. — `task test`
- [ ] 4. В `internal/tui/bulk_test.go` (новый файл) добавить `TestBulk_EmptyDispatchesPerCursor`: с `len(selected) == 0`, cursor на task, вызвать `m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})` → выполнить Cmd, проверить что cursor task имеет `Status == StatusCompleted`. (Это эквивалент T-1 теста, но через новый dispatcher.) — `task test`
- [ ] 5. В `internal/tui/bulk.go` добавить функцию `dispatch(m Model, action bulkAction) (Model, tea.Cmd)`: если `len(m.selected) == 0`, fallthrough — вызвать соответствующий single-action хелпер (логика из текущих `completeSelected`/`cancelSelected`/etc.). Если `len(m.selected) < bulkConfirmThreshold`, собрать `ids := slices.Collect(maps.Keys(m.selected))`, вернуть `(m, runBulk(m.service, action, ids))`. Если `len(m.selected) >= bulkConfirmThreshold`, установить `m.confirm = &confirmState{action: action, ids: ids}`, вернуть `(m, nil)`. — `task test`
- [ ] 6. В `internal/tui/bulk.go` добавить функцию `runBulk(svc *app.Service, action bulkAction, ids []id.ID) tea.Cmd`: возвращает Cmd, который последовательно вызывает соответствующий service-метод для каждого ID, считая `succeeded` и `failed`. Если `errors.Is(err, context.Canceled)` или `errors.Is(err, storage.ErrDatabaseLocked)` — установить `fatal = true` и прервать. Вернуть `bulkResultMsg{...}`. — `task test`
- [ ] 7. В `internal/tui/app.go` в `handleKey` заменить четыре case'а (`Complete`/`Cancel`/`Delete`/`PinToday`) на вызовы `dispatch(m, bulkActionComplete)` и т.д. Удалить старые методы `completeSelected`/`cancelSelected`/`deleteSelected`/`pinSelected` ИЛИ переименовать их в private helpers, которые вызывает `dispatch` при пустом `selected` (как фолбэк per-cursor). — `task test`
- [ ] 8. В `internal/tui/app.go` в `Update` добавить новый `case bulkResultMsg`: если `msg.fatal == false`, очистить `m.selected = make(map[id.ID]struct{})` и вызвать `m.loadCurrentList()`; всегда установить `m.statusMsg` в формате `"<Action>: <succeeded>/<total> succeeded, <failed> failed"` если `failed > 0`, иначе `"<Action>: <succeeded> done"`. Если `fatal`, не очищать `selected`, поставить статус в `lastErr.Error()`. — `task test`
- [ ] 9. В `internal/tui/bulk_test.go` добавить `TestBulk_BelowThresholdNoConfirm`: выделить 4 task'и, нажать `c` → проверить `m.confirm == nil`, Cmd не nil. И `TestBulk_AtThresholdShowsConfirm`: выделить 5 task'ей, нажать `c` → проверить `m.confirm != nil`, Cmd == nil. — `task test`
- [ ] 10. В `internal/tui/bulk.go` добавить функцию `handleConfirmKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd)`: проверить `msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'y'`. Если да — взять `m.confirm.action` и `m.confirm.ids`, обнулить `m.confirm = nil`, вернуть `(m, runBulk(m.service, action, ids))`. Если нет — обнулить `m.confirm`, вернуть `(m, nil)`. — `task test`
- [ ] 11. В `internal/tui/app.go` в `handleKey` перед общим switch добавить early-return: `if m.confirm != nil { return handleConfirmKey(m, msg) }`. Этот early-return должен быть ПОСЛЕ filter early-return (m.filtering), чтобы фильтр имел приоритет. — `task test`
- [ ] 12. В `internal/tui/bulk_test.go` добавить `TestBulk_YConfirmsAndDispatches`: с `m.confirm != nil`, нажать `y` → `m.confirm == nil`, Cmd возвращает `bulkResultMsg`. И `TestBulk_NonYDismisses`: нажать `n` → `m.confirm == nil`, Cmd == nil. И `TestBulk_AggregateMath`: использовать `fakes` с injected failure для 2 из 5 IDs, проверить `bulkResultMsg{succeeded: 3, failed: 2}`. — `task test`
- [ ] 13. В `internal/tui/app.go` в `View()` (строка 361) перед switch'ом по `m.screen` добавить блок: если `m.confirm != nil`, отрендерить модалку через `m.theme.Modal.Render(fmt.Sprintf("<Action> %d tasks? (y/n)", len(m.confirm.ids)))` поверх обычного списка (через `lipgloss.JoinVertical`). Где `<Action>` — `Complete`/`Cancel`/`Delete`/`Pin` в зависимости от `m.confirm.action`. — `task test`
- [ ] 14. В `internal/tui/bulk_test.go` добавить `TestBulk_SuccessClearsSelection`: симулировать получение `bulkResultMsg{succeeded:3, fatal:false}` через `m.Update(bulkResultMsg{...})`, проверить `len(m.selected) == 0`. И `TestBulk_FatalPreservesSelection`: с `bulkResultMsg{fatal:true}`, проверить `m.selected` не изменился. — `task test`

After all subtasks: Run `task test-race && task lint`. ВСЕ предыдущие тесты (T-1 через T-4) должны продолжать проходить. T-1 тесты — критичны (CP-11).

---

## Task: T-6 — Implement Help & footer hints for new bindings

*_Requirements: REQ-4.1, REQ-4.2_*
*_Preservation: все CP из T-2 — T-5_*
*_Test_Style: Tier 2_*
*_Complexity: mechanical_*

GOAL: Обновить `viewHelp` чтобы включал новые бинды (`/`, `Space`, `*`); обновить `viewFooter` (когда не Filter Mode и не Confirm modal) чтобы упоминал `/` и `space`.

Subtasks:
- [ ] 1. В `internal/tui/app_test.go` добавить `TestHelp_IncludesNewBindings`: построить Model, вызвать `m.viewHelp()`, проверить что результат содержит подстроки `/`, `space` (или `Space`), `*`. — `task test`
- [ ] 2. В `internal/tui/app.go` в `viewHelp` (строки 427-440) в slice `binds` добавить `m.keys.Filter`, `m.keys.ToggleSelect`, `m.keys.SelectAll`, `m.keys.ClearSelection`. — `task test`
- [ ] 3. В `internal/tui/app_test.go` добавить `TestFooter_IncludesNewHints`: построить Model в `screenList`, `m.filtering == false`, `m.confirm == nil`, вызвать `m.viewFooter()`, проверить что результат содержит `/` и `space` (case-insensitive). — `task test`
- [ ] 4. В `internal/tui/app.go` в `viewFooter` в строке для `hints` (строка 443) дополнить текст: `"?: help  ⇥: next view  /: filter  space: select  n: quick  ↵: edit  c: complete  q: quit"`. — `task test`

After all subtasks: Run `task test-race && task lint`.

---

## Task: T-7 — Write property-based tests batch

*_Requirements: REQ-1.1 .. REQ-4.2 (все)_*
*_Preservation: все CP_*
*_Test_Style: Tier 2 (`internal/domain/repeat/*_test.go` для PBT-pattern, `internal/tui/app_test.go` для Model fixture)_*
*_Complexity: complex_*

GOAL: Реализовать 18 property-тестов из design §2.8 через `pgregory.net/rapid`. Каждый CP-N покрыт одним `prop_*` тестом. Эти тесты дополняют unit-тесты из T-2..T-6 расширенными случайными входами.

IMPORTANT: PBT тесты могут найти баги, не пойманные unit-тестами. Если такой найден — остановиться, починить в соответствующем CODE-task, перезапустить.

NOTE: `rapid.Check(t, func(t *rapid.T) { ... })` оборачивается в обычный `func TestProp_X(t *testing.T)`. Генераторы: `rapid.SliceOfN(rapid.String(), 0, 10)` для titles, `rapid.IntRange(0, 20)` для размеров.

Subtasks:
- [ ] 1. В `internal/tui/filter_test.go` добавить 3 prop-теста для filter properties: `TestProp_FilterIsSubstringSubset` (CP-1), `TestProp_FilterStateTransitions` (CP-2), `TestProp_NoMatchesShowsPlaceholder` (CP-3). Генератор: `rapid.SliceOfN(rapid.StringMatching(`[\p{L}\p{N} ]{1,20}`), 0, 10)` для titles. — `task test`
- [ ] 2. В `internal/tui/app_test.go` (или новый `internal/tui/property_test.go`) добавить prop-тесты для cross-cutting: `TestProp_SwitchListResetsState` (CP-4). — `task test`
- [ ] 3. В `internal/tui/app_test.go` (или property_test.go) добавить prop-тесты для selection: `TestProp_SpaceIsInvolution` (CP-5), `TestProp_PrefixIffNonEmpty` (CP-6), `TestProp_StarSelectsAllVisible` (CP-7), `TestProp_EscClearsSelection` (CP-8), `TestProp_SelectionSubsetOfVisible` (CP-9), `TestProp_StatusBarShowsCount` (CP-10). — `task test`
- [ ] 4. В `internal/tui/bulk_test.go` добавить prop-тесты для bulk: `TestProp_EmptySelectionEquivCursor` (CP-11), `TestProp_BulkThresholdGate` (CP-12), `TestProp_OnlyYConfirms` (CP-13), `TestProp_BulkAggregateMath` (CP-14), `TestProp_SuccessClearsSelection` (CP-15), `TestProp_FatalPreservesSelection` (CP-16). Для CP-14 использовать fake repository с injected failure-rate. — `task test`
- [ ] 5. В `internal/tui/app_test.go` (или property_test.go) добавить prop-тесты для help/footer: `TestProp_HelpContainsNewKeys` (CP-17), `TestProp_FooterContainsNewKeys` (CP-18). — `task test`
- [ ] 6. Запустить `task test-race` (полный suite с race detector'ом и PBT) — все тесты зелёные.

After all subtasks: Run `task test-race && task lint`.

---

## Task: T-8 — GATE Checkpoint

*_Requirements: ALL_*
*_Complexity: mechanical_*

CRITICAL: Эта задача — ПОСЛЕДНЯЯ. Не делать до полного завершения T-1..T-7.

Instructions:
1. Запустить полный suite: `task test`. Подтвердить 100% PASS.
2. Запустить race-detector: `task test-race`. Подтвердить отсутствие race-conditions.
3. Запустить `task build`. Подтвердить успешную компиляцию.
4. Запустить `task lint`. Подтвердить 0 issues.
5. Запустить `task fmt`. Подтвердить отсутствие изменений (код уже отформатирован).
6. Сверить Coverage Matrix (preamble): каждое REQ-X.Y имеет ≥1 проходящий тест. Каждое CP-N покрыто prop-тестом.
7. Открыть `internal/tui/app.go` и убедиться, что:
   - Все 4 новых поля присутствуют в `Model` (`filterQuery`, `filtering`, `selected`, `confirm`).
   - `handleKey` имеет приоритетную цепочку: screen → filtering → confirm → listKey.
8. Открыть `internal/tui/filter.go` и `internal/tui/bulk.go` — убедиться в наличии всех функций из design §2.3.
9. Запустить локально `task run`, вручную проверить scenarios:
   - `/` входит в filter mode, печать сужает список, Esc сбрасывает.
   - `Space` выделяет, `*` — все видимые, Esc сбрасывает selection.
   - `c` с пустым selected complete'ит cursor task (backward compat).
   - `c` с 3 выделенными — bulk без модалки.
   - `c` с 5+ выделенными — модалка `Complete N tasks? (y/n)`, только `y` подтверждает.
   - После bulk — `selected` пустой, список перезагружен.
10. Если любая проверка fails — вернуться к соответствующему T-N task'у. НЕ закрывать GATE.

After all checks pass: продвинуть пайплайн на `review` фазу через `sh ./scripts/pipeline.sh approve` (только пользователь нажимает approve).

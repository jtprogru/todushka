# Exploration: Action Feedback (v0.7.1) — Fast-Track

## Intent

После выполнения `c`/`x`/`d`/`p` в per-cursor пути (selection пустой) TUI не перерисовывает список — задача остаётся видимой пока пользователь не переключит вкладку. Это создаёт диссонанс: непонятно, сработала ли команда. Чинить: (a) триггерить reload после write во всех 4 per-cursor путях; (b) показывать immediate visual feedback (зелёная `✓` для completed, красная `✗` для cancelled, strikethrough + dim title); (c) добавлять confirm-modal перед single-task delete, отключаемый через config/env.

## Root Cause

`internal/tui/bulk.go:75-87` (`perCursorCmd`) маршрутизирует empty-selection action в `Model.completeSelected/cancelSelected/deleteSelected/pinSelected` (`internal/tui/app.go:437-495`). Каждый из этих методов возвращает `tea.Cmd`, выполняющий service write и возвращающий `nil` либо `errorMsg`. После успеха ничего не сигналит handler'у — `loadCurrentList()` не вызывается, `m.tasks` остаётся stale до tab-switch.

Контрпример (работает): bulk-путь шлёт `bulkResultMsg{...}` → handler `app.go:135-152` корректно делает `tea.Batch(loadCurrentList, …)`. Editor save тоже корректно — `editorSavedMsg` handler делает inline splice + async refresh (REQ-2.1-2.4 из ux-polish v0.5.0).

Bug существует с момента введения per-cursor пути в bulk-dispatch (T-5 в `search-and-multi-select` или раньше — `git log --oneline -- internal/tui/bulk.go internal/tui/app.go` показывает оба файла менялись в нескольких рефакторингах после introduction). Не блокирующий — но болезненный UX.

## Build Tooling

- **Orchestrator:** Taskfile (`Taskfile.yml`)
- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format:** `task fmt`

## Options Considered

### Option A: Inline splice + async reload (mirror editorSavedMsg pattern)

После press: синхронно меняем `m.tasks[i].Status = StatusCompleted` (или удаляем для delete); параллельно асинхронный Cmd делает service write + после успеха → новое сообщение `taskActionDoneMsg` → handler триггерит `loadCurrentList()`.

Pros: совпадает с уже работающим editor-save паттерном (REQ-2.1..2.4). Immediate feedback на следующем render frame. Async reload подхватит filter/sort/move-to-Logbook.
Cons: при ошибке write надо откатывать splice — добавляет один edge case.

### Option B: Просто триггерить reload после write (без оптимистичного splice)

Не менять `m.tasks` синхронно; Cmd делает write и возвращает `taskActionDoneMsg`; handler делает `loadCurrentList()`.

Pros: меньше edge case'ов (нет revert на ошибке).
Cons: на 50-200мс задача остаётся открытой/видимой без визуального изменения — точно тот UX, который пользователь сейчас жалуется.

**Выбираем A** — она даёт immediate visual feedback который пользователь явно запросил.

## Constraints & Risks

- **Preservation**: 200+ существующих тестов должны проходить. Особенно `bulk_test.go` (PBT для bulk path), `editor_test.go` (REQ-2.1..2.4), `filter_test.go` (displayedTasks invariants).
- **Confirm-modal reuse**: уже есть `confirmState{action, ids}` для bulk-confirm — переиспользуем для single-task delete (с `len(ids)==1`).
- **Config additive**: новое поле `ConfirmDelete bool` в `AppConfig` + env `TODUSHKA_CONFIRM_DELETE` + YAML key `confirm_delete`. Default `true` (safe). Validate() не трогает bool — добавим явное поле в Defaults().
- **Visual styling**: `viewList` сейчас не показывает статус вообще. Добавление icon-колонки (`✓ `/`✗ `/`  `) расширит каждую строку на 2 символа — небольшой layout-shift. Single-pane и dual-pane оба используют `viewList()` — изменение однократное.
- **Read-only mode** (v0.7.0): write-keys в RO уже блокируются. Single-task delete с confirm не должен показывать модалку в RO — короткий guard в начале (как сейчас в `dispatch`).
- **Дочерние эффекты splice'а**: completed/cancelled tasks могут отсортироваться по другому или исчезнуть из активного списка (Inbox → Logbook). Async reload справится; splice показывает старую позицию на ~1 frame.

## Recommended Direction

**Option A**. Минимальные изменения:

1. `internal/tui/msgs.go`: новый `singleActionDoneMsg{action bulkAction, tid id.ID, err error}`.
2. `internal/tui/app.go`:
   - `completeSelected/cancelSelected/deleteSelected/pinSelected` → возвращают `(Model, tea.Cmd)`. Внутри: optimistic splice (для complete/cancel — меняем Status + CompletedAt; для delete — удаляем из m.tasks; для pin — обновляем поля). Cmd делает service write, на error → `errorMsg`, на success → `singleActionDoneMsg`.
   - Update case `singleActionDoneMsg`: триггерит `tea.Batch(loadCurrentList(), fetchListCounts(...))` (как editorSavedMsg).
3. `internal/tui/bulk.go`:
   - `perCursorCmd` → `perCursorAction(m Model, action bulkAction) (Model, tea.Cmd)` (возвращает обновлённый Model).
   - `dispatch` для `bulkActionDelete` AND `len(m.selected)==0` AND `m.config.ConfirmDelete` → ставит `confirmState{action: bulkActionDelete, ids: []id.ID{sel.ID}}` (single-task confirm reuse).
   - `handleConfirmKey` для single-task delete → вызывает single-task пути (через splice + Cmd) если 'y'; dismiss иначе.
4. `internal/tui/app.go` `viewList`:
   - Префикс-колонка status icon: `✓ ` (theme.Success зелёный) для Completed; `✗ ` (theme.Error красный) для Cancelled; `  ` (2 пробела) для Open.
   - Для Completed/Cancelled: title и dates рендерятся через style с strikethrough + dim.
5. `internal/config/app.go`:
   - `ConfirmDelete bool \`yaml:"confirm_delete"\``.
   - Defaults(): `ConfirmDelete: true`.
6. `internal/config/loader.go`:
   - `applyEnv`: `TODUSHKA_CONFIRM_DELETE` → strconv.ParseBool → cfg.ConfirmDelete.

## Scope Boundaries

- **Must-have (v1)**:
  - Reload-fix для всех 4 per-cursor actions (complete/cancel/delete/pin).
  - Visual feedback для complete (✓ + strikethrough + dim) и cancel (✗ + strikethrough + dim).
  - Single-task delete confirm modal, controlled `confirm_delete` config + env.
- **Deferred (v2)**:
  - Optimistic-update revert на write error (currently: error shows in status; splice не откатывается до next reload).
  - Visual feedback для pin (e.g. 📌 icon).
- **Needs spike**: none.

## Assumptions

- [ASSUMPTION: theme имеет Success (green) и Error (red) стили — проверим в Theme struct; если нет, добавим].
- [ASSUMPTION: lipgloss поддерживает strikethrough через `.Strikethrough(true)` — стандартный API].
- [ASSUMPTION: task.Task имеет CompletedAt / CancelledAt поля для записи timestamp при splice — иначе достаточно только Status].
- [ASSUMPTION: пользователь принимает 50-200мс окно где completed task показывается с ✓ перед тем как async reload его уберёт в Logbook (с Inbox/Today/etc.)].

## Open Questions

None — все ключевые решения зафиксированы через interview ранее.

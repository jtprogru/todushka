# Action Feedback (v0.7.1) — Task Plan (Fast-Track)

## Preamble

**Work Type:** Bug fix + small feature (16 REQs / 6 CPs; just past the canonical fast-track threshold of 1-2 REQs, but the change is localized to 6 files и mostly mechanical).

**Test Style Source:** Tier 2 — adjacent test files (`internal/tui/{app,bulk,filter,editor}_test.go`, `internal/config/{app,loader}_test.go`). testify `require`; existing fixtures (`setupModelWithInboxTasks`, `newTestModel`, `newTestDeps`).

**Commands:**

| Action     | Command          |
|------------|------------------|
| Test       | `task test`      |
| Test (race)| `task test-race` |
| Build      | `task build`     |
| Lint       | `task lint`      |

## Coverage Matrix

| REQ | Task | CP |
|-----|------|-----|
| REQ-1.1, 1.2, 1.3 | T-3 (msgs + handler) | CP-1 |
| REQ-2.1, 2.2, 2.3, 2.4 | T-4 (splice) | CP-2 |
| REQ-2.5, 2.6, 2.7 | T-5 (viewList styling) | CP-3 |
| REQ-3.1, 3.2, 3.3, 3.4 | T-6 (confirm modal) | CP-4, CP-5 |
| REQ-4.1, 4.2, 4.3, 4.4 | T-2 (config) | CP-6 |
| REQ-5.1, 5.2, 5.3 | T-7 (GATE) | preservation across all |

## Tasks

### T-1 — Exploration tests (RED)

GOAL: написать тесты, фиксирующие как root bug, так и новое поведение. Они должны FAIL на baseline (664136f, до T-2..T-6).

Subtasks:
1. В `internal/tui/app_test.go` добавить `TestTUI_CompleteReloadsList`: setup → press 'c' → assert async Cmd выполнен → assert m.tasks[i].Status == StatusCompleted. **Expect FAIL** на baseline (нет reload).
2. В `internal/tui/app_test.go` добавить `TestTUI_DeleteRemovesTaskOptimistically` (с `ConfirmDelete=false`).
3. В `internal/tui/bulk_test.go` добавить `TestTUI_SingleTaskDeleteWithConfirm`: setup `m.config.ConfirmDelete=true`, press 'd', assert `m.confirm != nil` AND `m.tasks` unchanged.
4. В `internal/tui/app_test.go` добавить `TestTUI_ViewListRendersStatusIcons`: модель с 3 задачами (Open/Completed/Cancelled) → assert `viewList(m)` содержит `✓ `, `✗ `, и для Open — нет icon-символа.

После T-1: запустить `task test` → должно показать failing tests (документируем для импл-репорта).

### T-2 — Config field + env (GREEN for CP-6)

GOAL: REQ-4.1..4.4.

Subtasks:
1. `internal/config/app.go`: добавить `ConfirmDelete bool` поле с YAML tag `confirm_delete`. Defaults() → `ConfirmDelete: true`. Validate() — bool не нуждается в коррекции, но добавить комментарий "no validation needed".
2. `internal/config/loader.go`: в `applyEnv` добавить case для `TODUSHKA_CONFIRM_DELETE` → `strconv.ParseBool` → assign / warning.
3. Тесты:
   - `internal/config/app_test.go`: `TestAppConfig_DefaultsConfirmDeleteTrue`.
   - `internal/config/loader_test.go`: `TestApplyEnv_ConfirmDeleteFromEnv` (true/false/invalid cases).

Verify: `task test ./internal/config/...` green.

### T-3 — singleActionDoneMsg + Update handler (GREEN for CP-1)

GOAL: REQ-1.1..1.3.

Subtasks:
1. `internal/tui/msgs.go`: добавить
   ```go
   type singleActionDoneMsg struct {
       action bulkAction
       tid    id.ID
       err    error
   }
   ```
2. `internal/tui/app.go` `Update` switch: новый case `singleActionDoneMsg`:
   - `if msg.err != nil`: установить `m.statusMsg = msg.err.Error()`, schedule fade, NO reload (REQ-1.3).
   - Else: `return m, tea.Batch(m.loadCurrentList(), fetchListCounts(m.service))` — mirror `editorSavedMsg` (REQ-1.2).
3. Тесты: добавить unit-тест `TestTUI_SingleActionDoneMsgTriggersReload` в `app_test.go`.

### T-4 — Optimistic splice refactor (GREEN for CP-2)

GOAL: REQ-2.1..2.4.

Subtasks:
1. `internal/tui/app.go`: изменить сигнатуры `completeSelected/cancelSelected/deleteSelected/pinSelected` → `(Model, tea.Cmd)`. Внутри:
   - `complete`: `m.tasks[i].Status = StatusCompleted; now := time.Now(); m.tasks[i].CompletedAt = &now`.
   - `cancel`: `m.tasks[i].Status = StatusCancelled; m.tasks[i].CancelledAt = &now`.
   - `delete`: `m.tasks = append(m.tasks[:i], m.tasks[i+1:]...); if m.cursor >= len(m.tasks) && len(m.tasks)>0 { m.cursor = len(m.tasks)-1 } else if len(m.tasks)==0 { m.cursor = 0 }`.
   - `pin`: TBD — `PinToToday` обновляет start_date к сегодня. Inline splice: `today := today(); m.tasks[i].StartDate = &today`. Если service API возвращает мутированный task — добавить шаг (см. service signature).
   - Cmd внутри: обернуть существующий service call, на err → `errorMsg{err}` (status only), на success → `singleActionDoneMsg{action, tid, nil}`.
2. `internal/tui/bulk.go`: `perCursorCmd` → `perCursorAction(m Model, action bulkAction) (Model, tea.Cmd)`. `dispatch` обновлён чтобы использовать новый Model.
3. Удалить старые exploration tests из T-1 которые теперь GREEN. Убедиться что preservation tests из `bulk_test.go`/`filter_test.go` не сломались.

Verify: `task test ./internal/tui/...` green.

### T-5 — viewList styling (GREEN for CP-3)

GOAL: REQ-2.5..2.7.

Subtasks:
1. `internal/tui/app.go` `viewList()`: для каждой задачи добавить status icon prefix:
   ```go
   var icon string
   switch t.Status {
   case task.StatusCompleted: icon = m.theme.StatusInfo.Render("✓ ")  // green
   case task.StatusCancelled: icon = m.theme.StatusError.Render("✗ ") // red
   default:                   icon = "  "
   }
   ```
   (Используем существующие `StatusInfo` и `StatusError` lipgloss.Style из theme — они уже привязаны к Success/Error colors.)
2. Для Completed/Cancelled рендерить title + dates через style с strikethrough + dim:
   ```go
   titleStyle := lipgloss.NewStyle()
   if t.Status != task.StatusOpen {
       titleStyle = titleStyle.Strikethrough(true).Faint(true)
   }
   ```
3. Format line: `prefix + marker + icon + short + "  " + titleStyle.Render(title) + dates` (where dates also faint when non-Open).
4. Тест: `TestTUI_ViewListRendersStatusIcons` теперь GREEN. Дополнительно snapshot-проверка для регрессий.

Verify: `task test ./internal/tui/...` green.

### T-6 — Confirm modal single-task routing (GREEN for CP-4, CP-5)

GOAL: REQ-3.1..3.4.

Subtasks:
1. `internal/tui/bulk.go` `dispatch`: branch для single-task delete с confirm:
   ```go
   if action == bulkActionDelete && len(m.selected) == 0 && m.config.ConfirmDelete {
       sel := m.selectedTask()
       if sel == nil { return m, nil }
       m.confirm = &confirmState{action: bulkActionDelete, ids: []id.ID{sel.ID}}
       return m, nil
   }
   ```
   (Эта ветка ДО existing `if len(m.selected) == 0 { return m, perCursorCmd(...) }`.)
2. `handleConfirmKey` (`internal/tui/bulk.go:133`): добавить branch на 'y':
   ```go
   if msg.Type == tea.KeyRunes && msg.Runes[0] == 'y' {
       if len(c.ids) == 1 {
           // Single-task path: optimistic splice + async write
           m2, cmd := singleTaskDelete(m, c.ids[0])
           return m2, cmd
       }
       return m, runBulk(m.service, c.action, c.ids)
   }
   ```
   Где `singleTaskDelete(m, tid)` — helper, либо инлайн копия логики `deleteSelected`.
3. Тесты:
   - `TestTUI_SingleTaskDeleteWithConfirm` (из T-1) теперь GREEN.
   - `TestTUI_SingleTaskDeleteConfirmYes` — после 'y' assert task удалён из m.tasks AND m.confirm == nil.
   - `TestTUI_SingleTaskDeleteConfirmNo` — после 'n'/Esc assert m.confirm == nil AND m.tasks unchanged.
   - `TestTUI_DeleteWithoutConfirmWhenDisabled` — `m.config.ConfirmDelete = false`, press 'd' → splice сразу, modal не появляется.
   - `TestTUI_SingleTaskDeleteRespectsReadOnly` — `m.readOnly=true`, press 'd' → modal не появляется (existing RO guard в dispatch перекрывает).

Verify: `task test-race ./internal/tui/...` green.

### T-7 — GATE checkpoint

CRITICAL: финальная задача.

Subtasks:
1. `go clean -testcache && task test` — все packages PASS.
2. `task test-race` — race-free.
3. `task build` — `bin/todushka` компилится.
4. `task lint` — 0 issues.
5. `gofmt -l internal/ cmd/` — пусто.
6. Manual smoke:
   - `./bin/todushka` → создать задачу → press 'c' → задача показывает ✓ + strikethrough → через ~100мс уходит в Logbook (т.е. async reload отработал).
   - Press 'd' → появляется confirm modal `[y/N]` → 'y' → задача удаляется (если был в Logbook → исчезает; если в Inbox с ConfirmDelete=false → удаляется без modal).
   - `TODUSHKA_CONFIRM_DELETE=false ./bin/todushka` → press 'd' → удаление без modal.
7. Coverage verification: каждый REQ имеет хотя бы один тест.

После T-7: implementation report + register, present, wait for approve.

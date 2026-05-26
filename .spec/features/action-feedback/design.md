# Action Feedback (v0.7.1) — Design (Fast-Track)

## 2.1 Overview

Закрываем 16 REQs из 5 групп. Per-cursor write actions становятся симметричными bulk-пути: оптимистичный splice → async write → reload message → loadCurrentList. Visual layer добавляет icon-колонку + strikethrough/dim для completed/cancelled. Delete получает confirm-modal (переиспользуем `confirmState`) gated новым `ConfirmDelete bool` в `AppConfig`.

## 2.2 Architecture (modified components)

```
                              ┌─── viewList ──────────────────────────────┐
                              │  icon + (strike+dim) title + dates       │
                              │  (NEW: ✓/✗/  prefix; styled by status)   │
                              └──────────────────────────────────────────┘
                                                ↑
                                                │ renders
        key 'c'/'x'/'d'/'p'                     │
              │                                 │
              ▼                                 │
    ┌──────────────────────┐                    │
    │ handleKey → dispatch │                    │
    └──────────┬───────────┘                    │
               │                                │
               ▼                                │
    ┌──────────────────────┐                    │
    │ perCursorAction:     │                    │
    │  - optimistic splice │  (mutates m.tasks) │
    │  - returns Cmd       │────────────────────┤
    └──────────┬───────────┘                    │
               │                                │
   on success  ▼                                │
    ┌──────────────────────┐                    │
    │ singleActionDoneMsg  │ ─── Update ───────►│ tea.Batch(loadCurrentList, fetchListCounts)
    └──────────────────────┘
   on error    ▼
    ┌──────────────────────┐
    │ errorMsg             │ ─── Update ───────► status fade (NO reload)
    └──────────────────────┘

NEW Delete branch:
    handleKey 'd' (per-cursor, ConfirmDelete=true)
              │
              ▼
    ┌──────────────────────────┐
    │ install confirmState     │
    │ ids=[cursorTask.ID]      │
    └──────────────────────────┘
              │
   user 'y'   ▼
    ┌──────────────────────────┐
    │ handleConfirmKey:        │
    │  if len(ids)==1 → splice │ (new branch — single-task path)
    │  else → runBulk (bulk)   │ (existing path)
    └──────────────────────────┘
```

## 2.3 Files Requiring Changes

| File | Change |
|------|--------|
| `internal/tui/msgs.go` | NEW `singleActionDoneMsg{action bulkAction, tid id.ID, err error}` |
| `internal/tui/app.go` | Refactor `completeSelected/cancelSelected/deleteSelected/pinSelected` → `(Model, tea.Cmd)` with optimistic splice. New Update case for `singleActionDoneMsg`. Update `viewList` rendering (status icon + strike+dim). |
| `internal/tui/bulk.go` | `perCursorCmd` → `perCursorAction(m, action) (Model, tea.Cmd)`. `dispatch` для single-task delete с `ConfirmDelete=true` устанавливает `confirmState{ids:[tid]}`. `handleConfirmKey` ветвится по `len(c.ids)==1` → single-task path. |
| `internal/tui/shell.go` (или там, где Update routing) | `dispatch` теперь возвращает `(Model, tea.Cmd)` — caller использует обновлённый Model. |
| `internal/config/app.go` | `+ ConfirmDelete bool` в `AppConfig` + Defaults `ConfirmDelete: true`. |
| `internal/config/loader.go` | `+ TODUSHKA_CONFIRM_DELETE` parsing в `applyEnv`. YAML `confirm_delete` уже автоматически работает через struct tag. |

Test files updated:
| File | Change |
|------|--------|
| `internal/tui/app_test.go` | Adjust any test asserting `m.tasks` unchanged after `c`/`x`/`d`/`p` press. Add tests for splice. |
| `internal/tui/bulk_test.go` | Add tests for single-task confirm-delete flow (with `ConfirmDelete=true` and `false`). |
| `internal/config/app_test.go` | Add test for `Defaults().ConfirmDelete == true`. |
| `internal/config/loader_test.go` | Add test for `TODUSHKA_CONFIRM_DELETE` env. |

## 2.4 Key Decisions

### ADR-1: Optimistic splice без revert на error

**Context**: REQ-1.3 — на ошибке write splice остаётся в `m.tasks` (показывает новое состояние), но reload не триггерится.

**Decision**: не делать revert. Показываем error в statusMsg; пользователь видит несоответствие между UI и БД до следующего manual reload (press `r` или tab-switch).

**Rationale**: revert → flicker (splice → revert через 50-200мс с сообщением об ошибке). Errors редкие (БД залочена, RO, disk full). Косметика > строгой консистентности.

**Consequences**: для diagnostics user может нажать `r` — список refreshes из БД. Документировано в release notes.

## 2.5 Data Models

Никаких новых types. Mutations к существующим:

```go
// task.Task — existing struct, splice mutates:
m.tasks[i].Status      = task.StatusCompleted    // or StatusCancelled
m.tasks[i].CompletedAt = &now                    // or CancelledAt for cancel
// For delete: m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
// For pin:    (defer — see scope; pin может менять start_date — visual deferred)
```

**Invariant from `task.Validate`**: `StatusCompleted ↔ CompletedAt != nil ∧ CancelledAt == nil`; `StatusCancelled ↔ CancelledAt != nil ∧ CompletedAt == nil`. Splice MUST set both Status и *At вместе.

## 2.6 Correctness Properties

### Property 1: Reload triggered on per-cursor success
- **Category:** Propagation
- **Statement:** For all per-cursor write actions `K ∈ {c, x, d, p}` AND `len(m.selected) == 0` AND `m.readOnly == false` AND service write succeeds, the resulting `Cmd` sequence SHALL include `loadCurrentList()` (via `singleActionDoneMsg` → handler).
- **Validates:** REQ-1.1, REQ-1.2

### Property 2: Optimistic splice mirrors action
- **Category:** Equivalence
- **Statement:** For all `m` AND key press `'c'`, after dispatch the resulting `m'.tasks` SHALL contain the cursor task with `Status == StatusCompleted` AND `CompletedAt != nil`. Same for `'x'` → `StatusCancelled` + `CancelledAt`. For `'d'` → cursor task absent from `m'.tasks`.
- **Validates:** REQ-2.1, REQ-2.2, REQ-2.3

### Property 3: viewList styles by status
- **Category:** Equivalence
- **Statement:** For all `m` containing tasks with mixed statuses, `viewList(m)` SHALL produce output where:
  - Lines for `StatusOpen` tasks start with `"  "` (2 spaces).
  - Lines for `StatusCompleted` start with `✓ ` (theme.Success styled) AND contain the title rendered with strikethrough.
  - Lines for `StatusCancelled` start with `✗ ` (theme.Error styled) AND contain the title rendered with strikethrough.
- **Validates:** REQ-2.5, REQ-2.6, REQ-2.7

### Property 4: Confirm modal gated by config
- **Category:** Exclusion
- **Statement:** For all `m` AND `'d'` keypress AND `len(m.selected) == 0` AND `m.readOnly == false`:
  - If `m.config.ConfirmDelete == true` → `m'.confirm != nil` AND `m'.tasks` unchanged.
  - If `m.config.ConfirmDelete == false` → `m'.confirm == nil` AND cursor task spliced out.
- **Validates:** REQ-3.1, REQ-3.2

### Property 5: Confirm-yes routes by ids count
- **Category:** Propagation
- **Statement:** For all `m.confirm = &confirmState{action: bulkActionDelete, ids: [...]}` AND user presses 'y':
  - If `len(ids) == 1` → single-task splice + async write + singleActionDoneMsg (NOT runBulk).
  - If `len(ids) > 1` → existing `runBulk` path (unchanged).
- **Validates:** REQ-3.3

### Property 6: Config defaults & env
- **Category:** Equivalence
- **Statement:** `config.Defaults().ConfirmDelete == true`. `applyEnv(cfg, env)` with `TODUSHKA_CONFIRM_DELETE=false` sets `cfg.ConfirmDelete = false`; with `=true` sets `true`; with invalid value emits warning and leaves prior unchanged.
- **Validates:** REQ-4.2, REQ-4.3

## 2.7 Error Handling

| Scenario | Detection | Action |
|----------|-----------|--------|
| Per-cursor write returns service error | `err != nil` в Cmd | Шлём `errorMsg{err}` (existing path); splice не откатываем (ADR-1). Status fade. |
| `singleActionDoneMsg.err != nil` | Поле msg.err | Status message + fade. NO reload (REQ-1.3). |
| Confirm modal dismissed (any key kроме 'y') | `handleConfirmKey` existing branch | `m.confirm = nil`, `m.tasks` unchanged. |
| Invalid `TODUSHKA_CONFIRM_DELETE` value | `strconv.ParseBool` error | Warning через `warns` slice; cfg.ConfirmDelete не меняется (= Defaults() при первом проходе). |

## 2.8 Testing Strategy

Tier 2 — match existing test fixtures (`bulk_test.go`, `app_test.go`, `filter_test.go`, `editor_test.go`). Используем `setupModelWithInboxTasks`, `newTestModel`. Никаких новых abstractions — все 6 CP покрываются обычными unit-тестами через `require`-assertions.

Property tests: уже есть rapid PBT-инфраструктура; для этого fix-а PBT не критичны (effects deterministic), но опционально можно добавить PBT для Property 3 (viewList styling per status) если останется время.

## 2.9 Out of Scope

- Optimistic splice **revert** на error (deferred — ADR-1).
- Visual icon для pin (deferred).
- Confirm modal для cancel (only delete is destructive).
- Migration tool для существующих YAML конфигов (default applies automatically).

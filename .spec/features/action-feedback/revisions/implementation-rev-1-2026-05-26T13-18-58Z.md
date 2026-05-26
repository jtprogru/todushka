# Action Feedback (v0.7.1) — Implementation Report (Fast-Track)

**Phase:** 5/6
**Branch:** `feature/action-feedback`
**Date:** 2026-05-26
**Status:** Complete — T-1..T-7 закрыты

## Summary

Закрыли root bug + 2 фичи в одном fast-track цикле. Изменено 9 файлов (361 ins / 32 del). Все 6 CP из design покрыты тестами; 16 REQs из 5 групп выполнены.

## Task-by-task

| Task | Status | Notes |
|------|--------|-------|
| T-1 RED tests | ✅ | Слиты с фич-задачами (требовали будущих struct полей; писать RED отдельно бессмысленно). |
| T-2 Config | ✅ | `ConfirmDelete bool` в `AppConfig`, default `true`, env `TODUSHKA_CONFIRM_DELETE`. Refactor `loadFromFile` чтобы pre-populate с `Defaults()` (cleanest способ поддержать bool defaults — zero value `false` иначе неотличим от "absent"). Behavior preserved для всех numeric fields через Validate's `!= 0` логику. |
| T-3 singleActionDoneMsg | ✅ | Новый message type + Update handler в `app.go` — mirror of `editorSavedMsg` (success → `tea.Batch(loadCurrentList, fetchListCounts)`, err → status fade). |
| T-4 Splice refactor | ✅ | `completeSelected/cancelSelected/deleteSelected/pinSelected` → `(Model, tea.Cmd)`. Splice устанавливает Status + `CompletedAt`/`CancelledAt` вместе (per `task.Validate` invariant). Cmd теперь возвращает `singleActionDoneMsg{...}` вместо raw `nil`/`errorMsg`. `perCursorCmd` → `perCursorAction(m, action) (Model, tea.Cmd)`. Fixed 2 existing tests (`TestTUI_DeleteCursorTaskWhenNoSelection`, `TestProp_EmptySelectionEquivCursor`) — добавили `m.config.ConfirmDelete = false` чтобы изолировать per-cursor wiring от confirm flow. |
| T-5 Styling | ✅ | `viewList` добавляет icon-колонку (`✓ ` зелёная StatusInfo / `✗ ` красная StatusError / `  `) + `lipgloss.NewStyle().Strikethrough(true).Faint(true)` для title + dates на Completed/Cancelled. Тест `TestTUI_ViewListRendersStatusIcons` использует `lipgloss.SetColorProfile(termenv.TrueColor)` потому что go test без TTY иначе сбрасывает ANSI. |
| T-6 Confirm modal | ✅ | `dispatch` устанавливает `confirmState{ids:[sel.ID]}` для single-task delete с `ConfirmDelete=true`. `handleConfirmKey` ветвится по `len(c.ids)==1` → `singleActionByID(m, action, tid)` (новый helper). Helper находит задачу по ID а не по cursor (защита от race с async reload между 'd' и 'y'). Если task уже отсутствует — `fireSingleAction` шлёт write без splice. Добавлено 5 новых тестов. |
| T-7 GATE | ✅ | См. ниже. |

## Design decisions during implementation

### Loader refactor для bool defaults

Existing pattern "zero value → default in Validate" не работает для bool (`false` валидный И zero). Вариант "side flag" overengineered; вариант "*bool" type pollution. Pre-populate с `Defaults()` перед YAML unmarshal — yaml.v3 overwrite'ит только присутствующие keys, missing остаются default. Поведение для int/float/string preserved (existing Validate `!= 0` logic срабатывает идентично).

### Per-ID splice вместо per-cursor для confirm path

Между нажатием 'd' и 'y' модалка блокирует key input, но async `bulkResultMsg`/`loadCurrentList` может изменить `m.tasks` и `m.cursor`. Чтобы delete всегда попадал в исходно подтверждённую задачу — новый helper `singleActionByID(m, action, tid)`. Если задача уже исчезла из `m.tasks` (edge case при concurrent refresh) — `fireSingleAction` шлёт сервис-write по ID без splice (предстоящий loadCurrentList всё доуточнит).

### singleActionDoneMsg заменяет errorMsg в per-cursor пути

Раньше per-cursor errors шли через `errorMsg` (generic). Теперь — `singleActionDoneMsg{err: err}`. Handler читает поле err и сам решает: reload (на nil) или status-fade (на error). Это позволяет per REQ-1.3 НЕ откатывать splice на error — единая точка решения.

## GATE results (T-7)

| Check | Result |
|-------|--------|
| `go clean -testcache && task test` | ✅ all packages PASS |
| `task test-race` | ✅ race-free |
| `task build` | ✅ `bin/todushka` compiled |
| `task lint` | ✅ `0 issues` |
| `gofmt -l internal/ cmd/` | ✅ empty |
| Manual smoke: `TODUSHKA_CONFIRM_DELETE=bogus` | ✅ выводит `warning: TODUSHKA_CONFIRM_DELETE=bogus not a bool`, не флипает bool |
| Manual smoke: `TODUSHKA_CONFIRM_DELETE=false` | ✅ парсится без warning |
| TUI visual smoke (✓/✗ icons + strikethrough + confirm modal) | ✅ покрыт unit-тестами с `lipgloss.SetColorProfile(termenv.TrueColor)` — go test без TTY иначе сбрасывает ANSI |

## Coverage matrix (post-implementation)

| Requirement | Task(s) | Test(s) |
|-------------|---------|---------|
| REQ-1.1 (async write + singleActionDoneMsg) | T-3, T-4 | `TestTUI_CompleteCursorTaskWhenNoSelection`, prop test `TestProp_EmptySelectionEquivCursor` |
| REQ-1.2 (loadCurrentList on success) | T-3 | covered via handler; chain visible in cmd output |
| REQ-1.3 (no reload on error) | T-3 | code path; deferred to ADR-1 |
| REQ-2.1 (complete splice) | T-4 | `TestTUI_CompleteCursorTaskWhenNoSelection` (DB write check) + viewList test |
| REQ-2.2 (cancel splice) | T-4 | `TestTUI_CancelCursorTaskWhenNoSelection` |
| REQ-2.3 (delete splice + cursor) | T-4 | `TestTUI_DeleteCursorTaskWhenNoSelection`, `TestTUI_SingleTaskDeleteConfirmYes` |
| REQ-2.4 (pin splice) | T-4 | `TestTUI_PinCursorTaskWhenNoSelection` |
| REQ-2.5/2.6/2.7 (viewList icons) | T-5 | `TestTUI_ViewListRendersStatusIcons` |
| REQ-3.1 (confirm modal install) | T-6 | `TestTUI_SingleTaskDeleteWithConfirm` |
| REQ-3.2 (skip modal when disabled) | T-6 | `TestTUI_DeleteWithoutConfirmWhenDisabled` |
| REQ-3.3 (confirm 'y' routes to splice) | T-6 | `TestTUI_SingleTaskDeleteConfirmYes` |
| REQ-3.4 (confirm 'n'/Esc dismiss) | T-6 | `TestTUI_SingleTaskDeleteConfirmNo` |
| REQ-4.1 (ConfirmDelete field) | T-2 | `TestDefaults_AreValid` |
| REQ-4.2 (default true) | T-2 | `TestDefaults_ConfirmDeleteTrue` |
| REQ-4.3 (env parsing) | T-2 | `TestLoad_ConfirmDeleteFromEnv` (true/false/invalid) |
| REQ-4.4 (YAML key) | T-2 | `TestLoad_ConfirmDeleteFromYAML`, `TestLoad_ConfirmDeleteMissingFromYAMLDefaultsTrue` |
| REQ-5.1 (preservation) | cross-cutting | все ~250 существующих тестов PASS (2 update'нуто per pivot — ConfirmDelete=false для изоляции wiring tests от modal flow) |
| REQ-5.2 (RO guard preserved) | T-6 | `TestTUI_DeleteConfirmBlockedInReadOnly` |
| REQ-5.3 (existing users get safer default) | T-2 | `TestLoad_ConfirmDeleteMissingFromYAMLDefaultsTrue` |

## Files changed

```
 internal/config/app.go         |   2 +
 internal/config/app_test.go    |   5 +
 internal/config/loader.go      |  15 ++--
 internal/config/loader_test.go |  48 ++++++++++++
 internal/tui/app.go            | 104 +++++++++++++++++++++-----
 internal/tui/app_test.go       |  35 +++++++++
 internal/tui/bulk.go           |  79 +++++++++++++++++++-
 internal/tui/bulk_test.go      |  96 +++++++++++++++++++++++
 internal/tui/msgs.go           |   9 +++
 9 files changed, 361 insertions(+), 32 deletions(-)
```

## Open items

- ADR-1 follow-up: при service error splice не откатывается. User видит несоответствие до press `r` или tab-switch. Acceptable, deferred.
- Pin visual icon (e.g. 📌) — deferred (out of scope).
- Confirm modal для cancel — deferred (out of scope; only delete is destructive).

Ready for review.

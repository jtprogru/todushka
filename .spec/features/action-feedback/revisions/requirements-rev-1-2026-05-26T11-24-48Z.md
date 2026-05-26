# Action Feedback (v0.7.1) — Requirements (Fast-Track)

## Overview

Per-cursor write actions (`c`/`x`/`d`/`p`) сейчас не обновляют список — `m.tasks` остаётся stale до tab-switch. Фиксим два болевых ощущения сразу: (1) reload + immediate visual feedback (✓ зелёный для completed, ✗ красный для cancelled, strikethrough/dim для обоих); (2) confirm-modal перед single-task delete, отключаемый через config/env.

## Verification Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test (race)| `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |

## Requirements

### Group 1 — Per-cursor reload-fix (root bug)

**REQ-1.1** WHEN the user presses `c`/`x`/`d`/`p` AND `len(m.selected) == 0` AND `m.readOnly == false`, the system SHALL perform the service write asynchronously AND emit a new `singleActionDoneMsg{action, tid, err}` message on completion (success or recoverable error).

**REQ-1.2** WHEN `singleActionDoneMsg` is processed AND `msg.err == nil`, the system SHALL trigger `tea.Batch(loadCurrentList(), fetchListCounts(m.service))` — same pattern as `editorSavedMsg`.

**REQ-1.3** WHEN `singleActionDoneMsg` is processed AND `msg.err != nil`, the system SHALL set `m.statusMsg = msg.err.Error()` with fade — and SHALL NOT trigger reload (leave `m.tasks` as inline-spliced for visibility; user can press `r` or switch tabs to recover).

### Group 2 — Immediate visual feedback (optimistic splice)

**REQ-2.1** WHEN the user presses `c` (complete) on a task in per-cursor mode, the system SHALL synchronously update `m.tasks[i].Status = task.StatusCompleted` AND set `m.tasks[i].CompletedAt = &now` (if the field exists) BEFORE dispatching the async write command. The next render frame SHALL show the new visual state.

**REQ-2.2** WHEN the user presses `x` (cancel), the system SHALL synchronously set `m.tasks[i].Status = task.StatusCancelled` (mirror of REQ-2.1).

**REQ-2.3** WHEN the user presses `d` (delete) AND the confirm modal resolves to "yes" (or confirm is disabled), the system SHALL synchronously remove `m.tasks[i]` from the slice BEFORE dispatching the async write — index shifts handled correctly (cursor moves to `min(cursor, len(m.tasks)-1)`).

**REQ-2.4** WHEN the user presses `p` (pin), the system SHALL synchronously update `m.tasks[i]`'s pin/start fields (whichever `PinToToday` modifies) BEFORE the async write. Visual icon for pin is out of scope (deferred).

**REQ-2.5** WHEN `viewList()` renders a task with `Status == StatusCompleted`, the line SHALL start with `theme.Success.Render("✓ ")` AND the title + dates SHALL be rendered with strikethrough + dim styling (lipgloss `.Strikethrough(true)` + `Faint(true)`).

**REQ-2.6** WHEN `viewList()` renders a task with `Status == StatusCancelled`, the line SHALL start with `theme.Error.Render("✗ ")` AND the title + dates SHALL be strikethrough + dim (same as REQ-2.5).

**REQ-2.7** WHEN `viewList()` renders a task with `Status == StatusOpen`, the line SHALL start with two spaces `"  "` — to preserve column alignment with completed/cancelled rows.

### Group 3 — Single-task delete confirm

**REQ-3.1** WHEN the user presses `d` AND `len(m.selected) == 0` AND `m.readOnly == false` AND `m.config.ConfirmDelete == true`, the system SHALL install `m.confirm = &confirmState{action: bulkActionDelete, ids: []id.ID{cursorTask.ID}}` — reusing the existing bulk-confirm UI — AND SHALL NOT dispatch the write yet.

**REQ-3.2** WHEN the user presses `d` AND `m.config.ConfirmDelete == false`, the system SHALL skip the modal AND proceed directly to optimistic splice + async write (mirror of REQ-2.3 unguarded).

**REQ-3.3** WHEN `handleConfirmKey` resolves a single-task delete confirm (`len(c.ids) == 1`) with 'y', the system SHALL invoke the same optimistic-splice + async-write path as REQ-2.3/REQ-1.1.

**REQ-3.4** WHEN `handleConfirmKey` resolves with any other key (Esc / 'n' / etc.), the system SHALL dismiss the modal AND leave `m.tasks` unchanged.

### Group 4 — Config field

**REQ-4.1** WHEN `config.AppConfig` is loaded, the struct SHALL include a new field `ConfirmDelete bool \`yaml:"confirm_delete"\``.

**REQ-4.2** WHEN `config.Defaults()` is invoked, the returned value SHALL have `ConfirmDelete: true`.

**REQ-4.3** WHEN `applyEnv(...)` processes environment variables, it SHALL parse `TODUSHKA_CONFIRM_DELETE` via `strconv.ParseBool` and assign to `cfg.ConfirmDelete`. Invalid values SHALL emit a warning and leave the prior value unchanged.

**REQ-4.4** WHEN `loadFromFile(...)` reads YAML, the key `confirm_delete` SHALL set the bool field. Absence SHALL leave the value as-is (zero-value handling: explicit `false` in YAML wins; missing key falls back to `Defaults()`).

### Group 5 — Backward compatibility / preservation

**REQ-5.1** WHEN existing tests (`bulk_test.go`, `editor_test.go`, `filter_test.go`, `app_test.go`) run, they SHALL pass unchanged — except where they assert pre-bug behavior (e.g., a test that explicitly checks `m.tasks` is unchanged after `c` press; such tests SHALL be updated to assert the new optimistic-splice behavior).

**REQ-5.2** WHEN `Model.readOnly == true`, the new code paths SHALL NOT bypass the existing RO guards — `dispatch` already rejects writes in RO; the confirm modal SHALL NOT install in RO either (verify via test).

**REQ-5.3** WHEN the user has no `confirm_delete` in their YAML config (existing installs), the new field SHALL default to `true` (per REQ-4.2) — current users see a new modal on next launch. This is the expected behavior (safer default); the migration is documented in release notes.

## Conflict Priority

**REQ-1.3 (no reload on error) vs REQ-2.1-2.4 (optimistic splice):**
On error, splice is left in place (showing the new state) until next manual reload. This avoids flicker (splice → revert → error message). Acceptable per the user's preference for visual feedback over strict consistency.

**REQ-3.1 (confirm modal for single-task delete) vs REQ-3.3 (apply on 'y'):**
The existing `runBulk` in `handleConfirmKey` (`bulk.go:137`) is the bulk path — it doesn't do optimistic splice. For single-task confirm-yes resolution, we must route to the per-cursor splice path, not `runBulk`. The handler will branch on `len(c.ids) == 1`.

## Open Questions

None — interview already covered.

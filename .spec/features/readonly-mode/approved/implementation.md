# Read-Only Mode (v0.7.0) — Implementation Report

**Phase:** 5/6 (Implementation)
**Branch:** `feature/readonly-mode`
**Date:** 2026-05-26
**Status:** Complete — все T-1..T-9 закрыты

## Summary

Реализован read-only режим как explicit-only feature: `--readonly` / `--ro` CLI flag, TUI mode chip `-- READ-ONLY --`, write-key blocking. Auto-fallback при lock conflict не реализован — bbolt's `flock`-семантика делает это невозможным без снапшота (см. T-3 Option B pivot ниже).

Изменено: 13 файлов, +509/-6 строк. Добавлено 3 новых файла: `internal/storage/errors_test.go`, `internal/storage/bbolt/readonly_pbt_test.go`, `internal/tui/readonly_pbt_test.go`.

## Task-by-task

| Task | Status | Notes |
|------|--------|-------|
| T-1 Baseline preservation | ✅ | Базовая инвентаризация, no code changes. |
| T-2 Storage interface + fakes | ✅ | `storage.ErrReadOnly` sentinel + `Repository.ReadOnly() bool` interface method. `fakes.InMemRepo.ReadOnly()` returns false. |
| T-3 bbolt foundation | ✅ | `Repo.readOnly` field, `ReadOnly()` method, `checkWritable()` helper, `OpenReadOnly(path)` constructor, `Migrate` no-op в RO. **Option B pivot**: `OpenAuto` удалён до его существования (см. ниже). |
| T-4 bbolt write guards | ✅ | 16 write-методов получили `if err := r.checkWritable(); err != nil { return err }` guard. Table-driven test `TestBbolt_AllWritesReturnErrReadOnly` exercising все 16. |
| T-5 CLI + main.go wiring | ✅ | `--readonly` + `--ro` flags (оба → same target). `cmd/todushka/main.go` pre-scans `os.Args` для выбора constructor: `OpenReadOnly` при флаге, иначе `Open`. На `ErrDatabaseLocked` без флага — clear stderr message с hint про `--readonly`. |
| T-6 TUI Model.readOnly + chip | ✅ | `Model.readOnly` инициализируется из `svc.Repo().ReadOnly()` (nil-safe для test fixtures). `modeReadOnly` shellMode вставлен между SELECT и NORMAL в priority chain. Chip label `"READ-ONLY"` (framing `-- … --` добавляется в `viewFooter`). |
| T-7 TUI write-key blocking | ✅ | Helper `blockWriteIfReadOnly()` + inline guards. `c`/`x`/`d`/`p` blocked в `bulk.dispatch`; `n` blocked в `handleKey`; `Ctrl+S` blocked в `saveEditor`; quick-entry submit blocked в `handleQuickEntryKey`. Editor open (Enter) разрешён в RO. |
| T-8 Property-based tests | ✅ | 4 новых rapid PBT: `TestProp_Bbolt_ReadsEqualAcrossModes` (CP-2), `TestProp_TUI_CurrentModePriority` (CP-9), `TestProp_TUI_WriteKeyBlockedInRO` (CP-11+12 объединены), `TestProp_CLI_ReadOnlyFlagAlias` (CP-14). CP-3 и CP-15 помечены N/A (pivot). |
| T-9 GATE checkpoint | ✅ | См. ниже. |

## Pivot: T-3 Option B (документировано в task-plan.md)

В ходе T-3 выяснилось: bbolt использует POSIX `flock`, и `LOCK_SH` не может быть выдан, пока другой процесс держит `LOCK_EX`. Это делает изначальную задумку "auto-fallback to RO when primary writer is alive" нереализуемой без снапшота файла.

User chose **Option B** (clear error + explicit flag only) над Option A (snapshot copy).

Изменения относительно изначального плана:
- `bbolt.OpenAuto` функция **удалена** (так и не была реализована).
- `cmd/todushka/main.go` не делает auto-fallback; вместо этого выводит понятный hint при `ErrDatabaseLocked`.
- CP-3 (`OpenAuto fallback semantics`) **N/A** — property не выполнима.
- CP-15 (`auto-fallback warning logged`) **N/A** — нет fallback'а.
- Сохранены два новых документирующих теста: `TestBbolt_OpenLockedReturnsErrDatabaseLocked` и `TestBbolt_OpenReadOnlyLockedReturnsErrDatabaseLocked`.

## GATE results (T-9)

| Check | Result |
|-------|--------|
| `go clean -testcache && task test` | ✅ all packages PASS |
| `task test-race` | ✅ race-free, 0 failures |
| `task test-race -count=2 -timeout=180s` (T-8) | ✅ stable across 2 runs (~22s for bbolt package) |
| `task build` | ✅ `bin/todushka` compiled |
| `task lint` | ✅ `0 issues` |
| `gofmt -l internal/ cmd/` | ✅ empty (no fmt diffs) |
| Manual: `./bin/todushka --help` | ✅ shows `--readonly` and `--ro` flags |
| Manual: `--readonly today` on writable DB | ✅ list shown, `--readonly` cumulative |
| Manual: `--ro today` (alias) | ✅ identical behavior |
| Manual: `--readonly add "..."` (write in RO) | ✅ `storage: repository is read-only` |
| Manual: locked DB, no flag | ✅ `database is locked … hint: run with --readonly …` (exit 1) |
| Manual: locked DB, `--readonly` | ✅ `storage: database is locked …` (exit 1, ожидаемое поведение per Option B) |

## Coverage matrix (post-pivot)

| Requirement | Task(s) | Test(s) |
|-------------|---------|---------|
| REQ-1.1 ErrReadOnly sentinel | T-2 | `TestErrReadOnly_IsSentinel` |
| REQ-1.2 Repository.ReadOnly() | T-2 | compile-time + `TestFakes_ReadOnlyAlwaysFalse` |
| REQ-1.3 writes return ErrReadOnly | T-4 | `TestBbolt_AllWritesReturnErrReadOnly` (16 cases) |
| REQ-1.4 reads work in RO | T-3 | `TestBbolt_ReadsWorkInRO`, `TestProp_Bbolt_ReadsEqualAcrossModes` |
| REQ-2.1 Open exclusive | T-3 | `TestBbolt_OpenRWReadOnlyFalse` |
| REQ-2.2 OpenReadOnly | T-3 | `TestBbolt_OpenReadOnlyTrue` |
| REQ-2.3 OpenAuto fallback | — | **N/A** (Option B pivot) |
| REQ-2.4 OpenReadOnly requires file | T-3 | `TestBbolt_OpenReadOnlyMissingFile` |
| REQ-2.5 Migrate no-op в RO | T-3 | `TestBbolt_MigrateNoOpInRO` |
| REQ-2.6 ReadOnly() reflects mode | T-3 | `TestBbolt_OpenRWReadOnlyFalse` + `TestBbolt_OpenReadOnlyTrue` |
| REQ-3.1 fakes ReadOnly == false | T-2 | `TestFakes_ReadOnlyAlwaysFalse` |
| REQ-4.1 --readonly + --ro flags | T-5 | `TestCLI_ReadOnlyFlagParsed`, `TestCLI_ReadOnlyAliasRO`, `TestProp_CLI_ReadOnlyFlagAlias` |
| REQ-4.2 explicit flag forces RO | T-5 | manual smoke (`--readonly today`) |
| REQ-4.3 default behavior | T-5 (revised) | manual smoke (locked-DB error message) |
| REQ-5.1 main.go wiring | T-5 | manual smoke |
| REQ-6.1 Model.readOnly from Repo | T-6 | `TestTUI_ModelReadOnlyReflectsRepo` |
| REQ-6.2 currentMode priority | T-6 | `TestTUI_CurrentModeReadOnly`, `TestTUI_CurrentModePriorityRespected`, `TestProp_TUI_CurrentModePriority` |
| REQ-6.3 mode chip label | T-6 | `TestTUI_ModeChipReadOnly` |
| REQ-7.1 write keys blocked | T-7 | `TestTUI_WriteKeyBlockedInRO_Complete`, `TestTUI_QuickEntryBlockedInRO`, `TestProp_TUI_WriteKeyBlockedInRO` |
| REQ-7.2 editor open in RO | T-7 | `TestTUI_EditorOpensInRO` |
| REQ-7.3 editor save error | T-7 | `TestTUI_EditorSaveBlockedInRO` |
| REQ-7.4 quick-entry blocked | T-7 | `TestTUI_QuickEntryBlockedInRO` |
| REQ-7.5 bulk dispatch blocked | T-7 | `TestTUI_BulkDispatchBlockedInRO` |
| REQ-8.1-8.3 backward compat | cross-cutting | все 200+ pre-existing тестов PASS |

## Files changed

```
 cmd/todushka/main.go                     |  31 ++++++-
 internal/cli/cli_test.go                 |  50 +++++++++++
 internal/cli/deps.go                     |   1 +
 internal/cli/root.go                     |   5 ++
 internal/storage/bbolt/bbolt.go          |  90 ++++++++++++++++++-
 internal/storage/bbolt/bbolt_test.go     | 146 +++++++++++++++++++++++++++++++
 internal/storage/fakes/inmemrepo.go      |   4 +
 internal/storage/fakes/inmemrepo_test.go |   5 ++
 internal/storage/repository.go           |   6 ++
 internal/tui/app.go                      |  29 ++++++
 internal/tui/bulk.go                     |   6 ++
 internal/tui/shell.go                    |  11 ++-
 internal/tui/shell_test.go               | 131 +++++++++++++++++++++++++++
 13 files changed, 509 insertions(+), 6 deletions(-)
```

New files (untracked):
- `.spec/features/readonly-mode/{explore,requirements,design,task-plan,implementation}.md`
- `internal/storage/errors_test.go`
- `internal/storage/bbolt/readonly_pbt_test.go`
- `internal/tui/readonly_pbt_test.go`

## Open items / follow-up

Никаких known issues. Готово к review (phase 6).

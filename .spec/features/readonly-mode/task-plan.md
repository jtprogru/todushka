# Read-Only Mode (v0.7.0) — Task Plan

## ⚠️ Pivot during T-3 (2026-05-26): Option B chosen

During T-3 implementation, a property of bbolt's `flock` semantics on Unix
surfaced: **`LOCK_SH` (shared/read-only) cannot be granted while another
process holds `LOCK_EX` (exclusive/writer)**. This makes the original
"auto-fallback to RO when primary writer is alive" goal impossible without
snapshotting the database file.

User chose **Option B** (clearer error + explicit `--readonly` only) over
Option A (snapshot-copy fallback) — see explore.md notes added later.

### Changes vs original plan

| What | Status |
|---|---|
| `storage.ErrReadOnly` sentinel | ✅ keep (T-2) |
| `Repository.ReadOnly() bool` | ✅ keep (T-2) |
| `bbolt.Repo.readOnly` field | ✅ keep (T-3) |
| `bbolt.ReadOnly()` method | ✅ keep (T-3) |
| `bbolt.checkWritable()` helper | ✅ keep (T-3, used in T-4) |
| `bbolt.OpenReadOnly(path)` | ✅ keep (T-3) — still useful when DB unlocked + `--readonly` flag |
| `bbolt.OpenAuto(path)` | ❌ **REMOVED** — could not deliver intended semantic |
| `bbolt.Migrate` no-op in RO | ✅ keep (T-3) |
| `TestBbolt_OpenAutoUnlockedReturnsWrite` | ❌ removed |
| `TestBbolt_OpenAutoLockedFallsBackToReadOnly` | ❌ removed (was failing per flock) |
| New tests `TestBbolt_OpenLockedReturnsErrDatabaseLocked` + `TestBbolt_OpenReadOnlyLockedReturnsErrDatabaseLocked` | ✅ added (T-3, documents real behavior) |
| T-4 bbolt write checkWritable guards | ✅ keep — `--readonly` still produces a writable-blocked repo |
| T-5 `--readonly` + `--ro` flag | ✅ keep, **but no auto-fallback in main.go**. Default `Open`; on `ErrDatabaseLocked` print clear message and exit. With `--readonly`: use `OpenReadOnly`. |
| T-6 TUI `Model.readOnly` + mode chip | ✅ keep |
| T-7 TUI write-key blocking | ✅ keep |
| T-8 PBT — CP-3 (OpenAuto fallback) | ❌ **REMOVED** — property cannot hold |
| Other PBTs (CP-1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15) | ✅ keep |
| Auto-fallback warning logging (ADR-6) | ❌ N/A — no auto-fallback |

### What v0.7.0 actually delivers

- `todushka --readonly` (or `--ro`) **explicitly** opens the DB in shared RO mode. Useful when:
  - No other instance has the DB open (most common case for audit/inspect scripts).
  - You want to guarantee a session won't write — even if you have permissions to do so.
- Without flag, if another instance holds the lock → user sees clearer error message:
  ```
  Error: database is locked by another todushka process.
         Close the other instance and try again, or run with --readonly
         once the writer has exited (note: a writer holding the database
         currently prevents read-only opens too due to file-locking).
  ```
- TUI write keys (`c`, `x`, `d`, `p`, `n`, editor save, bulk dispatch) blocked in RO with status message.
- Mode chip `-- READ-ONLY --` visible when RO.

## Preamble

### Work Type Classification

**Pure feature** with **significant preservation surface**: 200+ существующих тестов должны проходить. `Repository` interface получает новый метод — все impls (bbolt, fakes) обновляются. 16+ write methods в bbolt получают `checkWritable()` guard — mechanical но widespread.

### Test Style Source

**Tier 2** — adjacent tests
- `internal/storage/bbolt/bbolt_test.go`, `internal/storage/fakes/*_test.go`, `internal/cli/cli_test.go`, `internal/tui/*_test.go`.
- testify `require`; existing fixtures (`newTestModel`, `setupModelWithInboxTasks`, `newTestDeps`).
- For lock conflict: use direct `bolt.Open(path, ..., &bolt.Options{Timeout: 0})` чтобы взять exclusive lock + then test OpenAuto fallback.
- Property tests: `pgregory.net/rapid`.

### Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test race  | `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

### Coverage Matrix

| Requirement | Task(s) | CP |
|-------------|---------|----|
| REQ-1.1 ErrReadOnly sentinel | T-2 | — |
| REQ-1.2 Repository.ReadOnly() interface | T-2 | — |
| REQ-1.3 writes return ErrReadOnly | T-3, T-4 | CP-1 |
| REQ-1.4 reads work always | T-3 | CP-2 |
| REQ-2.1 Open exclusive (existing) | T-3 (verify preservation) | CP-5 |
| REQ-2.2 OpenReadOnly | T-3 | CP-4 |
| REQ-2.3 OpenAuto fallback | T-3 | CP-3 |
| REQ-2.4 OpenReadOnly requires file | T-3 | — |
| REQ-2.5 Migrate no-op в RO | T-3 | CP-6 |
| REQ-2.6 ReadOnly() reflects mode | T-3 | CP-5 |
| REQ-3.1 fakes ReadOnly == false | T-2 | CP-7 |
| REQ-4.1 --readonly + --ro flags | T-5 | CP-14 |
| REQ-4.2 explicit flag forces RO | T-5 | — |
| REQ-4.3 default uses OpenAuto | T-5 | — |
| REQ-5.1 main.go wiring | T-5 | CP-15 |
| REQ-6.1 Model.readOnly from svc.Repo | T-6 | CP-8 |
| REQ-6.2 currentMode priority chain | T-6 | CP-9 |
| REQ-6.3 mode chip READ-ONLY label | T-6 | CP-10 |
| REQ-7.1 write keys blocked + status | T-7 | CP-11, CP-12 |
| REQ-7.2 editor open allowed in RO | T-7 | — |
| REQ-7.3 editor save error | T-7 | CP-13 |
| REQ-7.4 quick entry submit blocked | T-7 | CP-11 |
| REQ-7.5 bulk dispatch blocked | T-7 | CP-11 |
| REQ-8.1 backward compat tests | T-1, T-9 | — |
| REQ-8.2 writable repo → false | T-2, T-3 | CP-5 |
| REQ-8.3 RO repo writes error | T-4 | CP-1 |

26 REQs → 9 tasks → 15 CPs.

---

## Task Order

```
T-1 GREEN (baseline preservation)
  → T-2 CODE (storage interface + fakes ReadOnly)
    → T-3 CODE (bbolt foundation: readOnly field + ReadOnly + checkWritable + OpenReadOnly + OpenAuto + Migrate guard)
      → T-4 CODE (bbolt write methods: insert checkWritable in 16+ methods)
        → T-5 CODE (CLI --readonly flag + main.go wiring)
          → T-6 CODE (TUI Model.readOnly + shellMode + currentMode + chip)
            → T-7 CODE (TUI write-key blocking)
              → T-8 GREEN (PBT batch — 15 CPs)
                → T-9 GATE (Checkpoint)
```

---

## Task: T-1 — Baseline preservation

*_Requirements: REQ-8.1_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Подтвердить что existing tests проходят на baseline.

Subtasks:
- [ ] 1. `go clean -testcache && task test-race` — all packages PASS.
- [ ] 2. `task lint` — 0 issues.

After all subtasks: baseline established.

---

## Task: T-2 — Storage interface + fakes ReadOnly

*_Requirements: REQ-1.1, REQ-1.2, REQ-3.1_*
*_Preservation: existing storage tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Добавить `ErrReadOnly` sentinel в storage package; добавить `ReadOnly() bool` метод в Repository interface; реализовать `ReadOnly()` для fakes (всегда false).

Subtasks:

1. В `internal/storage/repository.go` добавить:
```go
ErrReadOnly = errors.New("storage: repository is read-only")
```
в `var (...)` block рядом с existing errors.

2. В `internal/storage/repository.go` добавить метод в `Repository` interface (в конец):
```go
// ReadOnly reports whether this repository was opened in read-only mode.
// When true, all write methods return ErrReadOnly. Read methods work
// regardless.
ReadOnly() bool
```

3. В `internal/storage/fakes/inmemrepo.go` (или wherever struct defined) добавить:
```go
// ReadOnly reports whether this fake repository is read-only.
// In-memory fakes are always writable.
func (r *Repo) ReadOnly() bool { return false }
```
(Имя receiver type — посмотреть существующий код. `Repo` или другое.)

4. Run `task test` — все existing tests должны passing. Если bbolt не имеет ReadOnly() — будет compile error (it doesn't satisfy Repository interface). Это нормально — fix в T-3.

5. Прокомментировать: на этом этапе `task build` НЕ компилируется (bbolt missing ReadOnly). Это OK, T-3 fix'нет. Verify только что storage и fakes пакеты компилируются standalone: `go test ./internal/storage/... ./internal/storage/fakes/...`.

6. Добавить test в `internal/storage/fakes/inmemrepo_test.go` (или новый file):
```go
func TestFakes_ReadOnlyAlwaysFalse(t *testing.T) {
    r := fakes.New()
    require.False(t, r.ReadOnly())
}
```
И test в `internal/storage/repository_test.go` (если есть) ИЛИ create:
```go
func TestErrReadOnly_IsSentinel(t *testing.T) {
    require.True(t, errors.Is(storage.ErrReadOnly, storage.ErrReadOnly))
    require.NotNil(t, storage.ErrReadOnly.Error())
}
```

After all subtasks: interface + fakes complete. Compile broken globally until T-3.

---

## Task: T-3 — bbolt foundation (readOnly field + constructors + Migrate guard)

*_Requirements: REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4, REQ-2.5, REQ-2.6, REQ-1.4, REQ-8.2_*
*_Preservation: T-1, T-2 + existing bbolt tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Расширить `bbolt.Repo` с `readOnly` field; добавить `ReadOnly()`, `checkWritable()`, `OpenReadOnly`, `OpenAuto`. `Migrate` no-op в RO. Existing `Open` unchanged.

Subtasks:

1. В `internal/storage/bbolt/bbolt.go` добавить field в `Repo` struct:
```go
type Repo struct {
    db       *bolt.DB
    path     string
    readOnly bool  // [NEW]
}
```

2. В `internal/storage/bbolt/bbolt.go` добавить methods:
```go
// ReadOnly reports whether the repo was opened in read-only mode.
func (r *Repo) ReadOnly() bool { return r.readOnly }

// checkWritable returns storage.ErrReadOnly if read-only.
// Called at the start of every write method.
func (r *Repo) checkWritable() error {
    if r.readOnly {
        return storage.ErrReadOnly
    }
    return nil
}
```

3. В `internal/storage/bbolt/bbolt.go` добавить `OpenReadOnly`:
```go
// OpenReadOnly opens the database in shared read-only mode. Multiple
// read-only opens of the same path are allowed by bbolt's lock semantics.
// Returns an error if the file does not exist (cannot create in RO mode).
func OpenReadOnly(path string) (*Repo, error) {
    if _, err := os.Stat(path); err != nil {
        return nil, fmt.Errorf("bbolt: open read-only: %w", err)
    }
    db, err := bolt.Open(path, 0o600, &bolt.Options{
        Timeout:  openTimeout,
        ReadOnly: true,
    })
    if err != nil {
        if errors.Is(err, bolterrors.ErrTimeout) {
            return nil, storage.ErrDatabaseLocked
        }
        return nil, fmt.Errorf("bbolt: open read-only: %w", err)
    }
    return &Repo{db: db, path: path, readOnly: true}, nil
}
```

4. В `internal/storage/bbolt/bbolt.go` добавить `OpenAuto`:
```go
// OpenAuto attempts to open in exclusive write mode. On lock conflict
// (storage.ErrDatabaseLocked), retries via OpenReadOnly. Other errors
// are returned as-is.
func OpenAuto(path string) (*Repo, error) {
    r, err := Open(path)
    if err == nil {
        return r, nil
    }
    if errors.Is(err, storage.ErrDatabaseLocked) {
        return OpenReadOnly(path)
    }
    return nil, err
}
```

5. В существующей `Open` функции — verify что `readOnly: false` явно установлен в Repo literal. Если struct literal — добавить `readOnly: false` (optional but explicit).

6. В `internal/storage/bbolt/bbolt.go` Migrate function — добавить early-return:
```go
func (r *Repo) Migrate(ctx context.Context, target int) error {
    if r.readOnly {
        return nil // No-op in read-only mode (REQ-2.5)
    }
    // existing migration logic
    ...
}
```
Найти существующую Migrate function (вероятно в migrations.go или bbolt.go). Если она в bbolt.go — поправить inline. Если в migrations.go — open + edit.

7. В `internal/storage/bbolt/bbolt.go` `Open` функция вызывает `r.ensureBuckets()` и `r.Migrate(...)`. В RO mode мы НЕ хотим `ensureBuckets`. Но `Open` is unchanged — это для write mode. `OpenReadOnly` NOT calls ensureBuckets. Verify в коде:

OpenReadOnly должен:
- Open with ReadOnly=true option
- Skip ensureBuckets (RO can't create buckets)
- Skip Migrate (would fail anyway, but also guarded)
- Return Repo

Update OpenReadOnly accordingly. Look at existing `Open` body, copy structure but skip ensureBuckets/Migrate calls. Just return after opening.

8. Add tests в `internal/storage/bbolt/bbolt_test.go`:
```go
func TestBbolt_OpenRWReadOnlyFalse(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    r, err := Open(path)
    require.NoError(t, err)
    defer r.Close()
    require.False(t, r.ReadOnly())
}

func TestBbolt_OpenReadOnlyTrue(t *testing.T) {
    // First create the DB via write mode
    path := filepath.Join(t.TempDir(), "test.db")
    rw, err := Open(path)
    require.NoError(t, err)
    require.NoError(t, rw.Close())

    ro, err := OpenReadOnly(path)
    require.NoError(t, err)
    defer ro.Close()
    require.True(t, ro.ReadOnly())
}

func TestBbolt_OpenReadOnlyMissingFile(t *testing.T) {
    path := filepath.Join(t.TempDir(), "missing.db")
    _, err := OpenReadOnly(path)
    require.Error(t, err)
}

func TestBbolt_OpenAutoUnlockedReturnsWrite(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    r, err := OpenAuto(path)
    require.NoError(t, err)
    defer r.Close()
    require.False(t, r.ReadOnly())
}

func TestBbolt_OpenAutoLockedFallsBackToReadOnly(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    // Acquire write lock first
    first, err := Open(path)
    require.NoError(t, err)
    defer first.Close()
    // Second attempt — should fallback to RO
    second, err := OpenAuto(path)
    require.NoError(t, err)
    defer second.Close()
    require.True(t, second.ReadOnly())
}

func TestBbolt_MigrateNoOpInRO(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    rw, err := Open(path)
    require.NoError(t, err)
    require.NoError(t, rw.Close())
    ro, err := OpenReadOnly(path)
    require.NoError(t, err)
    defer ro.Close()
    require.NoError(t, ro.Migrate(context.Background(), storage.CurrentSchemaVersion))
}
```

9. Run `task test-race && task lint`. ВНИМАНИЕ: всё ещё может не компилироваться из-за того что bbolt write methods не имеют checkWritable — но Repository interface satisfied (ReadOnly() есть). T-4 завершит write methods. Tests T-3 должны passing.

After all subtasks: bbolt foundation готова.

---

## Task: T-4 — bbolt write methods checkWritable guards

*_Requirements: REQ-1.3, REQ-8.3_*
*_Preservation: T-1..T-3_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Add `if err := r.checkWritable(); err != nil { return err }` at the start of every write method. 16+ methods. Mechanical но widespread.

Subtasks:

1. List all write methods in bbolt package:
```
grep -n "^func (r \*Repo) " internal/storage/bbolt/*.go
```
Identify write ones: TaskCreate, TaskUpdate, TaskDelete, ProjectCreate, ProjectUpdate, ProjectDelete, HeadingCreate, HeadingUpdate, HeadingDelete, AreaCreate, AreaUpdate, AreaDelete, TagCreate, TagUpsert, TagRename, TagDelete. Note: Migrate guarded in T-3 already.

2. For each write method, insert at start (after opening brace, before existing logic):
```go
if err := r.checkWritable(); err != nil {
    return err
}
```
For methods that return additional values (e.g., `TagUpsert(ctx, name) (tag.Tag, error)`), use:
```go
if err := r.checkWritable(); err != nil {
    return tag.Tag{}, err
}
```

3. Run `task build` — should compile. `task test` — existing tests should pass (writable mode default, no behavior change).

4. Add tests в `internal/storage/bbolt/bbolt_test.go` для each write method в RO:
```go
func TestBbolt_TaskCreateErrReadOnly(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    rw, err := Open(path)
    require.NoError(t, err)
    require.NoError(t, rw.Close())
    ro, err := OpenReadOnly(path)
    require.NoError(t, err)
    defer ro.Close()
    err = ro.TaskCreate(context.Background(), task.Task{ID: id.New(), Title: "x"})
    require.ErrorIs(t, err, storage.ErrReadOnly)
}
```

Repeat pattern for each write method, or write parameterized table-driven test:
```go
func TestBbolt_AllWritesReturnErrReadOnly(t *testing.T) {
    path := filepath.Join(t.TempDir(), "test.db")
    rw, err := Open(path)
    require.NoError(t, err)
    require.NoError(t, rw.Close())
    ro, err := OpenReadOnly(path)
    require.NoError(t, err)
    defer ro.Close()
    ctx := context.Background()

    tests := []struct {
        name string
        op   func() error
    }{
        {"TaskCreate", func() error { return ro.TaskCreate(ctx, task.Task{ID: id.New(), Title: "x"}) }},
        {"TaskUpdate", func() error { return ro.TaskUpdate(ctx, task.Task{ID: id.New()}) }},
        {"TaskDelete", func() error { return ro.TaskDelete(ctx, id.New(), false) }},
        // ... всех 16+
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            require.ErrorIs(t, tc.op(), storage.ErrReadOnly)
        })
    }
}
```

5. Run `task test-race && task lint`. All passes.

After all subtasks: bbolt write protection complete.

---

## Task: T-5 — CLI --readonly flag + main.go wiring

*_Requirements: REQ-4.1, REQ-4.2, REQ-4.3, REQ-5.1_*
*_Preservation: T-1..T-4 + existing cli tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Добавить `--readonly` and `--ro` persistent flags; route в `cli.Deps.ReadOnly`; в `main.go` choose constructor.

Subtasks:

1. В `internal/cli/deps.go` добавить `ReadOnly bool` field в `Deps` struct. Default false.

2. В `internal/cli/root.go` найти `PersistentFlags()` section. Добавить:
```go
root.PersistentFlags().BoolVar(&deps.ReadOnly, "readonly", false, "Open database in read-only mode")
root.PersistentFlags().BoolVar(&deps.ReadOnly, "ro", false, "Alias for --readonly")
```
Both target same field.

ВНИМАНИЕ: cobra может protest на duplicate flag value targeting same var — verify in test. Alternative: register only `--readonly`, then alias через `cobra.Command.Aliases` (which is for command aliases, not flag aliases — won't work). Best: use shorthand `-r`:
```go
root.PersistentFlags().BoolVarP(&deps.ReadOnly, "readonly", "r", false, "Open database in read-only mode")
```
Then `--readonly` and `-r` both work. `--ro` not a standard short form (single char convention). However, user explicitly asked for `--ro` alias. Try: register `--readonly` with shorthand `-r`, plus separate `BoolVar` for `--ro`:
```go
root.PersistentFlags().BoolVarP(&deps.ReadOnly, "readonly", "r", false, "Open database in read-only mode")
root.PersistentFlags().BoolVar(&deps.ReadOnly, "ro", false, "Alias for --readonly")
```
If cobra rejects → use only `--readonly` and document `--ro` is unavailable. Verify behavior.

3. В `cmd/todushka/main.go` найти `bbolt.Open(filepath.Join(dataDir, "db"))` (line 32). Заменить:
```go
dbPath := filepath.Join(dataDir, "db")
var repo *bbolt.Repo
if /* deps.ReadOnly access — but deps аren't constructed yet */ {
    repo, err = bbolt.OpenReadOnly(dbPath)
} else {
    repo, err = bbolt.OpenAuto(dbPath)
}
```

WAIT — `deps` is created AFTER `bbolt.Open`. Need to:
- Parse flags first to know if --readonly set, OR
- Always call `OpenAuto` (which auto-fallbacks) and ignore the flag at this layer.

Approach: parse flags via cobra's pre-RunE OR just always use OpenAuto AND let TUI respect flag too.

Best: pre-parse just the `--readonly` and `--ro` flag values directly via `os.Args` scan OR use cobra's PreRunE.

Cleanest: defer DB open into root cobra `RunE` / `PersistentPreRunE` instead of main.go. But that changes architecture significantly.

Simplest: pre-parse os.Args for "--readonly" / "--ro":
```go
readonlyFlag := false
for _, a := range os.Args[1:] {
    if a == "--readonly" || a == "--ro" || a == "-r" {
        readonlyFlag = true
        break
    }
}
dbPath := filepath.Join(dataDir, "db")
var repo *bbolt.Repo
if readonlyFlag {
    repo, err = bbolt.OpenReadOnly(dbPath)
    if err == nil {
        fmt.Fprintln(os.Stderr, "warning: opened in read-only mode (--readonly)")
    }
} else {
    repo, err = bbolt.OpenAuto(dbPath)
    if err == nil && repo.ReadOnly() {
        fmt.Fprintln(os.Stderr, "warning: database locked by another process, opened in read-only mode")
    }
}
```

Yes, simple manual scan is acceptable. Persistent flag в `deps.ReadOnly` is still registered for help-text visibility и для cobra ecosystem.

4. После `repo` constructed: оставить existing `defer repo.Close()`, `svc := app.New(...)`, `deps := cli.DefaultDeps(svc)`, `deps.LaunchTUI = tui.Run`, `cli.Execute(deps)`.

5. В `internal/cli/cli_test.go` добавить:
```go
func TestCLI_ReadOnlyFlagParsed(t *testing.T) {
    deps, _, _ := newTestDeps(t, nil)
    deps.LaunchTUI = func(*app.Service, config.AppConfig) error { return nil }
    root := NewRootCmd(deps)
    root.SetArgs([]string{"--readonly"})
    require.NoError(t, root.Execute())
    require.True(t, deps.ReadOnly)
}
```
(Or similar pattern with existing fixtures.)

Note: cobra binds flag value to `&deps.ReadOnly` (pointer). On `Execute()`, `deps.ReadOnly` updated. But `deps` passed by value into closures — closures may capture old copy. Use the same pattern as for config in tui-shell — flag → PersistentPreRunE captures into shared var.

Look at existing root.go to see how config flag was wired in tui-shell — match that pattern.

6. Run `task test-race && task lint`. All passes.

After all subtasks: CLI integration complete.

---

## Task: T-6 — TUI Model.readOnly + shellMode + chip

*_Requirements: REQ-6.1, REQ-6.2, REQ-6.3_*
*_Preservation: T-1..T-5 + existing tui tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Add `Model.readOnly` field; `modeReadOnly` shellMode const; update `currentMode` priority; mode chip renders `READ-ONLY`.

Subtasks:

1. В `internal/tui/app.go` добавить `readOnly bool` field в `Model` struct (после `confirm` или другого logical place).

2. В `internal/tui/app.go` `NewModel` — добавить argument OR auto-detect from svc.Repo().ReadOnly():
   - Option A: signature change `NewModel(svc, theme, cfg) Model` already exists; add field initialization:
     ```go
     readOnly: svc.Repo().ReadOnly(),
     ```
   - This auto-detects from injected service. No signature change.

3. В `internal/tui/shell.go` добавить `modeReadOnly` в enum:
```go
const (
    modeNormal shellMode = iota
    modeFilter
    modeSelect
    modeConfirm
    modeEditor
    modeHelp
    modeReadOnly  // [NEW]
)
```

4. В `internal/tui/shell.go` `modeLabel()` switch — добавить case:
```go
case modeReadOnly:
    return "READ-ONLY"
```

5. В `internal/tui/shell.go` `currentMode()` — обновить priority chain. Перед `default: return modeNormal`:
```go
case m.readOnly:
    return modeReadOnly
```
Цепочка должна стать:
```go
switch {
case m.screen == screenHelp:
    return modeHelp
case m.screen == screenEditor:
    return modeEditor
case m.confirm != nil:
    return modeConfirm
case m.filtering:
    return modeFilter
case len(m.selected) > 0:
    return modeSelect
case m.readOnly:  // [NEW] — priority between SELECT and NORMAL
    return modeReadOnly
default:
    return modeNormal
}
```

6. В `internal/tui/shell.go` `modeKeyHints()` для `modeReadOnly` — добавить case:
```go
case modeReadOnly:
    return []string{"/: filter", "↵: view", "?: help", "q: quit"}
```
(Только read hints — нет complete/edit/etc.)

7. Add tests в `internal/tui/shell_test.go`:
```go
func TestTUI_ModelReadOnlyReflectsRepo(t *testing.T) {
    // Use a fake repo that reports ReadOnly() == true
    // Since fakes always returns false, we need a custom test repo.
    // Easiest: wrap fakes with a roFake { *fakes.Repo; readOnly bool } that override ReadOnly().
    // Or: use bbolt directly with RO mode.
    // For now: extract logic into helper function and test directly.
    require.Equal(t, true, true) // placeholder; concrete test уточняется в импл.
}

func TestTUI_CurrentModeReadOnly(t *testing.T) {
    m := newTestModel(t)
    m.readOnly = true
    require.Equal(t, modeReadOnly, currentMode(m))
}

func TestTUI_CurrentModePriorityRespected(t *testing.T) {
    m := newTestModel(t)
    m.readOnly = true
    m.filtering = true
    require.Equal(t, modeFilter, currentMode(m), "filter overrides RO")

    m.filtering = false
    m.confirm = &confirmState{}
    require.Equal(t, modeConfirm, currentMode(m), "confirm overrides RO")
}

func TestTUI_ModeChipReadOnly(t *testing.T) {
    m := newTestModel(t)
    m.readOnly = true
    out := m.viewFooter()
    require.Contains(t, out, "-- READ-ONLY --")
}
```

For the test using `svc.Repo().ReadOnly()`, write a helper test that constructs Model with a fake repo wrapper.

8. Run `task test-race && task lint`. All passes.

After all subtasks: TUI mode chip готов.

---

## Task: T-7 — TUI write-key blocking

*_Requirements: REQ-7.1, REQ-7.2, REQ-7.3, REQ-7.4, REQ-7.5_*
*_Preservation: T-1..T-6_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Block write keys in RO mode. Editor open allowed but save fails. Quick-entry submit blocked. Bulk dispatch blocked at dispatcher.

Subtasks:

1. В `internal/tui/app.go` добавить helper:
```go
// blockWriteIfReadOnly returns true if the model is in RO mode and sets
// the status message accordingly. Callers should return early on true.
func (m *Model) blockWriteIfReadOnly() bool {
    if !m.readOnly {
        return false
    }
    m.statusMsg = "read-only mode: writes disabled"
    m.statusUntil = time.Now().Add(statusFadeDuration)
    return true
}
```

2. В `internal/tui/bulk.go` `dispatch` — добавить early return:
```go
func dispatch(m Model, action bulkAction) (Model, tea.Cmd) {
    if m.blockWriteIfReadOnly() {
        return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
    }
    // existing logic
}
```
This blocks complete/cancel/delete/pin (when dispatched via key handlers).

Note: blockWriteIfReadOnly is method on `*Model`. `dispatch` takes Model by value. Need adjustment:
```go
func dispatch(m Model, action bulkAction) (Model, tea.Cmd) {
    if m.readOnly {
        m.statusMsg = "read-only mode: writes disabled"
        m.statusUntil = time.Now().Add(statusFadeDuration)
        return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
    }
    // existing logic
}
```
Inline the check for value-type Model functions.

3. В `internal/tui/app.go` `saveEditor` метод — добавить early-return при RO:
```go
func (m Model) saveEditor() (tea.Model, tea.Cmd) {
    if m.readOnly {
        m.editor.err = "read-only mode: writes disabled"
        return m, nil
    }
    // existing logic
}
```

4. В `internal/tui/app.go` `handleQuickEntryKey` для tea.KeyEnter case — check RO before emitting quickEntrySubmittedMsg:
```go
case tea.KeyEnter:
    raw := m.quickInput.Value()
    m.screen = screenList
    if m.readOnly {
        m.statusMsg = "read-only mode: writes disabled"
        m.statusUntil = time.Now().Add(statusFadeDuration)
        return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
    }
    return m, tea.Batch(
        func() tea.Msg { return quickEntrySubmittedMsg{raw: raw} },
        m.loadCurrentList(),
    )
```

5. Add tests в `internal/tui/shell_test.go` или `bulk_test.go`:
```go
func TestTUI_WriteKeyBlockedInRO_Complete(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.readOnly = true
    m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
    mm := m2.(Model)
    require.Contains(t, mm.statusMsg, "read-only")
    // Task remains open (no completion)
    // Verify task.Status == StatusOpen via svc.Repo().TaskGet
}

func TestTUI_EditorOpensInRO(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.readOnly = true
    m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    require.Equal(t, screenEditor, m2.(Model).screen)
}

func TestTUI_EditorSaveBlockedInRO(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.readOnly = true
    // Open editor
    m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    mm := m2.(Model)
    require.Equal(t, screenEditor, mm.screen)
    // Press Ctrl+S to save
    m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
    mm = m3.(Model)
    require.Contains(t, mm.editor.err, "read-only")
}

func TestTUI_BulkDispatchBlockedInRO(t *testing.T) {
    m, _, tasks := setupModelWithInboxTasks(t, "x", "y")
    m.readOnly = true
    for _, tk := range tasks {
        m.selected[tk.ID] = struct{}{}
    }
    m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
    mm := m2.(Model)
    require.Contains(t, mm.statusMsg, "read-only")
    require.Nil(t, mm.confirm, "confirm modal NOT opened in RO")
}
```

6. Run `task test-race && task lint`. All passes.

After all subtasks: write-key blocking complete.

---

## Task: T-8 — Property-based tests batch

*_Requirements: ALL_*
*_Preservation: ALL_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать 15 PBT из design §2.6.

Subtasks:

1. В `internal/storage/bbolt/bbolt_test.go` или новый file добавить PBT для storage (CP-1..CP-7).
2. В `internal/tui/shell_test.go` или новый file добавить PBT для TUI (CP-8..CP-13).
3. В `internal/cli/cli_test.go` PBT для flag equivalence (CP-14).
4. В `cmd/todushka/main_test.go` (если возможно) или integration test — CP-15 stderr warning.
5. Run `task test-race -count=2 -timeout=180s`. Stability check.

After all subtasks: 15 CPs covered.

---

## Task: T-9 — GATE Checkpoint

*_Requirements: ALL_*
*_Complexity: mechanical_*

CRITICAL: ПОСЛЕДНЯЯ задача.

Instructions:

1. `go clean -testcache && task test` — all packages PASS.
2. `task test-race` — race-free.
3. `task build` — bin/todushka compiles.
4. `task lint` — 0 issues.
5. `gofmt -l internal/ cmd/` — empty.
6. Coverage matrix verification.
7. Manual smoke:
   - Start primary: `./bin/todushka` (или через taskfile).
   - Start secondary в другом terminal: видит `-- READ-ONLY --` chip; writes blocked.
   - `./bin/todushka --readonly` forces RO.
   - `./bin/todushka --ro` — alias.
8. Manual smoke editor в RO: открывается, Ctrl+S → m.editor.err shows "read-only mode".
9. Если что-то fails — вернуться к T-N.

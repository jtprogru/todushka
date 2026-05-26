# Exploration: Read-Only Mode for Second Instance

## Intent

Сейчас при попытке запустить вторую копию todushka на том же datapath:

```
todushka: storage: database is locked by another todushka process
```

Это происходит потому что bbolt берёт **exclusive lock** на файл. Пользовательский запрос (из первоначального списка 4 пунктов в ux-polish): либо явное сообщение, либо запустить read-only режим для мониторинга. Был отложен из ux-polish в отдельную фичу — read-only — потому что требует storage interface changes.

Цель v0.7.0: при lock-conflict пробовать открыть DB в **shared read-only** режиме; TUI запускается и позволяет навигацию/просмотр; все write-операции возвращают user-friendly error через status bar; mode chip показывает `[RO]` indicator. Когда primary instance закрывается — secondary остаётся в read-only (не auto-upgrade — слишком сложно). Пользователь видит ясно что он в read-only.

## Investigation

### bbolt locking mechanism

`bolt.Open(path, mode, options)`:
- Default: exclusive lock (writer mode). Conflict → blocks until timeout, returns `bolterrors.ErrTimeout`.
- `options.ReadOnly = true`: shared lock (reader mode). Multiple readers OK. Writer can't open while readers exist (need to upgrade).

Текущее (`internal/storage/bbolt/bbolt.go:59`):
```go
db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: openTimeout})
```

Timeout 200ms. На lock conflict returns `storage.ErrDatabaseLocked`.

### Repository interface

`internal/storage/repository.go` определяет 25+ методов:
- Reads: `TaskGet`, `TaskList`, `ProjectGet`, etc. — должны работать в read-only.
- Writes: `TaskCreate`, `TaskUpdate`, `TaskDelete`, etc. — должны возвращать error.

Текущие методы возвращают error на bbolt errors. Нужно добавить новый sentinel `storage.ErrReadOnly` и flag на Repository.

### Service layer

`internal/app/service.go` использует Repo() напрямую. Не нужно менять — пробрасывает errors as-is.

### TUI layer

Mode indicator (`internal/tui/shell.go`): сейчас `viewFooter` показывает `-- NORMAL --` chip. Нужно добавить `[RO]` prefix или отдельный chip когда `svc.Repo().ReadOnly() == true`.

Key handlers: при попытке write (complete, cancel, delete, pin, edit save, quick entry, bulk) — нужно проверить readonly и не разрешать. Или просто let it fail at service level и показать error в status bar — это проще (no new key-level guards).

### Multi-instance lifecycle

- **Primary** запускается first — открывает RW.
- **Secondary** не может взять exclusive lock → fallback к RO.
- **Secondary** живёт пока user явно не закроет.
- **Primary** закрывается — secondary НЕ auto-upgrade. Это требует:
  - file watching (changes on disk),
  - lock upgrade attempts,
  - notification to user,
  - state reload.
  Слишком сложно для v1. Defer.

## Build Tooling

Unchanged: `task test`, `task test-race`, `task build`, `task lint`, `task fmt`.

## Options Considered

### Option A — Full read-only mode

- `storage.ErrReadOnly` sentinel.
- `Repository.ReadOnly() bool` method.
- bbolt fallback: try write mode, on timeout try ReadOnly=true.
- All write methods check `r.readOnly`, return `ErrReadOnly`.
- TUI shows `[RO]` mode chip; write attempts produce status bar error.

**Pros:**
- Решает user pain полностью.
- Useful для мониторинга / debug.

**Cons:**
- Storage interface change (new method).
- 25+ method changes в bbolt impl (each write checks readonly).
- TUI changes (mode chip + error feedback).

**Complexity:** Medium-Large.

### Option B — Just clearer error message + suggest CLI

При lock conflict — improved error:
```
Error: another todushka instance has the database open.
       Close it first to use the TUI.
       Read-only inspection: use `todushka today`, `todushka export`, etc.
```

**Pros:**
- Малый scope.

**Cons:**
- Не решает запрос мониторинга через TUI.

### Option C — Smart `--readonly` CLI flag

Force read-only on demand: `todushka --readonly` всегда открывает в RO независимо от lock state.

**Pros:**
- Explicit user intent.
- Можно совместить с Option A.

**Cons:**
- Без auto-fallback по конфликту lock, пользователь должен помнить флаг.

## Recommended Direction

**Option A (full read-only) + Option C (explicit `--readonly` flag).**

- Auto-fallback: lock timeout → retry ReadOnly=true → if succeeds, start TUI в RO.
- Explicit flag: `todushka --readonly` принудительно открывает RO (полезно для scripts/automation).
- TUI mode chip: `-- READ-ONLY --` (новая mode), priority выше NORMAL но ниже HELP/EDITOR/CONFIRM/FILTER/SELECT.
- Write attempt в RO → status bar shows "read-only mode: writes disabled" (5sec fade).

### Подробности

**Storage:**
- New `storage.ErrReadOnly` error.
- `Repository.ReadOnly() bool` interface method (add to interface).
- bbolt `Repo` struct gets `readOnly bool` field; all write methods check it first.
- Fakes (`inmemrepo.go`) implement `ReadOnly()` returning `false` always.

**bbolt Open:**
- New `OpenReadOnly(path string) (*Repo, error)` — separate constructor.
- New `OpenAuto(path string) (*Repo, error)` — tries write, falls back to RO on timeout.

**CLI:**
- New `--readonly` persistent flag в root cobra.
- `main.go` chooses constructor based on flag.

**TUI:**
- New `shellMode` constant `modeReadOnly`.
- `currentMode` priority: HELP > EDITOR > CONFIRM > FILTER > SELECT > **READ-ONLY** > NORMAL. (Read-only is sticky — overrides Normal.)
- Mode chip rendering supports new mode label.
- Write actions: when `repo.ReadOnly() == true` AND user presses write key (`c`/`x`/`d`/`p`/`Enter` for editor save/`n` for quick entry/`*` doesn't write itself, but bulk ops do):
  - Block before service call: set `m.statusMsg = "read-only mode: writes disabled"` + fade.
  - Editor: open OK, but save → error.
  - Quick entry: open OK, but submit → error.

## Scope Boundaries

### Must-have (v1)

- `storage.ErrReadOnly` sentinel.
- `Repository.ReadOnly() bool` interface + impl.
- bbolt: write methods check readOnly; `OpenAuto` fallback to RO on lock timeout.
- CLI: `--readonly` persistent flag.
- `cmd/todushka/main.go`: use `OpenAuto` (unless `--readonly` forces RO).
- TUI: `modeReadOnly` shellMode; mode chip renders RO; write keys blocked with status message.
- Backward compat: all existing tests passing; fakes default to RW.

### Deferred (v0.8+)

- **Auto-upgrade** when primary instance closes (file watching, lock retry).
- **Inotify-based real-time refresh** в RO instance (currently читает кэш + manual refresh).
- **Per-method readonly checks vs centralized middleware** (cleaner refactor).

### Needs spike

- **bbolt ReadOnly semantics:** что happens когда мы открыли DB в RO mode и затем writer открывает её? RO mode держит shared lock — writer заблокирован. OK, semantically correct. На macOS via `flock(2)` это standard behavior.

## Constraints & Risks

- **25+ method changes в bbolt impl** — каждая write method должна check `r.readOnly`. Verbose but mechanical.

- **Service layer untouched** — errors пробрасываются. Хорошо: isolation maintained.

- **Editor in RO** — user может открыть редактор (read), но `Ctrl+S` save → error. Should we disable editor entirely в RO? Recommend: open OK (allows viewing all fields), save returns error.

- **Quick entry in RO** — same: allow open, error on submit.

- **Bulk ops in RO** — already check at dispatcher level OR fail at service. Recommend: дополнительный check в `dispatch` — если RO, immediately set status "read-only" без opening confirm modal.

- **Test coverage:** Need to test:
  - bbolt Open with explicit ReadOnly=true succeeds when DB unlocked.
  - bbolt fallback path: simulate lock taken → OpenAuto returns RO Repo.
  - Write methods return ErrReadOnly when readOnly==true.
  - TUI: pressing write key in RO mode shows status "read-only".
  - Mode chip shows `READ-ONLY` when active.
  - Editor save in RO returns error properly displayed.

- **Atomicity of multi-process write detection:** bbolt uses fcntl/flock; reliable on POSIX. Windows может вести себя иначе, но мы строим on darwin/linux primarily.

- **bbolt opening in RO can't create file:** if path doesn't exist, RO open fails. Need to ensure `OpenAuto` only falls back to RO when file exists.

- **Existing `Migrate` only runs on writeable Repo** — in RO, schema migration must skip. Need to guard.

## Recommended Direction Recap

Option A + Option C. Implementation order:

1. **Storage interface:** add `ReadOnly() bool` + `storage.ErrReadOnly`.
2. **bbolt impl:** `readOnly` field; `OpenReadOnly`; `OpenAuto` (try write → RO fallback); 25+ write methods check; skip Migrate in RO.
3. **fakes:** return false from ReadOnly.
4. **CLI:** `--readonly` persistent flag.
5. **main.go:** wire OpenAuto / OpenReadOnly based on flag.
6. **TUI shellMode:** modeReadOnly + currentMode priority.
7. **TUI write key handlers:** check readonly, set status message, return early.
8. **Tests** — unit + property.

## Open Questions

1. **CLI flag name:** `--readonly`, `--read-only`, или `--ro` (alias)? **Recommend `--readonly` + `--ro` alias.**
2. **Mode chip label:** `READ-ONLY`, `READONLY`, `RO`, или `READ ONLY`? **Recommend `READ-ONLY`** (consistent format).
3. **Editor open в RO — allow или block?** **Recommend allow open** (view-mode), block save с error.
4. **Status message duration в RO:** standard 5s fade или dedicated longer? **Recommend standard 5s.**
5. **`--readonly` без lock conflict — что делать?** Even если DB unlocked — открыть в RO. Useful для read-only audits. **Recommend yes, always honor flag.**
6. **Auto-fallback default OR opt-in via flag?** **Recommend auto-fallback always** (without flag): пользователь expect "it just works". `--readonly` only for explicit intent.

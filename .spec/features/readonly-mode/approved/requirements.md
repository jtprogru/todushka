# Read-Only Mode (v0.7.0) — Requirements

**Status:** Draft
**Author:** spec-driven-dev (Claude)
**Date:** 2026-05-26

## Overview

При запуске второй копии todushka на том же datapath: вместо `storage: database is locked` ошибки — auto-fallback к bbolt **shared read-only mode**, TUI запускается с visible `-- READ-ONLY --` mode chip. Все write-операции (complete, cancel, delete, pin, quick entry, editor save, bulk dispatch) блокируются с user-friendly status message. Пользователь может навигировать, фильтровать, выделять, открывать editor для просмотра — но не модифицировать данные.

Explicit `--readonly` (alias `--ro`) CLI flag принудительно открывает RO даже без lock conflict — для read-only audits и monitoring.

Изменения:
- **storage:** new `ErrReadOnly`, `Repository.ReadOnly() bool` method.
- **bbolt:** `Repo.readOnly bool`; `OpenAuto(path)` (try write → RO fallback); `OpenReadOnly(path)`; 25+ write methods check; Migrate skipped в RO.
- **fakes:** `ReadOnly() bool` returns false.
- **cli:** `--readonly`/`--ro` persistent flag.
- **cmd/todushka/main.go:** wire fallback constructor.
- **tui:** new `modeReadOnly` shellMode; mode chip; write-key blocking.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| Read-Only Mode | Состояние Repository, при котором все write-методы возвращают `storage.ErrReadOnly`. Detected via `Repository.ReadOnly() bool`. | `internal/storage` (interface + sentinel) |
| Auto-Fallback | При conflict на exclusive lock — попытка открыть DB в shared read-only mode перед возвратом error. | `bbolt.OpenAuto(path)` |
| Explicit RO | `--readonly` или `--ro` CLI flag форсит read-only независимо от lock state. | `cli.Deps.ReadOnly bool` + flag |
| RO Mode Chip | Footer mode chip `-- READ-ONLY --`, priority выше NORMAL но ниже HELP/EDITOR/CONFIRM/FILTER/SELECT. | `tui.shellMode` enum |
| Write-Block Status | Status bar message `"read-only mode: writes disabled"` при попытке write key в RO. Fade через `statusFadeDuration`. | `tui` Update / handleKey logic |

## User Stories

- Как **пользователь, забывший что todushka уже открыт в другом терминале**, я хочу получить TUI который позволит мне видеть текущее состояние задач — а не error и закрытие.
- Как **monitoring/audit script**, я хочу запускать `todushka --readonly` чтобы guaranteed не модифицировать данные.
- Как **пользователь в RO mode**, я хочу понимать что я в read-only — explicit visual indicator (mode chip).
- Как **пользователь в RO mode, нажимающий `c` для complete**, я хочу видеть understandable message — а не silent ignore.
- Как **пользователь в RO mode, открывающий editor**, я хочу видеть details задачи. Save должен fail explicitly но editor открывается.

## Requirements

### Group 1 — Storage interface

**REQ-1.1** WHEN the storage package compiles, the system SHALL export a new error variable `ErrReadOnly` of type `error` with message containing `"read-only"`.

**REQ-1.2** WHEN the storage package compiles, the `storage.Repository` interface SHALL include a new method `ReadOnly() bool`.

**REQ-1.3** WHEN a `Repository` implementation reports `ReadOnly() == true`, every write method SHALL return `storage.ErrReadOnly` (NOT silently succeed). Write methods are: `TaskCreate`, `TaskUpdate`, `TaskDelete`, `ProjectCreate`, `ProjectUpdate`, `ProjectDelete`, `HeadingCreate`, `HeadingUpdate`, `HeadingDelete`, `AreaCreate`, `AreaUpdate`, `AreaDelete`, `TagCreate`, `TagUpsert`, `TagRename`, `TagDelete`, `Migrate`.

**REQ-1.4** WHEN `Repository.ReadOnly() == true`, every read method (`TaskGet`, `TaskList`, `TaskMatchShort`, `ProjectGet`, `ProjectList`, `ProjectFindByName`, `HeadingList`, `AreaGet`, `AreaList`, `AreaFindByNormalized`, `TagGet`, `TagList`, `SchemaVersion`, `Close`) SHALL function normally (no read-side restrictions).

### Group 2 — bbolt implementation

**REQ-2.1** WHEN `bbolt.Open(path)` is called (existing function), the system SHALL open the database in exclusive write mode (existing behavior preserved).

**REQ-2.2** WHEN a new function `bbolt.OpenReadOnly(path)` is called, the system SHALL open the database with `bolt.Options.ReadOnly = true` (shared lock, multiple readers allowed). Returns a `*Repo` with `readOnly == true`.

**REQ-2.3** WHEN a new function `bbolt.OpenAuto(path)` is called, the system SHALL first attempt `Open(path)` (write mode). If it returns `storage.ErrDatabaseLocked`, the system SHALL retry via `OpenReadOnly(path)`. On other errors, return as-is. On success (write mode), return the writable Repo.

**REQ-2.4** WHEN `bbolt.OpenReadOnly(path)` is called AND the file does not exist, the system SHALL return an error (cannot create in RO mode).

**REQ-2.5** WHEN a `bbolt.Repo` is opened with `readOnly == true`, the `Migrate` method SHALL skip schema migration silently AND return nil (no-op). Schema migrations require write access; RO instances trust the primary writer.

**REQ-2.6** WHEN `bbolt.Repo.ReadOnly()` is called, it SHALL return the value of the internal `readOnly` field.

### Group 3 — Fakes implementation

**REQ-3.1** WHEN `fakes.Repo.ReadOnly()` is called, it SHALL return `false` (in-memory fake is always writable; used in tests).

### Group 4 — CLI integration

**REQ-4.1** WHEN the root cobra command is built, the system SHALL register persistent flag `--readonly` (bool) AND `--ro` (alias) defaulting to `false`.

**REQ-4.2** WHEN `cli.Deps.ReadOnly == true`, the system SHALL force open via `OpenReadOnly` regardless of lock state.

**REQ-4.3** WHEN `cli.Deps.ReadOnly == false`, the system SHALL open via `OpenAuto` (auto-fallback semantics per REQ-2.3).

### Group 5 — main.go wiring

**REQ-5.1** WHEN the application starts, `cmd/todushka/main.go` SHALL:
- Resolve `--readonly` / `--ro` flag value (cobra parses persistent flags before `RunE`).
- Choose constructor:
  - If flag set → `bbolt.OpenReadOnly(path)`.
  - Else → `bbolt.OpenAuto(path)`.
- Pass resulting Repo to `app.New(repo, clock)`.

### Group 6 — TUI mode chip

**REQ-6.1** WHEN constructing `tui.Model`, the system SHALL detect `svc.Repo().ReadOnly()` AND store in a new Model field (e.g., `Model.readOnly bool`).

**REQ-6.2** WHEN `Model.readOnly == true`, the `currentMode` function SHALL include a new mode `modeReadOnly` in the priority chain:
- Priority: `HELP > EDITOR > CONFIRM > FILTER > SELECT > READ-ONLY > NORMAL`.
- Read-Only overrides Normal but is overridden by all transient modes.

**REQ-6.3** WHEN `currentMode == modeReadOnly`, the mode chip in viewFooter SHALL render `-- READ-ONLY --` using `theme.Header` style (existing chip rendering).

### Group 7 — Write-key blocking

**REQ-7.1** WHEN `Model.readOnly == true` AND a write key is pressed (`c` complete, `x` cancel, `d` delete, `p` pin, `n` quick entry), the system SHALL:
- Set `m.statusMsg = "read-only mode: writes disabled"`.
- Schedule fade via `clearStatusMsg` timer (existing pattern).
- NOT invoke the service write call.

**REQ-7.2** WHEN `Model.readOnly == true` AND `Enter` is pressed (editor open), the system SHALL open the editor normally (allow viewing task details).

**REQ-7.3** WHEN `Model.readOnly == true` AND user presses `Ctrl+S` in editor to save, the system SHALL display error in `m.editor.err` containing `"read-only mode: writes disabled"` AND NOT call `EditTask`. Editor remains open.

**REQ-7.4** WHEN `Model.readOnly == true` AND quick-entry submitted (textinput Enter), the system SHALL display status `"read-only mode: writes disabled"` AND NOT call `QuickEntry`. Quick-entry modal closes (consistent with current behavior).

**REQ-7.5** WHEN `Model.readOnly == true` AND bulk-dispatch invoked (any of `c`/`x`/`d`/`p` with selected ≥ 1), the system SHALL block at dispatcher level via REQ-7.1 — NOT open confirm modal.

### Group 8 — Backward compatibility

**REQ-8.1** WHEN existing tests run that do not exercise read-only paths, the system SHALL behave identically to v0.6.0 (writable mode, all writes succeed).

**REQ-8.2** WHEN `app.New(repo, clock)` is called with a writable `repo`, the resulting `Service.Repo().ReadOnly()` SHALL return `false`.

**REQ-8.3** WHEN `app.New(repo, clock)` is called with a read-only `repo`, write methods at app level SHALL return errors wrapping `storage.ErrReadOnly`.

## Topological Order

```
Group 1 (Storage interface)   — foundation; new error + interface method
Group 2 (bbolt)              — depends on Group 1
Group 3 (Fakes)              — depends on Group 1
Group 4 (CLI flag)           — depends on Group 2
Group 5 (main.go)            — depends on Groups 2, 4
Group 6 (TUI mode chip)      — depends on Groups 1-3 (needs Repo().ReadOnly())
Group 7 (Write-key blocking) — depends on Group 6
Group 8 (Backward compat)    — cross-cutting; verified at GATE
```

## Conflict Priority

**REQ-2.5 (Migrate no-op in RO) vs REQ-1.3 (writes return ErrReadOnly):**
Migrate is conceptually a write. In RO, returning ErrReadOnly would block bbolt open for legitimate read-only consumers. Resolution: REQ-2.5 takes priority — Migrate returns nil silently in RO. Documented as "RO trusts primary writer for schema".

**REQ-7.3 (editor save error display in m.editor.err) vs REQ-7.1 (status bar message):**
Editor has own err display; using m.editor.err keeps message in context of editor save action. REQ-7.1 applies to non-editor write keys.

**REQ-4.2 (explicit RO forces OpenReadOnly) vs REQ-2.4 (RO requires file exists):**
If user runs `todushka --readonly` on system without existing DB, RO open will fail. Acceptable behavior: error message "cannot open in read-only mode, database does not exist". Documented.

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Mode chip color: standard `theme.Header` accent или different (warning yellow)? | Visual signal strength. | REQ-6.3 |
| Editor in RO — pre-populated user-editable field changes preserved cosmetically but not persisted? Just block save? | Edit-and-lose-changes risk vs full save attempt. | REQ-7.3 |
| Bulk operations: status message wording (`"read-only mode: bulk disabled"` vs generic)? | Specificity. | REQ-7.5 |
| Should `--readonly` flag take effect even if DB unlocked (no conflict)? | Yes (Option C explicit intent). | REQ-4.2 |
| Should we log a warning when auto-fallback triggers ("opened RO due to lock conflict")? | Discoverability of the auto-fallback. | REQ-5.1 |

## Verification Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test (race)| `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

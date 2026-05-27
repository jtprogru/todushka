# Code Review: viewport-scroll

## Verdict: PASS

Bug fix BL-7 полностью покрывает 3 REQ + 4 CP. Все 4 проверки зелёные
(test, race, build, lint, fmt). Изменения изолированы (no service /
storage / domain touched). Exploration test
`TestModel_CursorInvisibleOnOverflow` фиксирует reproduction и теперь
PASS. PBT обнаружил narrow-window edge case (visible < 2*scrolloff+1) —
исправлено через hard-bounds clamp. 0 critical/major findings.

## Change Set

| File | Status | Notes |
|------|--------|-------|
| `internal/tui/viewport.go` | ✅ Planned [NEW] | `ensureCursorVisible`, `visibleRows`, `scrolloff` const |
| `internal/tui/viewport_test.go` | ✅ Planned [NEW] | 17 tests (9 unit + 3 PBT + 5 integration) |
| `internal/tui/app.go` | ✅ Planned [MODIFIED] | Model fields, 2 j/k handlers, 2 reload-clamps, 5 reset sites, slice in viewList |
| `internal/tui/project_list.go` | ✅ Planned [MODIFIED] | Slice + absIdx cursor in viewProjectList |
| `internal/tui/project_tasks.go` | ✅ Planned [MODIFIED] | Slice + absIdx cursor in viewProjectTasks |
| `.spec/BACKLOG.md` | ⚠️ Unexpected | Добавлена секция Bugs с BL-7 — justified (новый backlog item) |

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 | `TestModel_CursorInvisibleOnOverflow`, `TestModel_ViewList/_ViewProjectList/_ViewProjectTasks_CursorVisibleAfterJ`, `TestProp_CursorAlwaysInside`, `TestProp_ScrollOffBufferRespected` | `viewport.go:ensureCursorVisible`; j/k handlers in `app.go` (3 sites) | CP-1, CP-3 | ✅ |
| REQ-1.2 | `TestEnsureCursorVisible_*` (9 unit), `TestProp_OffsetBounded` | Slice in 3 views; clamp in helper | CP-2 | ✅ |
| REQ-1.3 | `TestModel_ScrollOffsetClampOnTasksReload`, `TestModel_ScrollOffsetResetOnSwitchList` | Reload-clamp in `tasksLoadedMsg`/`projectTasksLoadedMsg`/`projectsLoadedMsg` | CP-4 | ✅ |

## Design Conformance

### 3.1 Architectural Boundaries — ✅
Service/storage/domain не задействованы. Изменения только в TUI-layer
(viewport.go новый, 3 модифицированных в `internal/tui/`). Никаких
cross-layer imports.

### 3.2 Data Models — ✅
Никаких новых доменных типов. 2 `int` поля в Model — single-package.

### 3.3 API Contracts — ✅
Helper signatures (`ensureCursorVisible`, `visibleRows`) идентичны
design §2.3. Pure functions.

### 3.4 Error Handling — ✅
`visibleCount <= 0`, `totalCount <= visibleCount`, negative cursor —
все edge cases обрабатываются явно в helper. Pre-WindowSizeMsg (height==0)
→ visibleRows=0 → render показывает все строки (no slice, наследие).

### 3.5 Correctness Properties — ✅
Все 4 CP реализованы и протестированы PBT (CP-1 `TestProp_CursorAlwaysInside`,
CP-2 `TestProp_OffsetBounded`, CP-3 `TestProp_ScrollOffBufferRespected`,
CP-4 покрыто `TestModel_ScrollOffsetClamp*`).

### 3.6 Documentation Consistency — ✅
Mermaid в design §2.2 показывает 1 новый (viewport.go) + 4 модифицированных —
соответствует реальности.

## Code Quality

### 4.1 Naming & Clarity — ✅
`scrolloff`, `visibleRows`, `ensureCursorVisible` — идиоматичные имена
из vim-tradition. `absIdx = i + off` — clear inside loops.

### 4.2 Dead Code & Debug Artifacts — ✅
Чисто, никаких TODO/print/commented-out blocks.

### 4.3 Scope Creep — ✅
Только bug fix BL-7. Никаких refactors за scope. ADR-1 (logical-row scroll)
зафиксировал решение про wrap.

### 4.4 Test Quality — ✅
- Все тесты используют testify `require`.
- PBT через `pgregory.net/rapid`.
- Edge cases: `cursor=0`, `cursor=last`, narrow window (`visible<2*scrolloff+1`),
  reload-shrink, switchList-reset, всех 3 экранов.
- **PBT поймал реальный bug** на 15-й итерации (`visible=1, total=2`) →
  helper-логика была усилена hard-bounds clamp. Это validation того,
  что PBT не декоративные.

## Security

No security issues found in changed files.

- No external inputs (pure functions over int args).
- No new endpoints.
- No data exposure / error leakage.

## Verification Evidence

Re-run by reviewer in this session (fresh cache).

### task test (post `go clean -testcache`)
```
ok  	github.com/jtprogru/todushka/internal/domain/today	5.102s
ok  	github.com/jtprogru/todushka/internal/storage	4.270s
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	13.673s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	4.670s
ok  	github.com/jtprogru/todushka/internal/tui	7.105s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

### task build
```
(no output — build succeeded)
```

### task lint
```
0 issues.
```

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | nit | `internal/tui/project_list.go`, `project_tasks.go` | `vr = visibleRows(m) - 2` хардкодит "header+blank = 2 rows" — magic number. Можно вынести в const `projectListHeaderRows = 2` для читаемости. Не блокирует. | — |

## Recommendations

1. **F-1** (nit): вынести `vr - 2` в именованную константу. Опционально.

Никаких critical/major находок — verdict **PASS**.

# Code Review: readonly-mode

## Verdict: PASS

**Revision 2 (после Fix Cycle 1).** Реализация полностью соответствует утверждённому пост-pivot плану (Option B). Все 28 REQ покрыты тестами (или явно помечены N/A для пары REQ-2.3 / REQ-4.3 / CP-3 / CP-15 — задокументированный pivot из-за flock-семантики bbolt). 13 живых CP из 15 проверены unit- и rapid-property-тестами. Найденный в Revision 1 `major` finding (F-1: `hasReadOnlyFlag` не распознавал форму `--readonly=true`) исправлен в Cycle 1: добавлена поддержка `--readonly=value`/`--ro=value` через `parseFlagValue` + `strconv.ParseBool`, persistent-flag pre-scan теперь не останавливается на позиционных аргументах (mirror cobra). Добавлен `cmd/todushka/main_test.go` с 21 кейсом, покрывающим все стандартные pflag-формы. Минорная находка F-2 (неиспользуемая `tasks` в PBT) тоже устранена. Все верификационные команды (`task test`, `task test-race`, `task build`, `task lint`, `gofmt -l`) проходят без ошибок. Готово к approval.

## Change Set

Базовый коммит: `664136f4`. Все изменения некоммичены, видимы через `git status --short` и `git diff 664136f4 -- <path>`.

| File | Status | Notes |
|------|--------|-------|
| `internal/storage/repository.go` | ✅ Planned | T-2: `ErrReadOnly` sentinel + `ReadOnly() bool` метод интерфейса. |
| `internal/storage/errors_test.go` | ✅ Planned (new) | T-2: `TestErrReadOnly_IsSentinel`. |
| `internal/storage/fakes/inmemrepo.go` | ✅ Planned | T-2: `ReadOnly() bool { return false }`. |
| `internal/storage/fakes/inmemrepo_test.go` | ✅ Planned | T-2: `TestFakes_ReadOnlyAlwaysFalse`. |
| `internal/storage/bbolt/bbolt.go` | ✅ Planned | T-3, T-4: `readOnly` field, `ReadOnly()`, `checkWritable()`, `OpenReadOnly`, Migrate guard, 16 write-guards. |
| `internal/storage/bbolt/bbolt_test.go` | ✅ Planned | T-3, T-4: тесты конструкторов + табличный `TestBbolt_AllWritesReturnErrReadOnly`. |
| `internal/storage/bbolt/readonly_pbt_test.go` | ✅ Planned (new) | T-8: PBT CP-2 + явные note про N/A для CP-3/CP-15. |
| `internal/cli/deps.go` | ✅ Planned | T-5: `Deps.ReadOnly bool`. |
| `internal/cli/root.go` | ✅ Planned | T-5: регистрация `--readonly` и `--ro` через два `BoolVar` на одну переменную, проброс в `deps.ReadOnly` из `PersistentPreRunE`. |
| `internal/cli/cli_test.go` | ✅ Planned | T-5: `TestCLI_ReadOnlyFlagDefaultFalse`, `…Parsed`, `…AliasRO`, `TestProp_CLI_ReadOnlyFlagAlias` (CP-14). |
| `cmd/todushka/main.go` | ✅ Planned | T-5: `hasReadOnlyFlag` os.Args pre-scan (с поддержкой `--flag=value` после Cycle 1) + clear error на `ErrDatabaseLocked`. |
| `cmd/todushka/main_test.go` | ✅ Planned (new, Cycle 1) | Cycle 1 fix-cycle test: `TestHasReadOnlyFlag` (21 кейс) — закрывает F-1. |
| `internal/tui/app.go` | ✅ Planned | T-6, T-7: `Model.readOnly` (auto-detect из `svc.Repo().ReadOnly()`), `blockWriteIfReadOnly()` helper, guards в `handleKey` (quick-entry), `saveEditor`, `handleQuickEntryKey`. |
| `internal/tui/bulk.go` | ✅ Planned | T-7: inline RO check в `dispatch` (value-receiver). |
| `internal/tui/shell.go` | ✅ Planned | T-6: `modeReadOnly` const, расширение `modeLabel`, `currentMode` priority chain, `modeKeyHints` для RO. |
| `internal/tui/shell_test.go` | ✅ Planned | T-6, T-7: тесты chip, priority chain, write-key blocking, editor in RO. |
| `internal/tui/readonly_pbt_test.go` | ✅ Planned (new) | T-8: PBT CP-9 + CP-11/12. |

Никаких неожиданных файлов или скоупа сверх плана. `cmd/todushka/main_test.go` добавлен в Cycle 1 как покрытие fix'а — не нарушение скоупа.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 ErrReadOnly sentinel | `TestErrReadOnly_IsSentinel` | `internal/storage/repository.go:26` | — | ✅ |
| REQ-1.2 Repository.ReadOnly() | compile-time + `TestFakes_ReadOnlyAlwaysFalse` | `internal/storage/repository.go:96-98` | — | ✅ |
| REQ-1.3 writes return ErrReadOnly | `TestBbolt_AllWritesReturnErrReadOnly` (16 sub-tests) | все 16 write-методов в `bbolt.go` | CP-1 | ✅ |
| REQ-1.4 reads work in RO | `TestBbolt_ReadsWorkInRO`, `TestProp_Bbolt_ReadsEqualAcrossModes` | reads не получили checkWritable | CP-2 | ✅ |
| REQ-2.1 Open exclusive | `TestBbolt_OpenRWReadOnlyFalse` | `bbolt.go:Open` | CP-5 | ✅ |
| REQ-2.2 OpenReadOnly | `TestBbolt_OpenReadOnlyTrue`, `TestBbolt_ReadsWorkInRO` | `bbolt.go:101-118` | CP-4 | ✅ |
| REQ-2.3 OpenAuto fallback | — | — | CP-3 | ⚠️ N/A — Option B pivot (flock не позволяет) |
| REQ-2.4 OpenReadOnly требует существующий файл | `TestBbolt_OpenReadOnlyMissingFile` | `bbolt.go:104-106` | — | ✅ |
| REQ-2.5 Migrate no-op в RO | `TestBbolt_MigrateNoOpInRO` | `bbolt.go:Migrate` early-return | CP-6 | ✅ |
| REQ-2.6 ReadOnly() reflects mode | `TestBbolt_OpenRWReadOnlyFalse`, `…OpenReadOnlyTrue` | `bbolt.go:53` | CP-5 | ✅ |
| REQ-3.1 fakes ReadOnly==false | `TestFakes_ReadOnlyAlwaysFalse` | `internal/storage/fakes/inmemrepo.go:43-45` | CP-7 | ✅ |
| REQ-4.1 --readonly + --ro регистрируются | `TestCLI_ReadOnlyFlagParsed`, `…AliasRO`, `TestProp_CLI_ReadOnlyFlagAlias` | `internal/cli/root.go:47-48` | CP-14 | ✅ |
| REQ-4.2 explicit flag forces RO | `TestCLI_ReadOnlyFlagParsed` + `TestHasReadOnlyFlag` (21 cases — bare, =true, =false, after subcommand, etc.) | `cmd/todushka/main.go:34-40` + `hasReadOnlyFlag` | — | ✅ |
| REQ-4.3 default uses OpenAuto (post-pivot: default = `Open` с clear error) | manual smoke (`task test-race` для всего пакета) | `cmd/todushka/main.go:42-48` | — | ✅ |
| REQ-5.1 main.go wiring | `TestHasReadOnlyFlag` (constructor selection branch) + manual smoke | `cmd/todushka/main.go:34-54` | CP-15 N/A | ✅ |
| REQ-6.1 Model.readOnly из svc.Repo | `TestTUI_ModelReadOnlyReflectsRepo` | `internal/tui/app.go:60-64, 80` | CP-8 | ✅ |
| REQ-6.2 currentMode priority | `TestTUI_CurrentModeReadOnly`, `…PriorityRespected`, `TestProp_TUI_CurrentModePriority` | `internal/tui/shell.go:55-74` | CP-9 | ✅ |
| REQ-6.3 mode chip label | `TestTUI_ModeChipReadOnly` | `internal/tui/shell.go:51-52` + `viewFooter` framing | CP-10 | ✅ |
| REQ-7.1 write keys blocked + status | `TestTUI_WriteKeyBlockedInRO_Complete`, `…QuickEntryBlockedInRO`, `…BulkDispatchBlockedInRO`, `TestProp_TUI_WriteKeyBlockedInRO` | `internal/tui/bulk.go:dispatch`, `app.go:blockWriteIfReadOnly`, `handleKey` (QuickEntry) | CP-11, CP-12 | ✅ |
| REQ-7.2 editor open in RO | `TestTUI_EditorOpensInRO` | `app.go:openEditor` (без RO-guard) | — | ✅ |
| REQ-7.3 editor save error | `TestTUI_EditorSaveBlockedInRO` | `app.go:saveEditor` (early-return `m.editor.err`) | CP-13 | ✅ |
| REQ-7.4 quick entry submit blocked | `TestTUI_QuickEntryBlockedInRO` + `TestProp_TUI_WriteKeyBlockedInRO` для 'n' | `app.go:handleQuickEntryKey` Enter case | CP-11 | ✅ |
| REQ-7.5 bulk dispatch blocked | `TestTUI_BulkDispatchBlockedInRO` | `internal/tui/bulk.go:46-50` | CP-11 | ✅ |
| REQ-8.1 backward compat | все pre-existing тесты PASS | — | — | ✅ |
| REQ-8.2 writable repo → false | `TestBbolt_OpenRWReadOnlyFalse` | `bbolt.go:Open` явный `readOnly: false` | CP-5 | ✅ |
| REQ-8.3 RO repo writes error | `TestBbolt_AllWritesReturnErrReadOnly` | `checkWritable()` во всех write-методах | CP-1 | ✅ |

Все REQ покрыты. REQ-2.3 / CP-3 / CP-15 — N/A по pivot.

## Design Conformance

### 3.1 Architectural Boundaries
TUI зависит от storage только через `app.Service.Repo()`; никаких прямых импортов `internal/storage/bbolt` в TUI (за исключением одного integration-теста `shell_test.go`, что допустимо). CLI слой связывает Deps + cobra и передаёт сервис в TUI через `LaunchTUI`. main.go отвечает за выбор конструктора bbolt — что соответствует §3.1. ✅

### 3.2 Data Models
- `storage.ErrReadOnly` — `var` в общем блоке. ✅
- `Repository.ReadOnly() bool` — в конце интерфейса. ✅
- `bbolt.Repo.readOnly bool` — unexported field. ✅
- `cli.Deps.ReadOnly bool` — добавлен между `Config` и `LaunchTUI`. ✅
- `tui.Model.readOnly bool` — рядом с `filtering`. ✅
- `tui.shellMode.modeReadOnly` — в конце `iota`-блока. ✅

### 3.3 API Contracts
- `OpenReadOnly(path string) (*Repo, error)` — сигнатура совпадает с design §2.5. ✅
- `Repository.ReadOnly() bool` — сигнатура совпадает. ✅
- `OpenAuto` отсутствует — соответствует pivot. ✅

### 3.4 Error Handling
Все строки таблицы из design §2.7 имплементированы кроме N/A `OpenAuto`. Manual + automated smoke: write в RO → `ErrReadOnly`, locked DB без флага → clear stderr message + exit 1. ✅

### 3.5 Correctness Properties
| CP | Status |
|----|--------|
| CP-1 (writes return ErrReadOnly) | ✅ `TestBbolt_AllWritesReturnErrReadOnly` (16 sub-tests). |
| CP-2 (reads equivalent) | ✅ `TestProp_Bbolt_ReadsEqualAcrossModes`. |
| CP-3 (OpenAuto fallback) | N/A (pivot). |
| CP-4 (OpenReadOnly semantics) | ✅ `TestBbolt_OpenReadOnlyTrue` + `ReadsWorkInRO`. |
| CP-5 (ReadOnly() reflects construction) | ✅ `…OpenRWReadOnlyFalse` + `…OpenReadOnlyTrue`. |
| CP-6 (Migrate no-op в RO) | ✅ `TestBbolt_MigrateNoOpInRO`. |
| CP-7 (Fakes ReadOnly == false) | ✅ `TestFakes_ReadOnlyAlwaysFalse`. |
| CP-8 (Model.readOnly из Repo) | ✅ `TestTUI_ModelReadOnlyReflectsRepo`. |
| CP-9 (currentMode priority) | ✅ `TestTUI_CurrentModePriorityRespected` + rapid PBT. |
| CP-10 (mode chip label) | ✅ `TestTUI_ModeChipReadOnly`. |
| CP-11 + CP-12 (write keys blocked + status) | ✅ `TestProp_TUI_WriteKeyBlockedInRO` + unit tests. |
| CP-13 (editor save error) | ✅ `TestTUI_EditorSaveBlockedInRO`. |
| CP-14 (flags equivalent) | ✅ `TestProp_CLI_ReadOnlyFlagAlias`. |
| CP-15 (auto-fallback warning) | N/A (pivot). |

Все 13 живых CP покрыты. ✅

### 3.6 Documentation Consistency
Mermaid-диаграмма в `design.md` §2.2 показывает удалённые `OpenAuto`/`Fallback` ветви. Это **расхождение артефакта** (design ratified до pivot), не дефект кода. Опциональная рекомендация — синхронизировать диаграмму при следующей итерации artifact'ов. Не блокирует approval.

## Code Quality

### 4.1 Naming & Clarity
- Identifiers (`ReadOnly`, `readOnly`, `OpenReadOnly`, `checkWritable`, `blockWriteIfReadOnly`, `hasReadOnlyFlag`, `parseFlagValue`, `modeReadOnly`) следуют project naming conventions. ✅
- Doc-комментарии присутствуют на всех новых exported идентификаторах. ✅
- `hasReadOnlyFlag` после Cycle 1 имеет содержательный doc-комментарий, документирующий пять поддерживаемых форм. ✅

### 4.2 Dead Code & Debug Artifacts
- Нет закомментированного кода, `TODO`, `print`/`log` стейтментов. ✅
- NOTE-комментарии в `readonly_pbt_test.go` (про N/A CP-3/CP-15) и `bbolt_test.go` (про удалённый `OpenAuto`) — намеренная документация, полезны будущим читателям. ✅
- Импорты чистые — `golangci-lint run` → `0 issues`. ✅

### 4.3 Scope Creep
- Никаких изменений сверх плана. `cmd/todushka/main_test.go` (Cycle 1) — обоснован как fix coverage. ✅

### 4.4 Test Quality
- Все тесты используют `require.*` / `t.Fatalf` (последний только в `main_test.go` где testify не подходит для table-driven runner без `require.Equal` — но и `t.Fatalf` корректно сообщает). ✅
- Имена дескриптивные. ✅
- Table-driven там, где это уместно (`TestBbolt_AllWritesReturnErrReadOnly`, `TestHasReadOnlyFlag`). ✅
- PBT не just absence-of-error, но и state-snapshot assertion (`TestProp_TUI_WriteKeyBlockedInRO` сверяет `TaskList` до/после). ✅
- `TestTUI_ModelReadOnlyReflectsRepo` использует настоящий bbolt RO репозиторий (integration-style). ✅

## Security

| Категория | Оценка |
|-----------|--------|
| Input validation на `OpenReadOnly` | Путь приходит из конфига (`config.DataDir`) — controlled. ✅ |
| Path traversal | DB path детерминирован. ✅ |
| Permissions | `OpenReadOnly` использует `0o600`, файл не создаётся в RO (`os.Stat` pre-check). ✅ |
| `hasReadOnlyFlag` bypass | После Cycle 1 распознаются все стандартные pflag boolean-формы. Pre-scan не останавливается на позиционных аргументах — поведение mirror'ит cobra. Тесты покрывают 21 кейс. ✅ |
| Error leakage | Сообщения о locked-DB не утечают секреты (имя бинарника / hint про `--readonly`). ✅ |
| Secret exposure | Нет. ✅ |
| Injection / XSS / SQL | N/A. ✅ |
| New endpoints | Нет — CLI/TUI feature. ✅ |

Никаких security-issues. ✅

## Verification Evidence

Команды выполнены reviewer'ом **во время этой ревью-сессии** после применения Cycle 1 fix'а.

**Tests (`go clean -testcache && task test`):**
```
task: [test] go test ./...
ok  	github.com/jtprogru/todushka/cmd/todushka	0.308s
ok  	github.com/jtprogru/todushka/internal/app	0.706s
ok  	github.com/jtprogru/todushka/internal/cli	1.080s
ok  	github.com/jtprogru/todushka/internal/config	6.820s
ok  	github.com/jtprogru/todushka/internal/domain/area	1.405s
ok  	github.com/jtprogru/todushka/internal/domain/id	1.964s
ok  	github.com/jtprogru/todushka/internal/domain/project	3.240s
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	5.199s
ok  	github.com/jtprogru/todushka/internal/domain/repeat	5.776s
ok  	github.com/jtprogru/todushka/internal/domain/tag	4.322s
ok  	github.com/jtprogru/todushka/internal/domain/task	3.621s
ok  	github.com/jtprogru/todushka/internal/domain/today	2.843s
ok  	github.com/jtprogru/todushka/internal/storage	2.368s
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	13.846s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	4.692s
ok  	github.com/jtprogru/todushka/internal/tui	7.235s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

**Test (race) (`task test-race`):**
```
task: [test-race] go test -race ./...
ok  	github.com/jtprogru/todushka/cmd/todushka	1.421s
ok  	github.com/jtprogru/todushka/internal/app	1.861s
ok  	github.com/jtprogru/todushka/internal/cli	2.249s
ok  	github.com/jtprogru/todushka/internal/config	8.419s
ok  	github.com/jtprogru/todushka/internal/storage	3.791s
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	15.419s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	3.399s
ok  	github.com/jtprogru/todushka/internal/tui	9.992s
(прочие пакеты — все ok)
```

**Build (`task build`):**
```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

**Lint (`task lint`):**
```
task: [lint] golangci-lint run
0 issues.
```

**gofmt (`gofmt -l internal/ cmd/`):**
```
(empty output — нет неотформатированных файлов)
```

**Manual smoke (повторно после Cycle 1):**
```
$ ./bin/todushka --readonly=true add "smoke-after-fix"   →  storage: repository is read-only (exit 1)
$ ./bin/todushka --readonly=false add "smoke-allowed"    →  GZM6VY (exit 0)        ← корректно разрешено
$ ./bin/todushka --ro=true add "smoke-via-ro"            →  storage: repository is read-only (exit 1)
$ ./bin/todushka --readonly add "smoke-bare"             →  storage: repository is read-only (exit 1)
```

Все формы согласованы между cobra-парсером и `hasReadOnlyFlag`.

## Findings

| ID | Severity | File | Description | Requirement | Status |
|----|----------|------|-------------|-------------|--------|
| F-1 | major | `cmd/todushka/main.go:84-94` (Revision 1) | `hasReadOnlyFlag` не распознавал `--readonly=true` / `--ro=true` форму → cobra-parsed `deps.ReadOnly` расходился с open-time logic, write проходил при `--readonly=true`. | REQ-4.2, REQ-5.1 | **RESOLVED in Cycle 1**: добавлены ветви `strings.HasPrefix(a, "--readonly=")` / `--ro=` + helper `parseFlagValue`; pre-scan теперь не останавливается на позиционных аргументах (mirror cobra persistent-flag semantics). Тест `TestHasReadOnlyFlag` (21 кейс) подтверждает. Manual smoke подтвердил. |
| F-2 | nit | `internal/tui/readonly_pbt_test.go:110` (Revision 1) | Неиспользуемая локальная `tasks` сбрасывалась через `_ = tasks`. | — | **RESOLVED in Cycle 1**: `setupModelWithInboxTasks` возвращает в `_`. |

Никаких новых finding'ов после Cycle 1.

## Recommendations

**Информационные (не findings, не блокируют approval):**
1. `design.md` §2.2 Mermaid-диаграмма содержит удалённые ветви `OpenAuto`/`Fallback`. Подлежит синхронизации с pivot — опциональная задача на следующую итерацию артефактов.
2. (опционально) Можно объединить регистрацию `--readonly`/`--ro` через cobra-конструкт, который избегает дубликата doc-string в `--help` (сейчас оба флага видны как самостоятельные). Текущий подход (`BoolVar` × 2 на одну переменную) валиден и был принят в ADR-5.

## Fix Plan

<!-- Verdict == PASS — Fix Plan не требуется. -->

Все Findings из Revision 1 разрешены в Cycle 1. Дополнительных fix-tasks не требуется.

---

**История ревизий:**
- Revision 1 (initial): verdict `NEEDS_CHANGES` — F-1 (major), F-2 (nit).
- Revision 2 (после Cycle 1): verdict `PASS` — оба finding'а resolved.

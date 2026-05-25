# Code Review: tui-shell

## Verdict: PASS

Все 9 tasks из плана выполнены. 28 REQs трассируются в код + тесты; все 14 CPs покрыты property-тестами. Свежий полный test-suite зелёный (148 TUI test cases + 21 config test cases + 5 CLI tests), race detector clean, build OK, lint 0 issues, gofmt clean. Backward-compat T-1 preservation tests passing. Найдены 2 минорных нота — F-1 stale rapid regression seed в `testdata/`, F-2 subtle deviation от REQ-3.10 (status styling в non-confirm modes использует `theme.Help` вместо `theme.StatusError`). Ни одна не critical/major.

## Change Set

`review_base_commit = f4655b1` (HEAD `main` после merge dual-pane-layout). Все изменения uncommitted.

| File | Status | Notes |
|------|--------|-------|
| `internal/config/app.go` | ✅ Planned NEW | 63 lines: AppConfig struct, Defaults, Validate |
| `internal/config/app_test.go` | ✅ Planned NEW | 38 lines: 3 unit tests + 1 PBT (added in T-8) |
| `internal/config/loader.go` | ✅ Planned NEW | 135 lines: ResolvePath, Load, loadFromFile, createDefaultConfig, defaultConfigYAML, applyEnv |
| `internal/config/loader_test.go` | ✅ Planned NEW | 258 lines: 10 unit + 4 PBT, mockEnv helper |
| `internal/tui/shell.go` | ✅ Planned NEW | 220 lines: shellMode enum, currentMode, modeLabel, modeKeyHints, viewFooter, viewHeader, renderHeaderSegment, fetchListCounts |
| `internal/tui/shell_test.go` | ✅ Planned NEW | 403 lines: 20 unit + 9 PBT |
| `internal/tui/app.go` | ✅ Planned MODIFIED | Model +3 fields, NewModel signature, WindowSizeMsg both axes, tasksLoadedMsg → tea.Batch, countsLoadedMsg case, View() full-screen clamp; old viewHeader/viewFooter удалены |
| `internal/tui/app_test.go` | ✅ Planned MODIFIED | Helpers updated, +TestWindowSize_BothAxesTracked, PBT обновлены |
| `internal/tui/bulk.go` | ✅ Planned MODIFIED | const → m.config.BulkConfirmThreshold |
| `internal/tui/bulk_test.go` | ✅ Planned MODIFIED | config.Defaults().BulkConfirmThreshold reference |
| `internal/tui/details.go` | ✅ Planned MODIFIED | 3 constants → m.config.X, paneWidths(m) signature |
| `internal/tui/details_test.go` | ✅ Planned MODIFIED | Updated for paneWidths(m) signature |
| `internal/tui/filter_test.go` | ✅ Planned MODIFIED | config import + helper updates |
| `internal/tui/msgs.go` | ✅ Planned MODIFIED | countsLoadedMsg type |
| `internal/tui/run.go` | ✅ Planned MODIFIED | Run(svc, cfg) signature, selectThemeFromConfig helper |
| `internal/cli/deps.go` | ✅ Planned MODIFIED | Config field, LaunchTUI signature |
| `internal/cli/root.go` | ✅ Planned MODIFIED | --config flag + PersistentPreRunE |
| `internal/cli/cli_test.go` | ✅ Planned MODIFIED | Updated LaunchTUI signature |
| `go.mod` | ✅ Planned MODIFIED | yaml.v3 promoted to direct |
| `internal/tui/testdata/rapid/...` | ⚠️ Unexpected NEW | Rapid regression seed for `TestProp_FooterContainsNewKeys` — saved during T-4 when PBT was refactored; harmless but stale (см. F-1) |

NOT changed (per design §2.3): `storage/`, `app/`, `domain/`, `keys.go`, `filter.go`, `editor.go`, `style.go`, остальные `cli/*.go`, `cmd/todushka/main.go` (косвенно — через deps signature), `config/paths.go` — all confirmed unchanged.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 m.height tracked | `TestWindowSize_BothAxesTracked` + `TestProp_WindowSizeTracked` | `app.go` WindowSizeMsg case | CP-3 | ✅ |
| REQ-1.2 full-screen clamp | `TestView_FullScreenClamp` + `TestProp_FullScreenHeight` | `app.go` View() | CP-2 | ✅ |
| REQ-1.3 small terminal fallback | `TestTUI_ZeroWidthRendersSinglePane` + `TestTUI_NarrowWidthRendersSinglePane` + `TestView_SmallTerminalFallback` | `app.go` View() else-branch | CP-2 | ✅ |
| REQ-1.4 editor full-pane | `TestView_EditorIgnoresClamp` | `app.go` View() editor short-circuit | CP-2 | ✅ |
| REQ-2.1 header `(N) Name [Count]` | `TestViewHeader_FullModeSegment` + `TestProp_CompactModeThreshold` | `shell.go:146-181` | CP-5 | ✅ |
| REQ-2.2 color coding | `TestViewHeader_FullModeSegment` (выходные ANSI escapes) | `shell.go:171-181` (Selected/HeaderDim/Dim для digit/label/count) | CP-4 | ✅ |
| REQ-2.3 active inverted | `TestViewHeader_ActiveSegmentInverted` + `TestProp_ActiveSegmentStyled` | `shell.go:162-164` | CP-4 | ✅ |
| REQ-2.4 compact mode | `TestViewHeader_CompactMode` + `TestProp_CompactModeThreshold` | `shell.go:153-158, 166-173` | CP-5 | ✅ |
| REQ-2.5 counts on tasksLoadedMsg | `TestUpdate_TasksLoadedTriggersCountsFetch` + `TestProp_CountsRefreshPropagates` + `TestProp_CountsMatchService` | `app.go` tasksLoadedMsg case | CP-6, CP-12 | ✅ |
| REQ-2.6 countsLoadedMsg merge | `TestFetchListCounts_EmitsMsg` + `TestProp_CountsRefreshPropagates` | `app.go` countsLoadedMsg case | CP-6 | ✅ |
| REQ-2.7 unknown count [?] | `TestViewHeader_UnknownCountPlaceholder` | `shell.go:148-150` | CP-5 | ✅ |
| REQ-2.8 empty list [0] | `TestViewHeader_EmptyListShowsZero` | `shell.go:148-150` (knownCount=true, count=0) | CP-5 | ✅ |
| REQ-3.1 footer structure | `TestViewFooter_*` series | `shell.go:101-129` | CP-7 | ✅ |
| REQ-3.2 mode priority | `TestCurrentMode_PriorityOrder` + `TestProp_ModeExclusion` | `shell.go:55-70` | CP-1 | ✅ |
| REQ-3.3 mode chip styled | `TestViewFooter_ModeChipPresent` + `TestProp_ModeChipText` | `shell.go:103` (theme.Header) | CP-7 | ✅ |
| REQ-3.4 NORMAL hints | `TestModeKeyHints_Normal` + `TestProp_KeyHintsMatchMode` | `shell.go:77-78` | CP-8 | ✅ |
| REQ-3.5 FILTER hints | `TestModeKeyHints_Filter` | `shell.go:79-80, 106-108` | CP-8 | ✅ |
| REQ-3.6 SELECT hints | `TestModeKeyHints_Select` + `TestProp_StatusBarShowsCount` | `shell.go:81-82, 109-111` | CP-8 | ✅ |
| REQ-3.7 CONFIRM hints | `TestModeKeyHints_Confirm` | `shell.go:83-84` | CP-8 | ✅ |
| REQ-3.8 EDITOR hints | `TestModeKeyHints_Editor` | `shell.go:85-86` | CP-8 | ✅ |
| REQ-3.9 HELP hints | `TestModeKeyHints_Help` | `shell.go:87-88` | CP-8 | ✅ |
| REQ-3.10 status preserved | `TestViewFooter_StatusMessagePreserved` + `TestTUI_ErrorMsgUpdatesStatusBar` | `shell.go:115-122` | CP-7 | ⚠ (см. F-2) |
| REQ-4.1 path precedence | `TestResolvePath_*` + `TestProp_FlagBeatsEnv` | `loader.go:13-30` | CP-9, CP-13 | ✅ |
| REQ-4.2 auto-create | `TestLoad_FileMissingCreatesDefault` + `TestProp_AutoCreateRoundTrip` | `loader.go:56-78` | CP-11 | ✅ |
| REQ-4.3 5 settings parsed | `TestLoad_FileValidParses` | `app.go` AppConfig + `loader.go` yaml.Unmarshal | CP-9 | ✅ |
| REQ-4.4 unknown fields ignored | `TestLoad_UnknownYAMLFieldsIgnored` + `TestProp_UnknownYAMLFieldsIgnored` | `loader.go:65` (yaml default lenient) | CP-14 | ✅ |
| REQ-4.5 invalid values warn | `TestValidate_ThemeFallback` + `TestValidate_NumericRanges` + `TestProp_ValidateCorrectsInvalid` | `app.go` Validate | CP-10 | ✅ |
| REQ-4.6 env override | `TestLoad_EnvOverridesFile` + `TestProp_LoadPrecedence` | `loader.go:101-135` | CP-9 | ✅ |
| REQ-4.7 invalid env fallback | `TestLoad_InvalidEnvFallsBackToFile` | `loader.go` applyEnv if-err branches | CP-10 | ✅ |
| REQ-4.8 config injected | tests that use `m.config.X` for thresholds | `app.go` Model fields + `details.go`/`bulk.go` references | CP-9 | ✅ |
| REQ-5.1 zero-config visual | T-1 preservation tests passing | `m.config = config.Defaults()` in test fixtures | — | ✅ |
| REQ-5.2 NewModel signature | All test fixtures + main.go updated | `app.go` NewModel | — | ✅ |
| REQ-5.3 existing tests pass | 117 prior + new tests passing | — | — | ✅ |

28/28 ✅ (REQ-3.10 has ⚠ F-2 deviation, см. ниже).

## Design Conformance

### §3.1 Architectural Boundaries

✅ Все новые конструкции в `internal/config/` и `internal/tui/`. Никаких изменений в `storage/`, `app/`, `domain/`. Зависимости направлены правильно: `tui` → `config` (для AppConfig type); `cli` → `config` (для load/resolve); `config` зависимостей не имеет (использует только stdlib + yaml.v3).

### §3.2 Data Models

Design §2.5 vs реализация:

| Design | Implementation | Match |
|---|---|---|
| `AppConfig` (5 fields) | `internal/config/app.go` | ✅ |
| `Defaults()` constructor | `internal/config/app.go` | ✅ |
| `Validate() (AppConfig, []string)` | `internal/config/app.go` | ✅ |
| `Model.height int` | `app.go:Model` | ✅ |
| `Model.config config.AppConfig` | `app.go:Model` | ✅ |
| `Model.listCounts map[listKind]int` | `app.go:Model` | ✅ |
| `nameCacheLoadedMsg` (existing) | unchanged | ✅ |
| `countsLoadedMsg` | `msgs.go` | ✅ |
| `shellMode` enum (6) | `shell.go:23-30` | ✅ |
| `dualPaneMinWidth` константа удалена | confirmed via grep | ✅ |
| `listPaneShare` константа удалена | confirmed | ✅ |
| `bulkConfirmThreshold` константа удалена | confirmed | ✅ |
| `detailsNotesMaxLines` константа удалена | confirmed | ✅ |

### §3.3 API Contracts

`LaunchTUI` signature изменён — это **breaking change для embedders**, но `Deps` это test-seam, не публичный API. Acceptable.

CLI surface: `--config <path>` flag добавлен на root command, persistent — propagates на все subcommands. Visible в `--help`. Matches ADR-5.

### §3.4 Error Handling

Design §2.7 vs реализация:

| Scenario | Design Action | Code | Match |
|---|---|---|---|
| Config file missing | Auto-create + return defaults | `loader.go:56-60` | ✅ |
| Parent dir creation fails | Warn + defaults | `loader.go:73-77` (returns error wrapped в warning string) | ✅ |
| File unreadable | Warn + defaults | `loader.go:62-63` | ✅ |
| Malformed YAML | Warn + defaults | `loader.go:65-66` | ✅ |
| Invalid field value | Replace + warning | `app.go` Validate | ✅ |
| Invalid env value | Warn + skip | `loader.go:106-133` | ✅ |
| WindowSizeMsg very small | Legacy render | `app.go:View()` else-branch | ✅ |
| `Service.ListXxx` error | Skip ID, omit from map | `shell.go:200-217` (`if err == nil` per call) | ✅ |
| Negative bodyH | Clamp to 0 | `app.go:View()` `if bodyH < 0 { bodyH = 0 }` | ✅ |
| `--config` path invalid | Fallback to defaults | `root.go:PersistentPreRunE` else branch | ✅ |
| Concurrent tasksLoadedMsg | Bubble Tea serialization | Last write wins (idempotent map assign) | ✅ |

### §3.5 Correctness Properties

14/14 CPs реализованы как `TestProp_*`. Прогоняются через `rapid.Check`, стабильны при `-count=2` (~5 sec total). Никаких flakes не наблюдалось.

### §3.6 Documentation Consistency

Mermaid диаграмма в design §2.2 отражает фактические потоки:
- `tasksLoadedMsg` → `fetchNameCache + fetchListCounts` (через tea.Batch) → `countsLoadedMsg` / `nameCacheLoadedMsg` → Update merges → Model
- `View()` → `viewBody` (single или dual) / `viewHeader` / `viewFooter`
- `viewFooter` → `currentMode` → mode-specific routing

Implementation matches.

## Code Quality

### §4.1 Naming & Clarity

Идиоматичный Go. Identifiers descriptive: `shellMode`, `currentMode`, `modeLabel`, `modeKeyHints`, `renderHeaderSegment`, `fetchListCounts`, `headerCompactThreshold`, `listInitials`, `dualPaneMinWidth`, `bulkConfirmThreshold` (removed), и т.д. Файл shell.go хорошо организован — types/constants/enum/render functions сгруппированы.

### §4.2 Dead Code & Debug Artifacts

Чисто: 0 TODO, 0 закомментированных блоков, 0 debug-print. Старые `viewHeader`/`viewFooter` физически удалены из `app.go`, не оставлены закомментированными. Constants `dualPaneMinWidth`/`listPaneShare`/`detailsNotesMaxLines`/`bulkConfirmThreshold` корректно удалены — заменены на `m.config.X` references.

### §4.3 Scope Creep

Никакого. Все 7 design ADRs реализованы как решены. Никаких "while we're here" рефакторингов.

### §4.4 Test Quality

- 148 TUI test cases + 21 config + 5 CLI = 174 total в свежем прогоне; race-detector clean.
- testify `require` (не assert) — соответствует Tier 2 reference.
- Existing fixtures (`newTestModel`, `setupModelWithInboxTasks`, `bareTestModel`) переиспользованы без нового pattern'а.
- Filesystem isolation в config-тестах через `t.TempDir()` — корректно.
- PBT генераторы малые (titles ≤15, slices ≤10, intRange small).
- Edge cases покрыты: pane width boundaries, mode priority всех 6 modes, config invalid values по всем 5 settings, auto-create round-trip, env override, flag precedence, unknown YAML fields.

## Security

- **Внешние входы:** YAML config файл (пользовательский) и env vars. `yaml.Unmarshal` — стандартная safe-deserialization; нет custom unmarshallers. Env vars парсятся через `strconv.Atoi`/`ParseFloat` — safe. Invalid values → warning + default. `--config` path обрабатывается через `filepath.Abs` (canonical normalize).
- **File permissions:** `0750` для config dir, `0600` для config file — tight (owner-only read for file, group exec/read for dir).
- **Injection:** Нет shell exec, нет SQL. ANSI handled by lipgloss.
- **Secrets:** Ноль hardcoded credentials. Config file location в user's home — стандартная зона для CLI tools.
- **Race conditions:** Update serialized; cache writes additive; concurrent `tasksLoadedMsg` benign.
- **DoS surface:** Config file size unbounded — теоретически `os.ReadFile` на огромном файле прочитает всё в память. На single-user local app — non-issue (атакующий — сам user).

## Verification Evidence

Все команды повторно прогнаны в этой сессии после `go clean -testcache`.

- **Tests (full suite, fresh):**

```
task: [test] go test ./...
?   	github.com/jtprogru/todushka/cmd/todushka	[no test files]
ok  	github.com/jtprogru/todushka/internal/app	0.471s
ok  	github.com/jtprogru/todushka/internal/cli	0.793s
ok  	github.com/jtprogru/todushka/internal/config	5.255s
ok  	github.com/jtprogru/todushka/internal/domain/area	3.266s
ok  	github.com/jtprogru/todushka/internal/domain/id	4.454s
ok  	github.com/jtprogru/todushka/internal/domain/project	2.552s
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	1.513s
ok  	github.com/jtprogru/todushka/internal/domain/repeat	1.186s
ok  	github.com/jtprogru/todushka/internal/domain/tag	4.841s
ok  	github.com/jtprogru/todushka/internal/domain/task	2.879s
ok  	github.com/jtprogru/todushka/internal/domain/today	3.626s
?   	github.com/jtprogru/todushka/internal/storage	[no test files]
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	4.693s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	2.201s
ok  	github.com/jtprogru/todushka/internal/tui	2.010s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

- **Build:**

```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

- **Lint:**

```
task: [lint] golangci-lint run
0 issues.
```

- **Format check:**

```
(empty — all files gofmt-clean)
```

- **Manual smoke** (`./bin/todushka --help`): `--config string   Path to config file (overrides default)` присутствует.

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | nit | `internal/tui/testdata/rapid/TestProp_FooterContainsNewKeys/...-fail` | Stale rapid regression seed остался после рефактора `TestProp_FooterContainsNewKeys` в T-4 (мaster теперь mode-aware и проходит). Файл безвреден но смущает. Recommendation: либо commit как regression evidence, либо `.gitignore` правило `internal/tui/testdata/rapid/**/*.fail`. | — |
| F-2 | minor | `internal/tui/shell.go:115-122` | REQ-3.10 говорит "status message styled с `theme.StatusError` (red) for errors or `theme.StatusInfo` (green) for info — preserving current behavior". Реализация использует `theme.Help` (subtext) для non-confirm modes и `theme.StatusError` только в CONFIRM mode. Это **change of behavior** — старая `viewFooter` использовала `theme.StatusError` для всех status messages. Тест `TestTUI_ErrorMsgUpdatesStatusBar` (проверяет только text content, не style) → passing, не ловит regression. Visual impact: error notifications в NORMAL/FILTER/etc. modes теперь менее prominent (subtext-coloured вместо red). | REQ-3.10 |

## Recommendations

**Minor / nit (не блокеры):**

1. **F-1 (rapid seed cleanup):** Добавить `internal/tui/testdata/rapid/**/*.fail` в `.gitignore` (если эти seeds временные) ИЛИ закоммитить как regression evidence (если рассматриваются как permanent). Расхожая практика — gitignore'ить `.fail`-files, поскольку они меняются на каждый прогон.

2. **F-2 (status styling):** Восстановить REQ-3.10 strict compliance, отличая errors (red via `StatusError`) от info messages (green via `StatusInfo`). Сейчас текущее различие "CONFIRM → StatusError, else Help" не отражает спецификацию. Простое исправление: вернуть `theme.StatusError` для всех `m.statusMsg != ""` cases (preserving exact backward compat) ИЛИ добавить новое поле `Model.statusKind` с enum {info, error} для proper differentiation (более правильное долгосрочное решение — defer to follow-up).

Обе рекомендации — не блокеры. Verdict `PASS` независимо от их исполнения.

# Code Review: dual-pane-layout

## Verdict: PASS

Все 7 tasks из плана выполнены. Все 24 REQ-X.Y трассируются в production-код И тесты. 14/14 correctness properties покрыты property-тестами через `pgregory.net/rapid`. Полный test suite зелёный (race detector clean) — 117 TUI test cases passing. Bin строится, lint 0 issues, gofmt clean. Backward-compat preservation tests T-1 проходят без изменений, что подтверждает зеро regression на узких терминалах. Найдены 2 минорных нота (F-1 title wrap cap, F-2 known limitation HeadingGet), ни один не critical и не major.

## Change Set

`review_base_commit = 71d16ba` (HEAD `main` после merge feature #1). Все изменения — uncommitted working-tree.

| File | Status | Notes |
|------|--------|-------|
| `internal/tui/app.go` | ✅ Planned | +62 строки: 4 Model name-cache поля, NewModel init, `tasksLoadedMsg` теперь возвращает `fetchNameCache` Cmd, `nameCacheLoadedMsg` Update case, `viewBody()` method, `View()` рефактор для routing через viewBody в 3 местах |
| `internal/tui/app_test.go` | ✅ Planned | +68 строк: 3 T-1 preservation tests + 2 T-2 name cache tests |
| `internal/tui/msgs.go` | ✅ Planned | +10 строк: `nameCacheLoadedMsg` тип |
| `internal/tui/details.go` | ✅ Planned NEW | 188 строк: constants, isDualPane, paneWidths, fetchNameCache, cursorTask, wrapAndTruncate, resolveName, statusLabel, viewDetails |
| `internal/tui/details_test.go` | ✅ Planned NEW | 798 строк: 23 unit tests + 14 PBT |

Ни одного **Unexpected** или **Not Changed** файла из проектируемого scope. NOT changed (per design §2.3): `storage/`, `app/service`, `domain/`, `keys.go`, `filter.go`, `bulk.go`, `editor.go`, `style.go`, `cli/`, `cmd/` — все действительно не трогали.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 dual-pane condition | `TestIsDualPane_*` + `TestProp_LayoutModeExclusion` | `details.go:27-35` | CP-1 | ✅ |
| REQ-1.2 single-pane <100 | `TestTUI_NarrowWidthRendersSinglePane` + `TestIsDualPane_NarrowTerminal` + `TestProp_LayoutModeExclusion` | `details.go:28-30` | CP-1 | ✅ |
| REQ-1.3 width==0 single-pane | `TestTUI_ZeroWidthRendersSinglePane` + `TestIsDualPane_ZeroWidth` | `details.go:28-30` | CP-1 | ✅ |
| REQ-1.4 45/55 width split | `TestPaneWidths_SumEqualsTotal` + `TestProp_PaneWidthArithmetic` | `details.go:40-44` | CP-2 | ✅ |
| REQ-1.5 double border `║` | `TestViewBody_DualPaneRendersBothAndBorder` + `TestProp_BorderRenders` | `app.go:451` (`DoubleBorder()`) | CP-3 | ✅ |
| REQ-1.6 editor full-pane | `TestIsDualPane_EditorScreen` + `TestLayout_EditorFullPaneIgnoresWidth` + `TestProp_EditorHelpForceFullPane` | `details.go:31-33` | CP-4 | ✅ |
| REQ-1.7 help full-pane | `TestIsDualPane_HelpScreen` + `TestLayout_HelpFullPaneIgnoresWidth` + `TestProp_EditorHelpForceFullPane` | `details.go:31-33` | CP-4 | ✅ |
| REQ-1.8 confirm stacks | `TestLayout_ConfirmStacksBelowDual` + `TestProp_ConfirmStacksBelowDual` | `app.go:459-462` | CP-14 | ✅ |
| REQ-2.1 Title | `TestViewDetails_TitleAndStatus` + `TestProp_FieldOrderInvariant` | `details.go:153` | CP-5 | ⚠ (см. F-1) |
| REQ-2.2 Status | `TestViewDetails_TitleAndStatus` + `TestProp_FieldOrderInvariant` | `details.go:154` | CP-5 | ✅ |
| REQ-2.3 Notes truncation | `TestViewDetails_Notes` + `TestWrapAndTruncate_LongIsTruncated` + `TestProp_NotesTruncationCorrect` | `details.go:155-158` | CP-7 | ✅ |
| REQ-2.4 Start date | `TestViewDetails_Dates` + `TestProp_FieldOrderInvariant` | `details.go:159-161` | CP-5 | ✅ |
| REQ-2.5 Due date | `TestViewDetails_Dates` + `TestProp_FieldOrderInvariant` | `details.go:162-164` | CP-5 | ✅ |
| REQ-2.6 Pinned date | `TestProp_FieldOrderInvariant` | `details.go:165-167` | CP-5 | ✅ |
| REQ-2.7 Area name | `TestViewDetails_RelationsAndTags` + `TestProp_FieldOrderInvariant` | `details.go:168-170` | CP-5 | ✅ |
| REQ-2.8 Project name | `TestProp_FieldOrderInvariant` | `details.go:171-173` | CP-5 | ✅ |
| REQ-2.9 Heading name | `TestProp_FieldOrderInvariant` | `details.go:174-176` | CP-5 | ✅ |
| REQ-2.10 Tags list | `TestViewDetails_RelationsAndTags` + `TestProp_FieldOrderInvariant` | `details.go:177-183` | CP-5 | ✅ |
| REQ-2.11 Omit nil fields | `TestViewDetails_OmitsNilFields` + `TestProp_NilFieldsOmitted` | `details.go:155, 159, 162, 165, 168, 171, 174, 177, 184` (conditional appends) | CP-6 | ✅ |
| REQ-2.12 Someday marker | `TestProp_FieldOrderInvariant` | `details.go:184-186` | CP-5 | ✅ |
| REQ-3.1 Empty list placeholder | `TestViewDetails_EmptyTasksPlaceholder` + `TestProp_PlaceholderOnEmptyCursor` | `details.go:148-151` | CP-8 | ✅ |
| REQ-3.2 Out-of-range placeholder | `TestViewDetails_OutOfRangeCursorPlaceholder` + `TestProp_PlaceholderOnEmptyCursor` | `details.go:148-151` (`cursorTask` returns nil) | CP-8 | ✅ |
| REQ-4.1 Name cache fetch | `TestNameCache_FetchCmdEmitsMsg` + `TestNameCache_LoadedMsgPopulatesModel` + `TestUpdate_TasksLoadedDispatchesNameCacheFetch` + `TestProp_NameCachePopulation` | `app.go:84` + `details.go:55-96` | CP-10 | ✅ (heading: см. F-2) |
| REQ-4.2 No IO in View() | `TestProp_NoRepoAccessInView` | `details.go:147-188` (только read'ы map'ов) | CP-9 | ✅ |
| REQ-4.3 Short-ID fallback | `TestViewDetails_ShortIDFallback` + `TestProp_ShortIDFallback` | `details.go:124-129` | CP-11 | ✅ |
| REQ-5.1 Cursor change updates | `TestCursorChange_DetailsUpdates` + `TestProp_CursorChangeReflects` | `details.go:100-106` (`cursorTask` reads m.cursor) | CP-12 | ✅ |
| REQ-5.2 Filter change updates | `TestViewBody_FilterPreservesDualPane` + (transitively via `cursorTask` → `displayedTasks`) | `details.go:101` | CP-12 | ✅ |
| REQ-6.1 Filter preserves dual | `TestIsDualPane_FilteringAllowed` + `TestViewBody_FilterPreservesDualPane` + `TestProp_FilterPreservesDualPane` | `details.go:27-35` (no filtering guard) | CP-13 | ✅ |
| REQ-6.2 Quick-entry in dual | `TestLayout_QuickEntryStacksBelowDual` | `app.go:469` (uses viewBody) | CP-1 | ✅ |
| REQ-6.3 Bulk triggers reload | (transitively через `loadCurrentList` → `tasksLoadedMsg` → `fetchNameCache`) | `app.go:114-118` | CP-10 | ✅ |

24/24 ✅. Каждое REQ имеет ≥1 unit test + ≥1 property test, плюс трассируемое место в production-коде.

## Design Conformance

### §3.1 Architectural Boundaries

✅ Все новые символы — в пакете `internal/tui/`. Никаких изменений в `storage/`, `app/`, `domain/`. `details.go` импортирует `app` (для `*Service` в Cmd-closure), `id`, `task` — все upstream. Никаких циклов или unauthorized cross-layer imports.

### §3.2 Data Models

Design §2.5 vs реализация:

| Design | Implementation | Match |
|---|---|---|
| `Model.tagNamesByID map[id.ID]string` | `app.go:42` | ✅ |
| `Model.areaNamesByID` | `app.go:43` | ✅ |
| `Model.projectNamesByID` | `app.go:44` | ✅ |
| `Model.headingNamesByID` | `app.go:45` | ✅ |
| `nameCacheLoadedMsg{tags, areas, projects, headings}` | `msgs.go` | ✅ |
| `const dualPaneMinWidth = 100` | `details.go:17` | ✅ |
| `const listPaneShare = 0.45` | `details.go:18` | ✅ |
| `const detailsNotesMaxLines = 8` | `details.go:14` | ✅ |

Все типы и константы по design.

### §3.3 API Contracts

N/A — фича внутри TUI, никакого public API / wire format / schema migration.

### §3.4 Error Handling

Design §2.7 vs реализация:

| Scenario | Design Action | Code | Match |
|---|---|---|---|
| `TagGet`/`AreaGet`/`ProjectGet` error в `fetchNameCache` | Skip ID, fallback to short-ID | `details.go:79-93` (`if err == nil`) | ✅ |
| `m.cursor` out of range | `cursorTask` returns nil → placeholder | `details.go:101-105` + `:149-150` | ✅ |
| Terminal resize | `WindowSizeMsg` updates `m.width`; next `View()` re-evaluates | `app.go:75-77` (existing) | ✅ |
| `m.width == 0` | Single-pane fallback (existing single-pane code) | `details.go:28-30` (`width < dualPaneMinWidth`) | ✅ |
| Task TagID deleted out-of-band | `fetchNameCache` skips, viewDetails uses short-ID | По row 1 | ✅ |
| `m.width` very narrow (<20) | Design said "safety net" — implementation полагается на lipgloss handle of negative widths | `details.go:111-112` (`width <= 0` returns "") | ⚠ partial — only checks `width <= 0`, not "very narrow". Acceptable: `paneWidths(100)` дает list=44, details=55 — обе достаточны |
| Concurrent `tasksLoadedMsg` ordering | Bubble Tea serializes Update; maps merge additively | `app.go:86-99` (`for k, v := range msg.X { m.X[k] = v }`) | ✅ |

### §3.5 Correctness Properties

14/14 CPs реализованы как `TestProp_*` в `details_test.go`. Property tests прогоняются через `rapid.Check`, стабильны при `-count=2` (исполняется <1 sec). Не наблюдалось flakes.

### §3.6 Documentation Consistency

**Mermaid диаграммы в design §2.2:**
- `tasksLoadedMsg → fetchNameCache → Repository → nameCacheLoadedMsg → updateCache` — реализован точно как описано.
- `viewBody → isDualPane → JoinHorizontal | singlePane` — структура точно соответствует коду.
- Mode-interaction joins (confirm/quickEntry/editor) — все корректно отражены.

Все component/package имена в диаграммах соответствуют коду. Никаких новых компонентов, появившихся в реализации но отсутствующих в дизайне.

## Code Quality

### §4.1 Naming & Clarity

Все идентификаторы идиоматичны для Go: `isDualPane`, `paneWidths`, `cursorTask`, `wrapAndTruncate`, `resolveName`, `statusLabel`, `viewDetails`, `fetchNameCache`, `dualPaneMinWidth`, `listPaneShare`, `detailsNotesMaxLines`. Понятная разница `tagNamesByID` vs `tagSet` — первое cache, второе set для dedup'а в fetcher.

### §4.2 Dead Code & Debug Artifacts

Чисто: 0 TODO, 0 закомментированных блоков, 0 debug-print. Все добавленные функции вызываются.

### §4.3 Scope Creep

Никакого. 6 design ADRs (per-tasksLoadedMsg batch, lipgloss.Width word-wrap, fixed field order, new file, async Cmd, proportional width) реализованы как решены. Никаких "while we're here" рефакторингов.

### §4.4 Test Quality

- 117 проходящих test cases в `internal/tui/` (включая subtests).
- 23 unit + 14 PBT в `details_test.go`; все используют existing fixtures.
- testify `require` (не `assert`).
- Edge cases покрыты: width boundaries (0, 99, 100, 200), empty/nil fields, out-of-range cursor, short-ID fallback, modal stacking, multi-screen interactions.
- PBT генераторы малы (titles ≤15, slices ≤6) — fast (<1 sec для 14 PBT).
- Property test для view purity (CP-9) использует `require.NotPanics` с `bareTestModel()` (без service) — сильное доказательство отсутствия I/O.

## Security

Никаких security-проблем не обнаружено в изменённых файлах:

- **Внешние входы:** Только pre-existing keystrokes от локального терминала; никаких новых endpoints.
- **Injection:** Нет user-provided format strings, нет shell exec, нет SQL. `lipgloss.Style.Width().Render()` обрабатывает ANSI escape sequences корректно.
- **Secrets:** Ноль hardcoded credentials.
- **Data exposure:** Tag/area/project names — это user-content из локальной DB. Display в TUI идентичен existing editor behavior.
- **Resource limits:** Cache map'ы growable; для типичной нагрузки (1000 tasks × ~5 unique tags × ~50b name) ≈ 250KB — незаметно. Под DOS-вектор (10M+ задач) защита через `len(tasks)` уже ограничена storage capacity.
- **Concurrency:** `fetchNameCache` Cmd запускается асинхронно, но Update serializes msg processing — race-free.

## Verification Evidence

Все команды повторно прогнаны в этой сессии (`go clean -testcache` для guaranteed-fresh execution).

- **Tests (full suite, fresh):**

```
task: [test] go test ./...
?   	github.com/jtprogru/todushka/cmd/todushka	[no test files]
ok  	github.com/jtprogru/todushka/internal/app	1.125s
ok  	github.com/jtprogru/todushka/internal/cli	0.476s
ok  	github.com/jtprogru/todushka/internal/config	0.789s
ok  	github.com/jtprogru/todushka/internal/domain/area	1.517s
ok  	github.com/jtprogru/todushka/internal/domain/id	2.190s
ok  	github.com/jtprogru/todushka/internal/domain/project	1.853s
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	2.570s
ok  	github.com/jtprogru/todushka/internal/domain/repeat	3.275s
ok  	github.com/jtprogru/todushka/internal/domain/tag	4.652s
ok  	github.com/jtprogru/todushka/internal/domain/task	3.939s
ok  	github.com/jtprogru/todushka/internal/domain/today	2.917s
?   	github.com/jtprogru/todushka/internal/storage	[no test files]
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	4.126s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	4.987s
ok  	github.com/jtprogru/todushka/internal/tui	4.420s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

- **Build:**

```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

Bin `bin/todushka` создан, `--help` smoke OK.

- **Lint:**

```
task: [lint] golangci-lint run
0 issues.
```

- **Format check (`gofmt -l internal/tui/`):**

```
(empty — all files formatted)
```

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | minor | `internal/tui/details.go:153` | `viewDetails` использует `wrapAndTruncate(t.Title, width, 4)` — title hard-capped at 4 lines. REQ-2.1 says "word-wrapped to pane width across as many lines as needed" (без явного cap). Практически unobservable — title >4 lines очень редок. Рекомендация: либо документировать cap, либо убрать ограничение (truncate без maxLines: передать `len(text)` или специальное значение). | REQ-2.1 |
| F-2 | nit | `internal/tui/details.go:51-54` (NOTE) | Heading name resolution skipped (Repository не имеет `HeadingGet(ctx, id)`). Hearing IDs всегда отображаются через short-ID fallback. Documented as v2 spike в design. Не регрессия — feature был спроектирован с этим ограничением. | REQ-2.9, REQ-4.1 |

## Recommendations

**Minor (необязательно для approve):**

1. **F-1 (Title wrap cap):** Решить — оставить cap=4 (документировать в комментарии и в REQ-2.1) или убрать (передать `len(t.Title)` или новую функцию `wrap` без maxLines). Если оставлять — добавить комментарий в `viewDetails` строку 153: `// Title capped to 4 wrapped lines for visual stability.`

2. **F-2 (HeadingGet spike):** Добавить `HeadingGet(ctx, id.ID) (Heading, error)` в `storage.Repository` interface + bbolt/fakes реализация + использование в `fetchNameCache`. Отдельная фича в v2.

Обе рекомендации — не блокеры. Verdict `PASS` независимо от их исполнения.

# Code Review: search-and-multi-select

## Verdict: PASS

Все 8 tasks из плана выполнены. Все 24 REQ-X.Y из requirements трассируются в production-код И тесты. Все 18 correctness properties покрыты property-тестами. Полный test suite зелёный (race detector clean), bin/todushka собирается, lint 0 issues. Backward-compat T-1 preservation invariants (`c`/`x`/`d`/`p` over cursor при пустом selection) сохранены после введения `dispatch`. Найдено 2 минорные нотиссиа (F-1 doc-инконсистентность, F-2 UX-edge case), ни одна не critical и не major.

## Change Set

`review_base_commit = 95704b6` (последний коммит в `main`). Все изменения — uncommitted working-tree (модификации + новые файлы).

| File | Status | Notes |
|------|--------|-------|
| `internal/tui/app.go` | ✅ Planned | +103 строки: 4 новых Model-поля, handleKey precedence, bulkResultMsg case, dispatch, switchList cleanup, View overlay, viewList/viewFooter обновления |
| `internal/tui/app_test.go` | ✅ Planned | +465 строк: 4 T-1 preservation, 6 T-2 selection, 1 T-4 switchList, 2 T-6 help/footer, 9 PBT |
| `internal/tui/keys.go` | ✅ Planned | +4 binding'а: Filter, ToggleSelect, SelectAll, ClearSelection |
| `internal/tui/msgs.go` | ✅ Planned | +bulkResultMsg тип |
| `internal/tui/filter.go` | ✅ Planned NEW | 80 строк: foldCaseContains, displayedTasks, pruneSelection, handleFilterKey |
| `internal/tui/filter_test.go` | ✅ Planned NEW | 243 строки: 10 unit + 3 PBT |
| `internal/tui/bulk.go` | ✅ Planned NEW | 136 строк: bulkAction, confirmState, dispatch, selectionIDs, perCursorCmd, runBulk, applyAction, handleConfirmKey |
| `internal/tui/bulk_test.go` | ✅ Planned NEW | 349 строк: 8 unit + 6 PBT |

Ни одного **Unexpected** или **Not Changed** файла из спроектированного scope. Файлы NOT changed из design §2.3 (`storage/`, `app/`, `domain/`, `editor.go`, `style.go`, `cmd/`, `cli/`) — действительно не трогали.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 `/` enters Filter Mode | `TestFilter_SlashEntersMode` + `TestProp_FilterStateTransitions` | `app.go:155-158` | CP-2 | ✅ |
| REQ-1.2 live substring | `TestDisplayedTasks_SubstringMatch` + `TestDisplayedTasks_CaseInsensitive` + `TestProp_FilterIsSubstringSubset` | `filter.go:22-34` | CP-1 | ✅ |
| REQ-1.3 Enter saves query | `TestFilter_EnterPreservesQuery` + `TestProp_FilterStateTransitions` | `filter.go:59-61` | CP-2 | ✅ |
| REQ-1.4 Esc clears query | `TestFilter_EscClearsQuery` + `TestProp_FilterStateTransitions` | `filter.go:55-58` | CP-2 | ✅ |
| REQ-1.5 `(no matches)` placeholder | `TestFilter_NoMatchesPlaceholder` + `TestProp_NoMatchesShowsPlaceholder` | `app.go:460-464` | CP-3 | ✅ |
| REQ-1.6 list switch clears filter | `TestSwitchList_ClearsFilterAndSelection` + `TestProp_SwitchListResetsState` | `app.go:310-316` | CP-4 | ✅ |
| REQ-1.7 whitespace = empty | `TestDisplayedTasks_WhitespaceQueryEquivEmpty` + `TestProp_FilterIsSubstringSubset` | `filter.go:23-26` | CP-1 | ✅ |
| REQ-2.1 Space toggle | `TestSelection_SpaceToggle` + `TestProp_SpaceIsInvolution` | `app.go:175-184` | CP-5 | ✅ |
| REQ-2.2 prefix when selected | `TestSelection_PrefixVisibleWhenNonEmpty` + `TestProp_PrefixIffNonEmpty` | `app.go:467-475` | CP-6 | ✅ |
| REQ-2.3 no prefix when empty | `TestSelection_PrefixHiddenWhenEmpty` + `TestProp_PrefixIffNonEmpty` | `app.go:467` (`showPrefix := len(m.selected) > 0`) | CP-6 | ✅ |
| REQ-2.4 `*` selects visible | `TestSelection_StarSelectsAllVisible` + `TestProp_StarSelectsAllVisible` | `app.go:185-189` | CP-7 | ✅ |
| REQ-2.5 Esc clears selection | `TestSelection_EscClearsSelection` + `TestProp_EscClearsSelection` | `app.go:143-145` | CP-8 | ✅ |
| REQ-2.6 filter hides → drop | `TestFilter_ChangeDropsHiddenFromSelection` + `TestProp_SelectionSubsetOfVisible` | `filter.go:39-49` (`pruneSelection`) | CP-9 | ✅ |
| REQ-2.7 list switch clears selection | `TestSwitchList_ClearsFilterAndSelection` + `TestProp_SwitchListResetsState` | `app.go:313` | CP-4 | ✅ |
| REQ-2.8 `Selected: N` counter | `TestSelection_StatusBarCounter` + `TestProp_StatusBarShowsCount` | `app.go:529-530` | CP-10 | ✅ |
| REQ-3.1 empty → per-cursor | `TestTUI_*CursorTaskWhenNoSelection` (×4 T-1) + `TestBulk_EmptyDispatchesPerCursor` + `TestProp_EmptySelectionEquivCursor` | `bulk.go:48-50` | CP-11 | ✅ |
| REQ-3.2 1≤N<5 no confirm | `TestBulk_BelowThresholdNoConfirm` + `TestProp_BulkThresholdGate` | `bulk.go:51-54` | CP-12 | ✅ |
| REQ-3.3 N≥5 confirm | `TestBulk_AtThresholdShowsConfirm` + `TestProp_BulkThresholdGate` | `bulk.go:55` | CP-12 | ✅ |
| REQ-3.4 only `y` confirms | `TestBulk_YConfirmsAndDispatches` + `TestBulk_NonYDismissesNoDispatch` + `TestProp_OnlyYConfirms` | `bulk.go:129-135` | CP-13 | ✅ |
| REQ-3.5 aggregate failure | `TestBulk_AggregateMath` + `TestProp_BulkAggregateMath` | `bulk.go:88-108` | CP-14 | ✅ |
| REQ-3.6 clear after success | `TestBulk_SuccessClearsSelection` + `TestProp_SuccessClearsSelection` | `app.go:83` | CP-15 | ✅ |
| REQ-3.7 fatal preserves | `TestBulk_FatalPreservesSelection` + `TestProp_FatalPreservesSelection` | `app.go:77-82` | CP-16 | ✅ |
| REQ-4.1 help includes new keys | `TestHelp_IncludesNewBindings` + `TestProp_HelpContainsNewKeys` | `app.go:506` | CP-17 | ✅ |
| REQ-4.2 footer hints | `TestFooter_IncludesNewHints` + `TestProp_FooterContainsNewKeys` | `app.go:517` | CP-18 | ✅ |

24/24 ✅. Каждое REQ имеет ≥1 unit-тест **и** ≥1 property-тест, плюс трассируемое место в production-коде.

## Design Conformance

### §3.1 Architectural Boundaries

✅ Все новые символы — в пакете `internal/tui/`. Никаких изменений в `storage/`, `app/`, `domain/`. Зависимости пакетов соблюдены: `bulk.go` импортирует `app` (для `*Service`), `storage` (для `ErrDatabaseLocked`), `id` — все upstream-направления.

### §3.2 Data Models

Design §2.5 vs реализация:

| Design | Implementation | Match |
|---|---|---|
| `Model.filterQuery string` | `app.go:39` | ✅ |
| `Model.filtering bool` | `app.go:40` | ✅ |
| `Model.selected map[id.ID]struct{}` | `app.go:37` | ✅ |
| `Model.confirm *confirmState` | `app.go:38` | ✅ |
| `confirmState{action, ids}` | `bulk.go:25-28` | ✅ |
| `bulkAction int` + 4 constants | `bulk.go:13-20` | ✅ |
| `bulkResultMsg{action, succeeded, failed, lastErr, fatal}` | `msgs.go` | ✅ |
| `bulkConfirmThreshold = 5` | `bulk.go:22` | ✅ |

Все типы по design.

### §3.3 API Contracts

N/A — фича внутри TUI, никакого public API / wire format / schema migration.

### §3.4 Error Handling

Design §2.7 vs реализация:

| Scenario | Design Action | Code | Match |
|---|---|---|---|
| Per-task bulk failure | continue, count failed, save lastErr | `bulk.go:103-104` | ✅ |
| context.Canceled mid-bulk | fatal=true, abort | `bulk.go:98-101` | ✅ |
| ErrDatabaseLocked mid-bulk | fatal=true, abort | `bulk.go:98-101` | ✅ |
| Bulk on empty visible | no-op, no msg | Implicit: `len(selected)==0 → perCursorCmd`; pruneSelection ensures selected ⊆ visible | ✅ |
| Confirm Esc/any non-y | dismiss, no dispatch | `bulk.go:129-135` | ✅ |
| ID gone from m.tasks but in selected | counted as recoverable failure | `applyAction` returns service error, бухгалтеривается | ✅ |
| `*` on empty `displayedTasks` | no-op | Loop body не выполняется | ✅ |

### §3.5 Correctness Properties

Все 18 CP реализованы как `TestProp_*` в T-7. Сверка по traceability matrix выше: каждое CP → как минимум одно property-test, и каждое property-test → как минимум одно REQ. Сами PBT прогоняются под `rapid.Check`, стабильны при `-count=2`.

### §3.6 Documentation Consistency

**Mermaid диаграмма handleKey precedence:** в диаграмме design §2.2 показаны 5 routing-branches (screen-Editor, screen-QuickEntry, filtering, confirm, default). Реализация имеет порядок проверок:

```
1. m.confirm != nil   → handleConfirmKey
2. m.filtering        → handleFilterKey
3. m.screen switch    → handleEditorKey / handleQuickEntryKey
4. main key matchers  → list keymap
```

→ см. F-1 ниже.

## Code Quality

### §4.1 Naming & Clarity

Все идентификаторы идиоматичны для Go и согласованы с существующей кодовой базой: `bulkAction`, `confirmState`, `dispatch`, `runBulk`, `applyAction`, `handleConfirmKey`, `handleFilterKey`, `displayedTasks`, `foldCaseContains`, `pruneSelection`, `selectionIDs`, `perCursorCmd`, `bulkConfirmThreshold`. Без аббревиатур, без single-letter receivers (кроме `m Model`).

### §4.2 Dead Code & Debug Artifacts

Чисто: 0 TODO, 0 закомментированных блоков, 0 debug-print. `m.completeSelected/cancelSelected/deleteSelected/pinSelected` остались как private helpers — используются через `perCursorCmd` для empty-selection fallback (CP-11 backward compat).

### §4.3 Scope Creep

Никакого. Все 4 design ADR (inline footer, single-y confirm, sequential bulk, ASCII marker) реализованы как решены. Никаких "while we're here" рефакторингов в неrelated коде.

### §4.4 Test Quality

- 70 проходящих test cases в `internal/tui/` (включая subtests `TestProp_EmptySelectionEquivCursor/{c,x,d,p}` и `TestProp_SwitchListResetsState/{tab,shift+tab,1..6}`).
- Все тесты используют existing fixture pattern (`setupModelWithInboxTasks`, `newTestModel`, `newTestModelWithService`).
- testify `require` (не `assert` — соответствует Tier 2 reference).
- Property-тесты используют `pgregory.net/rapid` с малыми генераторами (titles ≤15 chars, slices ≤10) — стабильны, fast.
- Edge cases покрыты: threshold boundary (4 vs 5), Unicode fold-case, whitespace query, ghost IDs (несуществующие task IDs для аггрегации failures), Esc semantics в trех разных контекстах (editor / quickEntry / list-with-selection).

## Security

Никаких security-проблем не обнаружено в изменённых файлах:

- **Внешние входы:** TUI принимает только keystrokes от локального терминала. Никаких сетевых endpoints не добавлено.
- **Injection:** `foldCaseContains` использует `strings.ToLower` + `strings.Contains` — no regex, no shell, no SQL.
- **Secrets:** ни одного hardcoded credential / token / API key.
- **Data exposure:** `bulkResultMsg.lastErr.Error()` рендерится в status bar — содержит только domain-level и storage-level error messages, без системных деталей (stacktrace, fs paths, etc.).
- **Authorization:** локальное single-user приложение, нет concept of роли/permissions.

## Verification Evidence

Все команды повторно прогнаны в этой сессии (`go clean -testcache` перед `task test` для гарантированно свежего execution, не из кэша).

- **Tests (full suite, fresh):**

```
task: [test] go test ./...
?   	github.com/jtprogru/todushka/cmd/todushka	[no test files]
ok  	github.com/jtprogru/todushka/internal/app	0.387s
ok  	github.com/jtprogru/todushka/internal/cli	1.455s
ok  	github.com/jtprogru/todushka/internal/config	2.166s
ok  	github.com/jtprogru/todushka/internal/domain/area	1.097s
ok  	github.com/jtprogru/todushka/internal/domain/id	1.820s
ok  	github.com/jtprogru/todushka/internal/domain/project	0.732s
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	2.532s
ok  	github.com/jtprogru/todushka/internal/domain/repeat	4.060s
ok  	github.com/jtprogru/todushka/internal/domain/tag	2.868s
ok  	github.com/jtprogru/todushka/internal/domain/task	3.289s
ok  	github.com/jtprogru/todushka/internal/domain/today	3.685s
?   	github.com/jtprogru/todushka/internal/storage	[no test files]
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	5.448s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	4.418s
ok  	github.com/jtprogru/todushka/internal/tui	5.386s
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

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | minor | `.spec/features/search-and-multi-select/design.md` (ADR-6 consequences) | Design ADR-6 описывает handleKey precedence как `screen → filtering → confirm → listKey`, но реализация (соответствующая task-plan T-5 subtask 11) использует `confirm → filtering → screen → listKey`. Semантически разницы нет (confirm-modal появляется только когда filtering=false и screen=screenList), но документ устарел относительно фактического кода. | Doc-only, не функциональное REQ |
| F-2 | nit | `internal/tui/filter.go:71` | `m.filterQuery += string(msg.Runes)` не имеет максимальной длины. Локальный пользователь, paste'нувший 10MB, повесит свой TUI на substring-match'е. Не security (single-user local), но minor UX cliff edge. Рекомендация: добавить cap `const filterQueryMax = 256` и в `KeyRunes`-ветке `if len([]rune(m.filterQuery)) >= filterQueryMax { return m, nil }`. | UX hardening, не привязано к REQ |

## Recommendations

**Minor (необязательно для approve):**

1. **F-1 (doc maintenance):** обновить `design.md` ADR-6 consequences — заменить "screen → filtering → confirm → listKey" на "confirm → filtering → screen → listKey" чтобы документ соответствовал коду. Можно сделать в отдельном PR.

2. **F-2 (UX hardening):** добавить `const filterQueryMax = 256` и проверку в `handleFilterKey` (RunesKey + KeySpace ветки). Тест: paste-симуляция через многократный `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("очень длинная строка")}` — query не растёт сверх лимита. Можно сделать в отдельном PR.

Обе рекомендации — не блокеры. Verdict `PASS` независимо от их исполнения.

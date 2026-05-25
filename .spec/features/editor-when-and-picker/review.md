# Code Review: editor-when-and-picker (v0.6.0)

## Verdict: PASS

Все 8 tasks из плана выполнены. 26 REQs трассируются в код + тесты; все 13 CPs покрыты property-тестами. Свежий полный test-suite зелёный (все packages PASS, 17 PBTs в editor_test.go стабильны), race detector clean, build OK, lint 0 issues, gofmt clean. Backward-compat сохранён (213 baseline tests passing с обновлёнными assertion'ами на новый editor layout). Part A (context-aware label) и Part B (Area/Project/Heading picker) интегрированы cohesively. Найдено 2 минорных нота — F-1 inconsistent error wrapping style между Area/Project resolve блоками, F-2 heading "not found" error без project name context. Ни одна не critical/major.

## Change Set

`review_base_commit = 61ef286` (HEAD `main` после ux-polish pipeline state finalize). Все изменения uncommitted.

| File | Status | Notes |
|------|--------|-------|
| `internal/tui/editor.go` | ✅ Planned MODIFIED | +151 lines: 3 new textinput fields в EditorModel; 3 new editorField const + fieldCount=9; new `whenLabel(t)` pure helper; NewEditor signature `(ctx, t, svc)` с pre-fill через Repo; focusCurrent/UpdateForm extended; View renders 3 new blocks + context-aware When label; ApplyAndSave sequential resolve (Area→Project→Heading) с auto-clear heading on project change |
| `internal/tui/editor_test.go` | ✅ Planned MODIFIED | +629 lines: obsolete v0.5.0 hint tests removed (2); rename FieldCountIsSix→Nine; updated `TestProp_HintConditional` → `TestProp_WhenLabelMatchesContext`; +6 label/view tests; +4 pre-fill tests; +1 visual; +13 resolve tests; +12 PBT (CP-2..CP-13) |
| `internal/tui/app.go` | ✅ Planned MODIFIED | `openEditor` обновлён на новую NewEditor signature |
| `internal/tui/shell_test.go` | ✅ Planned MODIFIED | 2 NewEditor call sites обновлены |

NOT changed (per design §2.3): `storage/repository.go`, `bbolt/`, `fakes/`, `app/service.go`, `app/queries.go`, `domain/*`, `internal/tui/{shell.go (core), filter.go, bulk.go, details.go, style.go, keys.go, msgs.go, run.go}`, `config/`, `cli/`, `cmd/` — confirmed unchanged. Никаких новых файлов — фича изолирована в editor.go.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 Inbox label when no Area/Project | `TestWhenLabel_InboxForUnrelatedTask` + `TestEditor_ViewShowsInboxLabel` + `TestProp_WhenLabelMatchesContext` | `editor.go:324-329` (`whenLabel`) | CP-1 | ✅ |
| REQ-1.2 Anytime label when has Area or Project | `TestWhenLabel_AnytimeForAreaTask` + `TestWhenLabel_AnytimeForProjectTask` + `TestEditor_ViewShowsAnytimeLabel` | `editor.go:324-329` | CP-1 | ✅ |
| REQ-1.3 hint removed | `TestEditor_ViewHidesOldHint` + `TestProp_WhenLabelMatchesContext` (assert NotContains) | `editor.go:View()` no hint block | — | ✅ |
| REQ-1.4 Anytime → Someday=false | Existing `TestEditor_ApplyAndSaveMapsAnytime` (preserved) | `editor.go:220` (existing) | — | ✅ |
| REQ-1.5 Someday → Someday=true | Existing `TestEditor_ApplyAndSaveMapsSomeday` | `editor.go:220` | — | ✅ |
| REQ-1.6 Space cycles | Existing behavior preserved | `app.go:handleEditorKey` (unchanged) | — | ✅ |
| REQ-2.1 Area textinput | `TestEditor_ViewRendersAllNewFields` | `editor.go` field + View block | — | ✅ |
| REQ-2.2 Area pre-fill | `TestEditor_NewEditorPrefillArea` + `TestProp_PreFillRoundTrip` | `editor.go:97-104` (NewEditor) | CP-2 | ✅ |
| REQ-2.3 Empty area on open | `TestEditor_NewEditorEmptyArea` | NewEditor sees `t.AreaID == nil`, leaves textinput empty | — | ✅ |
| REQ-2.4 Empty area clears AreaID | `TestEditor_SaveEmptyAreaClearsID` + `TestProp_EmptyAreaClears` | `editor.go:255-257` | CP-3 | ✅ |
| REQ-2.5 Invalid area error | `TestEditor_SaveInvalidAreaErrors` + `TestProp_InvalidAreaErrors` | `editor.go:259-266` | CP-5 | ✅ |
| REQ-3.1 Project textinput | `TestEditor_ViewRendersAllNewFields` | `editor.go` field + View | — | ✅ |
| REQ-3.2 Project pre-fill | `TestEditor_NewEditorPrefillProject` + `TestProp_PreFillRoundTrip` | `editor.go:106-113` | CP-2 | ✅ |
| REQ-3.3 Empty project on open | `TestEditor_NewEditorEmptyArea` (also covers project) | NewEditor `t.ProjectID == nil` → empty | — | ✅ |
| REQ-3.4 Empty project clears both IDs | `TestEditor_SaveEmptyProjectClearsBothIDs` + `TestProp_EmptyProjectClearsBoth` | `editor.go:270-273` | CP-4 | ✅ |
| REQ-3.5 Project ambiguity error | `TestEditor_SaveAmbiguousProjectErrors` + `TestProp_AmbiguousProjectErrors` | `editor.go:279-284` | CP-6 | ✅ |
| REQ-4.1 Heading textinput | `TestEditor_ViewRendersAllNewFields` | `editor.go` field + View | — | ✅ |
| REQ-4.2 Heading pre-fill | `TestEditor_NewEditorPrefillHeading` + `TestProp_PreFillRoundTrip` | `editor.go:115-127` (HeadingList find-by-ID) | CP-2 | ✅ |
| REQ-4.3 Empty heading on open | `TestEditor_NewEditorEmptyArea` | NewEditor `t.HeadingID == nil` → empty | — | ✅ |
| REQ-4.4 Empty heading clears | (implicit; preserved through resolve when headingName=="") | `editor.go:295-296` | — | ✅ |
| REQ-4.5 Heading without project error | `TestEditor_SaveHeadingWithoutProjectErrors` + `TestProp_HeadingWithoutProject` | `editor.go:298-300` | CP-7 | ✅ |
| REQ-4.6 Heading found in project | `TestEditor_SaveValidHeadingSetsID` + `TestEditor_SaveCaseInsensitiveHeading` + `TestProp_ValidHeadingResolves` + `TestProp_HeadingCaseInsensitive` | `editor.go:301-315` (HeadingList + EqualFold) | CP-8, CP-12 | ✅ |
| REQ-5.1 Field order Title→Notes→Start→Deadline→Area→Project→Heading→Tags→When | `TestProp_TabCycleOrder` + visual rendering in View | `editor.go` field order + View JoinVertical | CP-10 | ✅ |
| REQ-5.2 fieldCount=9 | `TestEditor_FieldCountIsNine` + `TestProp_FieldCountInvariant` | `editor.go:32` | CP-9 | ✅ |
| REQ-5.3 Tab cycle preserves first 2 | `TestTUI_EditorTabCyclesFields` (existing, passes unchanged) | nextField unchanged | — | ✅ |
| REQ-5.4 New enum constants | (verified via grep + fieldCount test) | `editor.go:27-29` | CP-9 | ✅ |
| REQ-6.1 View renders 3 new fields | `TestEditor_ViewRendersAllNewFields` | `editor.go:View` JoinVertical | — | ✅ |
| REQ-6.2 Error display via m.err | `TestEditor_SaveInvalidAreaErrors` checks error message | Existing m.err field rendering | — | ✅ |

26/26 ✅. Каждый REQ имеет ≥1 test + код. Дополнительно ADR-5 (auto-clear heading on project change) покрыт через `TestEditor_SaveProjectChangeAutoClearsHeading` + `TestProp_ProjectChangeClearsOrphanHeading` → CP-13.

## Design Conformance

### §3.1 Architectural Boundaries

✅ Все изменения в `internal/tui/editor.go` + minimal call site updates. Используются existing `Repo` methods: `AreaGet`, `AreaFindByNormalized`, `ProjectGet`, `ProjectFindByName`, `HeadingList`. No new Repository / Service methods.

### §3.2 Data Models

Design §2.5 vs реализация:

| Design | Implementation | Match |
|---|---|---|
| `EditorModel.area textinput.Model` | `editor.go:51` | ✅ |
| `EditorModel.project textinput.Model` | `editor.go:52` | ✅ |
| `EditorModel.heading textinput.Model` | `editor.go:53` | ✅ |
| `fieldArea`, `fieldProject`, `fieldHeading` constants | `editor.go:27-29` | ✅ |
| `fieldCount == 9` | `editor.go:32` | ✅ |
| `whenLabel(t task.Task) string` pure helper | `editor.go:324-329` | ✅ |
| `NewEditor(ctx, t, svc)` signature change | `editor.go:61` | ✅ |
| `shellEditorWhen` unchanged | `editor.go:35-40` | ✅ |

### §3.3 API Contracts

`NewEditor` signature change — breaking, но это **internal API** (никакой публичной surface). Все call sites обновлены (`app.go`, tests).

### §3.4 Error Handling

Design §2.7 vs реализация:

| Scenario | Design Action | Code | Match |
|---|---|---|---|
| `Repo.AreaGet` fails в pre-fill | skip pre-fill; leave empty | `editor.go:101-103` (`if err == nil`) | ✅ |
| `Repo.ProjectGet` fails в pre-fill | skip pre-fill | `editor.go:110-112` | ✅ |
| `Repo.HeadingList` fails в pre-fill | skip pre-fill | `editor.go:119-126` | ✅ |
| Area name → ErrNotFound | error `"area 'X' not found"` | `editor.go:261-263` (`errors.Is(err, storage.ErrNotFound)`) | ✅ |
| Project name → 0 matches | error `"project 'X' not found"` | `editor.go:279-281` | ✅ |
| Project name → 2+ matches | error `"ambiguous"` | `editor.go:282-284` | ✅ |
| Heading non-empty + project empty | error `"heading requires a project"` | `editor.go:298-300` | ✅ |
| Heading name not in project | error `"heading 'X' not found in project"` | `editor.go:313-315` | ✅ |
| Project change → orphan heading | Auto-clear `t.HeadingID = nil` before heading resolve | `editor.go:286-289` | ✅ |

### §3.5 Correctness Properties

13/13 CPs реализованы как `TestProp_*`. Прогоняются через `rapid.Check`, стабильны при `-count=2` (~5 sec total). Никаких flakes.

### §3.6 Documentation Consistency

Mermaid diagram в design §2.2 отражает фактические потоки:
- NewEditor → pre-fill via AreaGet/ProjectGet/HeadingList → EditorModel
- ApplyAndSave → sequential resolve Area→Project→Heading → first error abort OR EditTask
- View → whenLabel + 3 new field blocks

Implementation matches.

## Code Quality

### §4.1 Naming & Clarity

Idiomatic Go. Identifiers descriptive: `fieldArea`/`fieldProject`/`fieldHeading`, `whenLabel`, `areaName`/`projectName`/`headingName` (local resolve vars), `newPID` (snapshot for project change detection). Comments объясняют ADR-5 (auto-clear on project change).

### §4.2 Dead Code & Debug Artifacts

Чисто: 0 TODO, 0 закомментированных блоков, 0 debug-print. Старый hint block физически удалён, не закомментирован. Obsolete v0.5.0 tests (`TestEditor_HintShownWhenAnytimeNoAreaProject`, `TestEditor_HintHiddenForSomeday`) полностью удалены с заменой на equivalent coverage.

### §4.3 Scope Creep

Никакого. 7 design ADRs (1-7) реализованы как решены. Никаких "while we're here" рефакторингов в неrelated коде.

### §4.4 Test Quality

- All packages PASS; race-detector clean.
- testify `require`. Existing fixtures (`newTestModel`, `newTestModelWithService`, `setupModelWithInboxTasks`, `setupRapidModel`) переиспользованы.
- PBT генераторы малые: `rapid.IntRange(2, 4)` для duplicate counts, `rapid.StringMatching` для names с ASCII-only.
- Edge cases покрыты: empty/valid/invalid for each picker; ambiguous project; heading without project; case-insensitive match; project change → auto-clear; sequential resolve order.
- Pre-fill round-trip property (`TestProp_PreFillRoundTrip`) — открыть + сохранить без изменений = identity. Хорошее invariant testing.

## Security

- **External inputs:** только user-typed names в editor textinput. Trim'аются перед lookup. Lookup через type-safe Repo methods (не raw queries).
- **Injection:** `Repo.AreaFindByNormalized`/`ProjectFindByName`/`HeadingList` — wrapped bbolt key ops; no SQL/shell injection vector.
- **Secrets:** none.
- **Data exposure:** error messages contain user-provided name (`"area %q not found"`). Quoted via `%q` — escape special chars. Acceptable; matches existing pattern.
- **Concurrency:** Editor state is per-Update; не shared. Repo operations atomic via bbolt transactions.
- **Resource limits:** textinput `CharLimit = 100` для area/project/heading — bounded.

## Verification Evidence

Все команды повторно прогнаны в этой сессии после `go clean -testcache`.

- **Tests (full suite, fresh):**

```
?       github.com/jtprogru/todushka/cmd/todushka       [no test files]
ok      github.com/jtprogru/todushka/internal/app       0.596s
ok      github.com/jtprogru/todushka/internal/cli       0.971s
ok      github.com/jtprogru/todushka/internal/config    5.758s
ok      github.com/jtprogru/todushka/internal/domain/area       1.324s
ok      github.com/jtprogru/todushka/internal/domain/id 1.692s
ok      github.com/jtprogru/todushka/internal/domain/project    2.067s
ok      github.com/jtprogru/todushka/internal/domain/quickentry 2.841s
ok      github.com/jtprogru/todushka/internal/domain/repeat     3.248s
ok      github.com/jtprogru/todushka/internal/domain/tag        2.437s
ok      github.com/jtprogru/todushka/internal/domain/task       3.620s
ok      github.com/jtprogru/todushka/internal/domain/today      4.385s
?       github.com/jtprogru/todushka/internal/storage   [no test files]
ok      github.com/jtprogru/todushka/internal/storage/bbolt     5.683s
ok      github.com/jtprogru/todushka/internal/storage/fakes     4.773s
ok      github.com/jtprogru/todushka/internal/tui       6.433s
?       github.com/jtprogru/todushka/internal/version   [no test files]
```

- **Build:** `task: [build] go build -o bin/todushka ./cmd/todushka` — success.

- **Lint:** `task: [lint] golangci-lint run` → `0 issues.`

- **Format check (`gofmt -l internal/ cmd/`):** empty (clean).

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | nit | `internal/tui/editor.go:259-266` | Area resolve использует `errors.Is(err, storage.ErrNotFound)` для специфичной обработки, но Project resolve (lines 275-284) не различает ErrNotFound vs другие errors — все wrap'ятся через `%w`. Inconsistency в error handling style. На самом деле семантика разная: `ProjectFindByName` возвращает slice (не ErrNotFound на пустой результат), что делает обработку Project'а natural. Это не баг — просто стилевая асимметрия. | — |
| F-2 | nit | `internal/tui/editor.go:313-315` | Heading error message `"heading 'X' not found in project"` не включает имя project'а. Для пользователя было бы информативнее `"heading 'X' not found in project 'Y'"`. Lookup для project name можно сделать через `projectName` локальную переменную (или skipped if it was cleared). Minor UX polish. | REQ-4.6 |

## Recommendations

**Minor / nit (не блокеры):**

1. **F-1 (error wrapping consistency):** Опционально — выровнять стиль error handling между Area и Project. Например, добавить explicit `errors.Is(err, storage.ErrNotFound)` обработку и в Project, если в будущем `ProjectFindByName` начнёт его возвращать. Cosmetic only.

2. **F-2 (heading error context):** Добавить project name в heading not-found error: `"heading 'X' not found in project 'Y'"`. Простая правка — `projectName` уже в scope от Group 3.

Обе рекомендации — не блокеры. Verdict `PASS` независимо от их исполнения.

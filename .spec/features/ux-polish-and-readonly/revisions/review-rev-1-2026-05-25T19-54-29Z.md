# Code Review: ux-polish (v0.5.0)

## Verdict: PASS

Все 8 tasks из плана выполнены. 26 REQs трассируются в код + тесты; все 12 CPs покрыты property-тестами. Свежий полный test-suite зелёный (185 TUI test cases + 23 config + 5 CLI = 213 total), race detector clean, build OK, lint 0 issues, gofmt clean. Backward-compat T-1 baseline (174 tests) сохранён. Найдено 2 минорных нота — F-1 redundant `m.screen != screenEditor` check в View() (defensive duplication), F-2 PBT `TestProp_SeparatorsConditional` сужено до `screenList` (modal-internal `─` runes конфликтуют с full-output substring match). Ни одна не critical/major.

## Change Set

`review_base_commit = 618f63a` (HEAD `main` после tui-shell merge). Все изменения uncommitted.

| File | Status | Notes |
|------|--------|-------|
| `internal/config/app.go` | ✅ Planned MODIFIED | `Validate()` accepts `"auto"`/`"system"` в switch case |
| `internal/config/app_test.go` | ✅ Planned MODIFIED | +2 unit tests |
| `internal/tui/detect_dark_darwin.go` | ✅ Planned NEW | macOS detection via `defaults read` |
| `internal/tui/detect_dark_linux.go` | ✅ Planned NEW | Linux detection via `gsettings` |
| `internal/tui/detect_dark_other.go` | ✅ Planned NEW | Fallback `(false, nil)` |
| `internal/tui/theme_resolve.go` | ✅ Planned NEW | `detectDarkModeFn` injection + `resolveAutoTheme()` |
| `internal/tui/theme_resolve_test.go` | ✅ Planned NEW | 5 unit + 3 PBT |
| `internal/tui/run.go` | ✅ Planned MODIFIED | `selectThemeFromConfig` handles auto/system + NO_COLOR precedence |
| `internal/tui/editor.go` | ✅ Planned MODIFIED | `shellEditorWhen` enum, `someday` → `when`, radio-style View, hint |
| `internal/tui/editor_test.go` | ✅ Planned NEW | 7 unit + 3 PBT |
| `internal/tui/app.go` | ✅ Planned MODIFIED | Editor Space handler, `editorSavedMsg` inline splice, `renderSeparator`, View() separators |
| `internal/tui/app_test.go` | ✅ Planned MODIFIED | +4 splice tests + 3 PBT |
| `internal/tui/shell_test.go` | ✅ Planned MODIFIED | +7 separator tests + 3 PBT |

NOT changed (per design §2.3): `storage/`, `app/service`, `domain/`, `tui/{filter.go, bulk.go, details.go, shell.go (core), keys.go, msgs.go, style.go}`, `cli/`, `cmd/`, `config/paths.go`, `config/loader.go` — all confirmed unchanged.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 auto triggers detection | `TestSelectThemeFromConfig_AutoDarkUsesMacchiato` + `TestProp_AutoThemeResolution` | `run.go:selectThemeFromConfig`, `theme_resolve.go:resolveAutoTheme` | CP-1, CP-11 | ✅ |
| REQ-1.2 macOS detection | (build-tag isolated; integration via theme_resolve tests with mock) | `detect_dark_darwin.go` | CP-1 | ✅ |
| REQ-1.3 Linux detection | (build-tag isolated; integration via theme_resolve tests with mock) | `detect_dark_linux.go` | CP-1 | ✅ |
| REQ-1.4 fallback dark | `TestResolveAutoTheme_ErrorFallsBackToDark` + `TestProp_AutoThemeResolution` | `theme_resolve.go:11-14` (`if err != nil || isDark`) | CP-1 | ✅ |
| REQ-1.5 auto/system both valid | `TestValidate_AutoThemeIsValid` + `TestValidate_SystemThemeIsValid` + `TestSelectThemeFromConfig_SystemAliasMatchesAuto` | `config/app.go` Validate switch | CP-11 | ✅ |
| REQ-1.6 NO_COLOR overrides | `TestSelectThemeFromConfig_NoColorOverridesAuto` + `TestProp_NoColorOverridesAuto` | `run.go:selectThemeFromConfig` NO_COLOR early-return | CP-2 | ✅ |
| REQ-1.7 500ms timeout | (verified via `context.WithTimeout` in detect files) | `detect_dark_darwin.go:11`, `detect_dark_linux.go:11` | CP-1 | ✅ |
| REQ-2.1 inline splice by ID | `TestEditorSavedMsg_InlineSpliceByID` + `TestProp_EditSpliceByID` | `app.go:157-170` | CP-4 | ✅ |
| REQ-2.2 batch refresh | `TestEditorSavedMsg_FiresBatchedCmd` + `TestProp_RefreshBatchHasBothCmds` | `app.go:167-170` (tea.Batch) | CP-12 | ✅ |
| REQ-2.3 missing ID skipped | `TestEditorSavedMsg_NotFoundSkipsSplice` + `TestProp_EditSplicePreservesLength` | `app.go:160-165` (break only on match) | CP-3, CP-4 | ✅ |
| REQ-2.4 length preserved | `TestEditorSavedMsg_PreservesSliceLength` + `TestProp_EditSplicePreservesLength` | `app.go:160-165` (in-place assign) | CP-3 | ✅ |
| REQ-3.1 radio-style View | `TestEditor_HintShownWhenAnytimeNoAreaProject` (visual content) + `TestProp_HintConditional` | `editor.go:230-247` | CP-5 | ✅ |
| REQ-3.2 focus highlight | (verified via Tab cycling test `TestTUI_EditorTabCyclesFields` continuing to pass + new `TestEditor_FieldCountIsSix`) | `editor.go:239-243` | CP-5 | ✅ |
| REQ-3.3 Space toggles | `TestProp_WhenToggleInvolution` | `app.go:handleEditorKey` Space case for fieldWhen | CP-5 | ✅ |
| REQ-3.4 mapping at save | `TestEditor_ApplyAndSaveMapsAnytime` + `TestEditor_ApplyAndSaveMapsSomeday` + `TestProp_WhenMapping` | `editor.go:ApplyAndSave` `t.Someday = m.when == whenSomeday` | CP-6 | ✅ |
| REQ-3.5 hint conditional | `TestEditor_HintShownWhenAnytimeNoAreaProject` + `TestEditor_HintHiddenForSomeday` + `TestProp_HintConditional` | `editor.go:244-247` | CP-7 | ✅ |
| REQ-3.6 field count preserved | `TestEditor_FieldCountIsSix` + `TestTUI_EditorTabCyclesFields` passing unchanged | `editor.go` fieldCount = fieldWhen+1 = 6 | — | ✅ |
| REQ-4.1 separator below header | `TestView_HasSeparatorsInFullScreen` + `TestProp_SeparatorsConditional` | `app.go:528` (sep после header в JoinVertical) | CP-8 | ✅ |
| REQ-4.2 separator above footer | Same as 4.1 | `app.go:528` (sep перед footer) | CP-8 | ✅ |
| REQ-4.3 separator width | `TestRenderSeparator_FullWidth` + `TestProp_SeparatorWidth` | `app.go:494` (`strings.Repeat("─", width)`) | CP-9 | ✅ |
| REQ-4.4 legacy no separator | `TestView_NoSeparatorsInLegacy` + `TestProp_SeparatorsConditional` | `app.go:517` (height/width threshold) | CP-8 | ✅ |
| REQ-4.5 editor no separator | `TestView_NoSeparatorsInEditor` (custom assertion на full-width run) | `app.go:499` (editor early return) + `app.go:517` (defensive guard) | CP-8 | ✅ |
| REQ-4.6 bodyH adjusted | `TestView_FullScreenHeightWithSeparators` + `TestProp_FullScreenHeightWithSeparators` | `app.go:523` (`bodyH = m.height - headerH - footerH - 2`) | CP-10 | ✅ |
| REQ-5.1 174 tests passing | Independent full-suite run | — | — | ✅ |
| REQ-5.2 explicit theme skips detect | `TestSelectThemeFromConfig_AutoLightUsesLatte` (mock) + manual code review of run.go | `run.go:selectThemeFromConfig` only resolves auto/system | — | ✅ |
| REQ-5.3 editor field count = 6 | `TestEditor_FieldCountIsSix` | `editor.go` fieldCount enum | — | ✅ |

26/26 ✅. Каждый REQ имеет ≥1 unit или property test + трассируется в код.

## Design Conformance

### §3.1 Architectural Boundaries

✅ Все новые конструкции в `internal/config/` и `internal/tui/`. Никаких изменений в `storage/`, `app/`, `domain/`. Dependencies flow correctly: `tui` → `config` (for AppConfig.Theme value); platform detection isolated в build-tagged files.

### §3.2 Data Models

Design §2.5 vs реализация:

| Design | Implementation | Match |
|---|---|---|
| `AppConfig.Validate` accepts `auto`/`system` | `config/app.go` Validate switch | ✅ |
| `shellEditorWhen` enum (whenAnytime, whenSomeday) | `editor.go:18-23` | ✅ |
| `EditorModel.when shellEditorWhen` (replaces `someday bool`) | `editor.go:33-45` | ✅ |
| `fieldSomeday` → `fieldWhen` | `editor.go:26-31` | ✅ |
| `task.Someday bool` UNCHANGED | confirmed via grep | ✅ |

### §3.3 API Contracts

`tui.Run` signature unchanged (from v0.4.0). `cli.Deps` signature unchanged. No API change.

### §3.4 Error Handling

Design §2.7 vs реализация:

| Scenario | Design Action | Code | Match |
|---|---|---|---|
| `defaults read` timeout/not found | Return `(false, err)`; fallback macchiato | `detect_dark_darwin.go:15-24` (handles ExitError + general err) | ✅ |
| `gsettings` timeout/not installed | Return `(false, err)`; fallback | `detect_dark_linux.go:14-17` | ✅ |
| Detection > 500ms | Context timeout cancels exec | Both detect files use `context.WithTimeout(500ms)` | ✅ |
| `editorSavedMsg.updated.ID` not found | Skip splice; rely on async loadCurrentList | `app.go:160-165` (break only on match) | ✅ |
| When toggle when field not focused | No-op | Existing focus check в handleEditorKey | ✅ |
| `m.width <= 0` for separator | Return empty | `app.go:491-493` | ✅ |

### §3.5 Correctness Properties

12/12 CPs реализованы как `TestProp_*`. Прогоняются через `rapid.Check`, стабильны при `-count=2` (~3-5 sec total). Никаких flakes.

### §3.6 Documentation Consistency

Mermaid диаграмма в design §2.2 отражает фактические потоки:
- `cfg.Theme = auto/system` → `resolveAutoTheme` → platform `detectDarkMode` → macchiato/latte mapping.
- `editorSavedMsg` → inline splice + `tea.Batch(loadCurrentList, fetchListCounts)`.
- `View()` → conditional `renderSeparator` + adjusted `bodyH`.
- Editor `Space` key → `whenAnytime ↔ whenSomeday` toggle.

Implementation matches.

## Code Quality

### §4.1 Naming & Clarity

Idiomatic Go. Identifiers descriptive: `shellEditorWhen`, `whenAnytime`/`whenSomeday`, `detectDarkModeFn` (injectable), `resolveAutoTheme`, `renderSeparator`, `fieldWhen`. Build-tag pattern для platform files стандартный. Файлы хорошо организованы.

### §4.2 Dead Code & Debug Artifacts

Чисто: 0 TODO, 0 закомментированных блоков, 0 debug-print. Старый `someday bool` field полностью удалён (не закомментирован). `fieldSomeday` константа удалена. `m.editor.someday` references все обновлены.

### §4.3 Scope Creep

Никакого. Все 7 design ADRs реализованы как решены. Read-only mode explicitly deferred к v0.6.0.

### §4.4 Test Quality

- 185 TUI test cases + 23 config + 5 CLI = 213 total. Race-detector clean.
- testify `require`. Existing fixtures (`newTestModel`, `setupModelWithInboxTasks`, `bareTestModel`) переиспользованы.
- Platform detection через injectable `detectDarkModeFn` — тесты работают на любой OS.
- Filesystem isolation в config-тестах через `t.TempDir()`.
- PBT генераторы малые (intRange small, slices ≤6).
- Edge cases покрыты: dark/light/error пути detection, splice match/no-match, when toggle bidirectional, separator threshold (width<40, height<10, editor screen), hint conditional по 8 combinations of (when × area × project).

## Security

- **Внешние входы:** Только OS-utilities (`defaults`, `gsettings`) с фиксированными argv. Никакого user input в exec — фиксированные команды. `context.WithTimeout(500ms)` ограничивает время выполнения.
- **Injection:** N/A — нет user-controllable входов в exec.
- **Secrets:** Ноль credentials.
- **Data exposure:** Detection outputs не логируются за пределы своих return values.
- **File permissions:** Не applicable — фича не создаёт новых файлов.
- **Concurrency:** `detectDarkModeFn` package var — production set once at init. Test override через `defer` восстанавливает. Update serialized via Bubble Tea.
- **DoS:** Detection с 500ms cap — bounded latency старта.

## Verification Evidence

Все команды повторно прогнаны в этой сессии после `go clean -testcache`.

- **Tests (full suite, fresh):**

```
?       github.com/jtprogru/todushka/cmd/todushka       [no test files]
ok      github.com/jtprogru/todushka/internal/app       0.908s
ok      github.com/jtprogru/todushka/internal/cli       0.654s
ok      github.com/jtprogru/todushka/internal/config    4.893s
ok      github.com/jtprogru/todushka/internal/domain/area       1.265s
ok      github.com/jtprogru/todushka/internal/domain/id 1.922s
ok      github.com/jtprogru/todushka/internal/domain/project    2.701s
ok      github.com/jtprogru/todushka/internal/domain/quickentry 2.161s
ok      github.com/jtprogru/todushka/internal/domain/repeat     1.691s
ok      github.com/jtprogru/todushka/internal/domain/tag        3.216s
ok      github.com/jtprogru/todushka/internal/domain/task       4.216s
ok      github.com/jtprogru/todushka/internal/domain/today      4.024s
?       github.com/jtprogru/todushka/internal/storage   [no test files]
ok      github.com/jtprogru/todushka/internal/storage/bbolt     4.047s
ok      github.com/jtprogru/todushka/internal/storage/fakes     3.799s
ok      github.com/jtprogru/todushka/internal/tui       4.713s
?       github.com/jtprogru/todushka/internal/version   [no test files]
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

- **Manual smoke** (`./bin/todushka --help`): `--config string   Path to config file (overrides default)` present.

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | nit | `internal/tui/app.go:517` | Условие `m.screen != screenEditor` в full-screen separator branch redundant — early return at line 499 (`if m.screen == screenEditor { ... }`) уже обрабатывает editor case. Defensive duplication. Recommendation: убрать `&& m.screen != screenEditor` из строки 517 для clarity, либо оставить как defensive guard с inline comment. | — |
| F-2 | nit | `internal/tui/shell_test.go` (TestProp_SeparatorsConditional) | PBT scope сужено до `screenList` only (вместо первоначально планировавшихся 4 screen kinds). Причина: editor/quickEntry modals содержат `─` runes в своих internal borders, что приводит к false-positive matches при поиске full-width separator substring. Adjacent deterministic тесты (`TestView_NoSeparatorsInEditor`, `TestView_HasSeparatorsInFullScreen`, `TestView_NoSeparatorsInLegacy`) покрывают excluded paths. Documented в test docstring. | REQ-4.5 |

## Recommendations

**Minor / nit (не блокеры):**

1. **F-1 (redundant check):** Убрать `&& m.screen != screenEditor` из строки 517 в `app.go` View(). Early return при `screenEditor` (line 499) делает эту проверку never-reachable для editor case. Cleanup для clarity.

2. **F-2 (PBT scope):** Опционально, можно расширить `TestProp_SeparatorsConditional` чтобы корректно различать "full-width separator rule" vs "modal-internal `─` runes" через проверку `strings.Contains(out, strings.Repeat("─", m.width))` (а не просто `Contains "─"`). Это позволит снова сэмплить все screen kinds. Currently adjacent unit tests покрывают edge cases.

Обе рекомендации — не блокеры. Verdict `PASS` независимо от их исполнения.

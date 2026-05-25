# Implementation Report: UX Polish (v0.5.0)

## Scope

Read-only режим (item #4 из original запроса) **deferred** к v0.6.0 как отдельный feature pipeline `readonly-mode`. Этот документ покрывает только items 1, 2, 3, 5.

## Summary

Реализованы 4 UX улучшения:

1. **Theme auto-detect** — `config.theme = "auto"` или `"system"` запускает platform-specific детектор (macOS `defaults read`, Linux `gsettings`, иначе fallback dark) с 500ms timeout. NO_COLOR имеет absolute precedence.
2. **Instant refresh after edit** — `editorSavedMsg` inline-splices обновлённую задачу в `m.tasks` для мгновенного visual update + параллельный `tea.Batch(loadCurrentList, fetchListCounts)` для async refresh sort/counts/caches.
3. **Anytime toggle в editor** — поле `someday bool` заменено на `when shellEditorWhen` enum (Anytime|Someday). View рендерит radio-style `[•] Anytime / [ ] Someday`. Space toggles. Mapping: `Anytime → task.Someday=false`, `Someday → task.Someday=true`. Inline hint при отсутствии area/project.
4. **Section borders** — тонкие `─` separators между header/body и body/footer в full-screen mode. Editor и legacy mode — без separators.

Все 8 tasks выполнены. 26 REQs покрыты unit + property тестами; 12 CPs покрыты PBT.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format check:** `gofmt -l internal/ cmd/` (clean)

## Task Execution

- [x] **T-1** Baseline preservation — 174 tests confirmed green
- [x] **T-2** Config validation + platform detection foundation — 3 platform files + theme_resolve.go + 5 tests
- [x] **T-3** Theme integration (auto/system handling in selectThemeFromConfig) — 4 tests
- [x] **T-4** Editor When refactor — `someday bool` → `when shellEditorWhen` enum; radio-style View; Space toggle; hint. 7 new tests in `editor_test.go`. **0 existing tests broken.**
- [x] **T-5** Edit splice in editorSavedMsg — inline splice + batched refresh Cmd. 4 new tests
- [x] **T-6** Section separators — `renderSeparator` helper + View() integration + bodyH adjusted. 7 new tests
- [x] **T-7** Property-based tests batch — 12 PBTs (one per CP). Stable across `-count=2`
- [x] **T-8** GATE — all verification checks PASS

## Final Verification

- **Tests (full suite, fresh):**

```
?       github.com/jtprogru/todushka/cmd/todushka       [no test files]
ok      github.com/jtprogru/todushka/internal/app       0.522s
ok      github.com/jtprogru/todushka/internal/cli       0.911s
ok      github.com/jtprogru/todushka/internal/config    5.321s
ok      github.com/jtprogru/todushka/internal/domain/area       4.138s
ok      github.com/jtprogru/todushka/internal/domain/id 3.787s
ok      github.com/jtprogru/todushka/internal/domain/project    3.424s
ok      github.com/jtprogru/todushka/internal/domain/quickentry 1.262s
ok      github.com/jtprogru/todushka/internal/domain/repeat     2.029s
ok      github.com/jtprogru/todushka/internal/domain/tag        3.074s
ok      github.com/jtprogru/todushka/internal/domain/task       2.736s
ok      github.com/jtprogru/todushka/internal/domain/today      1.671s
?       github.com/jtprogru/todushka/internal/storage   [no test files]
ok      github.com/jtprogru/todushka/internal/storage/bbolt     5.030s
ok      github.com/jtprogru/todushka/internal/storage/fakes     2.389s
ok      github.com/jtprogru/todushka/internal/tui       5.086s
?       github.com/jtprogru/todushka/internal/version   [no test files]
```

185 TUI test cases + 23 config + 5 CLI = 213 total. Race detector clean.

- **Build:**

```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

`bin/todushka --help` shows `--config string   Path to config file (overrides default)`.

- **Lint:**

```
task: [lint] golangci-lint run
0 issues.
```

- **Format check (`gofmt -l internal/ cmd/`):**

```
(empty — all files gofmt-clean)
```

## Files Changed

### Created

- `internal/tui/detect_dark_darwin.go` — macOS `defaults read -g AppleInterfaceStyle` detection
- `internal/tui/detect_dark_linux.go` — Linux `gsettings get org.gnome.desktop.interface color-scheme` detection
- `internal/tui/detect_dark_other.go` — fallback `(false, nil)` for unsupported platforms
- `internal/tui/theme_resolve.go` — `detectDarkModeFn` injection + `resolveAutoTheme()`
- `internal/tui/theme_resolve_test.go` — 5 unit tests + 3 PBT
- `internal/tui/editor_test.go` — 7 unit tests + 3 PBT for editor When refactor

### Modified

- `internal/config/app.go` — `Validate()` accepts `"auto"`/`"system"` as valid Theme
- `internal/config/app_test.go` — 2 new tests
- `internal/tui/run.go` — `selectThemeFromConfig` handles auto/system with NO_COLOR precedence
- `internal/tui/editor.go` — `shellEditorWhen` enum; `fieldSomeday` → `fieldWhen`; `someday bool` → `when shellEditorWhen`; ApplyAndSave mapping; radio-style View + hint
- `internal/tui/app.go` — Editor space handler updated; `editorSavedMsg` inline splice + batched refresh; `renderSeparator` helper; View() full-screen with separators + editor exclusion
- `internal/tui/app_test.go` — 4 edit-splice tests + 3 PBT
- `internal/tui/shell_test.go` — 7 separator tests + 3 PBT

## Notes

- **HeadingGet limitation** (из предыдущей фичи dual-pane-layout) остаётся открытым; out of scope for v0.5.0.
- **Read-only mode** deferred к v0.6.0 как явная decision из explore phase. Branch `feature/ux-polish-and-readonly` будет shipped as v0.5.0; readonly начнётся в новом branch.
- **Backward compat preserved:** все 174 baseline tests passing. Editor field count == 6 (renamed `fieldSomeday → fieldWhen`, not removed). `task.Someday bool` domain field unchanged.
- **macOS smoke pending:** `bin/todushka --help` OK. Interactive smoke невозможен в текущей среде (требует терминал).
- **Coverage Matrix verified:** 26 REQs → ≥1 test each; 12 CPs → 1 PBT each.

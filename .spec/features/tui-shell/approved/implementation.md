# Implementation Report: TUI Shell

## Summary

Полностью реализованы 4 пункта пользовательского запроса:

1. **Full-screen rendering** — `m.height` отслеживается, `View()` клампит body к `m.height - header_h - footer_h` через `lipgloss.Style.Height()` при достаточных размерах (≥40×10). Legacy fallback для маленьких терминалов.

2. **Header counts + indicators** — каждый сегмент `(N) Name [Count]` с colour-кодированием (digit accent, label subtext, count dim); активный — inverted background; compact mode `(N)I[Count]` при `m.width < 80`. Counts cached in Model, refresh batch'ем при каждом `tasksLoadedMsg`.

3. **zellij-style status line** — mode chip `-- NORMAL --` / `-- FILTER --` / `-- SELECT --` / `-- CONFIRM --` / `-- EDITOR --` / `-- HELP --` слева; контекстные key hints разделённые ` │ `. Mode detection через приоритетную цепочку HELP > EDITOR > CONFIRM > FILTER > SELECT > NORMAL.

4. **Config + env vars + CLI flag** — YAML `$XDG_CONFIG_HOME/todushka/config.yaml` (auto-create при первом запуске с inline-commented defaults); 5 настроек: `theme`, `dual_pane_min_width`, `list_pane_share`, `bulk_confirm_threshold`, `notes_max_lines`. Env vars `TODUSHKA_*` override. CLI flag `--config <path>` через persistent root flag. Precedence: flag > env > file > defaults.

Все 9 tasks выполнены. 28 REQ покрыты unit + property тестами; 14 CPs покрыты PBT.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format check:** `gofmt -l internal/ cmd/` (clean)

## Task Execution

- [x] **T-1** Baseline preservation — 117 TUI tests + lint 0 issues подтверждены на baseline
- [x] **T-2** Config foundation — `AppConfig`, `Defaults`, `Validate`, `Load`, `ResolvePath`, `applyEnv`, auto-create. 13 tests; yaml.v3 promoted to direct dep
- [x] **T-3** Model расширение + signature change — `m.height`, `m.config`, `m.listCounts`; `NewModel(svc, theme, cfg)`; 4 constants → `m.config.X`. Все callers обновлены (test helpers, main, run, filter_test, details_test, bulk_test). 117 prior tests + 1 new (`TestWindowSize_BothAxesTracked`) passing
- [x] **T-4** Mode + footer — `shellMode` enum, `currentMode`, `modeKeyHints`, новый `viewFooter` в `shell.go`; старый удалён из `app.go`. 11 new tests; 1 PBT обновлён (`TestProp_FooterContainsNewKeys` теперь mode-aware)
- [x] **T-5** Header refactor — `renderHeaderSegment`, новый `viewHeader` в `shell.go`; старый удалён из `app.go`. 5 new tests
- [x] **T-6** Counts + full-screen clamp — `countsLoadedMsg`, `fetchListCounts` Cmd, integration в `tasksLoadedMsg` через `tea.Batch`, `View()` рефактор с height clamp. 5 new tests
- [x] **T-7** CLI `--config` flag — persistent flag на root, `PersistentPreRunE` загружает config через `config.Load`, передаёт в `LaunchTUI`. `Deps.Config` field, `Deps.LaunchTUI` signature изменён на `func(svc, cfg)`. `tui.Run(svc, cfg)` signature обновлён. CLI test fixtures обновлены
- [x] **T-8** Property-based tests batch — 14 PBT для всех CPs (9 в `shell_test.go`, 5 в `loader_test.go`)
- [x] **T-9** GATE — все verification checks PASS

## Final Verification

- **Tests (full suite, fresh):**

```
?       github.com/jtprogru/todushka/cmd/todushka       [no test files]
ok      github.com/jtprogru/todushka/internal/app
ok      github.com/jtprogru/todushka/internal/cli       1.411s
ok      github.com/jtprogru/todushka/internal/config    1.762s
ok      github.com/jtprogru/todushka/internal/domain/area
ok      github.com/jtprogru/todushka/internal/domain/id
ok      github.com/jtprogru/todushka/internal/domain/project
ok      github.com/jtprogru/todushka/internal/domain/quickentry
ok      github.com/jtprogru/todushka/internal/domain/repeat
ok      github.com/jtprogru/todushka/internal/domain/tag
ok      github.com/jtprogru/todushka/internal/domain/task
ok      github.com/jtprogru/todushka/internal/domain/today
?       github.com/jtprogru/todushka/internal/storage   [no test files]
ok      github.com/jtprogru/todushka/internal/storage/bbolt
ok      github.com/jtprogru/todushka/internal/storage/fakes
ok      github.com/jtprogru/todushka/internal/tui       6.157s
?       github.com/jtprogru/todushka/internal/version   [no test files]
```

148 TUI test cases + 21 config test cases pass (subtests counted). Race detector clean.

- **Build:**

```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

`bin/todushka --help` показывает `--config string   Path to config file (overrides default)`.

- **Lint:**

```
task: [lint] golangci-lint run
0 issues.
```

- **Format check:**

```
(empty — all files gofmt-clean)
```

## Files Changed

### Created

- `internal/config/app.go` (63 lines) — `AppConfig`, `Defaults`, `Validate`
- `internal/config/app_test.go` (38 lines) — 3 unit tests
- `internal/config/loader.go` (135 lines) — `ResolvePath`, `Load`, `loadFromFile`, `createDefaultConfig`, `defaultConfigYAML`, `applyEnv`
- `internal/config/loader_test.go` (258 lines) — 10 unit tests + 5 PBT, `mockEnv` helper
- `internal/tui/shell.go` (220 lines) — `shellMode` enum, `currentMode`, `modeLabel`, `modeKeyHints`, `viewFooter`, `viewHeader`, `renderHeaderSegment`, `fetchListCounts`, constants
- `internal/tui/shell_test.go` (403 lines) — 20 unit + 9 PBT
- `internal/tui/testdata/rapid/` — rapid regression seeds

### Modified

- `internal/tui/app.go` — Model fields (height, config, listCounts), NewModel signature, WindowSizeMsg tracks both axes, `tasksLoadedMsg` returns `tea.Batch(fetchNameCache, fetchListCounts)`, new `countsLoadedMsg` case, `View()` рефактор с height clamp, удаление старых `viewHeader`/`viewFooter`
- `internal/tui/app_test.go` — `newTestModel`/`newTestModelWithService`/`setupRapidModel`/`bareTestModel` обновлены с `config.Defaults()`; `TestProp_FooterContainsNewKeys` mode-aware refactor; `TestWindowSize_BothAxesTracked`
- `internal/tui/details.go` — удаление 3 constants → `m.config.X`; `paneWidths(m)` signature
- `internal/tui/details_test.go` — обновление тестов под новые сигнатуры
- `internal/tui/filter_test.go` — добавлен `config` импорт + обновлены NewModel вызовы
- `internal/tui/bulk.go` — удаление `const bulkConfirmThreshold`, использование `m.config.BulkConfirmThreshold`
- `internal/tui/bulk_test.go` — обновление под `config.Defaults().BulkConfirmThreshold`
- `internal/tui/msgs.go` — `countsLoadedMsg` тип
- `internal/tui/run.go` — `Run(svc, cfg)` signature, `selectThemeFromConfig` helper
- `internal/cli/deps.go` — `Config` field, `LaunchTUI` signature `func(svc, cfg)`
- `internal/cli/root.go` — `--config` persistent flag, `PersistentPreRunE` загружает config
- `internal/cli/cli_test.go` — обновление fixtures под новый LaunchTUI signature
- `go.mod` — `gopkg.in/yaml.v3 v3.0.1` promoted to direct dep

## Notes

- **PersistentPreRunE resilience:** если `config.ResolvePath` падает (e.g. в тестах без `HOME`), tихо fallback к `config.Defaults()`. Это позволяет CLI-тестам работать без специальной env setup.
- **No commits made** — task plan не меняется; coordinator (вы) делает finalization (atomic commits + merge + tag).
- **Coverage Matrix verified:** 28 REQs → ≥1 test each; 14 CPs → 1 PBT each.
- **Backward compat:** все T-1 preservation tests passing; bulk threshold tests passing с `config.Defaults().BulkConfirmThreshold == 5` (unchanged behavior).
- **Lint resilience:** linter переформатировал `deps.go` aligned-comment поля — это OK, intentional.
- **Manual smoke pending:** `./bin/todushka` interactive TUI требует terminal — не запускалось в этом отчёте. Compile + `--help` smoke OK.

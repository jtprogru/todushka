# TUI Shell — Task Plan

## Preamble

### Work Type Classification

**Pure feature** with significant **preservation surface**: 117 existing TUI tests must remain valid после изменения `NewModel` signature и header/footer форматов. T-1 фиксирует невидимые-снаружи аспекты (single-pane при `m.height == 0`, существующая bulk-confirm threshold=5 behavior). Изменения в `viewHeader`/`viewFooter` форматах ожидаемо потребуют обновления существующих assertion'ов на тексты — это allowed.

### Test Style Source

**Tier 2** — adjacent tests
- **Reference unit tests:** `internal/tui/app_test.go`, `internal/tui/details_test.go`, `internal/tui/filter_test.go`, `internal/tui/bulk_test.go`. Установленные fixtures: `newTestModel(t)`, `newTestModelWithService(t)`, `setupModelWithInboxTasks(t, ...)`, `bareTestModel()`.
- **Reference property tests:** `internal/tui/*_test.go` уже содержат `rapid.Check(t, func(rt *rapid.T) {...})` patterns.
- **Config tests:** новый pattern с `t.TempDir()` для isolated filesystem.
- **Key patterns:** testify `require`, прямой `Update(tea.Msg)` dispatch, table-driven, rapid PBT.

### Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test race  | `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

### Coverage Matrix

| Requirement | Task(s) | CP |
|-------------|---------|----|
| REQ-1.1 m.height tracked | T-3 | CP-3 |
| REQ-1.2 full-screen clamp | T-6 | CP-2 |
| REQ-1.3 small terminal fallback | T-1, T-6 | CP-2 |
| REQ-1.4 editor full-pane | T-6 | CP-2 |
| REQ-2.1 header `(N) Name [Count]` | T-5 | CP-5 |
| REQ-2.2 color coding (digit/label/count) | T-5 | CP-4 |
| REQ-2.3 active inverted | T-5 | CP-4 |
| REQ-2.4 compact mode `(N)I[Count]` | T-5 | CP-5 |
| REQ-2.5 counts refresh on tasksLoadedMsg | T-6 | CP-6 |
| REQ-2.6 countsLoadedMsg merge | T-6 | CP-6 |
| REQ-2.7 unknown count `[?]` | T-5 | CP-5 |
| REQ-2.8 empty list `[0]` | T-5 | CP-5 |
| REQ-3.1 footer structure | T-4 | CP-7 |
| REQ-3.2 mode priority | T-4 | CP-1 |
| REQ-3.3 mode chip styled | T-4 | CP-7 |
| REQ-3.4 NORMAL hints | T-4 | CP-8 |
| REQ-3.5 FILTER hints | T-4 | CP-8 |
| REQ-3.6 SELECT hints | T-4 | CP-8 |
| REQ-3.7 CONFIRM hints | T-4 | CP-8 |
| REQ-3.8 EDITOR hints | T-4 | CP-8 |
| REQ-3.9 HELP hints | T-4 | CP-8 |
| REQ-3.10 status preserved | T-4 | CP-7 |
| REQ-4.1 path precedence | T-2 | CP-9, CP-13 |
| REQ-4.2 auto-create | T-2 | CP-11 |
| REQ-4.3 5 settings parsed | T-2 | CP-9 |
| REQ-4.4 unknown fields ignored | T-2 | CP-14 |
| REQ-4.5 invalid values warn | T-2 | CP-10 |
| REQ-4.6 env override | T-2 | CP-9 |
| REQ-4.7 invalid env fallback | T-2 | CP-10 |
| REQ-4.8 config injected into Model | T-3 | CP-9 |
| REQ-5.1 zero-config visually compatible | T-1, T-9 | — |
| REQ-5.2 NewModel signature change | T-1, T-3 | — |
| REQ-5.3 existing tests pass | T-9 | — |

28 REQs → 9 tasks → 14 CPs. Каждый REQ покрыт ≥1 task. Каждый CP покрыт property-тестом в T-8.

---

## Task Order

```
T-1 GREEN (preservation: zero-config behavior + signature compat)
  → T-2 CODE (Config foundation: AppConfig + Loader + auto-create)
    → T-3 CODE (Model расширение + NewModel signature + constants → config refs)
      → T-4 CODE (Mode detection + zellij-style footer)
      → T-5 CODE (Header refactor: segments + counts placeholders)
        → T-6 CODE (Counts fetching + full-screen clamp + Update integration)
          → T-7 CODE (CLI --config flag + main.go wiring)
            → T-8 GREEN (Property-based tests batch — 14 CPs)
              → T-9 GATE (Checkpoint)
```

T-4 и T-5 могут разрабатываться параллельно (оба в `shell.go`, разные функции), но представлены sequentially для четкости.

---

## Task: T-1 — Write preservation tests for zero-config compatibility

*_Requirements: REQ-5.1, REQ-5.2_*
*_Test_Style: Tier 2 (`internal/tui/app_test.go`)_*
*_Complexity: standard_*

GOAL: Зафиксировать current behavior до изменения `NewModel` signature и формата header/footer. Эти тесты будут обновлены в последующих tasks (тестируют существующее поведение, которое поменяется), но прямо сейчас они проходят на baseline.

NOTE: Это не классические preservation tests — мы добавляем 2 теста, фиксирующие что (a) `task test` зелёный на baseline (sanity), (b) `Model.width == 0` приводит к single-pane (это уже покрыто в dual-pane preservation, но важно подтвердить, что новый full-screen clamp не сломает этот invariant).

DO NOT: Изменять production-код в этой задаче.

Subtasks:

- [ ] 1. В `internal/tui/app_test.go` добавить `TestTUI_BaselineSuiteGreen`: пустой тест с комментарием — sanity placeholder перед началом изменений (`t.Log("baseline")`; этот тест существует для документирования момента, до которого код был стабилен). Опционально пропустить если кажется ненужным.

- [ ] 2. Запустить `task test-race` — все 117 существующих тестов проходят. Зафиксировать output как evidence baseline.

After all subtasks: Run `task lint` (no changes expected).

---

## Task: T-2 — Implement Config foundation

*_Requirements: REQ-4.1, REQ-4.2, REQ-4.3, REQ-4.4, REQ-4.5, REQ-4.6, REQ-4.7_*
*_Preservation: existing `internal/config` package tests + `internal/cli` tests_*
*_Test_Style: Tier 2 + new `t.TempDir()` pattern for filesystem isolation_*
*_Complexity: complex_*

GOAL: Создать `AppConfig` тип, `Defaults`, `Validate`, `Load`, `ResolvePath` функции с поддержкой YAML, env override, CLI flag (CLI integration — отдельно в T-7), auto-create при отсутствии файла.

IMPORTANT: Сделать `gopkg.in/yaml.v3` direct dependency: `go get gopkg.in/yaml.v3` после первого использования в коде.

Subtasks:

- [ ] 1. Создать `internal/config/app.go` с:
  ```go
  package config

  type AppConfig struct {
      Theme                string  `yaml:"theme"`
      DualPaneMinWidth     int     `yaml:"dual_pane_min_width"`
      ListPaneShare        float64 `yaml:"list_pane_share"`
      BulkConfirmThreshold int     `yaml:"bulk_confirm_threshold"`
      NotesMaxLines        int     `yaml:"notes_max_lines"`
  }

  // Defaults returns the built-in fallback values.
  func Defaults() AppConfig {
      return AppConfig{
          Theme:                "macchiato",
          DualPaneMinWidth:     100,
          ListPaneShare:        0.45,
          BulkConfirmThreshold: 5,
          NotesMaxLines:        8,
      }
  }

  // Validate returns a corrected AppConfig (invalid fields replaced with
  // defaults) AND a slice of warning messages describing each correction.
  func (c AppConfig) Validate() (AppConfig, []string) {
      def := Defaults()
      var warns []string
      switch c.Theme {
      case "macchiato", "latte", "mono", "":
          // ok
      default:
          warns = append(warns, "invalid theme '"+c.Theme+"', using '"+def.Theme+"'")
          c.Theme = def.Theme
      }
      if c.Theme == "" {
          c.Theme = def.Theme
      }
      if c.DualPaneMinWidth < 40 {
          if c.DualPaneMinWidth != 0 {
              warns = append(warns, "dual_pane_min_width must be >= 40, using default")
          }
          c.DualPaneMinWidth = def.DualPaneMinWidth
      }
      if c.ListPaneShare <= 0 || c.ListPaneShare >= 1 {
          if c.ListPaneShare != 0 {
              warns = append(warns, "list_pane_share must be in (0, 1), using default")
          }
          c.ListPaneShare = def.ListPaneShare
      }
      if c.BulkConfirmThreshold < 1 {
          if c.BulkConfirmThreshold != 0 {
              warns = append(warns, "bulk_confirm_threshold must be >= 1, using default")
          }
          c.BulkConfirmThreshold = def.BulkConfirmThreshold
      }
      if c.NotesMaxLines < 1 {
          if c.NotesMaxLines != 0 {
              warns = append(warns, "notes_max_lines must be >= 1, using default")
          }
          c.NotesMaxLines = def.NotesMaxLines
      }
      return c, warns
  }
  ```
  — `task test`

- [ ] 2. Создать `internal/config/app_test.go` с тестами:
  ```go
  package config

  import (
      "testing"
      "github.com/stretchr/testify/require"
  )

  func TestDefaults_AreValid(t *testing.T) {
      c, warns := Defaults().Validate()
      require.Empty(t, warns)
      require.Equal(t, "macchiato", c.Theme)
      require.Equal(t, 100, c.DualPaneMinWidth)
      require.InDelta(t, 0.45, c.ListPaneShare, 1e-9)
      require.Equal(t, 5, c.BulkConfirmThreshold)
      require.Equal(t, 8, c.NotesMaxLines)
  }

  func TestValidate_ThemeFallback(t *testing.T) {
      c, warns := AppConfig{Theme: "unknown"}.Validate()
      require.Equal(t, "macchiato", c.Theme)
      require.Len(t, warns, 1)
  }

  func TestValidate_NumericRanges(t *testing.T) {
      c, warns := AppConfig{
          Theme:                "macchiato",
          DualPaneMinWidth:     10,
          ListPaneShare:        1.5,
          BulkConfirmThreshold: -1,
          NotesMaxLines:        -5,
      }.Validate()
      require.Len(t, warns, 4)
      require.Equal(t, 100, c.DualPaneMinWidth)
      require.InDelta(t, 0.45, c.ListPaneShare, 1e-9)
      require.Equal(t, 5, c.BulkConfirmThreshold)
      require.Equal(t, 8, c.NotesMaxLines)
  }
  ```
  — `task test`

- [ ] 3. Создать `internal/config/loader.go` со скелетом:
  ```go
  package config

  import (
      "errors"
      "fmt"
      "os"
      "path/filepath"
      "strconv"

      "gopkg.in/yaml.v3"
  )

  // ResolvePath returns the path to use for the config file.
  // Precedence: flag (if non-empty) > env("TODUSHKA_CONFIG") > $XDG_CONFIG_HOME/todushka/config.yaml
  func ResolvePath(env Env, flag string) (string, error) {
      if flag != "" {
          return filepath.Abs(flag)
      }
      if env == nil {
          env = OSEnv
      }
      if p := env("TODUSHKA_CONFIG"); p != "" {
          return filepath.Abs(p)
      }
      dir, err := resolveDir(env, "XDG_CONFIG_HOME", filepath.Join(".config"))
      if err != nil {
          return "", err
      }
      return filepath.Join(dir, "config.yaml"), nil
  }

  // Load resolves the config path and returns the final AppConfig
  // (after precedence cascade + validation) and any warnings encountered.
  func Load(path string, env Env) (AppConfig, []string, error) {
      if env == nil {
          env = OSEnv
      }
      cfg, warns, err := loadFromFile(path)
      if err != nil {
          // file read/parse error → warn, return defaults
          warns = append(warns, fmt.Sprintf("config: %v; using defaults", err))
          cfg = Defaults()
      }
      cfg, envWarns := applyEnv(cfg, env)
      warns = append(warns, envWarns...)
      cfg, validateWarns := cfg.Validate()
      warns = append(warns, validateWarns...)
      return cfg, warns, nil
  }

  // loadFromFile reads `path`. If file does not exist, auto-creates it with
  // defaults + inline comments and returns Defaults().
  func loadFromFile(path string) (AppConfig, []string, error) {
      data, err := os.ReadFile(path)
      if err != nil {
          if errors.Is(err, os.ErrNotExist) {
              if cerr := createDefaultConfig(path); cerr != nil {
                  return Defaults(), []string{fmt.Sprintf("could not create %s: %v", path, cerr)}, nil
              }
              return Defaults(), nil, nil
          }
          return Defaults(), nil, err
      }
      var cfg AppConfig
      if err := yaml.Unmarshal(data, &cfg); err != nil {
          return Defaults(), nil, fmt.Errorf("parse %s: %w", path, err)
      }
      // Merge with defaults: any zero-valued field is replaced by default below
      // in applyEnv → Validate. This keeps semantic "absent in YAML == use default".
      return cfg, nil, nil
  }

  func createDefaultConfig(path string) error {
      if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
          return err
      }
      content := defaultConfigYAML()
      return os.WriteFile(path, []byte(content), 0600)
  }

  func defaultConfigYAML() string {
      return `# todushka configuration. See https://github.com/jtprogru/todushka
  # for documentation. Edit values below to customize.

  # Color theme: macchiato | latte | mono
  theme: macchiato

  # Minimum terminal width (columns) to activate dual-pane layout.
  dual_pane_min_width: 100

  # Fraction of width allocated to the list pane in dual-pane mode (0.0 - 1.0).
  list_pane_share: 0.45

  # Minimum number of selected tasks before a bulk operation requires confirmation.
  bulk_confirm_threshold: 5

  # Maximum lines of task notes displayed in the details pane.
  notes_max_lines: 8
  `
  }

  func applyEnv(cfg AppConfig, env Env) (AppConfig, []string) {
      var warns []string
      if v := env("TODUSHKA_THEME"); v != "" {
          cfg.Theme = v
      }
      if v := env("TODUSHKA_DUAL_PANE_MIN_WIDTH"); v != "" {
          if n, err := strconv.Atoi(v); err == nil {
              cfg.DualPaneMinWidth = n
          } else {
              warns = append(warns, "TODUSHKA_DUAL_PANE_MIN_WIDTH="+v+" not an integer")
          }
      }
      if v := env("TODUSHKA_LIST_PANE_SHARE"); v != "" {
          if f, err := strconv.ParseFloat(v, 64); err == nil {
              cfg.ListPaneShare = f
          } else {
              warns = append(warns, "TODUSHKA_LIST_PANE_SHARE="+v+" not a float")
          }
      }
      if v := env("TODUSHKA_BULK_CONFIRM_THRESHOLD"); v != "" {
          if n, err := strconv.Atoi(v); err == nil {
              cfg.BulkConfirmThreshold = n
          } else {
              warns = append(warns, "TODUSHKA_BULK_CONFIRM_THRESHOLD="+v+" not an integer")
          }
      }
      if v := env("TODUSHKA_NOTES_MAX_LINES"); v != "" {
          if n, err := strconv.Atoi(v); err == nil {
              cfg.NotesMaxLines = n
          } else {
              warns = append(warns, "TODUSHKA_NOTES_MAX_LINES="+v+" not an integer")
          }
      }
      return cfg, warns
  }
  ```
  Также добавить `XDG_CONFIG_HOME` handling в `resolveDir` через existing `internal/config/paths.go`. Если `resolveDir` существующая функция работает с `xdgKey` параметром (видели в Read above) — переиспользовать as is. — `task test`

- [ ] 4. Создать `internal/config/loader_test.go` с тестами:
  - `TestLoad_FileMissingCreatesDefault`: использовать `t.TempDir()`, `Load(filepath.Join(tmp, "config.yaml"), envFn)` — assert файл создан + cfg = Defaults().
  - `TestLoad_FileValidParses`: pre-write YAML `theme: latte` + поля → парсится корректно.
  - `TestLoad_EnvOverridesFile`: env returns `TODUSHKA_NOTES_MAX_LINES=20` + file `notes_max_lines: 8` → result 20.
  - `TestLoad_FlagOverridesEnv`: через ResolvePath check (flag wins).
  - `TestLoad_UnknownYAMLFieldsIgnored`: YAML `feature_x: true\ntheme: macchiato` → no warns.
  - `TestLoad_MalformedYAMLReturnsDefaults`: YAML `not valid: : ::` → defaults + warn.
  - `TestLoad_InvalidEnvFallsBackToFile`: env `TODUSHKA_NOTES_MAX_LINES=abc` + file value → file used + warn.
  - `TestResolvePath_FlagAbsolute`: `ResolvePath(env, "/abs/path.yaml")` → `/abs/path.yaml`.
  - `TestResolvePath_EnvFallback`: env returns `TODUSHKA_CONFIG=/env/path` → that's used.
  - `TestResolvePath_XDGDefault`: empty flag/env → XDG path.
  Используйте helper `mockEnv(m map[string]string) Env` для encapsulation.
  — `task test`

- [ ] 5. Запустить `go get gopkg.in/yaml.v3` чтобы сделать direct dep (обновляет `go.mod`). Запустить `task test` + `task lint`. — `task test-race && task lint`

After all subtasks: All new config tests pass. Existing tests must keep passing.

---

## Task: T-3 — Расширить Model + новая NewModel signature + constants → config refs

*_Requirements: REQ-1.1, REQ-4.8_*
*_Preservation: T-1 tests + 117 existing TUI tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Добавить в `Model` поля `height int`, `config config.AppConfig`, `listCounts map[listKind]int`. Сменить signature `NewModel(svc, theme) → NewModel(svc, theme, cfg)`. Заменить hardcoded constants в `details.go` и `bulk.go` на `m.config.X` references. `WindowSizeMsg` обновляет ОБА axes.

CRITICAL: Эта задача затрагивает 117 существующих тестов через signature change. Стратегия:
- Создать `newTestModel(t)` (private) хелпер, который инжектит `config.Defaults()` — большинство тестов не меняется.
- Тесты с custom AppConfig получают `newTestModelWithConfig(t, cfg)` вариант.
- `setupModelWithInboxTasks` (test helper) обновляется чтобы вызывать новый signature.

Subtasks:

- [ ] 1. В `internal/tui/app.go` добавить поля в `Model`:
  ```go
  // After existing fields (after headingNamesByID):
  height     int
  config     config.AppConfig
  listCounts map[listKind]int
  ```
  Импорт `"github.com/jtprogru/todushka/internal/config"` если ещё не присутствует. — `task test` (компиляция ломается на NewModel)

- [ ] 2. В `internal/tui/app.go` обновить `NewModel`:
  ```go
  func NewModel(svc *app.Service, theme Theme, cfg config.AppConfig) Model {
      ti := textinput.New()
      ti.Placeholder = "what to do? — tokens: #tag @today @project !YYYY-MM-DD"
      ti.CharLimit = 256
      return Model{
          service:          svc,
          keys:             DefaultKeyMap(),
          theme:            theme,
          screen:           screenList,
          activeList:       listToday,
          quickInput:       ti,
          selected:         make(map[id.ID]struct{}),
          tagNamesByID:     make(map[id.ID]string),
          areaNamesByID:    make(map[id.ID]string),
          projectNamesByID: make(map[id.ID]string),
          headingNamesByID: make(map[id.ID]string),
          listCounts:       make(map[listKind]int),
          config:           cfg,
      }
  }
  ```
  Один аргумент — `cfg config.AppConfig`. — `task test` (продолжает ломаться: все вызовы NewModel в tests)

- [ ] 3. В `internal/tui/app_test.go` обновить хелперы:
  ```go
  func newTestModel(t *testing.T) Model {
      t.Helper()
      svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
      return NewModel(svc, NewTheme(), config.Defaults())
  }

  func newTestModelWithService(t *testing.T) (Model, *app.Service) {
      t.Helper()
      svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
      return NewModel(svc, NewTheme(), config.Defaults()), svc
  }
  ```
  Имп `"github.com/jtprogru/todushka/internal/config"`. — `task test`

- [ ] 4. В `internal/tui/app_test.go` функция `bareTestModel`/`setupRapidModel` если есть — обновить аналогично. Поиск через grep: `grep -n "NewModel(" internal/tui/`. Обновить все call-sites. — `task test`

- [ ] 5. В `cmd/todushka/main.go` обновить вызов `NewModel`. Поиск: `grep -n "NewModel(" cmd/`. Передать `config.Defaults()` пока что (T-7 заменит на реальную загрузку). — `task build`

- [ ] 6. В `internal/tui/app.go` `Update` обработать `WindowSizeMsg`:
  ```go
  case tea.WindowSizeMsg:
      m.width = msg.Width
      m.height = msg.Height
      return m, nil
  ```
  — `task test`

- [ ] 7. В `internal/tui/app_test.go` добавить тест `TestWindowSize_BothAxesTracked`:
  ```go
  func TestWindowSize_BothAxesTracked(t *testing.T) {
      m := newTestModel(t)
      m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
      mm := m2.(Model)
      require.Equal(t, 120, mm.width)
      require.Equal(t, 40, mm.height)
  }
  ```
  — `task test`

- [ ] 8. В `internal/tui/details.go` удалить `const dualPaneMinWidth`/`listPaneShare`/`detailsNotesMaxLines`. Обновить:
  ```go
  func isDualPane(m Model) bool {
      if m.width < m.config.DualPaneMinWidth {
          return false
      }
      ...
  }

  func paneWidths(m Model) (int, int) {
      list := int(float64(m.width-1) * m.config.ListPaneShare)
      details := m.width - 1 - list
      return list, details
  }
  ```
  Где функция вызывается с `m.width` напрямую (внутри `viewBody` в `app.go`) — заменить на `paneWidths(m)`. — `task test`

- [ ] 9. В `internal/tui/details.go` `viewDetails` функция:
  ```go
  lines = append(lines, wrapAndTruncate(t.Notes, width, m.config.NotesMaxLines))
  ```
  Удалить старый `const detailsNotesMaxLines = 8`. — `task test`

- [ ] 10. В `internal/tui/bulk.go` удалить `const bulkConfirmThreshold = 5`. В `dispatch`:
  ```go
  func dispatch(m Model, action bulkAction) (Model, tea.Cmd) {
      if len(m.selected) == 0 {
          return m, perCursorCmd(m, action)
      }
      ids := selectionIDs(m)
      if len(ids) < m.config.BulkConfirmThreshold {
          return m, runBulk(m.service, action, ids)
      }
      m.confirm = &confirmState{action: action, ids: ids}
      return m, nil
  }
  ```
  — `task test`

- [ ] 11. В `internal/tui/app.go` `viewBody` обновить вызовы `paneWidths(m.width)` → `paneWidths(m)`. — `task test`

- [ ] 12. Запустить `task test-race && task lint`. Все 117 тестов + new T-3 тест должны проходить.

After all subtasks: Backward compat сохранён через `config.Defaults()` injection. Тесты passing.

---

## Task: T-4 — Implement Mode detection + zellij-style footer

*_Requirements: REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4, REQ-3.5, REQ-3.6, REQ-3.7, REQ-3.8, REQ-3.9, REQ-3.10_*
*_Preservation: T-1, T-3 tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Создать `internal/tui/shell.go` с `shellMode` enum, `currentMode`, `modeLabel`, `modeKeyHints`, новый `viewFooter`. Заменить старый `viewFooter` в `app.go` (move logic в shell.go).

IMPORTANT: Существующие тесты `TestFooter_IncludesNewHints`, `TestSelection_StatusBarCounter` etc. — могут потребовать обновления text-assertion'ов. Обновите чтобы они matchед новому формату (`-- NORMAL --`, `key: hint` separated by `│`).

Subtasks:

- [ ] 1. Создать `internal/tui/shell.go`:
  ```go
  package tui

  import (
      "fmt"
      "strings"

      "github.com/charmbracelet/lipgloss"
  )

  type shellMode int

  const (
      modeNormal shellMode = iota
      modeFilter
      modeSelect
      modeConfirm
      modeEditor
      modeHelp
  )

  func (m shellMode) modeLabel() string {
      switch m {
      case modeNormal:
          return "NORMAL"
      case modeFilter:
          return "FILTER"
      case modeSelect:
          return "SELECT"
      case modeConfirm:
          return "CONFIRM"
      case modeEditor:
          return "EDITOR"
      case modeHelp:
          return "HELP"
      }
      return "?"
  }

  // currentMode determines the active mode following priority:
  // HELP > EDITOR > CONFIRM > FILTER > SELECT > NORMAL.
  func currentMode(m Model) shellMode {
      switch {
      case m.screen == screenHelp:
          return modeHelp
      case m.screen == screenEditor:
          return modeEditor
      case m.confirm != nil:
          return modeConfirm
      case m.filtering:
          return modeFilter
      case len(m.selected) > 0:
          return modeSelect
      default:
          return modeNormal
      }
  }

  // modeKeyHints returns context-aware key hints for the given mode.
  func modeKeyHints(mode shellMode) []string {
      switch mode {
      case modeNormal:
          return []string{"/: filter", "space: select", "n: quick", "↵: edit", "c: complete", "?: help", "q: quit"}
      case modeFilter:
          return []string{"↵: save", "esc: cancel"}
      case modeSelect:
          return []string{"c/x/d/p: bulk", "*: all", "esc: clear"}
      case modeConfirm:
          return []string{"y: yes", "any: cancel"}
      case modeEditor:
          return []string{"Tab: next", "Shift+Tab: prev", "Ctrl+S: save", "esc: cancel"}
      case modeHelp:
          return []string{"?: close"}
      }
      return nil
  }
  ```
  — `task test`

- [ ] 2. В `internal/tui/shell_test.go` (new) добавить тесты:
  ```go
  func TestCurrentMode_PriorityOrder(t *testing.T) {
      m := newTestModel(t)
      // HELP wins over CONFIRM
      m.screen = screenHelp
      m.confirm = &confirmState{}
      require.Equal(t, modeHelp, currentMode(m))
      // EDITOR wins over CONFIRM
      m.screen = screenEditor
      require.Equal(t, modeEditor, currentMode(m))
      // CONFIRM wins over FILTER
      m.screen = screenList
      m.filtering = true
      require.Equal(t, modeConfirm, currentMode(m))
      // FILTER wins over SELECT
      m.confirm = nil
      m.selected[id.New()] = struct{}{}
      require.Equal(t, modeFilter, currentMode(m))
      // SELECT wins over NORMAL
      m.filtering = false
      require.Equal(t, modeSelect, currentMode(m))
      // NORMAL default
      m.selected = make(map[id.ID]struct{})
      require.Equal(t, modeNormal, currentMode(m))
  }

  func TestModeKeyHints_Normal(t *testing.T) {
      h := modeKeyHints(modeNormal)
      joined := strings.Join(h, " ")
      require.Contains(t, joined, "/: filter")
      require.Contains(t, joined, "space: select")
      require.Contains(t, joined, "q: quit")
  }
  ```
  И аналогичные для Filter, Select, Confirm, Editor, Help. — `task test`

- [ ] 3. В `internal/tui/shell.go` добавить новый `viewFooter` method:
  ```go
  func (m Model) viewFooter() string {
      mode := currentMode(m)
      chip := m.theme.Header.Render(fmt.Sprintf(" -- %s -- ", mode.modeLabel()))

      // Per-mode hints
      hints := modeKeyHints(mode)
      if mode == modeFilter {
          // Show filter query inside hints
          hints = append([]string{"Filter: " + m.filterQuery + "_"}, hints...)
      }
      if mode == modeSelect {
          hints = append(hints, fmt.Sprintf("Selected: %d", len(m.selected)))
      }

      // Join hints with " │ "
      hintsRendered := m.theme.Help.Render(strings.Join(hints, " │ "))

      // Status message (right-aligned)
      var status string
      if m.statusMsg != "" {
          if mode == modeConfirm {
              status = m.theme.StatusError.Render(m.statusMsg)
          } else {
              status = m.theme.Help.Render(m.statusMsg)
          }
      }

      left := lipgloss.JoinHorizontal(lipgloss.Top, chip, " ", hintsRendered)
      if status != "" {
          return left + "  " + status
      }
      return left
  }
  ```
  ВАЖНО: Это **замена** существующего `viewFooter` в `app.go`. Удалить старую функцию из app.go (или переименовать; зависит от того что есть). — `task test` (старые footer-тесты могут упасть на assertion'ах текста — это ОК)

- [ ] 4. В `internal/tui/app.go` удалить старый `viewFooter` (теперь в shell.go). Подтвердить через `grep -n "viewFooter" internal/tui/app.go` что осталась только использующая ссылка. — `task test`

- [ ] 5. Обновить существующие тесты в `app_test.go` и `details_test.go` которые проверяют footer text:
  - `TestSelection_StatusBarCounter`: `viewFooter` теперь содержит `Selected: 3` в mode chip section — assert по-новому: `require.Contains(t, out, "Selected: 3")`.
  - `TestFooter_IncludesNewHints`: assert `require.Contains(t, out, "/: filter")` + `require.Contains(t, out, "space: select")`.
  - `TestTUI_ErrorMsgUpdatesStatusBar`: статус всё ещё в выходе — пересмотрите если тест проверяет position.
  - Любые другие тесты которые искали старый формат "?: help  ⇥: ..." — обновить.
  Поиск: `grep -rn 'viewFooter\(\)' internal/tui/*_test.go`. — `task test`

- [ ] 6. В `internal/tui/shell_test.go` добавить:
  ```go
  func TestViewFooter_ModeChipPresent(t *testing.T) {
      m := newTestModel(t)
      out := m.viewFooter()
      require.Contains(t, out, "-- NORMAL --")
  }

  func TestViewFooter_FilterModeChip(t *testing.T) {
      m := newTestModel(t)
      m.filtering = true
      out := m.viewFooter()
      require.Contains(t, out, "-- FILTER --")
  }

  func TestViewFooter_SelectModeChip(t *testing.T) {
      m, _, tasks := setupModelWithInboxTasks(t, "x")
      m.selected[tasks[0].ID] = struct{}{}
      out := m.viewFooter()
      require.Contains(t, out, "-- SELECT --")
  }

  func TestViewFooter_StatusMessagePreserved(t *testing.T) {
      m := newTestModel(t)
      m.statusMsg = "boom"
      out := m.viewFooter()
      require.Contains(t, out, "boom")
  }
  ```
  — `task test`

After all subtasks: Run `task test-race && task lint`. Все T-1, T-3 + обновлённые existing тесты passing.

---

## Task: T-5 — Implement Header refactor (segments + counts)

*_Requirements: REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4, REQ-2.7, REQ-2.8_*
*_Preservation: T-1, T-3, T-4 tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Создать новый `viewHeader` в `shell.go`. Каждый сегмент — `(N) Name [Count]` или `(N)I[Count]` в compact mode. Active list — wrapped в `theme.Header`. `[?]` placeholder при отсутствии count в cache.

Subtasks:

- [ ] 1. В `internal/tui/shell.go` добавить:
  ```go
  const headerCompactThreshold = 80

  var listInitials = map[listKind]string{
      listInbox:    "I",
      listToday:    "T",
      listUpcoming: "U",
      listAnytime:  "A",
      listSomeday:  "S",
      listLogbook:  "L",
  }

  // renderHeaderSegment formats one header segment.
  // active=true → entire segment styled with theme.Header (inverted bg).
  // compact=true → "(N)I[Count]"; compact=false → "(N) Name [Count]".
  // knownCount=false → display "[?]" instead of "[Count]".
  func renderHeaderSegment(theme Theme, idx int, label string, count int, knownCount bool, active, compact bool) string {
      n := fmt.Sprintf("(%d)", idx+1)
      countStr := "[?]"
      if knownCount {
          countStr = fmt.Sprintf("[%d]", count)
      }
      var raw string
      if compact {
          initial := label[:1] // first byte; safe for our list names (ASCII)
          if i, ok := listInitials[allLists[idx]]; ok {
              initial = i
          }
          raw = n + initial + countStr
      } else {
          raw = n + " " + label + " " + countStr
      }
      if active {
          return theme.Header.Render(raw)
      }
      // Composite styling: digit with accent, label with text, count with dim.
      var b strings.Builder
      if compact {
          b.WriteString(theme.Selected.Render(n))           // (N) — accent
          b.WriteString(theme.HeaderDim.Render(label[:1])) // I — subtext
          b.WriteString(theme.Dim.Render(countStr))         // [Count] — dim
      } else {
          b.WriteString(theme.Selected.Render(n))
          b.WriteString(" ")
          b.WriteString(theme.HeaderDim.Render(label))
          b.WriteString(" ")
          b.WriteString(theme.Dim.Render(countStr))
      }
      return b.String()
  }
  ```
  NOTE: Composite styling uses `theme.Selected` for accent-colored digit (Selected style = `Bold(true).Foreground(accent).Background(...)` — could be wrong color choice; consider creating a dedicated `theme.HeaderDigit` style if visual is wrong. For v1 use existing styles.) — `task test`

- [ ] 2. В `internal/tui/shell.go` добавить новый `viewHeader`:
  ```go
  func (m Model) viewHeader() string {
      compact := m.width > 0 && m.width < headerCompactThreshold
      labels := []string{"Inbox", "Today", "Upcoming", "Anytime", "Someday", "Logbook"}
      parts := make([]string, 0, len(allLists))
      for i, l := range allLists {
          count, known := m.listCounts[l]
          active := l == m.activeList
          parts = append(parts, renderHeaderSegment(m.theme, i, labels[i], count, known, active, compact))
      }
      return strings.Join(parts, " ")
  }
  ```
  — `task test`

- [ ] 3. В `internal/tui/app.go` удалить старый `viewHeader` (теперь в shell.go). Проверить через `grep -n "func (m Model) viewHeader" internal/tui/`. — `task test`

- [ ] 4. Обновить существующий `TestTUI_ViewContainsListLabels` если падает: новый формат содержит digit prefix и count, но labels всё ещё есть → должен пройти. Подтвердить. — `task test`

- [ ] 5. В `internal/tui/shell_test.go` добавить:
  ```go
  func TestViewHeader_FullModeSegment(t *testing.T) {
      m := newTestModel(t)
      m.width = 120
      m.listCounts[listInbox] = 24
      out := m.viewHeader()
      require.Contains(t, out, "(1)")
      require.Contains(t, out, "Inbox")
      require.Contains(t, out, "[24]")
  }

  func TestViewHeader_CompactMode(t *testing.T) {
      m := newTestModel(t)
      m.width = 70
      m.listCounts[listInbox] = 24
      out := m.viewHeader()
      require.Contains(t, out, "(1)I[24]")
      require.NotContains(t, out, "Inbox")
  }

  func TestViewHeader_UnknownCountPlaceholder(t *testing.T) {
      m := newTestModel(t)
      m.width = 120
      // listCounts is empty by default
      out := m.viewHeader()
      require.Contains(t, out, "[?]")
  }

  func TestViewHeader_EmptyListShowsZero(t *testing.T) {
      m := newTestModel(t)
      m.width = 120
      m.listCounts[listInbox] = 0
      out := m.viewHeader()
      require.Contains(t, out, "[0]")
  }

  func TestViewHeader_ActiveSegmentInverted(t *testing.T) {
      m := newTestModel(t)
      m.width = 120
      m.activeList = listInbox
      m.listCounts[listInbox] = 24
      out := m.viewHeader()
      // Active inversion via theme.Header — verify ANSI escape sequence presence.
      // theme.Header is Bold(true).Foreground(bg).Background(accent).Padding(0, 1)
      // — emits specific escape. Indirect check: assert visual length includes
      // padding spaces or just that output has more than plain text length.
      require.Greater(t, len(out), len("(1) Inbox [24]"))
  }
  ```
  — `task test`

After all subtasks: Run `task test-race && task lint`.

---

## Task: T-6 — Counts fetching + full-screen clamp

*_Requirements: REQ-1.2, REQ-1.3, REQ-1.4, REQ-2.5, REQ-2.6_*
*_Preservation: T-1, T-3, T-4, T-5 tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать `fetchListCounts` Cmd + `countsLoadedMsg` handling. Обновить `View()` для full-screen clamp при достаточных размерах.

Subtasks:

- [ ] 1. В `internal/tui/msgs.go` добавить:
  ```go
  type countsLoadedMsg struct {
      counts map[listKind]int
  }
  ```
  — `task test`

- [ ] 2. В `internal/tui/shell.go` добавить `fetchListCounts`:
  ```go
  func fetchListCounts(svc *app.Service) tea.Cmd {
      return func() tea.Msg {
          ctx := context.Background()
          counts := make(map[listKind]int, 6)
          if list, err := svc.ListInbox(ctx); err == nil {
              counts[listInbox] = len(list)
          }
          if list, err := svc.ListToday(ctx); err == nil {
              counts[listToday] = len(list)
          }
          if list, err := svc.ListUpcoming(ctx); err == nil {
              counts[listUpcoming] = len(list)
          }
          if list, err := svc.ListAnytime(ctx); err == nil {
              counts[listAnytime] = len(list)
          }
          if list, err := svc.ListSomeday(ctx); err == nil {
              counts[listSomeday] = len(list)
          }
          if list, err := svc.ListLogbook(ctx); err == nil {
              counts[listLogbook] = len(list)
          }
          return countsLoadedMsg{counts: counts}
      }
  }
  ```
  Импорты: `context`, `tea "github.com/charmbracelet/bubbletea"`, `"github.com/jtprogru/todushka/internal/app"`. — `task test`

- [ ] 3. В `internal/tui/shell_test.go` добавить:
  ```go
  func TestFetchListCounts_EmitsMsg(t *testing.T) {
      _, svc := newTestModelWithService(t)
      ctx := context.Background()
      _, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
      require.NoError(t, err)
      cmd := fetchListCounts(svc)
      msg := cmd()
      res, ok := msg.(countsLoadedMsg)
      require.True(t, ok)
      require.Equal(t, 1, res.counts[listInbox])
  }
  ```
  — `task test`

- [ ] 4. В `internal/tui/app.go` `Update` добавить новый case `countsLoadedMsg` (поставить рядом с `nameCacheLoadedMsg`):
  ```go
  case countsLoadedMsg:
      m.listCounts = msg.counts
      return m, nil
  ```
  — `task test`

- [ ] 5. В `internal/tui/app.go` `Update` обновить `tasksLoadedMsg` case чтобы возвращать `tea.Batch` с обоими Cmds:
  ```go
  case tasksLoadedMsg:
      m.tasks = msg.tasks
      if m.cursor >= len(m.tasks) {
          m.cursor = max(0, len(m.tasks)-1)
      }
      return m, tea.Batch(
          fetchNameCache(m.service, m.tasks),
          fetchListCounts(m.service),
      )
  ```
  — `task test`

- [ ] 6. В `internal/tui/shell_test.go` добавить:
  ```go
  func TestUpdate_TasksLoadedTriggersCountsFetch(t *testing.T) {
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
      require.NoError(t, err)
      m2, cmd := m.Update(tasksLoadedMsg{tasks: []task.Task{tk}})
      require.NotNil(t, cmd)
      _ = m2
      // Execute Cmd; expect tea.BatchMsg or sequential message
      msg := cmd()
      // Walk batch
      if batch, ok := msg.(tea.BatchMsg); ok {
          var foundCounts bool
          for _, sub := range batch {
              if sub == nil {
                  continue
              }
              if _, ok := sub().(countsLoadedMsg); ok {
                  foundCounts = true
              }
          }
          require.True(t, foundCounts, "countsLoadedMsg expected in batch")
          return
      }
      // direct emission
      _, ok := msg.(countsLoadedMsg)
      require.True(t, ok, "expected countsLoadedMsg, got %T", msg)
  }
  ```
  — `task test`

- [ ] 7. В `internal/tui/app.go` обновить `View()` для full-screen clamp:
  ```go
  func (m Model) View() string {
      // Editor takes full content area regardless of size — REQ-1.4.
      if m.screen == screenEditor {
          body := m.editor.View(m.theme, m.editorWidth())
          return lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
      }
      // Compute body for current screen.
      var body string
      if m.confirm != nil {
          modal := m.theme.Modal.Render(fmt.Sprintf("%s %d tasks? (y/n)", m.confirm.action.label(), len(m.confirm.ids)))
          body = lipgloss.JoinVertical(lipgloss.Left, m.viewBody(), modal)
      } else {
          switch m.screen {
          case screenHelp:
              body = m.viewHelp()
          case screenQuickEntry:
              body = lipgloss.JoinVertical(lipgloss.Left, m.viewBody(), m.viewQuickEntry())
          default:
              body = m.viewBody()
          }
      }
      // Full-screen clamp if terminal is large enough — REQ-1.2, REQ-1.3.
      if m.height >= 10 && m.width >= 40 {
          header := m.viewHeader()
          footer := m.viewFooter()
          headerH := lipgloss.Height(header)
          footerH := lipgloss.Height(footer)
          bodyH := m.height - headerH - footerH
          if bodyH < 0 {
              bodyH = 0
          }
          clampedBody := lipgloss.NewStyle().Height(bodyH).MaxHeight(bodyH).Render(body)
          return lipgloss.JoinVertical(lipgloss.Left, header, clampedBody, footer)
      }
      // Legacy fallback for small terminals.
      return lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
  }
  ```
  Импорт `lipgloss` уже есть. — `task test`

- [ ] 8. В `internal/tui/shell_test.go` добавить тесты на full-screen clamp:
  ```go
  func TestView_FullScreenClamp(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "x")
      m.width = 120
      m.height = 40
      out := m.View()
      require.Equal(t, 40, lipgloss.Height(out), "full-screen mode must produce exactly m.height lines")
  }

  func TestView_SmallTerminalFallback(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "x")
      m.width = 120
      m.height = 5
      out := m.View()
      // Legacy mode: no exact-height clamp; just verify no crash and reasonable output.
      require.NotEmpty(t, out)
  }

  func TestView_EditorIgnoresClamp(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "x")
      m.width = 200
      m.height = 40
      m.screen = screenEditor
      m.editor = NewEditor(m.tasks[0])
      out := m.View()
      // Editor present; full-screen height may not be exactly m.height.
      // Test just verifies no crash.
      require.NotEmpty(t, out)
  }
  ```
  — `task test`

After all subtasks: Run `task test-race && task lint`. Все предыдущие тесты passing.

---

## Task: T-7 — CLI `--config` flag + main.go wiring

*_Requirements: REQ-4.1, REQ-4.8_*
*_Preservation: existing CLI tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Добавить `--config <path>` persistent flag на root cobra command. Подключить `config.Load(path, env)` в `main.go` и передать результат в `NewModel`.

Subtasks:

- [ ] 1. Найти root cobra command: `grep -rn "cobra.Command" cmd/ internal/cli/ | head -5`. Обычно в `internal/cli/root.go` или `cmd/todushka/main.go`. Открыть тот файл. — (read-only)

- [ ] 2. В root command добавить persistent flag:
  ```go
  var configFlag string

  // в init или в build root cmd:
  rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "Path to config file (overrides default)")
  ```
  Если флаг локальный к main: добавить через package-level variable + getter `ConfigFlag()`. — `task test`

- [ ] 3. В `cmd/todushka/main.go` обновить TUI initialization чтобы:
  ```go
  configPath, err := config.ResolvePath(config.OSEnv, configFlag)
  if err != nil {
      // log warning, use Defaults
      cfg := config.Defaults()
  } else {
      cfg, warns, _ := config.Load(configPath, config.OSEnv)
      for _, w := range warns {
          // log w to LogPath
      }
      // use cfg
  }
  model := tui.NewModel(svc, theme, cfg)
  ```
  Импорты: `"github.com/jtprogru/todushka/internal/config"`. — `task build`

- [ ] 4. Запустить `./bin/todushka --help` после `task build`; проверить что `--config` отображён. Manual check.

- [ ] 5. В `internal/cli/...` или integration test (if any exist for the CLI) — добавить тест что флаг парсится. Если нет существующего CLI test infrastructure — pass. — `task test`

After all subtasks: Run `task test-race && task lint && task build`. Manual smoke: `./bin/todushka --help` includes `--config`.

---

## Task: T-8 — Property-based tests batch

*_Requirements: ALL_*
*_Preservation: ALL_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать 14 property-тестов из design §2.6 через `pgregory.net/rapid`, по одному на каждый CP-N.

Subtasks:

- [ ] 1. В `internal/tui/shell_test.go` добавить PBT для mode/footer/header (CP-1, CP-4, CP-5, CP-7, CP-8). Генераторы: `rapid.SampledFrom([]screenKind{...})`, `rapid.IntRange(40, 200)` для width, `rapid.Bool()` for filtering/etc. — `task test`

- [ ] 2. В `internal/tui/shell_test.go` добавить PBT для full-screen и counts (CP-2, CP-3, CP-6, CP-12). — `task test`

- [ ] 3. В `internal/config/app_test.go` или `loader_test.go` добавить PBT для config (CP-9, CP-10, CP-11, CP-13, CP-14). — `task test`

- [ ] 4. Запустить `task test-race -count=2` для проверки стабильности PBT.

After all subtasks: Все тесты зелёные, lint 0.

---

## Task: T-9 — GATE Checkpoint

*_Requirements: ALL_*
*_Complexity: mechanical_*

CRITICAL: Эта задача — ПОСЛЕДНЯЯ.

Instructions:

1. `go clean -testcache && task test` — все packages PASS.
2. `task test-race` — race-free.
3. `task build` — bin/todushka собирается.
4. `task lint` — 0 issues.
5. `gofmt -l internal/ cmd/` — no files.
6. Coverage matrix: каждое REQ проверено ≥1 тестом, каждое CP покрыто property-тестом.
7. Manual smoke: `./bin/todushka --help` показывает `--config` флаг.
8. Manual smoke: запустить TUI (если возможно) на терминале 120×40 — проверить header counts + zellij-style footer + dual-pane.
9. Если что-то не работает — вернуться к соответствующей T-N.

After all checks pass: ready for user approval; pipeline finishes.

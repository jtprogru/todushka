# UX Polish (v0.5.0) — Task Plan

## Preamble

### Work Type Classification

**Pure feature** with **preservation surface**: 174 existing tests must remain valid после rename `fieldSomeday → fieldWhen`, editor View text changes, separator rendering. Editor refactor — самая чувствительная зона (Tab cycling tests могут потребовать обновления).

### Test Style Source

**Tier 2** — adjacent tests
- **Reference unit tests:** `internal/tui/app_test.go`, `internal/tui/shell_test.go`, `internal/config/app_test.go`. Установленные fixtures: `newTestModel(t)`, `newTestModelWithService(t)`, `setupModelWithInboxTasks(t, ...)`, `bareTestModel()`.
- **Reference property tests:** `internal/tui/*_test.go` уже содержат rapid PBT (`pgregory.net/rapid`).
- **Platform-specific tests:** платформо-специфичная detection функция инжектируется через `detectDarkModeFn` package var.
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
| REQ-1.1 auto triggers detection | T-2, T-3 | CP-1 |
| REQ-1.2 macOS detection | T-2 | CP-1 |
| REQ-1.3 Linux detection | T-2 | CP-1 |
| REQ-1.4 fallback dark | T-2 | CP-1 |
| REQ-1.5 auto/system both valid | T-2 | CP-11 |
| REQ-1.6 NO_COLOR overrides | T-3 | CP-2 |
| REQ-1.7 500ms timeout | T-2 | CP-1 |
| REQ-2.1 inline splice by ID | T-5 | CP-4 |
| REQ-2.2 batch refresh | T-5 | CP-12 |
| REQ-2.3 missing ID skipped | T-5 | CP-3 |
| REQ-2.4 length preserved | T-5 | CP-3 |
| REQ-3.1 radio-style View | T-4 | CP-5 |
| REQ-3.2 focus highlight | T-4 | CP-5 |
| REQ-3.3 Space toggles | T-4 | CP-5 |
| REQ-3.4 mapping at save | T-4 | CP-6 |
| REQ-3.5 hint conditional | T-4 | CP-7 |
| REQ-3.6 field count preserved | T-4 | (backward compat) |
| REQ-4.1 separator below header | T-6 | CP-8 |
| REQ-4.2 separator above footer | T-6 | CP-8 |
| REQ-4.3 separator width | T-6 | CP-9 |
| REQ-4.4 legacy no separator | T-6 | CP-8 |
| REQ-4.5 editor no separator | T-6 | CP-8 |
| REQ-4.6 bodyH adjusted | T-6 | CP-10 |
| REQ-5.1 174 tests passing | T-1, T-8 | — |
| REQ-5.2 explicit themes skip detect | T-3 | — |
| REQ-5.3 editor field count = 6 | T-4 | — |

26 REQs → 8 tasks → 12 CPs. Каждый REQ покрыт ≥1 task; каждый CP — property-тестом в T-7.

---

## Task Order

```
T-1 GREEN (baseline preservation)
  → T-2 CODE (Config validation + platform detection foundation)
    → T-3 CODE (Theme integration: selectThemeFromConfig handles auto/system)
      → T-4 CODE (Editor When refactor: rename + hint + Space cycle)
        → T-5 CODE (Edit splice in editorSavedMsg)
          → T-6 CODE (Section separators in View)
            → T-7 GREEN (PBT batch — 12 CPs)
              → T-8 GATE (Checkpoint)
```

---

## Task: T-1 — Baseline preservation

*_Requirements: REQ-5.1_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Подтвердить что 174 теста проходят на baseline до изменений.

Subtasks:

- [ ] 1. Запустить `go clean -testcache && task test-race` — все packages PASS.
- [ ] 2. Запустить `task lint` — 0 issues.
- [ ] 3. Зафиксировать count: `go test ./internal/tui/ -v -count=1 2>&1 | grep -c "^--- PASS"` — текущее число.

After all subtasks: baseline established.

---

## Task: T-2 — Config validation + Platform detection foundation

*_Requirements: REQ-1.1, REQ-1.2, REQ-1.3, REQ-1.4, REQ-1.5, REQ-1.7_*
*_Preservation: existing config tests + lint_*
*_Test_Style: Tier 2 + platform-injectable_*
*_Complexity: complex_*

GOAL: Создать platform-specific detection files + `theme_resolve.go` + `resolveAutoTheme` function. Обновить `AppConfig.Validate` чтобы принимать `"auto"`/`"system"`.

Subtasks (each touches one file; run `task test` after each):

1. В `internal/config/app.go` `Validate()` обновить theme switch чтобы принимать `"auto"` и `"system"` как valid (no warning):
```go
switch c.Theme {
case "macchiato", "latte", "mono", "auto", "system", "":
    // ok
default:
    warns = append(warns, ...)
    c.Theme = def.Theme
}
```
Запустить existing config tests — все passing.

2. В `internal/config/app_test.go` добавить:
```go
func TestValidate_AutoThemeIsValid(t *testing.T) {
    c, warns := AppConfig{Theme: "auto"}.Validate()
    require.Empty(t, warns)
    require.Equal(t, "auto", c.Theme)
}

func TestValidate_SystemThemeIsValid(t *testing.T) {
    c, warns := AppConfig{Theme: "system"}.Validate()
    require.Empty(t, warns)
    require.Equal(t, "system", c.Theme)
}
```

3. Создать `internal/tui/detect_dark_darwin.go`:
```go
//go:build darwin

package tui

import (
    "context"
    "os/exec"
    "strings"
    "time"
)

// detectDarkMode reports whether macOS is in dark mode via:
//   defaults read -g AppleInterfaceStyle
// Returns (true, nil) when output is "Dark"; (false, nil) when command
// exits non-zero (which is macOS's way of saying "light mode"); (false, err)
// only on unexpected errors.
func detectDarkMode() (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    cmd := exec.CommandContext(ctx, "defaults", "read", "-g", "AppleInterfaceStyle")
    out, err := cmd.Output()
    if err != nil {
        // exit code 1 means "Light mode" (key doesn't exist) — not an error per se
        if _, ok := err.(*exec.ExitError); ok {
            return false, nil
        }
        return false, err
    }
    return strings.TrimSpace(string(out)) == "Dark", nil
}
```

4. Создать `internal/tui/detect_dark_linux.go`:
```go
//go:build linux

package tui

import (
    "context"
    "os/exec"
    "strings"
    "time"
)

// detectDarkMode reports whether Linux/GNOME is in dark mode via:
//   gsettings get org.gnome.desktop.interface color-scheme
// Output formats: 'prefer-dark' | 'prefer-light' | 'default'.
// Returns (true, nil) if 'dark' substring; (false, nil) if 'light';
// (false, err) on tool/timeout error.
func detectDarkMode() (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()
    cmd := exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
    out, err := cmd.Output()
    if err != nil {
        return false, err
    }
    s := strings.ToLower(strings.TrimSpace(string(out)))
    if strings.Contains(s, "dark") {
        return true, nil
    }
    return false, nil
}
```

5. Создать `internal/tui/detect_dark_other.go`:
```go
//go:build !darwin && !linux

package tui

// detectDarkMode for unsupported platforms returns (false, nil).
// Callers should fall back to a default theme (macchiato per design ADR).
func detectDarkMode() (bool, error) {
    return false, nil
}
```

6. Создать `internal/tui/theme_resolve.go`:
```go
package tui

// detectDarkModeFn is package-level so tests can override it. Production
// uses the platform-specific detectDarkMode from detect_dark_*.go.
var detectDarkModeFn = detectDarkMode

// resolveAutoTheme returns the theme name to use when AppConfig.Theme is
// "auto" or "system". Calls detectDarkModeFn(); maps:
//   dark  → "macchiato"
//   light → "latte"
//   error/unsupported → "macchiato" (default to dark)
func resolveAutoTheme() string {
    isDark, err := detectDarkModeFn()
    if err != nil || isDark {
        return "macchiato"
    }
    return "latte"
}
```

7. Создать `internal/tui/theme_resolve_test.go`:
```go
package tui

import (
    "errors"
    "testing"

    "github.com/stretchr/testify/require"
)

func TestResolveAutoTheme_DarkMode(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return true, nil }
    require.Equal(t, "macchiato", resolveAutoTheme())
}

func TestResolveAutoTheme_LightMode(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return false, nil }
    require.Equal(t, "latte", resolveAutoTheme())
}

func TestResolveAutoTheme_ErrorFallsBackToDark(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return false, errors.New("boom") }
    require.Equal(t, "macchiato", resolveAutoTheme())
}
```

After all subtasks: Run `task test-race && task lint`. All passes.

---

## Task: T-3 — Theme integration (auto/system handling)

*_Requirements: REQ-1.1, REQ-1.6, REQ-5.2_*
*_Preservation: T-2 tests + existing theme tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: `selectThemeFromConfig` (в `run.go`) теперь handle'ит `"auto"`/`"system"` через `resolveAutoTheme`. NO_COLOR имеет absolute precedence.

Subtasks:

1. В `internal/tui/run.go` обновить `selectThemeFromConfig`:
```go
func selectThemeFromConfig(name string, env func(string) string) Theme {
    if env == nil {
        env = func(string) string { return "" }
    }
    // NO_COLOR has absolute precedence (REQ-1.6)
    if env("NO_COLOR") != "" {
        return NewMonochromeTheme()
    }
    // Auto-detection for system/auto themes
    if name == "auto" || name == "system" {
        name = resolveAutoTheme()
    }
    switch name {
    case "latte", "light":
        return newColorTheme(latte)
    case "mono", "monochrome":
        return NewMonochromeTheme()
    case "macchiato", "dark":
        return newColorTheme(macchiato)
    }
    // Fallback to legacy env-based selection
    return SelectTheme(env)
}
```

2. В `internal/tui/theme_resolve_test.go` (или новом `run_test.go` если предпочитаемо) добавить:
```go
func TestSelectThemeFromConfig_AutoDarkUsesMacchiato(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return true, nil }
    th := selectThemeFromConfig("auto", func(string) string { return "" })
    require.Equal(t, "catppuccin-macchiato", th.Name)
}

func TestSelectThemeFromConfig_AutoLightUsesLatte(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return false, nil }
    th := selectThemeFromConfig("auto", func(string) string { return "" })
    require.Equal(t, "catppuccin-latte", th.Name)
}

func TestSelectThemeFromConfig_NoColorOverridesAuto(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return true, nil }
    env := func(k string) string {
        if k == "NO_COLOR" { return "1" }
        return ""
    }
    th := selectThemeFromConfig("auto", env)
    require.Equal(t, "monochrome", th.Name)
}

func TestSelectThemeFromConfig_SystemAliasMatchesAuto(t *testing.T) {
    orig := detectDarkModeFn
    defer func() { detectDarkModeFn = orig }()
    detectDarkModeFn = func() (bool, error) { return true, nil }
    th1 := selectThemeFromConfig("auto", func(string) string { return "" })
    th2 := selectThemeFromConfig("system", func(string) string { return "" })
    require.Equal(t, th1.Name, th2.Name)
}
```

After all subtasks: Run `task test-race && task lint`.

---

## Task: T-4 — Editor When refactor

*_Requirements: REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4, REQ-3.5, REQ-3.6_*
*_Preservation: existing 174 tests; field count must remain 6_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Заменить `someday bool` → `when shellEditorWhen` enum в `EditorModel`. Rename `fieldSomeday` → `fieldWhen`. Update View с radio-style; ApplyAndSave mapping; Space handler.

CRITICAL: 174 existing tests должны продолжать проходить с adjustment'ами только для editor visual assertions.

Subtasks:

1. В `internal/tui/editor.go` добавить `shellEditorWhen` enum (рядом с `editorField`):
```go
type shellEditorWhen int
const (
    whenAnytime shellEditorWhen = iota
    whenSomeday
)
```

2. В `EditorModel` struct: заменить `someday bool` на `when shellEditorWhen`. Заменить `fieldSomeday` в editorField const на `fieldWhen`.

3. В `NewEditor` инициализация: вместо `someday: t.Someday` использовать:
```go
when := whenAnytime
if t.Someday {
    when = whenSomeday
}
```
И присвоить `when: when` в struct literal.

4. В `ApplyAndSave` заменить `t.Someday = m.someday` на:
```go
t.Someday = m.when == whenSomeday
```

5. В `internal/tui/app.go` обновить `handleEditorKey` Space handler:
```go
case m.editor.focus == fieldWhen && msg.Type == tea.KeySpace:
    if m.editor.when == whenAnytime {
        m.editor.when = whenSomeday
    } else {
        m.editor.when = whenAnytime
    }
    return m, nil
```

6. В `internal/tui/editor.go` `View()` заменить блок someday (lines 218-226) на radio-style + hint:
```go
anytimeBullet := "[ ]"
somedayBullet := "[ ]"
if m.when == whenAnytime {
    anytimeBullet = "[•]"
} else {
    somedayBullet = "[•]"
}
whenSection := fmt.Sprintf("%s Anytime\n%s Someday", anytimeBullet, somedayBullet)
if m.focus == fieldWhen {
    whenSection = theme.Selected.Render("▶ When") + "\n" + theme.Selected.Render(whenSection)
} else {
    whenSection = theme.Dim.Render("  When") + "\n" + theme.Dim.Render(whenSection)
}
// Anytime hint when no Area/Project (REQ-3.5)
if m.when == whenAnytime && m.original.AreaID == nil && m.original.ProjectID == nil {
    whenSection += "\n" + theme.Dim.Render("(will appear in Inbox without Area/Project)")
}
```
И заменить `someday` на `whenSection` в `JoinVertical` ниже.

7. Обновить help text внизу View — было "Space: toggle Someday" → "Space: toggle When (Anytime/Someday)".

8. В `internal/tui/editor.go` или `app.go` найти любую другую ссылку на `fieldSomeday` или `m.editor.someday` — обновить (есть в `handleEditorKey` — fixed выше).

9. Запустить existing тесты:
- `TestTUI_EditorTabCyclesFields` — fieldWhen вместо fieldSomeday в expected? Проверить, что фактически тестирует. Update if needed.
- Все остальные тесты должны проходить (editor View text сменился, но никто не проверял конкретно "Someday" текст в editor View — проверить grep).

`grep -rn 'Someday\|fieldSomeday\|m.editor.someday\|m.someday' internal/tui/`. Update assertions where needed.

10. В `internal/tui/editor.go` test file (или app_test.go) добавить новые тесты:
```go
func TestEditor_WhenDefaultsAnytimeForOpenTask(t *testing.T) {
    tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen, Someday: false}
    ed := NewEditor(tk)
    require.Equal(t, whenAnytime, ed.when)
}

func TestEditor_WhenDefaultsSomedayForSomedayTask(t *testing.T) {
    tk := task.Task{ID: id.New(), Title: "x", Someday: true}
    ed := NewEditor(tk)
    require.Equal(t, whenSomeday, ed.when)
}

func TestEditor_SpaceTogglesWhen(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // open editor
    mm := m2.(Model)
    // Navigate to fieldWhen via Tab (5 fields before it: title, notes, start, deadline, tags)
    for i := 0; i < 5; i++ {
        mm2, _ := mm.Update(tea.KeyMsg{Type: tea.KeyTab})
        mm = mm2.(Model)
    }
    require.Equal(t, fieldWhen, mm.editor.focus)
    initial := mm.editor.when
    m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeySpace})
    mm = m3.(Model)
    require.NotEqual(t, initial, mm.editor.when)
}

func TestEditor_ApplyAndSaveMapsAnytime(t *testing.T) {
    _, svc, tasks := setupModelWithInboxTasks(t, "x")
    ed := NewEditor(tasks[0])
    ed.when = whenAnytime
    saved, err := ed.ApplyAndSave(context.Background(), svc)
    require.NoError(t, err)
    require.False(t, saved.Someday)
}

func TestEditor_ApplyAndSaveMapsSomeday(t *testing.T) {
    _, svc, tasks := setupModelWithInboxTasks(t, "x")
    ed := NewEditor(tasks[0])
    ed.when = whenSomeday
    saved, err := ed.ApplyAndSave(context.Background(), svc)
    require.NoError(t, err)
    require.True(t, saved.Someday)
}

func TestEditor_HintWhenAnytimeNoAreaProject(t *testing.T) {
    tk := task.Task{ID: id.New(), Title: "x", Someday: false}
    ed := NewEditor(tk)
    ed.when = whenAnytime
    out := ed.View(NewTheme(), 60)
    require.Contains(t, out, "will appear in Inbox")
}

func TestEditor_NoHintForSomeday(t *testing.T) {
    tk := task.Task{ID: id.New(), Title: "x", Someday: true}
    ed := NewEditor(tk)
    ed.when = whenSomeday
    out := ed.View(NewTheme(), 60)
    require.NotContains(t, out, "will appear in Inbox")
}

func TestEditor_TabCycleHas6Fields(t *testing.T) {
    require.Equal(t, 6, int(fieldCount))
}
```

After all subtasks: Run `task test-race && task lint`. All 174 + new tests passing.

---

## Task: T-5 — Edit splice in editorSavedMsg

*_Requirements: REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4_*
*_Preservation: T-1..T-4 tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: `editorSavedMsg` Update case теперь делает inline splice обновлённой задачи в `m.tasks` + параллельно fire `tea.Batch(loadCurrentList, fetchListCounts)`.

Subtasks:

1. В `internal/tui/app.go` найти `editorSavedMsg` case:
```go
case editorSavedMsg:
    m.screen = screenList
    return m, m.loadCurrentList()
```
Заменить на:
```go
case editorSavedMsg:
    m.screen = screenList
    // Inline splice for immediate visual update (REQ-2.1, 2.3, 2.4)
    for i := range m.tasks {
        if m.tasks[i].ID == msg.updated.ID {
            m.tasks[i] = msg.updated
            break
        }
    }
    return m, tea.Batch(
        m.loadCurrentList(),
        fetchListCounts(m.service),
    )
```

2. В `internal/tui/app_test.go` добавить:
```go
func TestEditorSavedMsg_InlineSpliceByID(t *testing.T) {
    m, _, tasks := setupModelWithInboxTasks(t, "old title", "other")
    updated := tasks[0]
    updated.Title = "new title"
    m2, _ := m.Update(editorSavedMsg{updated: updated})
    mm := m2.(Model)
    require.Equal(t, "new title", mm.tasks[0].Title, "inline splice replaces by ID")
    require.Len(t, mm.tasks, 2, "length preserved")
}

func TestEditorSavedMsg_NotFoundSkipsSplice(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "a", "b")
    ghost := task.Task{ID: id.New(), Title: "ghost"}
    m2, _ := m.Update(editorSavedMsg{updated: ghost})
    mm := m2.(Model)
    require.Len(t, mm.tasks, 2)
    require.NotEqual(t, "ghost", mm.tasks[0].Title)
    require.NotEqual(t, "ghost", mm.tasks[1].Title)
}

func TestEditorSavedMsg_PreservesSliceLength(t *testing.T) {
    m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c")
    require.Len(t, m.tasks, 3)
    upd := tasks[1]
    upd.Title = "modified"
    m2, _ := m.Update(editorSavedMsg{updated: upd})
    mm := m2.(Model)
    require.Len(t, mm.tasks, 3)
}

func TestEditorSavedMsg_FiresBatchedCmd(t *testing.T) {
    m, _, tasks := setupModelWithInboxTasks(t, "x")
    _, cmd := m.Update(editorSavedMsg{updated: tasks[0]})
    require.NotNil(t, cmd)
    // Cmd execution should produce a batch containing tasksLoadedMsg and countsLoadedMsg
    msg := cmd()
    if batch, ok := msg.(tea.BatchMsg); ok {
        foundCounts := false
        foundTasks := false
        for _, sub := range batch {
            if sub == nil { continue }
            switch sub().(type) {
            case countsLoadedMsg:
                foundCounts = true
            case tasksLoadedMsg:
                foundTasks = true
            }
        }
        require.True(t, foundCounts, "countsLoadedMsg expected in batch")
        require.True(t, foundTasks, "tasksLoadedMsg expected in batch")
        return
    }
    t.Fatalf("expected tea.BatchMsg, got %T", msg)
}
```

After all subtasks: Run `task test-race && task lint`.

---

## Task: T-6 — Section separators

*_Requirements: REQ-4.1, REQ-4.2, REQ-4.3, REQ-4.4, REQ-4.5, REQ-4.6_*
*_Preservation: T-1..T-5 tests; existing full-screen height tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Add `renderSeparator` helper в `app.go`; интегрировать в `View()` между header/body и body/footer в full-screen mode. Update `bodyH` calc.

CRITICAL: Existing `TestView_FullScreenClamp` тест проверяет `lipgloss.Height(out) == m.height`. С separator'ами body должен иметь `m.height - headerH - footerH - 2` рядов. Тест должен продолжать проходить (lipgloss.Height invariant сохраняется).

Subtasks:

1. В `internal/tui/app.go` добавить `renderSeparator`:
```go
// renderSeparator returns a one-line horizontal "─" rule of width characters,
// styled via theme.Help. Returns "" if width <= 0.
func renderSeparator(theme Theme, width int) string {
    if width <= 0 {
        return ""
    }
    return theme.Help.Render(strings.Repeat("─", width))
}
```

2. В `internal/tui/app.go` `View()` обновить full-screen branch. Текущая структура:
```go
if m.height >= 10 && m.width >= 40 {
    header := m.viewHeader()
    footer := m.viewFooter()
    headerH := lipgloss.Height(header)
    footerH := lipgloss.Height(footer)
    bodyH := m.height - headerH - footerH
    if bodyH < 0 { bodyH = 0 }
    clampedBody := lipgloss.NewStyle().Height(bodyH).MaxHeight(bodyH).Render(body)
    return lipgloss.JoinVertical(lipgloss.Left, header, clampedBody, footer)
}
```

Заменить на:
```go
if m.height >= 10 && m.width >= 40 && m.screen != screenEditor {
    header := m.viewHeader()
    footer := m.viewFooter()
    sep := renderSeparator(m.theme, m.width)
    headerH := lipgloss.Height(header)
    footerH := lipgloss.Height(footer)
    bodyH := m.height - headerH - footerH - 2 // 2 separators = 2 lines
    if bodyH < 0 { bodyH = 0 }
    clampedBody := lipgloss.NewStyle().Height(bodyH).MaxHeight(bodyH).Render(body)
    return lipgloss.JoinVertical(lipgloss.Left, header, sep, clampedBody, sep, footer)
}
```

NOTE: добавлено `m.screen != screenEditor` чтобы separators не отображались в editor (REQ-4.5).

3. В `internal/tui/app_test.go` (или shell_test.go) добавить:
```go
func TestRenderSeparator_FullWidth(t *testing.T) {
    s := renderSeparator(NewTheme(), 80)
    // strip ANSI to count "─" — but simpler: just check it contains 80 characters of "─"
    plain := strings.ReplaceAll(s, "\x1b", "") // strip basic escape; lipgloss adds more
    require.Contains(t, plain, strings.Repeat("─", 80))
}

func TestRenderSeparator_EmptyOnZero(t *testing.T) {
    s := renderSeparator(NewTheme(), 0)
    require.Empty(t, s)
}

func TestView_HasSeparatorsInFullScreen(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.width = 120
    m.height = 40
    out := m.View()
    require.Contains(t, out, "─")
}

func TestView_NoSeparatorsInLegacy(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.width = 120
    m.height = 5
    out := m.View()
    require.NotContains(t, out, "─", "legacy mode must not render section separators")
}

func TestView_NoSeparatorsInEditor(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.width = 120
    m.height = 40
    m.screen = screenEditor
    m.editor = NewEditor(m.tasks[0])
    out := m.View()
    require.NotContains(t, out, "─", "editor mode must not render section separators")
}

func TestView_FullScreenHeightWithSeparators(t *testing.T) {
    m, _, _ := setupModelWithInboxTasks(t, "x")
    m.width = 120
    m.height = 40
    out := m.View()
    require.Equal(t, 40, lipgloss.Height(out), "lipgloss height invariant preserved")
}
```

4. Обновить existing `TestView_FullScreenClamp` (в shell_test.go) если падает — assertion должна продолжать pass с новой bodyH calc.

5. Существующий test `TestView_EditorIgnoresClamp` — проверить что он passes с новым `m.screen != screenEditor` early-return.

After all subtasks: Run `task test-race && task lint`. Все T-1..T-5 tests + new T-6 tests passing.

---

## Task: T-7 — Property-based tests batch

*_Requirements: ALL_*
*_Preservation: ALL_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать 12 property-тестов из design §2.6 через `pgregory.net/rapid`, по одному на каждый CP-N.

Subtasks:

1. В `internal/tui/theme_resolve_test.go` (or new file) добавить PBT для theme resolution (CP-1, CP-2, CP-11):
   - `TestProp_AutoThemeResolution` — random isDark/error → predictable mapping.
   - `TestProp_NoColorOverridesAuto` — NO_COLOR fixed + random theme name → monochrome.
   - `TestProp_AutoSystemAliases` — both produce identical result.

2. В `internal/tui/app_test.go` (or new property file) добавить PBT для edit splice (CP-3, CP-4, CP-12):
   - `TestProp_EditSplicePreservesLength`
   - `TestProp_EditSpliceByID`
   - `TestProp_RefreshBatchHasBothCmds`

3. В `internal/tui/editor_test.go` (or app_test.go) добавить PBT для editor When (CP-5, CP-6, CP-7):
   - `TestProp_WhenToggleInvolution`
   - `TestProp_WhenMapping`
   - `TestProp_HintConditional`

4. В `internal/tui/app_test.go` (or shell_test.go) добавить PBT для separators (CP-8, CP-9, CP-10):
   - `TestProp_SeparatorsConditional` — random width/height/screen → expected separator presence.
   - `TestProp_SeparatorWidth` — random width → output contains exactly `width` `─` chars.
   - `TestProp_FullScreenHeightWithSeparators` — random valid dimensions → lipgloss.Height == m.height.

5. Запустить `task test-race -count=2 -timeout=120s ./internal/tui/ ./internal/config/` для стабильности.

After all subtasks: Run `task test-race && task lint`.

---

## Task: T-8 — GATE Checkpoint

*_Requirements: ALL_*
*_Complexity: mechanical_*

CRITICAL: ПОСЛЕДНЯЯ задача.

Instructions:

1. `go clean -testcache && task test` — all packages PASS.
2. `task test-race` — race-free.
3. `task build` — bin/todushka compiles.
4. `task lint` — 0 issues.
5. `gofmt -l internal/ cmd/` — empty.
6. Coverage matrix sanity: каждое REQ имеет ≥1 проходящий тест; каждое CP покрыто property-тестом.
7. Manual smoke: `./bin/todushka --help` — все subcommands и `--config` flag present.
8. Manual smoke (если возможно): запустить TUI на широком терминале, проверить:
   - Section separators видны между header/body/footer.
   - Editor: When toggle отображается; Space переключает между Anytime/Someday; hint показывается при отсутствии area/project.
   - После edit: задача в списке обновляется немедленно (no Tab needed).
   - `config: theme: auto` → автоматическая тема следует OS.
9. Если что-то не работает — вернуться к T-N.

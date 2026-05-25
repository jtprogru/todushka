# Dual-Pane Layout for TUI — Task Plan

## Preamble

### Work Type Classification

**Pure feature** — новая адаптивная layout-фича. Существующее single-pane поведение должно оставаться идентичным при `m.width == 0` или `m.width < 100` (REQ-1.2/1.3). T-1 preservation-тесты фиксируют это.

### Test Style Source

**Tier 2** — adjacent tests
- **Reference unit tests:** `internal/tui/app_test.go`, `internal/tui/filter_test.go`, `internal/tui/bulk_test.go` — testify `require`, `newTestModel(t)`/`setupModelWithInboxTasks(t, ...)`/`bareTestModel()` fixtures, прямой `Update(tea.KeyMsg/Msg{...})` dispatch.
- **Reference property tests:** `internal/tui/*_test.go` уже содержат rapid PBT (`pgregory.net/rapid`).
- **Key patterns:** assertions через `require`; Cmd verification через `msg := cmd(); require.IsType(t, expectedMsg, msg)`; table-driven с `name`-keyed cases.

### Commands

| Action     | Command          | Source         |
|------------|------------------|----------------|
| Test       | `task test`      | `Taskfile.yml` |
| Test race  | `task test-race` | `Taskfile.yml` |
| Build      | `task build`     | `Taskfile.yml` |
| Lint       | `task lint`      | `Taskfile.yml` |
| Format     | `task fmt`       | `Taskfile.yml` |

### Coverage Matrix

| Requirement | Task(s) | Correctness Property |
|-------------|---------|----------------------|
| REQ-1.1 dual-pane condition | T-4 | CP-1 (Exclusion: layout mode) |
| REQ-1.2 single-pane below threshold | T-1, T-4 | CP-1 |
| REQ-1.3 width==0 → single-pane | T-1, T-4 | CP-1 |
| REQ-1.4 45/55 width allocation | T-4 | CP-2 (Equivalence: pane width arithmetic) |
| REQ-1.5 double border `║` | T-4 | CP-3 (Propagation: border renders) |
| REQ-1.6 editor full-pane | T-5 | CP-4 (Absence: editor/help disable dual) |
| REQ-1.7 help full-pane | T-5 | CP-4 |
| REQ-1.8 confirm modal stacks | T-5 | CP-14 (Propagation: confirm stacks) |
| REQ-2.1 Title rendering | T-3 | CP-5 (Equivalence: fixed field order) |
| REQ-2.2 Status rendering | T-3 | CP-5 |
| REQ-2.3 Notes truncation | T-3 | CP-7 (Equivalence: notes truncation) |
| REQ-2.4 Start date | T-3 | CP-5 |
| REQ-2.5 Due date | T-3 | CP-5 |
| REQ-2.6 Pinned date | T-3 | CP-5 |
| REQ-2.7 Area name | T-3 | CP-5 |
| REQ-2.8 Project name | T-3 | CP-5 |
| REQ-2.9 Heading name | T-3 | CP-5 |
| REQ-2.10 Tags list | T-3 | CP-5 |
| REQ-2.11 Omit nil fields | T-3 | CP-6 (Absence: nil-field omission) |
| REQ-2.12 Someday marker | T-3 | CP-5 |
| REQ-3.1 Empty list placeholder | T-3 | CP-8 (Propagation: placeholder) |
| REQ-3.2 Out-of-range cursor placeholder | T-3 | CP-8 |
| REQ-4.1 Name cache fetch on tasksLoadedMsg | T-2 | CP-10 (Propagation: name cache populated) |
| REQ-4.2 No IO in View() | T-3 | CP-9 (Absence: no repo access) |
| REQ-4.3 Short-ID fallback | T-3 | CP-11 (Propagation: short-ID fallback) |
| REQ-5.1 Cursor change updates details | T-3, T-4 | CP-12 (Propagation: cursor change reflects) |
| REQ-5.2 Filter change updates details | T-4 | CP-12 |
| REQ-6.1 Filter preserves dual-pane | T-4 | CP-13 (Propagation: filter preserves dual) |
| REQ-6.2 Quick-entry in dual-pane | T-5 | CP-1 |
| REQ-6.3 Bulk triggers reload | T-5 | CP-10 (через существующий tasksLoadedMsg path) |

24 REQ → 7 tasks → 14 CPs. Каждый REQ покрыт ≥1 task, каждый CP получит property-test в T-6.

---

## Task Order

```
T-1 GREEN (preservation: single-pane при width==0 / width<100)
  → T-2 CODE (Name Cache foundation: Model fields + msg + Cmd + Update integration)
    → T-3 CODE (Details Pane: viewDetails + cursorTask + wrapAndTruncate + helpers)
      → T-4 CODE (Layout decision: isDualPane + pane widths + viewBody dispatcher)
        → T-5 CODE (Mode interactions: editor/help full-pane, confirm/quick-entry stacking)
          → T-6 GREEN (Property-based tests batch — 14 CPs)
            → T-7 GATE (Checkpoint)
```

---

## Task: T-1 — Write preservation tests for single-pane behavior

*_Requirements: REQ-1.2, REQ-1.3_*
*_Test_Style: Tier 2 (`internal/tui/app_test.go`)_*
*_Complexity: standard_*

GOAL: Зафиксировать текущее single-pane поведение TUI до введения dual-pane. Тесты должны проходить **до**, **во время** и **после** реализации.

IMPORTANT: Эти тесты лочат backward-compat ветку `m.width < 100` AND `m.width == 0`. После T-4 они должны продолжать проходить.

DO NOT: Изменять production-код в этой задаче.

Subtasks:

- [ ] 1. В `internal/tui/app_test.go` добавить `TestTUI_ZeroWidthRendersSinglePane`: создать model через `setupModelWithInboxTasks(t, "task one")`, **НЕ** устанавливать `m.width` (остаётся 0), вызвать `m.View()`, assert что output НЕ содержит `║` (double border glyph). — `task test`

- [ ] 2. В `internal/tui/app_test.go` добавить `TestTUI_NarrowWidthRendersSinglePane`: setup с 1 task, `m.width = 80`, `m.View()`, assert no `║`. — `task test`

- [ ] 3. В `internal/tui/app_test.go` добавить `TestTUI_NarrowWidthShowsListInBody`: setup с 2 task'и (`"alpha"`, `"beta"`), `m.width = 80`, `m.cursor = 0`, `m.View()`, assert output contains both `"alpha"` and `"beta"` (single-pane shows full list). — `task test`

- [ ] 4. Запустить `task test` — все три теста должны пройти на текущей кодовой базе (зелёный baseline; `m.View()` сейчас single-pane везде). Зафиксировать как preservation lock.

After all subtasks: Run `task lint` to confirm no style errors.

---

## Task: T-2 — Implement Name Cache foundation

*_Requirements: REQ-4.1_*
*_Preservation: T-1 tests + все существующие TUI тесты_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Добавить 4 name-cache map'а в `Model`, тип `nameCacheLoadedMsg`, функцию `fetchNameCache(svc, tasks) tea.Cmd`, и интеграцию в `tasksLoadedMsg` Update case.

IMPORTANT: На этом этапе viewDetails ещё нет — `nameCacheLoadedMsg` не используется в View. Тесты проверяют только Update-логику.

Subtasks:

- [ ] 1. В `internal/tui/app.go` добавить 4 поля в `Model` struct (после `filtering`): `tagNamesByID map[id.ID]string`, `areaNamesByID map[id.ID]string`, `projectNamesByID map[id.ID]string`, `headingNamesByID map[id.ID]string`. В `NewModel` инициализировать каждое через `make(map[id.ID]string)`. — `task test`

- [ ] 2. В `internal/tui/msgs.go` добавить тип:
  ```go
  type nameCacheLoadedMsg struct {
      tags     map[id.ID]string
      areas    map[id.ID]string
      projects map[id.ID]string
      headings map[id.ID]string
  }
  ```
  Импорт `"github.com/jtprogru/todushka/internal/domain/id"` если ещё нет. — `task test`

- [ ] 3. Создать новый файл `internal/tui/details.go` со скелетом package и функцией:
  ```go
  // fetchNameCache returns a Cmd that resolves all referenced tag/area/project/heading
  // IDs from tasks via the Repository, emitting nameCacheLoadedMsg with the result.
  // Per-ID errors are silently skipped — missing names fall back to short-IDs in views.
  func fetchNameCache(svc *app.Service, tasks []task.Task) tea.Cmd {
      return func() tea.Msg {
          ctx := context.Background()
          tags := make(map[id.ID]string)
          areas := make(map[id.ID]string)
          projects := make(map[id.ID]string)
          headings := make(map[id.ID]string)
          // collect distinct IDs
          tagSet := make(map[id.ID]struct{})
          areaSet := make(map[id.ID]struct{})
          projectSet := make(map[id.ID]struct{})
          headingSet := make(map[id.ID]struct{})
          for _, t := range tasks {
              for _, tg := range t.Tags {
                  tagSet[tg] = struct{}{}
              }
              if t.AreaID != nil { areaSet[*t.AreaID] = struct{}{} }
              if t.ProjectID != nil { projectSet[*t.ProjectID] = struct{}{} }
              if t.HeadingID != nil { headingSet[*t.HeadingID] = struct{}{} }
          }
          repo := svc.Repo()
          for tid := range tagSet {
              if tg, err := repo.TagGet(ctx, tid); err == nil { tags[tid] = tg.Name }
          }
          for aid := range areaSet {
              if a, err := repo.AreaGet(ctx, aid); err == nil { areas[aid] = a.Name }
          }
          for pid := range projectSet {
              if p, err := repo.ProjectGet(ctx, pid); err == nil { projects[pid] = p.Name }
          }
          for hid := range headingSet {
              // Headings don't have direct Get; iterate via parent project.
              // For simplicity in v1: skip heading resolution if not in cached projects.
              // (Real fix would be HeadingGet on the repository — defer to follow-up.)
              _ = hid
              _ = headings
          }
          return nameCacheLoadedMsg{tags: tags, areas: areas, projects: projects, headings: headings}
      }
  }
  ```
  Imports: `context`, `tea "github.com/charmbracelet/bubbletea"`, `app`, `id`, `task`. — `task test`

  NOTE on Headings: `Repository.HeadingList(ctx, projectID)` requires a project ID; there's no direct HeadingGet. For v1, leave `headings` empty (REQ-4.3 short-ID fallback applies) and add `HeadingGet` as a v2 spike (`Needs spike` in design Scope).

- [ ] 4. В `internal/tui/app_test.go` добавить `TestNameCache_LoadedMsgPopulatesModel`:
  ```go
  func TestNameCache_LoadedMsgPopulatesModel(t *testing.T) {
      m := newTestModel(t)
      tid := id.New()
      aid := id.New()
      msg := nameCacheLoadedMsg{
          tags:  map[id.ID]string{tid: "work"},
          areas: map[id.ID]string{aid: "home"},
          projects: map[id.ID]string{},
          headings: map[id.ID]string{},
      }
      m2, _ := m.Update(msg)
      mm := m2.(Model)
      require.Equal(t, "work", mm.tagNamesByID[tid])
      require.Equal(t, "home", mm.areaNamesByID[aid])
  }
  ```
  Test will fail until subtask 5. — `task test`

- [ ] 5. В `internal/tui/app.go` `Update` добавить case `nameCacheLoadedMsg` (после `tasksLoadedMsg` case): merge каждую map'у в соответствующее Model field (loop with assignment, не присваивание сразу — чтобы не терять накопленные имена из предыдущих fetch'ей):
  ```go
  case nameCacheLoadedMsg:
      for k, v := range msg.tags     { m.tagNamesByID[k] = v }
      for k, v := range msg.areas    { m.areaNamesByID[k] = v }
      for k, v := range msg.projects { m.projectNamesByID[k] = v }
      for k, v := range msg.headings { m.headingNamesByID[k] = v }
      return m, nil
  ```
  — `task test`

- [ ] 6. В `internal/tui/app.go` `Update` существующий `tasksLoadedMsg` case изменить — вместо `return m, nil` возвращать `tea.Batch` с `fetchNameCache`:
  ```go
  case tasksLoadedMsg:
      m.tasks = msg.tasks
      if m.cursor >= len(m.tasks) {
          m.cursor = max(0, len(m.tasks)-1)
      }
      return m, fetchNameCache(m.service, m.tasks)
  ```
  IMPORTANT: Это поменяет возврат с `(m, nil)` на `(m, cmd)`. Существующий тест `TestTUI_TasksLoadedPopulatesModel` сейчас игнорирует cmd (использует `m2, _ := m.Update(...)`) — должен продолжить проходить. — `task test`

- [ ] 7. В `internal/tui/app_test.go` добавить `TestNameCache_FetchCmdEmitsMsg`:
  ```go
  func TestNameCache_FetchCmdEmitsMsg(t *testing.T) {
      _, svc := newTestModelWithService(t)
      ctx := context.Background()
      a, err := svc.AddArea(ctx, "work")
      require.NoError(t, err)
      tg, err := svc.UpsertTag(ctx, "urgent")
      require.NoError(t, err)
      tk, err := svc.AddTask(ctx, app.AddTaskInput{
          Title:  "t1",
          AreaID: &a.ID,
          Tags:   []id.ID{tg.ID},
      })
      require.NoError(t, err)

      cmd := fetchNameCache(svc, []task.Task{tk})
      require.NotNil(t, cmd)
      msg := cmd()
      res, ok := msg.(nameCacheLoadedMsg)
      require.True(t, ok)
      require.Equal(t, "work", res.areas[a.ID])
      require.Equal(t, "urgent", res.tags[tg.ID])
  }
  ```
  — `task test`

After all subtasks: Run `task test-race && task lint`. T-1 tests + all baseline tests pass. Lint 0 issues.

---

## Task: T-3 — Implement Details Pane content

*_Requirements: REQ-2.1..2.12, REQ-3.1, REQ-3.2, REQ-4.2, REQ-4.3, REQ-5.1_*
*_Preservation: T-1 tests + T-2 tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать `viewDetails(m Model, width int) string`, `cursorTask(m Model) *task.Task`, `wrapAndTruncate(text string, width, maxLines int) string` в `details.go`. Все 11 полей в фиксированном порядке (Title → Status → Notes → Start → Due → Pinned → Area → Project → Heading → Tags → Someday). Omit nil/empty fields per REQ-2.11. Placeholder `(no task selected)` при пустом cursor.

IMPORTANT: Функции pure — никакого IO в render path. Все name lookup'ы — через `m.tagNamesByID` и т.д.

Subtasks:

- [ ] 1. В `internal/tui/details.go` добавить:
  ```go
  // cursorTask returns the task at m.cursor inside displayedTasks(m), or nil
  // when cursor is out of range or the visible list is empty.
  func cursorTask(m Model) *task.Task {
      disp := displayedTasks(m)
      if m.cursor < 0 || m.cursor >= len(disp) {
          return nil
      }
      return &disp[m.cursor]
  }
  ```
  — `task test`

- [ ] 2. В `internal/tui/details.go` добавить:
  ```go
  const detailsNotesMaxLines = 8

  // wrapAndTruncate soft-wraps text to width via lipgloss and truncates to
  // at most maxLines lines, appending "…" if truncation occurred. Empty input
  // returns "".
  func wrapAndTruncate(text string, width, maxLines int) string {
      if text == "" || width <= 0 {
          return ""
      }
      wrapped := lipgloss.NewStyle().Width(width).Render(text)
      lines := strings.Split(wrapped, "\n")
      if len(lines) <= maxLines {
          return wrapped
      }
      return strings.Join(lines[:maxLines], "\n") + "\n…"
  }
  ```
  Imports: `strings`, `"github.com/charmbracelet/lipgloss"`. — `task test`

- [ ] 3. Создать `internal/tui/details_test.go` с тестами для `wrapAndTruncate`:
  ```go
  func TestWrapAndTruncate_EmptyReturnsEmpty(t *testing.T) {
      require.Equal(t, "", wrapAndTruncate("", 40, 8))
  }

  func TestWrapAndTruncate_ShortPreserved(t *testing.T) {
      out := wrapAndTruncate("hello", 40, 8)
      require.Equal(t, "hello", strings.TrimRight(out, " "))
  }

  func TestWrapAndTruncate_LongIsTruncated(t *testing.T) {
      lines := make([]string, 20)
      for i := range lines { lines[i] = "line" }
      input := strings.Join(lines, "\n")
      out := wrapAndTruncate(input, 40, 8)
      require.Contains(t, out, "…")
      // count newlines: 8 lines = 8 lines, + "…" → 9 lines total
      require.Equal(t, 9, len(strings.Split(out, "\n")))
  }
  ```
  Imports: `strings`, `testing`, `"github.com/stretchr/testify/require"`. — `task test`

- [ ] 4. В `internal/tui/details.go` добавить:
  ```go
  // viewDetails renders the right pane content for dual-pane mode. Pure
  // function — reads only m and width. Returns "(no task selected)" when
  // cursorTask(m) is nil.
  func viewDetails(m Model, width int) string {
      t := cursorTask(m)
      if t == nil {
          return m.theme.Dim.Render("(no task selected)")
      }
      var lines []string
      // Title (full, wrapped)
      lines = append(lines, m.theme.Title.Render(wrapAndTruncate(t.Title, width, 4)))
      // Status
      lines = append(lines, "Status: "+statusLabel(t.Status))
      // Notes
      if t.Notes != "" {
          lines = append(lines, "")
          lines = append(lines, wrapAndTruncate(t.Notes, width, detailsNotesMaxLines))
      }
      // Dates
      if t.StartDate != nil {
          lines = append(lines, "Start:  "+t.StartDate.Format("2006-01-02"))
      }
      if t.Deadline != nil {
          lines = append(lines, "Due:    "+t.Deadline.Format("2006-01-02"))
      }
      if t.PinnedToday != nil {
          lines = append(lines, "Pinned: "+t.PinnedToday.Format("2006-01-02"))
      }
      // Relations
      if t.AreaID != nil {
          lines = append(lines, "Area:    "+resolveName(m.areaNamesByID, *t.AreaID))
      }
      if t.ProjectID != nil {
          lines = append(lines, "Project: "+resolveName(m.projectNamesByID, *t.ProjectID))
      }
      if t.HeadingID != nil {
          lines = append(lines, "Heading: "+resolveName(m.headingNamesByID, *t.HeadingID))
      }
      // Tags
      if len(t.Tags) > 0 {
          names := make([]string, 0, len(t.Tags))
          for _, tg := range t.Tags {
              names = append(names, resolveName(m.tagNamesByID, tg))
          }
          lines = append(lines, "Tags: "+strings.Join(names, ", "))
      }
      // Someday flag
      if t.Someday {
          lines = append(lines, m.theme.Dim.Render("Someday"))
      }
      return strings.Join(lines, "\n")
  }

  // resolveName looks up an ID in a name cache, falling back to id.Short(id)
  // when missing (REQ-4.3).
  func resolveName(cache map[id.ID]string, tid id.ID) string {
      if n, ok := cache[tid]; ok && n != "" {
          return n
      }
      return id.Short(tid)
  }

  // statusLabel maps Task.Status to a user-facing label.
  func statusLabel(s task.Status) string {
      switch s {
      case task.StatusOpen:
          return "Open"
      case task.StatusCompleted:
          return "Completed"
      case task.StatusCancelled:
          return "Cancelled"
      }
      return string(s)
  }
  ```
  Imports add: `task` (for Status), `id` (for Short). — `task test`

- [ ] 5. В `internal/tui/details_test.go` добавить `TestViewDetails_EmptyTasksPlaceholder`:
  ```go
  func TestViewDetails_EmptyTasksPlaceholder(t *testing.T) {
      m := newTestModel(t)
      m.tasks = nil
      out := viewDetails(m, 60)
      require.Contains(t, out, "(no task selected)")
  }
  ```
  Plus `TestViewDetails_OutOfRangeCursorPlaceholder`:
  ```go
  func TestViewDetails_OutOfRangeCursorPlaceholder(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "a", "b")
      m.cursor = 5
      out := viewDetails(m, 60)
      require.Contains(t, out, "(no task selected)")
  }
  ```
  — `task test`

- [ ] 6. В `internal/tui/details_test.go` добавить `TestViewDetails_TitleAndStatus`:
  ```go
  func TestViewDetails_TitleAndStatus(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "buy milk")
      out := viewDetails(m, 60)
      require.Contains(t, out, "buy milk")
      require.Contains(t, out, "Open")
  }
  ```
  — `task test`

- [ ] 7. В `internal/tui/details_test.go` добавить `TestViewDetails_Notes`:
  ```go
  func TestViewDetails_Notes(t *testing.T) {
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", Notes: "some note here"})
      require.NoError(t, err)
      m.tasks = []task.Task{tk}
      m.cursor = 0
      out := viewDetails(m, 60)
      require.Contains(t, out, "some note here")
  }
  ```
  — `task test`

- [ ] 8. В `internal/tui/details_test.go` добавить `TestViewDetails_Dates`:
  ```go
  func TestViewDetails_Dates(t *testing.T) {
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      start := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
      due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
      tk, err := svc.AddTask(ctx, app.AddTaskInput{
          Title:     "x",
          StartDate: &start,
          Deadline:  &due,
      })
      require.NoError(t, err)
      m.tasks = []task.Task{tk}
      m.cursor = 0
      out := viewDetails(m, 60)
      require.Contains(t, out, "Start:")
      require.Contains(t, out, "2026-05-25")
      require.Contains(t, out, "Due:")
      require.Contains(t, out, "2026-06-01")
  }
  ```
  — `task test`

- [ ] 9. В `internal/tui/details_test.go` добавить `TestViewDetails_OmitsNilFields`:
  ```go
  func TestViewDetails_OmitsNilFields(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "minimal")
      out := viewDetails(m, 60)
      require.NotContains(t, out, "Start:")
      require.NotContains(t, out, "Due:")
      require.NotContains(t, out, "Pinned:")
      require.NotContains(t, out, "Area:")
      require.NotContains(t, out, "Project:")
      require.NotContains(t, out, "Tags:")
      require.NotContains(t, out, "Someday")
  }
  ```
  — `task test`

- [ ] 10. В `internal/tui/details_test.go` добавить `TestViewDetails_RelationsAndTags`:
  ```go
  func TestViewDetails_RelationsAndTags(t *testing.T) {
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      a, err := svc.AddArea(ctx, "work")
      require.NoError(t, err)
      tg, err := svc.UpsertTag(ctx, "urgent")
      require.NoError(t, err)
      tk, err := svc.AddTask(ctx, app.AddTaskInput{
          Title:  "x",
          AreaID: &a.ID,
          Tags:   []id.ID{tg.ID},
      })
      require.NoError(t, err)
      m.tasks = []task.Task{tk}
      m.cursor = 0
      // Populate caches via the Cmd flow
      m.tagNamesByID[tg.ID] = "urgent"
      m.areaNamesByID[a.ID] = "work"
      out := viewDetails(m, 60)
      require.Contains(t, out, "Area:")
      require.Contains(t, out, "work")
      require.Contains(t, out, "Tags:")
      require.Contains(t, out, "urgent")
  }
  ```
  — `task test`

- [ ] 11. В `internal/tui/details_test.go` добавить `TestViewDetails_ShortIDFallback`:
  ```go
  func TestViewDetails_ShortIDFallback(t *testing.T) {
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      a, err := svc.AddArea(ctx, "work")
      require.NoError(t, err)
      tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: &a.ID})
      require.NoError(t, err)
      m.tasks = []task.Task{tk}
      m.cursor = 0
      // Do NOT populate areaNamesByID — should fall back to short-ID.
      out := viewDetails(m, 60)
      require.Contains(t, out, "Area:")
      require.Contains(t, out, id.Short(a.ID))
  }
  ```
  — `task test`

- [ ] 12. В `internal/tui/details_test.go` добавить `TestViewDetails_FieldOrder`:
  ```go
  func TestViewDetails_FieldOrder(t *testing.T) {
      // Build a task with every field populated; assert lines appear in
      // Title → Status → Notes → Start → Due → Pinned → Area → Project → Heading → Tags → Someday order.
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      a, err := svc.AddArea(ctx, "work")
      require.NoError(t, err)
      tg, err := svc.UpsertTag(ctx, "urgent")
      require.NoError(t, err)
      start := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
      due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
      pin := task.NewDate(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC))
      tk, err := svc.AddTask(ctx, app.AddTaskInput{
          Title: "T", Notes: "N", AreaID: &a.ID,
          StartDate: &start, Deadline: &due,
          Tags: []id.ID{tg.ID},
      })
      require.NoError(t, err)
      tk.PinnedToday = &pin
      tk.Someday = true
      m.tasks = []task.Task{tk}
      m.cursor = 0
      m.areaNamesByID[a.ID] = "work"
      m.tagNamesByID[tg.ID] = "urgent"
      out := viewDetails(m, 60)
      iStatus := strings.Index(out, "Status:")
      iStart := strings.Index(out, "Start:")
      iDue := strings.Index(out, "Due:")
      iPinned := strings.Index(out, "Pinned:")
      iArea := strings.Index(out, "Area:")
      iTags := strings.Index(out, "Tags:")
      iSomeday := strings.Index(out, "Someday")
      require.Less(t, iStatus, iStart)
      require.Less(t, iStart, iDue)
      require.Less(t, iDue, iPinned)
      require.Less(t, iPinned, iArea)
      require.Less(t, iArea, iTags)
      require.Less(t, iTags, iSomeday)
  }
  ```
  — `task test`

After all subtasks: Run `task test-race && task lint`. T-1, T-2 tests must still pass.

---

## Task: T-4 — Implement Layout decision and pane widths

*_Requirements: REQ-1.1, REQ-1.2, REQ-1.3, REQ-1.4, REQ-1.5, REQ-5.1, REQ-5.2, REQ-6.1_*
*_Preservation: T-1 tests + T-2 tests + T-3 tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать `isDualPane(m Model) bool`, pane width math, и `viewBody()` dispatcher в `app.go`. Интеграция с существующим `View()`.

CRITICAL: T-1 preservation tests должны продолжать проходить — единственное изменение в `View()` это добавление dual-pane ветки. Single-pane fallback идентичен текущему.

Subtasks:

- [ ] 1. В `internal/tui/details.go` добавить:
  ```go
  const (
      dualPaneMinWidth = 100
      listPaneShare   = 0.45
  )

  // isDualPane reports whether the renderer should use horizontal split.
  // Activated only when:
  //   - terminal width >= dualPaneMinWidth (so width > 0 too)
  //   - screen is screenList (editor and help force single-pane)
  // Filter mode and quick-entry overlay do NOT disable dual-pane.
  func isDualPane(m Model) bool {
      if m.width < dualPaneMinWidth {
          return false
      }
      if m.screen == screenEditor || m.screen == screenHelp {
          return false
      }
      return true
  }

  // paneWidths returns (listWidth, detailsWidth) for the dual-pane layout.
  // The 1-column border between panes is allocated separately. Invariant:
  // listWidth + 1 + detailsWidth == m.width.
  func paneWidths(totalWidth int) (int, int) {
      list := int(float64(totalWidth-1) * listPaneShare)
      details := totalWidth - 1 - list
      return list, details
  }
  ```
  — `task test`

- [ ] 2. В `internal/tui/details_test.go` добавить:
  ```go
  func TestIsDualPane_WideTerminal(t *testing.T) {
      m := newTestModel(t)
      m.width = 100
      require.True(t, isDualPane(m))
  }

  func TestIsDualPane_NarrowTerminal(t *testing.T) {
      m := newTestModel(t)
      m.width = 99
      require.False(t, isDualPane(m))
  }

  func TestIsDualPane_ZeroWidth(t *testing.T) {
      m := newTestModel(t)
      // m.width = 0 by default
      require.False(t, isDualPane(m))
  }

  func TestIsDualPane_EditorScreen(t *testing.T) {
      m := newTestModel(t)
      m.width = 200
      m.screen = screenEditor
      require.False(t, isDualPane(m))
  }

  func TestIsDualPane_HelpScreen(t *testing.T) {
      m := newTestModel(t)
      m.width = 200
      m.screen = screenHelp
      require.False(t, isDualPane(m))
  }

  func TestIsDualPane_FilteringAllowed(t *testing.T) {
      m := newTestModel(t)
      m.width = 200
      m.filtering = true
      require.True(t, isDualPane(m))
  }

  func TestPaneWidths_SumEqualsTotal(t *testing.T) {
      for _, w := range []int{100, 120, 150, 200, 300} {
          list, details := paneWidths(w)
          require.Equal(t, w, list+1+details, "width %d", w)
      }
  }
  ```
  — `task test`

- [ ] 3. В `internal/tui/app.go` добавить новый метод `viewBody`:
  ```go
  // viewBody dispatches between single-pane and dual-pane rendering.
  // Returns the body content (header and footer are wrapped separately by View).
  func (m Model) viewBody() string {
      if !isDualPane(m) {
          return m.viewList()
      }
      listW, detailsW := paneWidths(m.width)
      left := lipgloss.NewStyle().Width(listW).Render(m.viewList())
      right := lipgloss.NewStyle().
          Width(detailsW).
          Border(lipgloss.DoubleBorder(), false, false, false, true).
          BorderForeground(m.theme.Help.GetForeground()).
          PaddingLeft(1).
          Render(viewDetails(m, detailsW-2))  // -2 for border + padding
      return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
  }
  ```
  — `task test`

- [ ] 4. В `internal/tui/app.go` `View()` рефактор: использовать `viewBody()` для `screenList` default case:
  ```go
  func (m Model) View() string {
      if m.confirm != nil {
          modal := m.theme.Modal.Render(fmt.Sprintf("%s %d tasks? (y/n)", m.confirm.action.label(), len(m.confirm.ids)))
          body := lipgloss.JoinVertical(lipgloss.Left, m.viewBody(), modal)
          return lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
      }
      var body string
      switch m.screen {
      case screenHelp:
          body = m.viewHelp()
      case screenQuickEntry:
          body = lipgloss.JoinVertical(lipgloss.Left, m.viewBody(), m.viewQuickEntry())
      case screenEditor:
          body = m.editor.View(m.theme, m.editorWidth())
      default:
          body = m.viewBody()
      }
      return lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
  }
  ```
  ВАЖНО: `screenHelp` и `screenEditor` не вызывают `viewBody()` — они показывают full-pane содержимое (REQ-1.6/1.7). Все остальные пути проходят через `viewBody()`, и `viewBody()` сам решает single vs dual.
  — `task test`

- [ ] 5. В `internal/tui/details_test.go` добавить `TestViewBody_SinglePaneNoBorder`:
  ```go
  func TestViewBody_SinglePaneNoBorder(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "task one")
      m.width = 80
      out := m.viewBody()
      require.NotContains(t, out, "║")
      require.Contains(t, out, "task one")
  }
  ```
  Plus `TestViewBody_DualPaneRendersBothAndBorder`:
  ```go
  func TestViewBody_DualPaneRendersBothAndBorder(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "task one")
      m.width = 120
      out := m.viewBody()
      require.Contains(t, out, "║", "double-line border must be present")
      require.Contains(t, out, "task one", "list content must be present")
      require.Contains(t, out, "Status:", "details content must be present")
  }
  ```
  — `task test`

- [ ] 6. В `internal/tui/details_test.go` добавить `TestViewBody_FilterPreservesDualPane`:
  ```go
  func TestViewBody_FilterPreservesDualPane(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "task one")
      m.width = 120
      m.filtering = true
      m.filterQuery = "task"
      out := m.viewBody()
      require.Contains(t, out, "║", "filter mode must NOT disable dual-pane")
  }
  ```
  — `task test`

- [ ] 7. В `internal/tui/details_test.go` добавить `TestCursorChange_DetailsUpdates`:
  ```go
  func TestCursorChange_DetailsUpdates(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "alpha-task", "beta-task")
      m.width = 120
      m.cursor = 0
      out0 := m.viewBody()
      require.Contains(t, out0, "alpha-task")
      m.cursor = 1
      out1 := m.viewBody()
      require.Contains(t, out1, "beta-task")
      // Cursor 0 task title should not be the primary one in cursor=1 render's details portion.
      // (The list portion still contains both.) Verify Status line follows the new task.
      require.NotEqual(t, out0, out1)
  }
  ```
  — `task test`

After all subtasks: Run `task test-race && task lint`. ALL T-1, T-2, T-3 tests pass. T-1's `TestTUI_ZeroWidthRendersSinglePane` is critical — it locks REQ-1.3.

---

## Task: T-5 — Mode interactions (editor/help full-pane, confirm/quick-entry stacking)

*_Requirements: REQ-1.6, REQ-1.7, REQ-1.8, REQ-6.2, REQ-6.3_*
*_Preservation: all T-1..T-4 tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Проверить и закрепить тестами что editor / help — full-pane (даже на широком терминале), confirm modal стекается под весь split-body, quick-entry overlay стекается аналогично.

NOTE: T-4 уже реализовал `View()` с правильной диспетчеризацией. Эти тесты — verification, что behavior соответствует REQ-1.6, REQ-1.7, REQ-1.8, REQ-6.2.

Subtasks:

- [ ] 1. В `internal/tui/details_test.go` добавить `TestLayout_EditorFullPaneIgnoresWidth`:
  ```go
  func TestLayout_EditorFullPaneIgnoresWidth(t *testing.T) {
      m, _, tasks := setupModelWithInboxTasks(t, "x")
      m.width = 200
      // Open editor
      m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
      mm := m2.(Model)
      require.Equal(t, screenEditor, mm.screen)
      out := mm.View()
      require.NotContains(t, out, "║", "editor screen must NOT render double-pane border")
      _ = tasks
  }
  ```
  — `task test`

- [ ] 2. В `internal/tui/details_test.go` добавить `TestLayout_HelpFullPaneIgnoresWidth`:
  ```go
  func TestLayout_HelpFullPaneIgnoresWidth(t *testing.T) {
      m := newTestModel(t)
      m.width = 200
      m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
      mm := m2.(Model)
      require.Equal(t, screenHelp, mm.screen)
      out := mm.View()
      require.NotContains(t, out, "║", "help screen must NOT render double-pane border")
  }
  ```
  — `task test`

- [ ] 3. В `internal/tui/details_test.go` добавить `TestLayout_ConfirmStacksBelowDual`:
  ```go
  func TestLayout_ConfirmStacksBelowDual(t *testing.T) {
      m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c", "d", "e")
      m.width = 120
      for _, tk := range tasks {
          m.selected[tk.ID] = struct{}{}
      }
      // Trigger confirm modal via 'c' key
      m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
      mm := m2.(Model)
      require.NotNil(t, mm.confirm)
      out := mm.View()
      // Both dual-pane border AND modal text must be present
      require.Contains(t, out, "║", "dual-pane border must still render")
      require.Contains(t, out, "Complete 5 tasks?", "confirm modal must render")
      // Modal must appear AFTER the border (i.e. below split body in the output string)
      iBorder := strings.Index(out, "║")
      iModal := strings.Index(out, "Complete 5 tasks?")
      require.Less(t, iBorder, iModal, "modal must stack below split body")
  }
  ```
  — `task test`

- [ ] 4. В `internal/tui/details_test.go` добавить `TestLayout_QuickEntryStacksBelowDual`:
  ```go
  func TestLayout_QuickEntryStacksBelowDual(t *testing.T) {
      m, _, _ := setupModelWithInboxTasks(t, "x")
      m.width = 120
      m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
      mm := m2.(Model)
      require.Equal(t, screenQuickEntry, mm.screen)
      out := mm.View()
      require.Contains(t, out, "║", "dual-pane must still render under quick-entry")
      require.Contains(t, out, "Quick Entry", "quick-entry modal must render")
  }
  ```
  — `task test`

- [ ] 5. В `internal/tui/details_test.go` добавить `TestUpdate_TasksLoadedDispatchesNameCacheFetch`:
  ```go
  func TestUpdate_TasksLoadedDispatchesNameCacheFetch(t *testing.T) {
      m, svc := newTestModelWithService(t)
      ctx := context.Background()
      a, err := svc.AddArea(ctx, "work")
      require.NoError(t, err)
      tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: &a.ID})
      require.NoError(t, err)
      m2, cmd := m.Update(tasksLoadedMsg{tasks: []task.Task{tk}})
      mm := m2.(Model)
      require.Equal(t, []task.Task{tk}, mm.tasks)
      require.NotNil(t, cmd, "tasksLoadedMsg must trigger fetchNameCache Cmd")
      // Execute the cmd and verify it emits nameCacheLoadedMsg
      msg := cmd()
      // tea.Batch composes; if it's a batchMsg, find the nameCacheLoadedMsg.
      // Simpler: just check via assertion that a name cache resolution happened.
      // For this test, fetchNameCache is the sole Cmd in this path, so:
      if batch, ok := msg.(tea.BatchMsg); ok {
          // walk through
          for _, sub := range batch {
              if sub == nil { continue }
              if nm, ok := sub().(nameCacheLoadedMsg); ok {
                  require.Equal(t, "work", nm.areas[a.ID])
                  return
              }
          }
          t.Fatal("no nameCacheLoadedMsg in batch")
      }
      // direct emission case
      if nm, ok := msg.(nameCacheLoadedMsg); ok {
          require.Equal(t, "work", nm.areas[a.ID])
          return
      }
      t.Fatalf("unexpected msg type: %T", msg)
  }
  ```
  NOTE: `tea.BatchMsg` is the type emitted when a Cmd returns multiple commands; for v1 only `fetchNameCache` is dispatched, so msg should be `nameCacheLoadedMsg` directly. Defensive code handles both shapes. — `task test`

After all subtasks: Run `task test-race && task lint`. All T-1..T-4 tests still pass.

---

## Task: T-6 — Write property-based tests batch

*_Requirements: REQ-1.1..6.3 (все)_*
*_Preservation: все CP_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать 14 property-тестов из design §2.8 через `pgregory.net/rapid`. Каждый CP-N покрыт `TestProp_*`.

Subtasks:

- [ ] 1. В `internal/tui/details_test.go` добавить PBT для layout properties: `TestProp_LayoutModeExclusion` (CP-1), `TestProp_PaneWidthArithmetic` (CP-2), `TestProp_BorderRenders` (CP-3), `TestProp_EditorHelpForceFullPane` (CP-4). Генераторы: `rapid.IntRange(0, 300)` for width, `rapid.SampledFrom([]screenKind{screenList, screenEditor, screenHelp, screenQuickEntry})`. — `task test`

- [ ] 2. В `internal/tui/details_test.go` добавить PBT для details content: `TestProp_FieldOrderInvariant` (CP-5), `TestProp_NilFieldsOmitted` (CP-6), `TestProp_NotesTruncationCorrect` (CP-7), `TestProp_PlaceholderOnEmptyCursor` (CP-8). Build task с randomized subset полей (some nil, some populated). — `task test`

- [ ] 3. В `internal/tui/details_test.go` добавить `TestProp_NoRepoAccessInView` (CP-9). Use a wrapped service whose `Repo()` returns a `panicRepo` that fails on any method call; ensure `viewDetails(m, w)` does not panic when name caches are populated. NOTE: создать helper `panicRepo` implements `storage.Repository` interface but every method calls `t.Fatal`. — `task test`

- [ ] 4. В `internal/tui/details_test.go` добавить `TestProp_NameCachePopulation` (CP-10) и `TestProp_ShortIDFallback` (CP-11). Для CP-10 — random count of tasks с random refs к area/tag IDs; после `fetchNameCache` execution, все resolveable IDs в cache. Для CP-11 — random task с TagID НЕ в cache → `viewDetails` contains `id.Short(TagID)`. — `task test`

- [ ] 5. В `internal/tui/details_test.go` добавить `TestProp_CursorChangeReflects` (CP-12), `TestProp_FilterPreservesDualPane` (CP-13), `TestProp_ConfirmStacksBelowDual` (CP-14). — `task test`

- [ ] 6. Запустить `task test-race -count=2` для проверки стабильности PBT.

After all subtasks: Run `task test-race && task lint`. Все тесты зелёные.

---

## Task: T-7 — GATE Checkpoint

*_Requirements: ALL_*
*_Complexity: mechanical_*

CRITICAL: Эта задача — ПОСЛЕДНЯЯ. Не делать до полного завершения T-1..T-6.

Instructions:

1. Запустить полный suite: `task test`. Подтвердить 100% PASS.
2. Запустить race-detector: `task test-race`. Подтвердить отсутствие race-conditions.
3. Запустить `task build`. Подтвердить успешную компиляцию.
4. Запустить `task lint`. Подтвердить 0 issues.
5. Запустить `gofmt -l internal/tui/`. Подтвердить отсутствие необработанных файлов (goimports может быть недоступен — используем gofmt).
6. Сверить Coverage Matrix (preamble): каждое REQ-X.Y → ≥1 test pass; каждое CP-N → property-test passing.
7. Открыть `internal/tui/app.go` и подтвердить:
   - `Model` имеет 4 новых name-cache поля
   - `Update` имеет `nameCacheLoadedMsg` case
   - `tasksLoadedMsg` case возвращает `fetchNameCache` Cmd
   - `View()` использует `m.viewBody()` для screenList default + screenQuickEntry + confirm-overlay
   - `screenEditor` и `screenHelp` cases НЕ используют `viewBody()` (full-pane)
8. Открыть `internal/tui/details.go` и подтвердить наличие всех функций из design §2.3.
9. Запустить локально `task run` (если не interactive — `./bin/todushka --help` для smoke). Подтвердить компилируется.
10. Ручная проверка scenarios (можно опустить если запуск интерактивный недоступен):
    - На широком терминале (≥100 cols): split с details справа.
    - На узком терминале (<100): single-pane как раньше.
    - В editor (Enter): full-pane editor, нет split.
    - С N≥5 selected + 'c': confirm modal под split-body.
11. Если любая проверка fails — вернуться к соответствующему T-N task'у. НЕ закрывать GATE.

After all checks pass: продвинуть пайплайн на `review` фазу через `sh ./scripts/pipeline.sh approve` (пользователь нажимает approve).

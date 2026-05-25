# Editor When + Picker (v0.6.0) — Task Plan

## Preamble

### Work Type Classification

**Pure feature** с **preservation surface**: 213 existing tests должны проходить. Editor field count меняется 6→9. Существующий `TestEditor_FieldCountIsSix` обновляется. `NewEditor` signature меняется — затрагивает все call sites.

### Test Style Source

**Tier 2** — adjacent tests
- `internal/tui/editor_test.go`, `internal/tui/app_test.go` — установленные fixtures (`newTestModel`, `newTestModelWithService`, `setupModelWithInboxTasks`).
- Property tests: `pgregory.net/rapid`.

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
| REQ-1.1 Inbox label | T-4 | CP-1 |
| REQ-1.2 Anytime label | T-4 | CP-1 |
| REQ-1.3 hint removed | T-4 | — |
| REQ-1.4 Anytime → Someday=false | (existing behavior; preserved by T-1 + T-7) | — |
| REQ-1.5 Someday → Someday=true | (existing; preserved) | — |
| REQ-1.6 Space cycles | (existing; preserved) | — |
| REQ-2.1 Area textinput exists | T-2, T-5 | — |
| REQ-2.2 Area pre-fill | T-3 | CP-2 |
| REQ-2.3 Empty area on open | T-3 | — |
| REQ-2.4 Empty area clears | T-6 | CP-3 |
| REQ-2.5 Invalid area error | T-6 | CP-5 |
| REQ-3.1 Project textinput exists | T-2, T-5 | — |
| REQ-3.2 Project pre-fill | T-3 | CP-2 |
| REQ-3.3 Empty project on open | T-3 | — |
| REQ-3.4 Empty project clears both IDs | T-6 | CP-4 |
| REQ-3.5 Project ambiguous error | T-6 | CP-6 |
| REQ-4.1 Heading textinput exists | T-2, T-5 | — |
| REQ-4.2 Heading pre-fill | T-3 | CP-2 |
| REQ-4.3 Empty heading on open | T-3 | — |
| REQ-4.4 Empty heading clears | T-6 | — |
| REQ-4.5 Heading without project error | T-6 | CP-7 |
| REQ-4.6 Heading found in project | T-6 | CP-8, CP-12 |
| REQ-5.1 Field order | T-2 | CP-10 |
| REQ-5.2 fieldCount=9 | T-2 | CP-9 |
| REQ-5.3 Tab cycle preserves first 2 | T-2, T-1 (preservation) | CP-10 |
| REQ-5.4 New enum constants | T-2 | CP-9 |
| REQ-6.1 View renders 3 new fields | T-5 | — |
| REQ-6.2 Error display via m.err | T-6 | — |

26 REQs → 8 tasks → 13 CPs.

---

## Task Order

```
T-1 GREEN (baseline preservation)
  → T-2 CODE (enum extension + struct fields + focus/cycle helpers)
    → T-3 CODE (NewEditor pre-fill — signature change + Repo lookups)
      → T-4 CODE (whenLabel helper + View context-aware label — Part A done)
      → T-5 CODE (View renders 3 new field blocks)
        → T-6 CODE (ApplyAndSave sequential resolve — Part B core)
          → T-7 GREEN (Property-based tests batch — 13 CPs)
            → T-8 GATE (Checkpoint)
```

T-4 и T-5 — параллельные (оба touching View() but different sections), но представлены sequentially для clarity.

---

## Task: T-1 — Baseline preservation

*_Requirements: REQ-5.3_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Подтвердить что 213 тестов проходят на baseline до изменений.

Subtasks:
- [ ] 1. `go clean -testcache && task test-race` — all packages PASS.
- [ ] 2. `task lint` — 0 issues.

After all subtasks: baseline established.

---

## Task: T-2 — Editor enum extension + struct fields + focus/cycle

*_Requirements: REQ-2.1, REQ-3.1, REQ-4.1, REQ-5.1, REQ-5.2, REQ-5.4_*
*_Preservation: existing 213 tests (TestEditor_FieldCountIsSix gets updated)_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Расширить `editorField` enum (+ 3 const), struct (+ 3 textinput fields), и helper functions (`focusCurrent`, `nextField`, `prevField`, `UpdateForm`) под новый цикл.

Subtasks (each touches one file; run `task test` after each):

1. В `internal/tui/editor.go` обновить `editorField` const block:
```go
const (
    fieldTitle editorField = iota
    fieldNotes
    fieldStart
    fieldDeadline
    fieldArea       // [NEW]
    fieldProject    // [NEW]
    fieldHeading    // [NEW]
    fieldTags
    fieldWhen
    fieldCount      // = 9
)
```

2. В `internal/tui/editor.go` обновить `EditorModel` struct — добавить 3 поля `area`, `project`, `heading textinput.Model` между `deadline` и `tags`.

3. В `internal/tui/editor.go` обновить `focusCurrent`: добавить 3 case'а для новых полей. Каждый делает `m.X.Blur()` для всех других + `m.X.Focus()` для current.

4. В `internal/tui/editor.go` обновить `UpdateForm`: добавить 3 case'а dispatching keymsg to `m.area.Update(msg)`, `m.project.Update(msg)`, `m.heading.Update(msg)`.

5. `nextField` и `prevField` уже работают через arithmetic mod `fieldCount` — но verify: `nextField()` from `fieldDeadline` → `fieldArea` → `fieldProject` → `fieldHeading` → `fieldTags`. Run `task test` to ensure no breakage.

6. В `internal/tui/editor_test.go` обновить `TestEditor_FieldCountIsSix` → `TestEditor_FieldCountIsNine`:
```go
func TestEditor_FieldCountIsNine(t *testing.T) {
    require.Equal(t, 9, int(fieldCount))
}
```

7. Запустить `task test-race && task lint`. ВНИМАНИЕ: на этом этапе компиляция должна проходить (textinputs существуют но не инициализированы — `NewEditor` ещё не обновлён). NewEditor возвращает struct с нулевыми textinputs — UpdateForm для новых полей будет no-op. Existing tests continue passing.

After all subtasks: foundation готова для T-3.

---

## Task: T-3 — NewEditor pre-fill (signature change + Repo lookups)

*_Requirements: REQ-2.2, REQ-2.3, REQ-3.2, REQ-3.3, REQ-4.2, REQ-4.3_*
*_Preservation: T-1, T-2 tests_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Обновить `NewEditor` signature: `NewEditor(ctx context.Context, t task.Task, svc *app.Service) EditorModel`. Pre-fill 3 new textinput'ов через `Repo.AreaGet`, `Repo.ProjectGet`, `Repo.HeadingList` (find by HeadingID).

**CRITICAL:** Это меняет signature — все callers (app.go `openEditor`, tests) должны быть обновлены.

Subtasks:

1. В `internal/tui/editor.go` обновить `NewEditor(t task.Task)` → `NewEditor(ctx context.Context, t task.Task, svc *app.Service)`. В body инициализировать 3 new textinputs (placeholder, char limit), pre-fill через `svc.Repo().AreaGet`/`ProjectGet`/`HeadingList`. На errors — leave empty (per design §2.7):
```go
areaIn := textinput.New()
areaIn.Placeholder = "area name (empty = Inbox)"
areaIn.CharLimit = 100
if t.AreaID != nil {
    if a, err := svc.Repo().AreaGet(ctx, *t.AreaID); err == nil {
        areaIn.SetValue(a.Name)
    }
}
// аналогично для project, heading
```

2. В `internal/tui/editor.go` `EditorModel` literal — добавить `area: areaIn, project: projectIn, heading: headingIn`.

3. В `internal/tui/app.go` найти `openEditor` (`grep -n "NewEditor" internal/tui/app.go`). Обновить вызов:
```go
m.editor = NewEditor(context.Background(), *sel, m.service)
```
Imports: добавить `"context"` если не присутствует в app.go (вероятно уже есть).

4. В `internal/tui/app_test.go` найти все вызовы `NewEditor(`. Обновить — добавить `context.Background()` и сервис. Use `grep -rn "NewEditor(" internal/tui/`. Tests могут быть в:
- `TestTUI_EnterOpensEditor`
- `TestTUI_EditorEscClosesWithoutSave`
- `TestTUI_EditorTabCyclesFields`
- `editor_test.go` (новые tests T-4 ещё не написаны)

Обновите все вызовы — большинство уже have `newTestModelWithService(t)` доступ к `svc`. Если нет — нужно построить.

5. В `internal/tui/editor_test.go` добавить:
```go
func TestEditor_NewEditorPrefillArea(t *testing.T) {
    _, svc := newTestModelWithService(t)
    ctx := context.Background()
    a, err := svc.AddArea(ctx, "work")
    require.NoError(t, err)
    tk := task.Task{ID: id.New(), Title: "x", AreaID: &a.ID}
    ed := NewEditor(ctx, tk, svc)
    require.Equal(t, "work", ed.area.Value())
}

func TestEditor_NewEditorPrefillProject(t *testing.T) {
    _, svc := newTestModelWithService(t)
    ctx := context.Background()
    p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "todushka"})
    require.NoError(t, err)
    tk := task.Task{ID: id.New(), Title: "x", ProjectID: &p.ID}
    ed := NewEditor(ctx, tk, svc)
    require.Equal(t, "todushka", ed.project.Value())
}

func TestEditor_NewEditorEmptyArea(t *testing.T) {
    _, svc := newTestModelWithService(t)
    tk := task.Task{ID: id.New(), Title: "x"}
    ed := NewEditor(context.Background(), tk, svc)
    require.Empty(t, ed.area.Value())
    require.Empty(t, ed.project.Value())
    require.Empty(t, ed.heading.Value())
}
```

6. Heading pre-fill test — `TestEditor_NewEditorPrefillHeading`. Создать project + heading + task referencing them. Verify `ed.heading.Value() == heading.Name`. Сначала нужно убедиться что `Service.AddHeading` доступен. Looking at app/service.go — yes, `AddHeading(ctx, projectID, name)` exists.

7. Запустить `task test-race && task lint`. All 213 + 4 new tests passing.

After all subtasks: pre-fill готов; foundation extension complete.

---

## Task: T-4 — whenLabel + View context-aware label (Part A)

*_Requirements: REQ-1.1, REQ-1.2, REQ-1.3_*
*_Preservation: T-1, T-2, T-3 tests_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: Реализовать `whenLabel(t task.Task) string` helper. Обновить View() — заменить hardcoded "Anytime" на `whenLabel(m.original)` call. Удалить старый hint `(will appear in Inbox without Area/Project)`.

Subtasks:

1. В `internal/tui/editor.go` добавить `whenLabel`:
```go
// whenLabel returns the "When" section's primary label based on whether the
// task is currently routed to Inbox (no Area/Project) or Anytime (has either).
//
//   - "Inbox"   when t.AreaID == nil AND t.ProjectID == nil
//   - "Anytime" otherwise
func whenLabel(t task.Task) string {
    if t.AreaID == nil && t.ProjectID == nil {
        return "Inbox"
    }
    return "Anytime"
}
```

2. В `internal/tui/editor.go` `View()` найти существующий block (lines ~230-247 в текущем editor.go):
```go
whenBody := fmt.Sprintf("%s Anytime\n%s Someday", anytimeBullet, somedayBullet)
```
Заменить `"Anytime"` на context-aware label:
```go
primaryLabel := whenLabel(m.original)
whenBody := fmt.Sprintf("%s %s\n%s Someday", anytimeBullet, primaryLabel, somedayBullet)
```

3. В `internal/tui/editor.go` `View()` УДАЛИТЬ hint block:
```go
if m.when == whenAnytime && m.original.AreaID == nil && m.original.ProjectID == nil {
    whenSection += "\n" + theme.Dim.Render("(will appear in Inbox without Area/Project)")
}
```
Полностью убрать. Label теперь "Inbox" — этого достаточно.

4. В `internal/tui/editor_test.go` добавить:
```go
func TestWhenLabel_InboxForUnrelatedTask(t *testing.T) {
    require.Equal(t, "Inbox", whenLabel(task.Task{}))
}

func TestWhenLabel_AnytimeForAreaTask(t *testing.T) {
    aid := id.New()
    require.Equal(t, "Anytime", whenLabel(task.Task{AreaID: &aid}))
}

func TestWhenLabel_AnytimeForProjectTask(t *testing.T) {
    pid := id.New()
    require.Equal(t, "Anytime", whenLabel(task.Task{ProjectID: &pid}))
}

func TestEditor_ViewShowsInboxLabel(t *testing.T) {
    _, svc := newTestModelWithService(t)
    tk := task.Task{ID: id.New(), Title: "x"}
    ed := NewEditor(context.Background(), tk, svc)
    out := ed.View(NewTheme(), 80)
    require.Contains(t, out, "Inbox")
    require.NotContains(t, out, "Anytime")
}

func TestEditor_ViewShowsAnytimeLabel(t *testing.T) {
    _, svc := newTestModelWithService(t)
    ctx := context.Background()
    a, err := svc.AddArea(ctx, "work")
    require.NoError(t, err)
    tk := task.Task{ID: id.New(), Title: "x", AreaID: &a.ID}
    ed := NewEditor(ctx, tk, svc)
    out := ed.View(NewTheme(), 80)
    require.Contains(t, out, "Anytime")
}

func TestEditor_ViewHidesOldHint(t *testing.T) {
    _, svc := newTestModelWithService(t)
    tk := task.Task{ID: id.New(), Title: "x"}
    ed := NewEditor(context.Background(), tk, svc)
    out := ed.View(NewTheme(), 80)
    require.NotContains(t, out, "will appear in Inbox")
}
```

5. Существующие tests `TestEditor_HintShownWhenAnytimeNoAreaProject` и `TestEditor_HintHiddenForSomeday` (из v0.5.0) — нужно либо удалить (hint исчез), либо переписать. Прочитать их и решить:
- `TestEditor_HintShownWhenAnytimeNoAreaProject`: hint исчез → тест надо **удалить** (REQ-1.3 явно требует отсутствие hint).
- `TestEditor_HintHiddenForSomeday`: hint hidden was correct → теперь "will appear in Inbox" вообще не существует → удалить.

Заменить эти 2 тестa на 2 новых из subtask 4 (`TestEditor_ViewShowsInboxLabel`, `TestEditor_ViewShowsAnytimeLabel`).

6. Запустить `task test-race && task lint`.

After all subtasks: Part A complete — label honest.

---

## Task: T-5 — View renders 3 new field blocks

*_Requirements: REQ-6.1, REQ-6.2_*
*_Preservation: T-1..T-4_*
*_Test_Style: Tier 2_*
*_Complexity: standard_*

GOAL: В `View()` добавить визуальный рендеринг для Area, Project, Heading textinput'ов между Deadline и Tags.

Subtasks:

1. В `internal/tui/editor.go` `View()` найти JoinVertical block:
```go
body := lipgloss.JoinVertical(lipgloss.Left,
    theme.Title.Render("Edit task"),
    "",
    field("Title", m.title.View(), m.focus == fieldTitle),
    field("Notes", m.notes.View(), m.focus == fieldNotes),
    field("Start", m.start.View(), m.focus == fieldStart),
    field("Deadline", m.deadline.View(), m.focus == fieldDeadline),
    field("Tags", m.tags.View(), m.focus == fieldTags),
    whenSection,
    ...
)
```

Вставить 3 новых field'а между Deadline и Tags:
```go
    field("Deadline", m.deadline.View(), m.focus == fieldDeadline),
    field("Area", m.area.View(), m.focus == fieldArea),
    field("Project", m.project.View(), m.focus == fieldProject),
    field("Heading", m.heading.View(), m.focus == fieldHeading),
    field("Tags", m.tags.View(), m.focus == fieldTags),
    whenSection,
```

2. В `internal/tui/editor_test.go` добавить:
```go
func TestEditor_ViewRendersAllNewFields(t *testing.T) {
    _, svc := newTestModelWithService(t)
    tk := task.Task{ID: id.New(), Title: "x"}
    ed := NewEditor(context.Background(), tk, svc)
    out := ed.View(NewTheme(), 80)
    require.Contains(t, out, "Area")
    require.Contains(t, out, "Project")
    require.Contains(t, out, "Heading")
}
```

3. Запустить `task test-race && task lint`.

After all subtasks: visual rendering complete.

---

## Task: T-6 — ApplyAndSave sequential resolve (Part B core)

*_Requirements: REQ-2.4, REQ-2.5, REQ-3.4, REQ-3.5, REQ-4.4, REQ-4.5, REQ-4.6_*
*_Preservation: T-1..T-5_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Расширить `ApplyAndSave` — добавить sequential resolve для Area, Project, Heading. First error abort'ает. Auto-clear heading on project change.

CRITICAL: This is the core of Part B. Test all error paths.

Subtasks:

1. В `internal/tui/editor.go` `ApplyAndSave` — добавить блок resolve perед `svc.EditTask(ctx, t)`. Внутри:

```go
// Resolve Area
areaName := strings.TrimSpace(m.area.Value())
if areaName == "" {
    t.AreaID = nil
} else {
    a, err := svc.Repo().AreaFindByNormalized(ctx, areaName)
    if err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            return task.Task{}, fmt.Errorf("area %q not found", areaName)
        }
        return task.Task{}, fmt.Errorf("area %q: %w", areaName, err)
    }
    t.AreaID = &a.ID
}

// Resolve Project
projectName := strings.TrimSpace(m.project.Value())
if projectName == "" {
    t.ProjectID = nil
    t.HeadingID = nil
} else {
    matches, err := svc.Repo().ProjectFindByName(ctx, projectName)
    if err != nil {
        return task.Task{}, fmt.Errorf("project %q: %w", projectName, err)
    }
    if len(matches) == 0 {
        return task.Task{}, fmt.Errorf("project %q not found", projectName)
    }
    if len(matches) > 1 {
        return task.Task{}, fmt.Errorf("project %q is ambiguous (%d matches), use CLI to disambiguate", projectName, len(matches))
    }
    newPID := matches[0].ID
    // Detect project change → auto-clear heading (ADR-5)
    if m.original.ProjectID == nil || newPID != *m.original.ProjectID {
        t.HeadingID = nil
    }
    t.ProjectID = &newPID
}

// Resolve Heading
headingName := strings.TrimSpace(m.heading.Value())
if headingName == "" {
    t.HeadingID = nil
} else {
    if t.ProjectID == nil {
        return task.Task{}, fmt.Errorf("heading %q requires a project", headingName)
    }
    headings, err := svc.Repo().HeadingList(ctx, *t.ProjectID)
    if err != nil {
        return task.Task{}, fmt.Errorf("heading %q: %w", headingName, err)
    }
    var found bool
    for _, h := range headings {
        if strings.EqualFold(h.Name, headingName) {
            t.HeadingID = &h.ID
            found = true
            break
        }
    }
    if !found {
        return task.Task{}, fmt.Errorf("heading %q not found in project", headingName)
    }
}
```

Imports: `"errors"`, `"strings"` (probably present), `"github.com/jtprogru/todushka/internal/storage"`.

2. В `internal/tui/editor_test.go` добавить tests из design §2.8 Unit Tests table:
- `TestEditor_SaveEmptyAreaClearsID`
- `TestEditor_SaveValidAreaSetsID`
- `TestEditor_SaveInvalidAreaErrors`
- `TestEditor_SaveEmptyProjectClearsBothIDs`
- `TestEditor_SaveValidProjectSetsID`
- `TestEditor_SaveAmbiguousProjectErrors` — нужно построить 2 проекта с одним именем. Это требует, чтобы `Repo.ProjectFindByName` возвращал оба. Use `svc.AddProject` twice with same Name.
- `TestEditor_SaveInvalidProjectErrors`
- `TestEditor_SaveHeadingWithoutProjectErrors`
- `TestEditor_SaveValidHeadingSetsID`
- `TestEditor_SaveInvalidHeadingErrors`
- `TestEditor_SaveCaseInsensitiveHeading`
- `TestEditor_SaveProjectChangeAutoClearsHeading`
- `TestEditor_SaveSequentialResolveOrder`

Use `setupModelWithInboxTasks` fixture + `svc.AddArea`, `svc.AddProject`, `svc.AddHeading` для тестовых данных.

3. Запустить `task test-race && task lint`. All 213 baseline + new tests passing.

After all subtasks: Part B complete.

---

## Task: T-7 — Property-based tests batch

*_Requirements: ALL_*
*_Preservation: ALL_*
*_Test_Style: Tier 2_*
*_Complexity: complex_*

GOAL: Реализовать 13 PBT из design §2.8 через `pgregory.net/rapid`, по одному на каждый CP.

Subtasks:

1. В `internal/tui/editor_test.go` добавить PBTs CP-1..CP-3 (label, pre-fill, empty clear):
- `TestProp_WhenLabelInboxOrAnytime` — random task with nullable AreaID/ProjectID.
- `TestProp_PreFillRoundTrip` — tasks with valid IDs; NewEditor + ApplyAndSave preserves.
- `TestProp_EmptyAreaClears` — random task; empty area textinput → AreaID=nil.

2. В `internal/tui/editor_test.go` добавить PBTs CP-4..CP-8 (project/heading paths):
- `TestProp_EmptyProjectClearsBoth`
- `TestProp_InvalidAreaErrors`
- `TestProp_AmbiguousProjectErrors`
- `TestProp_HeadingWithoutProject`
- `TestProp_ValidHeadingResolves`

3. В `internal/tui/editor_test.go` добавить PBTs CP-9..CP-13:
- `TestProp_FieldCountInvariant` — `fieldCount == 9`.
- `TestProp_TabCycleOrder` — iterate `nextField` 9 раз → return to fieldTitle.
- `TestProp_SequentialErrorOrder` — invalid area + invalid project → first error == area.
- `TestProp_HeadingCaseInsensitive` — random case variation matches same heading.
- `TestProp_ProjectChangeClearsOrphanHeading` — original project A + heading, switch to B без heading typed → HeadingID cleared.

4. Запустить `task test-race -count=2 -timeout=120s ./internal/tui/` — стабильность PBT.

After all subtasks: all 13 CPs covered.

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
6. Coverage matrix verification.
7. Manual smoke (если возможно):
   - Открыть task без area/project — label "Inbox".
   - Открыть task с area — label "Anytime".
   - В editor написать area name, save → task переходит в Anytime list.
   - Стереть area, save → task возвращается в Inbox.
   - Опечатка в area name → error в editor's m.err.
8. Если что-то fails — вернуться к T-N.

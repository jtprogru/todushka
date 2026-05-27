# Project Navigation — Task Plan

**Work Type:** Pure feature
**Test Style Source:** Tier 2
- Reference files: `internal/tui/details_redesign_test.go`, `internal/tui/list_render_polish_test.go`, `internal/tui/app_test.go`, `internal/tui/editor_test.go`, `internal/app/service_test.go`
- Framework: `testing` + `github.com/stretchr/testify/require`
- PBT: `pgregory.net/rapid` (look at `internal/tui/testdata/rapid/` for existing property tests)
- Naming: `TestXxx_Scenario` (unit), `TestProp_Xxx` (PBT)
- Termenv discipline: `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup(... termenv.Ascii)` for color-aware tests
- Repository fake: `fakes.NewInMemRepo()` from `internal/storage/fakes`

**Commands:**

| Action     | Command          | Source        |
|------------|------------------|---------------|
| Test       | `task test`      | Taskfile.yml  |
| Test (race)| `task test-race` | Taskfile.yml  |
| Build      | `task build`     | Taskfile.yml  |
| Lint       | `task lint`      | Taskfile.yml  |
| Format     | `task fmt`       | Taskfile.yml  |

---

## Coverage Matrix

| Requirement | Task(s)         | Correctness Property |
|-------------|-----------------|----------------------|
| REQ-1.1     | T-2, T-3        | CP-1                 |
| REQ-1.2     | T-2, T-3        | CP-1                 |
| REQ-1.3     | T-2             | CP-14                |
| REQ-1.4     | T-3             | CP-2                 |
| REQ-2.1     | T-3             | CP-13 (rendering)    |
| REQ-2.2     | T-3             | —                    |
| REQ-2.3     | T-3             | CP-3                 |
| REQ-2.4     | T-3             | CP-3                 |
| REQ-2.5     | T-1             | CP-4                 |
| REQ-2.6     | T-3             | CP-3                 |
| REQ-2.7     | T-3             | CP-5                 |
| REQ-2.8     | T-3             | —                    |
| REQ-3.1     | T-4             | —                    |
| REQ-3.2     | T-4             | —                    |
| REQ-3.3     | T-4             | CP-10                |
| REQ-3.4     | T-4             | CP-11                |
| REQ-3.5     | T-4             | CP-11                |
| REQ-3.6     | T-4             | CP-11                |
| REQ-3.7     | T-4             | —                    |
| REQ-3.8     | T-5             | —                    |
| REQ-3.9     | T-5, T-1        | CP-6                 |
| REQ-3.10    | T-5             | —                    |
| REQ-4.1     | T-6             | CP-12                |
| REQ-4.2     | T-6             | CP-13                |
| REQ-4.3     | T-6             | —                    |
| REQ-4.4     | T-6             | CP-12                |
| REQ-4.5     | T-6             | CP-13 (badge)        |
| REQ-4.6     | T-6             | CP-15                |
| REQ-5.1     | T-4, T-5        | CP-9                 |
| REQ-5.2     | T-4, T-5        | —                    |
| REQ-5.3     | T-3, T-6        | —                    |
| REQ-6.1     | T-1             | CP-8                 |
| REQ-6.2     | T-1             | CP-6                 |
| REQ-6.3     | T-1             | CP-7                 |
| REQ-6.4     | T-1             | —                    |
| REQ-6.5     | T-1             | —                    |
| REQ-7.1     | T-1             | CP-4                 |
| REQ-7.2     | T-1             | CP-3                 |
| REQ-7.3     | T-1             | CP-3                 |

All 15 correctness properties (CP-1..CP-15) are linked to at least one task; property-based tests for all of them are collected in **T-7**.

---

## Tasks

### T-1: Service layer — DeleteProject + ListProjectsSorted
*_Requirements: 2.5, 3.9, 6.1, 6.2, 6.3, 6.4, 6.5, 7.1, 7.2, 7.3_*
*_Complexity: standard_*

GOAL: реализовать новые сервисные методы и их доменные ошибки. Тесты
изолированы от TUI, используют `fakes.NewInMemRepo`.

- **T-1.1 [GREEN]** Добавить unit-тесты в `internal/app/service_test.go` для DeleteProject.
  - File: `internal/app/service_test.go`.
  - Tests: `TestService_DeleteProject_NotFound`, `TestService_DeleteProject_NonEmpty_NoConfirm`, `TestService_DeleteProject_NonEmpty_Confirm`, `TestService_DeleteProject_Empty_Confirm`, `TestService_DeleteProject_Empty_NoConfirm`, `TestService_DeleteProject_ReadOnly`.
  - Use `fakes.NewInMemRepo()` + `app.New(repo, FixedClock{...})`.
  - Assertions:
    - NotFound: `errors.Is(err, app.ErrProjectNotFound)`.
    - NonEmpty_NoConfirm: `errors.Is(err, app.ErrProjectNotEmpty)`; задачи unchanged (compare slice before/after).
    - NonEmpty_Confirm: задачи имеют `ProjectID == nil && HeadingID == nil` после; проект имеет `DeletedAt != nil`.
    - Empty_Confirm: project soft-deleted.
    - ReadOnly: use a read-only test wrapper (use `storage/bbolt` open RO is overkill — добавить helper `fakes.NewInMemRepoReadOnly()` если нет, иначе использовать существующий механизм). Если нет — отметить `ErrReadOnly` через mock.
  - **Test_Style:** `internal/app/service_test.go` (testify/require, table-driven where applicable)
  - CRITICAL: запустить `task test` перед T-1.2 — все 6 тестов должны упасть (метод не существует).

- **T-1.2 [GREEN]** Добавить unit-тесты для ListProjectsSorted в `internal/app/service_test.go`.
  - Tests: `TestService_ListProjectsSorted_Basic`, `TestService_ListProjectsSorted_OnlyOpen`, `TestService_ListProjectsSorted_All`, `TestService_ListProjectsSorted_AreaFilter`, `TestService_CountProjectTasks`.
  - Assertions:
    - Basic: 3 проекта с Position=[10,5,5], Name=["b","A","c"] → ожидаем [Pos5+Name A, Pos5+Name c, Pos10 b] (sort by Position ASC, then Name ASC case-fold).
    - OnlyOpen: 3 проекта со Status=[Open, Completed, Cancelled] → возвращает только 1 Open.
    - All: возвращает все 3 (но Completed+Cancelled тоже).
    - AreaFilter: проекты в разных Area → фильтр по areaID.
    - CountProjectTasks: 3 задачи (2 Open, 1 Completed) → `(open=2, total=3)`.
  - CRITICAL: запустить `task test` — тесты должны упасть.

- **T-1.3 [CODE]** Реализовать `DeleteProject` и errors.
  *_Preservation: CP-6, CP-7, CP-8_*
  - File 1: `internal/app/errors.go` — добавить
    ```go
    var ErrProjectNotEmpty = errors.New("app: project has active tasks; confirm required")
    var ErrProjectNotFound = errors.New("app: project not found")
    ```
  - File 2: `internal/app/service.go` — добавить метод `DeleteProject(ctx, pid id.ID, confirm bool) error`.
    Signature & semantics — exactly from design §2.3 Interface Signatures.
    1. `_, err := s.repo.ProjectGet(ctx, pid)` → if `errors.Is(err, storage.ErrNotFound)` return `ErrProjectNotFound`.
    2. `tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{ProjectID: &pid})` — фильтрует не-deleted by default.
    3. If `!confirm` and `len(tasks) > 0`: return `ErrProjectNotEmpty`.
    4. For each `t` in tasks: `t.ProjectID = nil; t.HeadingID = nil; t.UpdatedAt = s.clock.Now()`; `s.repo.TaskUpdate(ctx, t)`; if err — fail-fast return err.
    5. `s.repo.ProjectDelete(ctx, pid, true)` (soft).
    6. Return nil.
  - DO NOT использовать транзакции (см. ADR-5).

- **T-1.4 [CODE]** Реализовать `ListProjectsSorted` + `countProjectTasks`.
  *_Preservation: CP-3, CP-4_*
  - File: `internal/app/queries.go`.
  - `ListProjectsSorted(ctx, areaID *id.ID, includeAllStatuses bool)`:
    1. `filter := storage.ProjectFilter{AreaID: areaID}`. If `!includeAllStatuses`, `filter.Statuses = []project.Status{project.StatusOpen}`.
    2. `projects, err := s.repo.ProjectList(ctx, filter)`.
    3. `sort.SliceStable(projects, func(i, j int) bool { ... })` — Position ASC, then strings.ToLower(Name) ASC.
    4. Return.
  - `countProjectTasks(ctx, pid id.ID) (open, total int, err error)`:
    1. `tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{ProjectID: &pid})`. Filter excludes soft-deleted by default.
    2. Loop: total++; if `t.Status == task.StatusOpen` → open++.
    3. Return.
  - Метод НЕ экспортируется (lowercase): он используется только из TUI Cmd внутри пакета `app`? Нет — TUI находится в другом пакете. Делаем экспортируемым: `CountProjectTasks`.

- **T-1.5 [VERIFY]** Прогнать `task test` и `task test-race`.
  - GOAL: убедиться, что T-1.1 + T-1.2 теперь зелёные. Все остальные тесты остаются зелёными (preservation).
  - IMPORTANT: если что-то red помимо новых тестов — это регрессия; диагностировать и пофиксить до T-2.

---

### T-2: TUI scaffolding — screenKind, modeProjects, keys, Model fields
*_Requirements: 1.1, 1.2, 1.3_*
*_Complexity: mechanical_*

GOAL: добавить структурные предпосылки в TUI: новые enum-значения,
поля Model, keybindings, shellMode — без реальной логики. После этой
задачи код компилируется, существующие тесты зелёные, но UI новый
функционал ещё не показывает.

- **T-2.1 [CODE]** Добавить `screenProjects`, `screenProjectTasks` в `screenKind`.
  *_Preservation: existing screenKind values; CP-1, CP-14_*
  - File: `internal/tui/msgs.go`.
  - Добавить в существующий `const ( ... )` блок после `screenEditor`:
    ```go
    screenProjects
    screenProjectTasks
    ```
  - Добавить новые msg types (только определения, ещё не emit'тится):
    ```go
    type projectsLoadedMsg struct {
        projects []project.Project
        counts   map[id.ID][2]int
    }
    type projectTasksLoadedMsg struct {
        projectID id.ID
        tasks     []task.Task
    }
    type projectSavedMsg struct {
        project project.Project
        created bool
    }
    type projectDeletedMsg struct {
        projectID id.ID
    }
    type projectStatusFilterMode int
    const (
        psfOpen projectStatusFilterMode = iota
        psfAll
    )
    ```
  - NOTE: `project.Project` уже импортируется в msgs.go.

- **T-2.2 [CODE]** Добавить новые поля в `Model`.
  *_Preservation: existing Model fields untouched_*
  - File: `internal/tui/app.go`.
  - Внутри `type Model struct { ... }` после `headingNamesByID` добавить:
    ```go
    projects            []project.Project
    projectCounts       map[id.ID][2]int
    projectCursor       int
    activeProjectID     *id.ID
    projectStatusFilter projectStatusFilterMode
    projectTasks        []task.Task
    projectEditor       ProjectEditorModel
    ```
  - В `NewModel` инициализировать `projectCounts: make(map[id.ID][2]int)`. Остальные поля Go init на zero-value — это OK.

- **T-2.3 [CODE]** Расширить `KeyMap`.
  *_Preservation: existing keys; CP-14_*
  - File: `internal/tui/keys.go`.
  - Добавить поля в `KeyMap`:
    ```go
    Projects          key.Binding
    ToggleAllStatuses key.Binding
    ```
  - В `DefaultKeyMap()` добавить (после `Filter`):
    ```go
    Projects:          key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "projects")),
    ToggleAllStatuses: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "toggle all statuses")),
    ```
  - CRITICAL: `P` — заглавная буква (Shift+p). НЕ `p` (которая занята PinToday).

- **T-2.4 [CODE]** Добавить `modeProjects` в `shellMode`.
  *_Preservation: existing modes; CP-14_*
  - File: `internal/tui/shell.go`.
  - В `const ( ... )` после `modeReadOnly` добавить `modeProjects`.
  - В `modeLabel()`: добавить case `modeProjects` → `"PROJECTS"`.
  - В `currentMode()`: добавить ветку (между EDITOR и CONFIRM)
    ```go
    case m.screen == screenProjects || m.screen == screenProjectTasks:
        return modeProjects
    ```
    NOTE: после HELP/EDITOR — потому что в editor открытом проекте mode должен быть EDITOR, а если confirm активен — CONFIRM. См. порядок проверки в существующей функции.
  - В `modeKeyHints()`: добавить case `modeProjects` →
    `[]string{"↵: open", "n: new", "e: edit", "d: delete", "/: filter", "a: status", "esc/P: back"}`.

- **T-2.5 [GREEN]** Добавить unit-тесты для shell mode и keys.
  *_Test_Style: internal/tui/app_test.go style_*
  - File: `internal/tui/project_navigation_test.go` (новый, для T-2..T-3 базовых тестов).
  - Tests:
    - `TestShellMode_Projects` — Model с `screen=screenProjects` → `currentMode == modeProjects` && `modeLabel == "PROJECTS"`.
    - `TestShellMode_ProjectTasks` — Model с `screen=screenProjectTasks` → `currentMode == modeProjects` (CP-14 partial).
    - `TestKeyMap_Projects_IsCapitalP` — `DefaultKeyMap().Projects.Keys()` содержит `"P"`, не содержит `"p"`.
    - `TestKeyMap_ToggleAllStatuses_IsA` — содержит `"a"`.

- **T-2.6 [VERIFY]** Прогнать `task test` + `task build`.
  - GOAL: убедиться что код компилируется, новые тесты зелёные,
    существующие тесты остались зелёные.

---

### T-3: Project list rendering + navigation
*_Requirements: 1.1, 1.2, 1.4, 2.1, 2.2, 2.3, 2.4, 2.6, 2.7, 2.8, 5.3_*
*_Complexity: standard_*

GOAL: пользователь нажимает `P` в screenList → видит список проектов;
j/k навигация, `/` фильтр, `a` toggle statuses, `Esc/P` назад.

- **T-3.1 [GREEN]** Тесты рендера и навигации.
  *_Test_Style: internal/tui/list_render_polish_test.go, app_test.go_*
  - File: `internal/tui/project_list_test.go` (новый).
  - Tests:
    - `TestViewProjectList_Empty` — `m.projects = nil` → output contains `"(no projects)"`.
    - `TestViewProjectList_NonEmpty_HasShortID` — 1 project, output contains `id.Short(p.ID)`.
    - `TestViewProjectList_NonEmpty_HasName` — output contains `p.Name`.
    - `TestViewProjectList_NonEmpty_HasCounts` — counts={pid:{2,3}} → output contains `[2/3]`.
    - `TestViewProjectList_NonEmpty_HasAreaName` — when `p.AreaID != nil` and `areaNamesByID[*p.AreaID] = "Work"` → output contains `"Work"`.
    - `TestViewProjectList_NonEmpty_StatusIcon` — Completed → `✓`, Cancelled → `✗`, Open → пустой icon (или `  `).
    - `TestViewProjectList_NonEmpty_Deadline` — `p.Deadline != nil` → output contains formatted date.
    - `TestModel_EnterScreenProjects_P` — Model `screen=screenList`, KeyMsg(`P`) → `screen == screenProjects`.
    - `TestModel_ExitScreenProjects_P` — Model `screen=screenProjects`, KeyMsg(`P`) → `screen == screenList`.
    - `TestModel_ExitScreenProjects_Esc` — Model `screen=screenProjects`, no confirm/filter, KeyMsg(Esc) → `screen == screenList`.
    - `TestModel_GTDKeysBlocked_InProjects` — Model `screen=screenProjects`, sequential KeyMsg("1","2","3","Tab") → `m.activeList` и `m.screen` не изменились.
    - `TestModel_ProjectsCursor_J_K` — projects=[p1,p2,p3], j j → cursor=2; k → 1; k k k (over-clamp) → 0.
    - `TestModel_ProjectsToggleStatusFilter_A` — KeyMsg(`a`) переключает `projectStatusFilter` Open ↔ All; expects fetchProjects Cmd re-emitted.
    - `TestModel_ProjectsFilter_Slash` — KeyMsg(`/`) → `m.filtering == true`; type "fo" → `m.filterQuery == "fo"`; `displayedProjects` filtered by Name (case-fold contains).

- **T-3.2 [CODE]** Реализовать `viewProjectList` + `displayedProjects` + `fetchProjects`.
  *_Preservation: existing render code; CP-2, CP-3, CP-5_*
  - File: `internal/tui/project_list.go` (новый).
  - Functions:
    1. `viewProjectList(m Model, width int) string` — рендер строк per project. Использует `m.theme.{Title,Selected,Dim,StatusInfo,StatusError,Help}`. Формат строки: `marker shortID name [open/total] (area: <name>) status_icon (deadline)`. Пустой список → `m.theme.Dim.Render("\n  (no projects)\n")`.
    2. `displayedProjects(m Model) []project.Project` — apply `m.filterQuery` (case-fold contains по `Name`). Аналогично `displayedTasks`.
    3. `fetchProjects(svc *app.Service, includeAll bool) tea.Cmd` — emit `projectsLoadedMsg`. Внутри Cmd: `s.ListProjectsSorted(ctx, nil, includeAll)`; для каждого `p`: `open, total, _ := s.CountProjectTasks(ctx, p.ID)`; counts[p.ID] = [2]int{open, total}. Errors → `errorMsg`.

- **T-3.3 [CODE]** Добавить ветки `handleKey` для screenProjects.
  *_Preservation: existing screenList behavior; CP-1, CP-2_*
  - File: `internal/tui/app.go` (modify `handleKey`).
  - В `handleKey`, ПОСЛЕ `switch m.screen { case screenQuickEntry: ... case screenEditor: ... }` добавить:
    ```go
    case screenProjects:
        return m.handleProjectsKey(msg)
    case screenProjectTasks:
        return m.handleProjectTasksKey(msg)
    ```
  - В ветке `screenList` блока `switch { ... }`: добавить case
    `key.Matches(msg, m.keys.Projects)` ПОСЛЕ existing Filter/NextView:
    ```go
    case key.Matches(msg, m.keys.Projects) && m.screen == screenList:
        m.screen = screenProjects
        m.projectCursor = 0
        return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)
    ```
  - Создать new method `(m Model) handleProjectsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd)`:
    1. If `m.confirm != nil` → handleConfirmKey.
    2. If `m.filtering` → handleFilterKey (но без вызова pruneSelection — для проектов нет selected).
    3. KeyMsg(Esc) or KeyMsg(`P`): `m.screen = screenList; return m, nil`.
    4. KeyMsg(j/k): cursor up/down, clamp.
    5. KeyMsg(`a`): toggle `m.projectStatusFilter`; return `fetchProjects(...)` Cmd.
    6. KeyMsg(`/`): `m.filtering = true; m.filterQuery = ""`.
    7. KeyMsg(`n`): handled in T-4.
    8. KeyMsg(`e`): handled in T-4.
    9. KeyMsg(`d`): handled in T-5.
    10. KeyMsg(Enter): handled in T-6.
    11. ALL OTHER KEYS (`1..6`, Tab, Shift+Tab, `c/x/d/p`/etc when no selection): no-op (return m, nil).
       CRITICAL: явно НЕ диспатчить через `m.switchList` или bulk-actions.

- **T-3.4 [CODE]** Обработать `projectsLoadedMsg` в Update.
  *_Preservation: existing msg handlers_*
  - File: `internal/tui/app.go` (modify `Update`).
  - Добавить case в основной `switch msg := msg.(type)`:
    ```go
    case projectsLoadedMsg:
        m.projects = msg.projects
        m.projectCounts = msg.counts
        if m.projectCursor >= len(displayedProjects(m)) {
            m.projectCursor = max(0, len(displayedProjects(m))-1)
        }
        return m, nil
    ```

- **T-3.5 [CODE]** Расширить `viewBody` для screenProjects.
  *_Preservation: screenList rendering_*
  - File: `internal/tui/app.go` (modify `viewBody`).
  - В начале функции (перед `if !isDualPane(m)`) добавить:
    ```go
    if m.screen == screenProjects {
        return viewProjectList(m, m.width)
    }
    if m.screen == screenProjectTasks {
        return viewProjectTasks(m, m.width)  // stub for T-6
    }
    ```
    NOTE: `viewProjectTasks` ещё не существует — добавить заглушку `func viewProjectTasks(m Model, w int) string { return "" }` в `project_tasks.go` для компиляции, реализация в T-6.

- **T-3.6 [VERIFY]** Прогнать `task test` + `task build`.

---

### T-4: Project editor modal — create + edit
*_Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 5.1, 5.2_*
*_Complexity: complex_*

GOAL: `n` создаёт пустой `ProjectEditorModel` и открывает его как modal;
`e` открывает с предзаполненными полями текущего проекта; `Ctrl+S`
валидирует и сохраняет; `Esc` отменяет.

- **T-4.1 [GREEN]** Тесты валидации и save-flow.
  *_Test_Style: internal/tui/editor_test.go_*
  - File: `internal/tui/project_editor_test.go` (новый).
  - Tests:
    - `TestProjectEditor_Save_EmptyName` — Name="" → err contains "name required"; editor.err не пуст; нет вызова AddProject (через mock-svc или через repo state).
    - `TestProjectEditor_Save_UnknownArea` — Name="x", Area="nosuch" → err contains "area".
    - `TestProjectEditor_Save_MalformedDeadline` — Deadline="abc" → err contains "deadline".
    - `TestProjectEditor_Save_Create_Valid` — original=nil, Name="New", Area="", Deadline="2026-12-31" → repo получает 1 new project; project имеет указанные поля.
    - `TestProjectEditor_Save_Edit_Valid` — original=existing, Name="Renamed" → repo update вызван; новое имя в repo.
    - `TestProjectEditor_New_OpensEmpty` — KeyMsg(`n`) в screenProjects → m.projectEditor.original==nil, name input пустой; screen unchanged (модалка — overlay, не отдельный screen).
    - `TestProjectEditor_Edit_OpensPrefilled` — KeyMsg(`e`) на курсоре → editor.name.Value() == project.Name; editor.original == &project.
    - `TestProjectEditor_Esc_DismissesWithoutSave` — Esc в editor → editor closed, project unchanged.
    - `TestModel_ReadOnly_N_Blocked` — readOnly=true + KeyMsg(`n`) → no editor open, statusMsg contains "read-only".
    - `TestModel_ReadOnly_E_Blocked` — readOnly=true + KeyMsg(`e`) → no editor open.
  - **Open question:** где модалка? Использовать `m.confirm`-style overlay или отдельный screenKind `screenProjectEditor`? **Decision:** новый ephemeral state `m.editingProject bool` (не отдельный screen, чтобы не плодить enum-значения). Модалка рендерится поверх `viewProjectList`.

- **T-4.2 [CODE]** Реализовать `ProjectEditorModel`.
  *_Preservation: EditorModel for tasks unchanged_*
  - File: `internal/tui/project_editor.go` (новый).
  - Struct fields per design §2.3.
  - Functions:
    - `NewProjectEditor(create bool, p *project.Project, svc *app.Service, ctx context.Context) ProjectEditorModel` — инициализация textinputs. Если `create` (или `p == nil`): все пустые. Если `p != nil`: name=p.Name, notes=p.Notes, area name (lookup via repo.AreaGet if p.AreaID != nil), deadline (format YYYY-MM-DD if p.Deadline != nil), autoClose=p.AutoClose. `original = p` (or nil for create).
    - `(m ProjectEditorModel) UpdateForm(msg tea.Msg) (ProjectEditorModel, tea.Cmd)` — диспатч в focused field. Space на focused autoClose toggle'ит bool (как `whenAnytime/whenSomeday` в task editor).
    - `(m ProjectEditorModel) focusCurrent() tea.Cmd`, `nextField`, `prevField` — по образцу task editor.
    - `(m ProjectEditorModel) ApplyAndSave(ctx, svc *app.Service) (project.Project, bool /*created*/, error)`:
      1. `name := strings.TrimSpace(m.name.Value())`. If empty → return `project.Project{}, false, errors.New("name required")`.
      2. `notes := strings.TrimSpace(m.notes.Value())`.
      3. Area: trim. If empty → AreaID=nil. Else `svc.Repo().AreaFindByNormalized(ctx, areaName)`. Not found → error.
      4. Deadline: trim. If empty → nil. Else `time.ParseInLocation("2006-01-02", ...)` → task.NewDate + ptr. Parse err → error.
      5. If `m.original == nil`: `p, err := svc.AddProject(ctx, app.AddProjectInput{Name: name, Notes: notes, AreaID: areaID, Deadline: deadlinePtr, AutoClose: m.autoClose})`. Return `p, true, err`.
      6. Else (edit): обновить поля original.* (Name/Notes/AreaID/Deadline/AutoClose); `svc.EditProject(ctx, *m.original)`. Return `*m.original, false, err`.
    - `View(theme Theme, width int) string` — по образцу `EditorModel.View`: Title "Edit project" / "New project"; fields with labels; helper text; modal-style frame.

- **T-4.3 [CODE]** Wire create/edit через `n`/`e` в `handleProjectsKey`.
  *_Preservation: CP-9, CP-11_*
  - File: `internal/tui/app.go` (modify `handleProjectsKey` and `Update`).
  - В handleProjectsKey добавить:
    1. KeyMsg(`n`): if `m.readOnly` → blockWriteIfReadOnly + status fade tick. Else: `m.projectEditor = NewProjectEditor(true, nil, m.service, ctx)`; `m.editingProject = true`; return m, projectEditor.focusCurrent().
    2. KeyMsg(`e`): if RO → block. Else: `p := projectAtCursor(m)`. If `p == nil` → no-op. Else `m.projectEditor = NewProjectEditor(false, p, m.service, ctx)`; `m.editingProject = true`.
  - Добавить `Model.editingProject bool` в struct.
  - Добавить новый метод `handleProjectEditorKey(m Model, msg tea.KeyMsg)`:
    1. Esc → `m.editingProject = false; m.projectEditor = ProjectEditorModel{}`.
    2. Ctrl+S → save через goroutine-Cmd: вызов `m.projectEditor.ApplyAndSave(ctx, m.service)`. Если err — `m.projectEditor.err = err.Error()`; модалка остаётся. Если ok — `m.editingProject = false`; trigger `fetchProjects(m.service, ...)`.
    3. Tab/Shift+Tab → nextField/prevField.
    4. По умолчанию: UpdateForm.
  - В `handleKey` (или `Update`?) добавить перед `switch m.screen`: если `m.editingProject` → `return handleProjectEditorKey(m, keymsg)`.

- **T-4.4 [CODE]** Обработать `projectSavedMsg` и render editor overlay.
  *_Preservation: viewBody dispatch_*
  - File 1: `internal/tui/app.go`.
  - Добавить в `Update` switch:
    ```go
    case projectSavedMsg:
        m.editingProject = false
        m.projectEditor = ProjectEditorModel{}
        return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)
    ```
  - В `View()` после `if m.confirm != nil { ... } else { switch ... }`, для случая `m.screen == screenProjects` и `m.editingProject == true`: рендерить
    `lipgloss.JoinVertical(viewProjectList(m, w), m.projectEditor.View(m.theme, m.editorWidth()))`.
    NOTE: либо рендерить overlay в `viewBody` для screenProjects: если editingProject → join список + editor view.

- **T-4.5 [VERIFY]** Прогнать `task test` + `task build`.

---

### T-5: Delete-project confirm flow
*_Requirements: 3.8, 3.9, 3.10, 5.1, 5.2_*
*_Complexity: standard_*

GOAL: `d` на проекте → confirm modal "Delete <name>? (y/n)" → `y`
вызывает `DeleteProject(confirm=true)` и reload; любая другая клавиша
снимает confirm.

- **T-5.1 [GREEN]** Тесты.
  *_Test_Style: internal/tui/bulk_test.go_*
  - File: `internal/tui/project_list_test.go` (extend).
  - Tests:
    - `TestModel_DeleteProject_DKey_OpensConfirm` — KeyMsg(`d`) на курсоре → `m.confirm != nil`; confirm.action — newly defined `bulkActionDeleteProject` или поле `confirm.projectID *id.ID`. **Decision:** добавить поле `confirm.projectID *id.ID`, рендер confirm учитывает оба варианта.
    - `TestModel_DeleteProject_Confirm_Y_Deletes` — confirm с projectID, KeyMsg(`y`) → service.DeleteProject вызван; m.confirm == nil; fetchProjects re-triggered.
    - `TestModel_DeleteProject_Confirm_N_Cancels` — KeyMsg(`n`) → m.confirm == nil; project unchanged.
    - `TestModel_ReadOnly_D_Blocked` — readOnly=true + KeyMsg(`d`) → no confirm, statusMsg = "read-only...".
    - `TestModel_DeleteProject_EmptyList_NoOp` — пустой список + `d` → no-op.

- **T-5.2 [CODE]** Расширить confirmState и handleConfirmKey.
  *_Preservation: existing task-delete confirm; CP-9_*
  - File 1: `internal/tui/bulk.go`.
  - В `type confirmState struct { ... }` добавить поле `projectID *id.ID`. Если `projectID != nil` → confirm работает над проектом, а не над задачами.
  - В `handleConfirmKey` после `if msg.Type == tea.KeyRunes ... 'y'`:
    ```go
    if c.projectID != nil {
        pid := *c.projectID
        return m, func() tea.Msg {
            if err := m.service.DeleteProject(context.Background(), pid, true); err != nil {
                return errorMsg{err}
            }
            return projectDeletedMsg{projectID: pid}
        }
    }
    ```
  - В `bulkAction.label()` добавить case (не нужно — мы не добавляем bulkAction, используем флаг projectID).
  - В `dispatch` — не трогаем (это для tasks).

- **T-5.3 [CODE]** Wire `d` в handleProjectsKey.
  *_Preservation: CP-9_*
  - File: `internal/tui/app.go` (extend `handleProjectsKey`).
  - KeyMsg(`d`): if RO → block. Else: `p := projectAtCursor(m)`. If nil → no-op. Else `m.confirm = &confirmState{projectID: &p.ID, action: <whatever>, ids: nil}`. Используем существующий confirm rendering pattern (виден в View(): "<action.label()> N tasks?"). Для project — кастомный label: добавить helper:
    ```go
    func (c *confirmState) labelFor(name string) string {
        if c.projectID != nil { return "Delete project " + name }
        return ...
    }
    ```
    или встроить inline в `View()`. Конкретно: в `View()` где рендерится `c.action.label() + " N tasks?"` — обработать `c.projectID != nil` отдельно: `"Delete project? (y/n)"` или с именем проекта.

- **T-5.4 [CODE]** Обработать `projectDeletedMsg`.
  *_Preservation: existing handlers_*
  - File: `internal/tui/app.go` (modify `Update`).
  - Добавить case:
    ```go
    case projectDeletedMsg:
        return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)
    ```

- **T-5.5 [VERIFY]** Прогнать `task test` + `task build`.

---

### T-6: Zoom-in — screenProjectTasks
*_Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 5.3_*
*_Complexity: complex_*

GOAL: Enter на проекте → отдельный экран с задачами этого проекта;
все task-actions работают; Esc возвращает в список проектов;
`P/Tab/1..6` игнорируются.

- **T-6.1 [GREEN]** Тесты zoom-flow.
  *_Test_Style: internal/tui/app_test.go, list_render_polish_test.go_*
  - File: `internal/tui/project_tasks_test.go` (новый).
  - Tests:
    - `TestModel_ZoomIntoProject` — screen=screenProjects, projects=[p], Enter → screen=screenProjectTasks; activeProjectID=&p.ID; projectTasksLoadedMsg triggered.
    - `TestModel_ZoomOut_Esc` — screen=screenProjectTasks → Esc → screen=screenProjects; projectCursor preserved.
    - `TestModel_PKey_IgnoredInTasksScreen` — screen=screenProjectTasks + KeyMsg(`P`) → screen unchanged.
    - `TestModel_TabKey_IgnoredInTasksScreen` — Tab → screen и activeList unchanged.
    - `TestModel_GTDKeys_IgnoredInTasksScreen` — `1..6` → no-op.
    - `TestViewProjectTasks_NonEmpty` — 1 task with project's ID → output content includes task title.
    - `TestViewProjectTasks_Empty` — no tasks → output contains `"(no tasks in this project)"`.
    - `TestViewProjectTasks_HeadingBadge` — task with HeadingID и `m.headingNamesByID[*hid] = "Section A"` → output contains `"[Section A]"` (или похожая нотация).

- **T-6.2 [CODE]** Реализовать `viewProjectTasks` + `fetchProjectTasks`.
  *_Preservation: viewList unchanged_*
  - File: `internal/tui/project_tasks.go` (заменить заглушку из T-3.5).
  - Functions:
    1. `fetchProjectTasks(svc *app.Service, pid id.ID) tea.Cmd` — emit `projectTasksLoadedMsg`. Внутри: `svc.Repo().TaskList(ctx, storage.TaskFilter{ProjectID: &pid})`. (NOTE: TaskFilter не имеет фильтра по `Statuses` тут — показываем все статусы внутри проекта, чтобы пользователь видел и done task'и; v2 может toggle'ить.)
    2. `viewProjectTasks(m Model, width int) string` — заголовок (имя текущего проекта), separator, потом список задач. Используем модифицированный inline-рендер строки, переиспользующий `wrapTitleColumn` для wrap, добавляющий inline-бейдж heading: `[<heading name>]` после title колонки, если `t.HeadingID != nil`. Empty list → `m.theme.Dim.Render("\n  (no tasks in this project)\n")`.

- **T-6.3 [CODE]** Wire Enter / Esc / key blocking.
  *_Preservation: CP-12, CP-15_*
  - File: `internal/tui/app.go` (extend `handleProjectsKey` and new `handleProjectTasksKey`).
  - В `handleProjectsKey` обработать KeyMsg(Enter): `p := projectAtCursor(m)`. If nil → no-op. Else `m.screen = screenProjectTasks; m.activeProjectID = &p.ID; m.cursor = 0; return m, fetchProjectTasks(m.service, p.ID)`.
  - Создать `handleProjectTasksKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd)`:
    1. If `m.confirm != nil` → handleConfirmKey.
    2. If `m.filtering` → handleFilterKey.
    3. KeyMsg(Esc): `m.screen = screenProjects; m.activeProjectID = nil; m.projectTasks = nil`. Trigger `fetchProjects` (counts могли поменяться).
    4. KeyMsg(`P`/Tab/Shift+Tab/`1..6`): no-op (CRITICAL).
    5. KeyMsg(j/k): cursor up/down (по `m.projectTasks`).
    6. KeyMsg(`/`): filter mode.
    7. KeyMsg(`c`/`x`/`d`/`p`/space/`*`/`r`/Enter/`n`): reuse существующие task action dispatchers, но cursor работает по `m.projectTasks`, а не `m.tasks`. **Detail:** проще всего временно подменить — например, `if m.screen == screenProjectTasks { m.tasks = m.projectTasks } ...` нет, это сломает Update. **Decision:** добавить helper `func currentTaskList(m Model) []task.Task { if m.screen == screenProjectTasks { return m.projectTasks }; return m.tasks }`. Использовать его в `displayedTasks`, `selectedTask`, `cursorTask`. Это маленькое изменение в filter.go/details.go.
       NOTE: re-cursor refers to projectTasks index in this screen. selected map is per-task ID so reuse OK.

- **T-6.4 [CODE]** Обработать `projectTasksLoadedMsg`.
  - File: `internal/tui/app.go` (Update).
  - Добавить case:
    ```go
    case projectTasksLoadedMsg:
        if m.activeProjectID == nil || msg.projectID != *m.activeProjectID {
            return m, nil  // stale msg, ignore
        }
        m.projectTasks = msg.tasks
        if m.cursor >= len(msg.tasks) {
            m.cursor = max(0, len(msg.tasks)-1)
        }
        return m, nil
    ```

- **T-6.5 [CODE]** Обновить `displayedTasks`, `cursorTask`, `selectedTask` для screenProjectTasks.
  *_Preservation: existing screenList task behavior_*
  - File 1: `internal/tui/filter.go` — `displayedTasks`: source = `currentTaskList(m)` вместо `m.tasks`.
  - File 2: `internal/tui/app.go` — `selectedTask`: source = `currentTaskList(m)`.
  - File 3: `internal/tui/details.go` — `cursorTask`: source = `displayedTasks(m)` (уже OK, transitively через `displayedTasks` change).
  - Helper:
    ```go
    func currentTaskList(m Model) []task.Task {
        if m.screen == screenProjectTasks {
            return m.projectTasks
        }
        return m.tasks
    }
    ```
    в `internal/tui/app.go` или `internal/tui/filter.go`.

- **T-6.6 [VERIFY]** Прогнать `task test` + `task build`.
  - CRITICAL: запустить ВСЕ существующие тесты (включая `app_test.go`, `details_redesign_test.go`, `list_render_polish_test.go`) — никаких регрессий быть не должно.

---

### T-7: Property tests + final gate
*_Requirements: all (15 CPs)_*
*_Complexity: standard_*

GOAL: добавить PBT'ы для всех 15 correctness properties; run full
test suite (incl. race); run lint; run build.

- **T-7.1 [GREEN]** Property tests группа 1 (CP-1..CP-5).
  *_Test_Style: internal/tui/testdata/rapid_*
  - File: `internal/tui/project_navigation_pbt_test.go` (новый).
  - Tests (см. design §2.8 PBT table):
    - `TestProp_ScreenEntryRoundTrip` (CP-1)
    - `TestProp_GTDKeysAbsent` (CP-2)
    - `TestProp_StatusFilterEquiv` (CP-3)
    - `TestProp_SortStable` (CP-4)
    - `TestProp_CursorBounded` (CP-5)

- **T-7.2 [GREEN]** Property tests группа 2 (CP-6..CP-10).
  - File: `internal/tui/project_navigation_pbt_test.go` (extend) или
    `internal/app/project_pbt_test.go` для service-level CPs.
  - Tests:
    - `TestProp_DeleteReassignsTasks` (CP-6) — в `internal/app/...`.
    - `TestProp_SoftDeleteInvisible` (CP-7) — в app.
    - `TestProp_EmptyProjectGuard` (CP-8) — в app.
    - `TestProp_ReadOnlyBlocks` (CP-9) — в tui.
    - `TestProp_EditorSaveRoundTrip` (CP-10) — в tui (требует svc + repo).

- **T-7.3 [GREEN]** Property tests группа 3 (CP-11..CP-15).
  - Tests:
    - `TestProp_EditorInvalidStaysOpen` (CP-11)
    - `TestProp_ZoomRoundTrip` (CP-12)
    - `TestProp_ProjectTasksFilter` (CP-13)
    - `TestProp_ModeLabelProjects` (CP-14)
    - `TestProp_PKeyIgnoredInTasks` (CP-15)

- **T-7.4 [GATE]** Final verification.
  - Run `task test` — ALL must pass.
  - Run `task test-race` — ALL must pass.
  - Run `task build` — succeeds.
  - Run `task lint` — no new warnings/errors.
  - Run `task fmt` — diff после fmt должен быть пустой.
  - IMPORTANT: при любом провале — остановиться, диагностировать, не идти в commit.
  - GOAL: confirm coverage matrix исчерпан, все 30 REQ покрыты, все 15 CP — green PBT.

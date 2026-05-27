# Viewport Scroll — Task Plan

**Work Type:** Bug fix (known reproduction: cursor invisible when list > screen)
**Test Style Source:** Tier 2
- Reference files: `internal/tui/project_navigation_pbt_test.go`, `internal/tui/app_test.go`
- Framework: `testing` + `testify/require` + `pgregory.net/rapid`

**Commands:**

| Action     | Command          | Source       |
|------------|------------------|--------------|
| Test       | `task test`      | Taskfile.yml |
| Test (race)| `task test-race` | Taskfile.yml |
| Build      | `task build`     | Taskfile.yml |
| Lint       | `task lint`      | Taskfile.yml |

## Coverage Matrix

| Requirement | Task(s)       | Correctness Property |
|-------------|---------------|----------------------|
| REQ-1.1     | T-1, T-2, T-3, T-4 | CP-1, CP-3      |
| REQ-1.2     | T-1, T-3      | CP-2                 |
| REQ-1.3     | T-1, T-4      | CP-4                 |

## Tasks

### T-1: ensureCursorVisible helper + tests
*_Requirements: 1.1, 1.2, 1.3_*
*_Complexity: standard_*

- **T-1.1 [RED]** Exploration: написать `TestModel_CursorInvisibleOnOverflow` —
  имитирует баг: 30 задач, height=10 (visibleRows ~6), 'j' 20 раз →
  ожидаем что `m.cursor` лежит внутри slice, отображаемого `viewList`.
  Это упадёт на текущем коде (без scroll-offset).
  - File: `internal/tui/viewport_test.go` (новый).
  - CRITICAL: запустить `task test`, убедиться что тест RED.

- **T-1.2 [GREEN]** Tests для `ensureCursorVisible`.
  - File: `internal/tui/viewport_test.go`.
  - Tests:
    - `TestEnsureCursorVisible_FitsInVisible` — total=5, visible=10 → 0.
    - `TestEnsureCursorVisible_CursorAtStart` — cursor=0 → 0.
    - `TestEnsureCursorVisible_CursorAtEnd` — cursor=19, total=20, visible=10 → 10 (cursor at last row).
    - `TestEnsureCursorVisible_CursorMovesDownIntoBuffer` — initial offset=0, total=20, visible=10, cursor=7 → offset shifts to make scrolloff=3 below.
    - `TestEnsureCursorVisible_CursorMovesUpIntoBuffer` — initial offset=10, cursor=12 → offset=9.
    - `TestEnsureCursorVisible_ClampOnReload` — offset=10, total=5, visible=10 → 0.
    - `TestEnsureCursorVisible_VisibleZero` — visibleCount=0 → 0 (no panic).
    - `TestEnsureCursorVisible_NegativeCursor` — cursor=-1 → treated as 0.
  - CRITICAL: компиляция упадёт — функции ещё нет. Это OK; T-1.3 даст GREEN.

- **T-1.3 [CODE]** Implement `ensureCursorVisible` + `visibleRows` + `scrolloff` const.
  *_Preservation: CP-1, CP-2, CP-3, CP-4_*
  - File: `internal/tui/viewport.go` (новый).
  - `const scrolloff = 3`.
  - `ensureCursorVisible(cursor, offset, visibleCount, scrolloff, totalCount int) int`:
    1. If `visibleCount <= 0` → return 0.
    2. If `totalCount <= visibleCount` → return 0.
    3. If `cursor < 0` → cursor = 0.
    4. `maxOffset := totalCount - visibleCount`.
    5. If `cursor - offset < scrolloff` → `offset = cursor - scrolloff`.
    6. If `cursor - offset > visibleCount - 1 - scrolloff` → `offset = cursor - visibleCount + 1 + scrolloff`.
    7. Clamp `offset` to `[0, maxOffset]`.
    8. Return `offset`.
  - `visibleRows(m Model) int`:
    1. If `m.height <= 0` → return 0.
    2. `headerH := lipgloss.Height(m.viewHeader())`.
    3. `footerH := lipgloss.Height(m.viewFooter())`.
    4. `body := m.height - headerH - footerH - 2`.
    5. If `body < 0` → return 0.
    6. Return `body`.
  - DO NOT include views/handlers yet — keep helper isolated.

- **T-1.4 [GREEN]** PBT для helper invariants.
  - File: `internal/tui/viewport_test.go`.
  - Tests:
    - `TestProp_OffsetBounded` — rapid: total/visible/cursor/offset/scrolloff → result in `[0, max(0, total-visible)]`.
    - `TestProp_CursorAlwaysInside` — when `visible > 0 && total > 0`, ensure cursor within `[result, result+visible)`.
    - `TestProp_ScrollOffBufferRespected` — when `total > visible + 2*scrolloff`, buffer must be at least scrolloff (unless cursor near edge).

- **T-1.5 [VERIFY]** Run `task test` — T-1.2 tests pass, T-1.4 PBTs pass.

---

### T-2: Model fields + handler integration
*_Requirements: 1.1_*
*_Complexity: mechanical_*

- **T-2.1 [CODE]** Add Model fields.
  *_Preservation: existing Model fields_*
  - File: `internal/tui/app.go`.
  - Add to Model struct (after `projectScrollOffset`):
    ```go
    scrollOffset        int
    projectScrollOffset int
    ```
  - Default zero in NewModel — no explicit init needed.

- **T-2.2 [CODE]** Update j/k handlers — 3 sites.
  *_Preservation: CP-1_*
  - File: `internal/tui/app.go`.
  - **Site 1: screenList** (inside main `handleKey`, j/k cases):
    ```go
    case key.Matches(msg, m.keys.Up):
        if m.cursor > 0 {
            m.cursor--
        }
        m.scrollOffset = ensureCursorVisible(m.cursor, m.scrollOffset, visibleRows(m), scrolloff, len(displayedTasks(m)))
        return m, nil
    case key.Matches(msg, m.keys.Down):
        if m.cursor < len(m.tasks)-1 {
            m.cursor++
        }
        m.scrollOffset = ensureCursorVisible(m.cursor, m.scrollOffset, visibleRows(m), scrolloff, len(displayedTasks(m)))
        return m, nil
    ```
  - **Site 2: screenProjects** (in `handleProjectsKey`):
    ```go
    case key.Matches(msg, m.keys.Up):
        if m.projectCursor > 0 { m.projectCursor-- }
        m.projectScrollOffset = ensureCursorVisible(m.projectCursor, m.projectScrollOffset, visibleRows(m), scrolloff, len(displayedProjects(m)))
        return m, nil
    // (and mirror for Down)
    ```
  - **Site 3: screenProjectTasks** (in `handleProjectTasksKey`):
    same as Site 1 (reuses m.cursor + m.scrollOffset).

- **T-2.3 [CODE]** Clamp on reload msgs.
  *_Preservation: CP-4_*
  - File: `internal/tui/app.go`.
  - In `tasksLoadedMsg` handler (after `m.tasks = msg.tasks` and cursor clamp):
    `m.scrollOffset = ensureCursorVisible(m.cursor, m.scrollOffset, visibleRows(m), scrolloff, len(m.tasks))`.
  - In `projectTasksLoadedMsg` handler (after mirror + cursor clamp):
    same call.
  - In `projectsLoadedMsg` handler (after projectCursor clamp):
    `m.projectScrollOffset = ensureCursorVisible(m.projectCursor, m.projectScrollOffset, visibleRows(m), scrolloff, len(displayedProjects(m)))`.

- **T-2.4 [CODE]** Reset on screen transitions / list switches.
  *_Preservation: existing state transitions_*
  - File: `internal/tui/app.go`.
  - In `switchList`: `m.scrollOffset = 0`.
  - In Projects entry (existing case): `m.projectScrollOffset = 0`.
  - In `handleProjectsKey` Esc/P: `m.projectScrollOffset = 0`.
  - In Enter (zoom into project): `m.scrollOffset = 0`.
  - In `handleProjectTasksKey` Esc (zoom-out): `m.scrollOffset = 0`.
  - In `handleProjectsKey` ToggleAllStatuses: `m.projectScrollOffset = 0`.
  - In filter activation (/) in any screen: do NOT reset — filter shrinks the list; rely on clamp via reload.

---

### T-3: Render slicing
*_Requirements: 1.2_*
*_Complexity: standard_*

- **T-3.1 [CODE]** Slice in viewList.
  *_Preservation: existing render, all wrap logic, strikethrough etc._*
  - File: `internal/tui/app.go` (`viewList` func).
  - After computing `disp := displayedTasks(m)`:
    ```go
    vr := visibleRows(m)
    off := m.scrollOffset
    if vr > 0 && len(disp) > vr {
        end := off + vr
        if end > len(disp) {
            end = len(disp)
        }
        if off > len(disp) {
            off = max(0, len(disp)-vr)
        }
        disp = disp[off:end]
    }
    ```
  - Cursor marker logic must use **relative index** within sliced disp:
    `if i + off == m.cursor { marker = ">" }` — or, simpler, adjust `m.cursor` substitute → use absolute cursor index relative to `off`.
    Concrete fix:
    ```go
    for i, t := range disp {
        absIdx := i + off
        if absIdx == m.cursor { marker = ... }
        ...
    }
    ```

- **T-3.2 [CODE]** Slice in viewProjectList.
  *_Preservation: render fields, formatting_*
  - File: `internal/tui/project_list.go`.
  - Same pattern: slice `disp` by `m.projectScrollOffset` + `visibleRows(m)`.
  - Cursor marker uses absolute `i + off` against `m.projectCursor`.

- **T-3.3 [CODE]** Slice in viewProjectTasks.
  *_Preservation: heading badge, render_*
  - File: `internal/tui/project_tasks.go`.
  - Same pattern with `m.scrollOffset` (reuses m.cursor).

- **T-3.4 [VERIFY]** Run `task test` — T-1.1 exploration test now GREEN (cursor visible after 20×j).

---

### T-4: Integration & reload tests
*_Requirements: 1.1, 1.3_*
*_Complexity: standard_*

- **T-4.1 [GREEN]** Integration tests for all three screens.
  - File: `internal/tui/viewport_test.go`.
  - Tests:
    - `TestModel_ViewList_CursorVisibleAfterJ` — 30 tasks, height=20, j x 20 → after each iter, abs(cursor) in [offset, offset+visible).
    - `TestModel_ViewProjectList_CursorVisibleAfterJ` — same for projects.
    - `TestModel_ViewProjectTasks_CursorVisibleAfterJ` — same for zoom.
    - `TestModel_ScrollOffsetResetOnSwitchList` — scroll in Inbox, Tab to Today → offset back to 0.
    - `TestModel_ScrollOffsetClampOnTasksReload` — offset=10, then tasksLoadedMsg with shrunk list → offset clamped.

- **T-4.2 [VERIFY]** Run `task test`, `task test-race`. All green.

---

### T-5: Final gate
*_Requirements: all_*
*_Complexity: mechanical_*

- **T-5.1 [GATE]** Run `task test`, `task test-race`, `task build`, `task lint`, `task fmt`. All green.

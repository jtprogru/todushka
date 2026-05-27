# Implementation Report: Viewport Scroll (BL-7)

## Summary

Bug fix BL-7 — добавлен viewport scroll-offset во все три list view'а
(`viewList`, `viewProjectList`, `viewProjectTasks`). Курсор всегда виден
с минимум 3-строчным буфером (vim-like scrolloff=3). Exploration test
`TestModel_CursorInvisibleOnOverflow` зафиксировал bug и теперь PASS.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`

## Task Execution

- [x] **T-1** ensureCursorVisible helper + tests — GREEN
  - Note: TestProp_CursorAlwaysInside упал на первой версии при
    visible < 2*scrolloff+1 (узкое окно, конфликт top/bottom buffer
    rules). Добавил hard-bounds clamp в `[minOffset, maxOffsetForCursor]`
    после буферных правил → cursor всегда внутри окна.
- [x] **T-2** Model fields + handler integration — GREEN
- [x] **T-3** Render slicing in 3 views — GREEN
  - Note: для `viewProjectList`/`viewProjectTasks` `vr = visibleRows(m) - 2`
    (header + blank line занимают 2 строки). Cursor marker использует
    absIdx = i + off для абсолютной позиции.
- [x] **T-4** Integration tests — GREEN (5 tests)
- [x] **T-5** Final gate — все 5 проверок зелёные

## Final Verification

### task test
```
ok  	github.com/jtprogru/todushka/cmd/todushka	(cached)
ok  	github.com/jtprogru/todushka/internal/app	(cached)
ok  	github.com/jtprogru/todushka/internal/cli	(cached)
ok  	github.com/jtprogru/todushka/internal/storage	(cached)
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	(cached)
ok  	github.com/jtprogru/todushka/internal/storage/fakes	(cached)
ok  	github.com/jtprogru/todushka/internal/tui	1.317s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

### task test-race (subset: tui + app)
```
ok  	github.com/jtprogru/todushka/internal/tui	5.149s
ok  	github.com/jtprogru/todushka/internal/app	1.621s
```

### task build
```
(no output — build succeeded)
```

### task lint
```
0 issues.
```

### task fmt
```
(no output — clean)
```

## Files Changed

### New (2)
- `internal/tui/viewport.go` — `ensureCursorVisible`, `visibleRows`, `scrolloff` const
- `internal/tui/viewport_test.go` — 17 tests (9 unit + 3 PBT + 5 integration)

### Modified (3)
- `internal/tui/app.go` — `Model.scrollOffset` + `Model.projectScrollOffset`; j/k handlers (2 sites: screenList, screenProjectTasks); reload-clamp in `tasksLoadedMsg` + `projectTasksLoadedMsg`; reset offsets in `switchList` / Projects-key / handleProjectsKey Esc-P-Enter-ToggleAllStatuses / handleProjectTasksKey Esc; slice in `viewList` with absIdx cursor marker
- `internal/tui/project_list.go` — slice in `viewProjectList` (vr-2 for header), absIdx cursor marker
- `internal/tui/project_tasks.go` — slice in `viewProjectTasks` (vr-2 for header), absIdx cursor marker

### Also modified (handler)
- `internal/tui/app.go` — `handleProjectsKey` j/k updated to set `m.projectScrollOffset`
- `internal/tui/app.go` — `projectsLoadedMsg` handler updated to clamp `m.projectScrollOffset`

## Notes

- **Narrow-window edge case:** when `visibleCount < 2*scrolloff + 1`, both
  buffer rules can fire simultaneously and conflict (top wants offset
  lower, bottom wants offset higher than cursor). PBT
  `TestProp_CursorAlwaysInside` caught this on first run with
  total=2/visible=1. Fix: compute `minOffset = max(0, cursor - visible + 1)`
  and `maxOffsetForCursor = min(cursor, maxOffset)`, and clamp into that
  interval after buffer adjustments. This guarantees cursor visibility
  regardless of how small `visibleCount` is.
- **`vr - 2` adjustment** in `viewProjectList`/`viewProjectTasks`: both
  render functions add `"Projects"` header + blank line (2 rows) before
  the data rows. `viewList` (screenList) has no such header. The subtraction
  is necessary so that integration tests pass without skip — otherwise
  the available rows for actual content drift below scrolloff and the
  bottom of the list visually overflows separator.
- **Logical-row scroll math:** decision from ADR-1 — wrapped multi-line
  titles count as 1 logical row in scroll math. Acceptable tradeoff for
  v1 (lipgloss `MaxHeight` clamp in `View()` remains as safety net).
- **`m.scrollOffset` shared** between screenList and screenProjectTasks
  because `m.tasks` is mirrored. Reset to 0 on every zoom-in / zoom-out.
- **No deviations** from task plan.

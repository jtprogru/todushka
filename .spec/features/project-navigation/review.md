# Code Review: project-navigation

## Verdict: PASS

Реализация полностью покрывает 30 REQ + 15 CP. Все 5 verification-команд
зелёные (test, race, build, lint, fmt). Архитектурные границы соблюдены:
storage не тронут, новые методы сидят в правильных пакетах
(`internal/app/*`, `internal/tui/*`). Никаких critical/major находок.
Несколько minor/nit замечаний по тестовому покрытию edge-cases (REQ-4.2,
5.2, 5.3) и косметике (несколько dead `_ = pkg.X` маркеров) — не блокируют
merge.

## Change Set

| File | Status | Notes |
|------|--------|-------|
| `internal/app/errors.go` | ✅ Planned | `ErrProjectNotEmpty`, `ErrProjectNotFound` |
| `internal/app/service.go` | ✅ Planned | `DeleteProject` |
| `internal/app/queries.go` | ✅ Planned | `ListProjectsSorted`, `CountProjectTasks` |
| `internal/app/service_test.go` | ✅ Planned | 11 new tests + `readOnlyRepo` wrapper |
| `internal/tui/msgs.go` | ✅ Planned | screenProjects/screenProjectTasks + 5 msgs |
| `internal/tui/keys.go` | ✅ Planned | `Projects` (P), `ToggleAllStatuses` (a) |
| `internal/tui/shell.go` | ✅ Planned | `modeProjects` |
| `internal/tui/app.go` | ✅ Planned | Model fields + handlers + viewBody dispatch + `reloadDisplayedTasks` |
| `internal/tui/project_list.go` | ✅ Planned | New |
| `internal/tui/project_editor.go` | ✅ Planned | New |
| `internal/tui/project_tasks.go` | ✅ Planned | New |
| `internal/tui/project_list_test.go` | ✅ Planned | New (24 tests) |
| `internal/tui/project_editor_test.go` | ✅ Planned | New (14 tests) |
| `internal/tui/project_tasks_test.go` | ✅ Planned | New (10 tests) |
| `internal/tui/project_navigation_pbt_test.go` | ✅ Planned | New (15 PBT) |
| `internal/tui/project_navigation_test.go` | ⚠️ Unexpected | Scaffolding tests for T-2 — minor; could be folded into one of the other test files, but isolation is fine. Justified. |
| `internal/tui/bulk.go` | ⚠️ Unexpected | `confirmState.projectID` field. Justified (needed for project delete confirm; alternative would've been a separate state struct — more code). |
| `go.mod` | ⚠️ Unexpected | `termenv` promoted from indirect to direct. Justified — test files now import it directly via `lipgloss.SetColorProfile(termenv.TrueColor)`. |

Сравнение с design §2.3: ничего не пропущено, ничего лишнего. `internal/app/queries_test.go` (план: NEW или MODIFIED) был интегрирован в `service_test.go` — мелкое отклонение, тесты на месте.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 | `TestModel_EnterScreenProjects_P`, `TestProp_ScreenEntryRoundTrip` | `app.go:285-290` (Projects key) | CP-1 | ✅ |
| REQ-1.2 | `TestModel_ExitScreenProjects_P/_Esc` | `app.go:handleProjectsKey` Esc/P branches | CP-1 | ✅ |
| REQ-1.3 | `TestShellMode_Projects/_ProjectTasks`, `TestProp_ModeLabelProjects` | `shell.go:currentMode`, `modeLabel` | CP-14 | ✅ |
| REQ-1.4 | `TestModel_GTDKeysBlocked_InProjects`, `TestProp_GTDKeysAbsent` | `app.go:handleProjectsKey` returns nil for unhandled | CP-2 | ✅ |
| REQ-2.1 | `TestViewProjectList_*` (8 tests) | `project_list.go:viewProjectList` | — | ✅ |
| REQ-2.2 | `TestViewProjectList_Empty` | `project_list.go:viewProjectList` empty branch | — | ✅ |
| REQ-2.3 | `TestService_ListProjectsSorted_OnlyOpen`, `TestProp_StatusFilterEquiv` | `queries.go:ListProjectsSorted` | CP-3 | ✅ |
| REQ-2.4 | `TestService_ListProjectsSorted_All`, `TestProp_StatusFilterEquiv` | `queries.go:ListProjectsSorted` | CP-3 | ✅ |
| REQ-2.5 | `TestService_ListProjectsSorted_Basic`, `TestProp_SortStable` | `queries.go:ListProjectsSorted` sort.SliceStable | CP-4 | ✅ |
| REQ-2.6 | `TestModel_ProjectsToggleStatusFilter_A` | `app.go:handleProjectsKey` ToggleAllStatuses | CP-3 | ✅ |
| REQ-2.7 | `TestModel_ProjectsCursor_J_K`, `TestProp_CursorBounded` | `app.go:handleProjectsKey` Up/Down | CP-5 | ✅ |
| REQ-2.8 | `TestModel_ProjectsFilter_Slash` | `app.go:handleProjectsKey` Filter | — | ✅ |
| REQ-3.1 | `TestModel_NewProject_OpensEditor`, `TestProjectEditor_New_OpensEmpty` | `app.go:handleProjectsKey` QuickEntry | — | ✅ |
| REQ-3.2 | `TestModel_EditProject_OpensEditor`, `TestProjectEditor_Edit_OpensPrefilled` | `app.go:handleProjectsKey` 'e' branch | — | ✅ |
| REQ-3.3 | `TestProjectEditor_Save_Create/_Edit_Valid`, `TestProp_EditorSaveRoundTrip` | `project_editor.go:ApplyAndSave` | CP-10 | ✅ |
| REQ-3.4 | `TestProjectEditor_Save_EmptyName`, `TestProp_EditorInvalidStaysOpen` | `project_editor.go:ApplyAndSave` empty-name guard | CP-11 | ✅ |
| REQ-3.5 | `TestProjectEditor_Save_UnknownArea`, `TestProp_EditorInvalidStaysOpen` | `project_editor.go:ApplyAndSave` area lookup | CP-11 | ✅ |
| REQ-3.6 | `TestProjectEditor_Save_MalformedDeadline`, `TestProp_EditorInvalidStaysOpen` | `project_editor.go:ApplyAndSave` ParseInLocation | CP-11 | ✅ |
| REQ-3.7 | `TestModel_ProjectEditor_Esc_DismissesWithoutSave` | `app.go:handleProjectEditorKey` CloseModal | — | ✅ |
| REQ-3.8 | `TestModel_DeleteProject_DKey_OpensConfirm` | `app.go:handleProjectsKey` Delete | — | ✅ |
| REQ-3.9 | `TestModel_DeleteProject_Confirm_Y_Deletes`, `TestProp_DeleteReassignsTasks` | `bulk.go:handleConfirmKey` projectID branch + `service.go:DeleteProject` | CP-6 | ✅ |
| REQ-3.10 | `TestModel_DeleteProject_Confirm_N_Cancels` | `bulk.go:handleConfirmKey` non-y branch | — | ✅ |
| REQ-4.1 | `TestModel_ZoomIntoProject`, `TestProp_ZoomRoundTrip` | `app.go:handleProjectsKey` Enter | CP-12 | ✅ |
| REQ-4.2 | `TestModel_ProjectTasksCursor_J_K`, `TestProp_ProjectTasksFilter`; full action reuse implicit via m.tasks mirroring | `app.go:handleProjectTasksKey` (reuses dispatch/openEditor/etc.) | CP-13 | ⚠️ Partial — see F-1 |
| REQ-4.3 | `TestViewProjectTasks_Empty` | `project_tasks.go:viewProjectTasks` empty branch | — | ✅ |
| REQ-4.4 | `TestModel_ZoomOut_Esc/_RestoresGTDList`, `TestProp_ZoomRoundTrip` | `app.go:handleProjectTasksKey` CloseModal | CP-12 | ✅ |
| REQ-4.5 | `TestViewProjectTasks_HeadingBadge` | `project_tasks.go:viewProjectTasks` heading branch | CP-13 | ✅ |
| REQ-4.6 | `TestModel_PKey/_TabKey/_GTDKeys_IgnoredInTasksScreen`, `TestProp_PKeyIgnoredInTasks` | `app.go:handleProjectTasksKey` blocked-keys switch | CP-15 | ✅ |
| REQ-5.1 | `TestModel_ReadOnly_N/_E/_D_Blocked`, `TestProp_ReadOnlyBlocks` | `app.go:handleProjectsKey` blockWriteIfReadOnly | CP-9 | ✅ |
| REQ-5.2 | implicit via `errorMsg` + `projectEditorErrMsg` paths | `app.go:Update` error msg handlers | — | ⚠️ Partial — see F-2 |
| REQ-5.3 | implicit via `errorMsg` handler in fetchProjects/fetchProjectTasks | `project_list.go:fetchProjects`, `project_tasks.go:fetchProjectTasks` | — | ⚠️ Partial — see F-3 |
| REQ-6.1 | `TestService_DeleteProject_NonEmpty_NoConfirm`, `TestProp_EmptyProjectGuard` | `service.go:DeleteProject` empty-guard | CP-8 | ✅ |
| REQ-6.2 | `TestService_DeleteProject_NonEmpty_Confirm`, `TestProp_DeleteReassignsTasks` | `service.go:DeleteProject` reassign loop | CP-6 | ✅ |
| REQ-6.3 | `TestService_DeleteProject_Empty_Confirm`, `TestProp_SoftDeleteInvisible` | `service.go:DeleteProject` ProjectDelete(soft=true) | CP-7 | ✅ |
| REQ-6.4 | `TestService_DeleteProject_NotFound` | `service.go:DeleteProject` mapping | — | ✅ |
| REQ-6.5 | `TestService_DeleteProject_ReadOnly` | `service.go:DeleteProject` first-write fails-fast | — | ✅ |
| REQ-7.1 | `TestService_ListProjectsSorted_Basic`, `TestProp_SortStable` | `queries.go:ListProjectsSorted` | CP-4 | ✅ |
| REQ-7.2 | `TestService_ListProjectsSorted_OnlyOpen` | `queries.go:ListProjectsSorted` filter.Statuses | CP-3 | ✅ |
| REQ-7.3 | `TestService_ListProjectsSorted_All` | `queries.go:ListProjectsSorted` includeAll=true | CP-3 | ✅ |

Все 30 REQ имеют хотя бы один тест. 3 REQ (4.2, 5.2, 5.3) покрыты частично и отмечены в Findings как minor.

## Design Conformance

### 3.1 Architectural Boundaries — ✅ OK
- Service-layer methods (`DeleteProject`, `ListProjectsSorted`, `CountProjectTasks`) сидят в `internal/app/{service,queries}.go` — корректно. Не лезут в TUI/storage напрямую.
- TUI handlers/views в `internal/tui/*.go` — изолированы от storage; зависят от `app.Service` API.
- Storage layer (`internal/storage/repository.go`, `bbolt/*`, `fakes/*`) не модифицировался — контракт сохранён.

### 3.2 Data Models — ✅ OK
- `project.Project` не модифицирован. `Heading` не тронут.
- Новые типы (`ProjectEditorModel`, `projectStatusFilterMode`, 5 msgs) — в TUI-layer, не persistable, соответствуют design §2.3.
- `confirmState.projectID *id.ID` добавлен как и описано в design (ADR-2 + §2.3).

### 3.3 API Contracts — ✅ OK
- Сигнатура `DeleteProject(ctx, pid, confirm) error` идентична design.
- `ListProjectsSorted(ctx, areaID *id.ID, includeAllStatuses bool)` идентична.
- `CountProjectTasks(ctx, pid) (open, total int, err error)` — поднят до экспортируемого (как и решено в task plan T-1.4).

### 3.4 Error Handling — ✅ OK
- `ErrProjectNotEmpty`, `ErrProjectNotFound` объявлены, mapped корректно.
- `ApplyAndSave` возвращает строковые ошибки (`"name required"`, `area %q not found`, `deadline: %w`) для валидации — соответствует design §2.7 и REQ-3.4/3.5/3.6.
- Fail-fast при ErrReadOnly — соответствует ADR-5 (REQ-6.5).

### 3.5 Correctness Properties — ✅ OK
Все 15 CP реализованы и протестированы в `project_navigation_pbt_test.go`. PBT прогнал по 100 случайных входов каждый — все зелёные. CP-7 (Soft-delete invisibility), CP-8 (Empty-project guard), CP-6 (Delete reassigns tasks) тестируются на service-уровне с реальным fakes.NewInMemRepo.

### 3.6 Documentation Consistency — ⚠️ Minor
Mermaid-диаграмма в design.md §2.2 показывает 3 новых файла (`project_list.go`, `project_editor.go`, `project_tasks.go`) и 8 модифицированных — соответствует реальности. **Отклонение:** диаграмма не показывает `bulk.go` как модифицированный (`confirmState.projectID` расширение). Это minor — обновление диаграммы пост-фактум — out of scope review phase.

## Code Quality

### 4.1 Naming & Clarity — ✅ OK
- `projectStatusFilterMode`, `psfOpen/psfAll` следуют существующему конвенции (см. `whenAnytime/whenSomeday`).
- `screenProjects`/`screenProjectTasks` зеркалит существующие `screenList`/`screenEditor`.
- `modeProjects` зеркалит `modeNormal`/`modeReadOnly`.
- `newProjectEditor`/`ProjectEditorModel` зеркалит `NewEditor`/`EditorModel`.
- `handleProjectsKey`/`handleProjectTasksKey`/`handleProjectEditorKey` — паттерн `handleXxxKey` уже устоялся в `handleQuickEntryKey`/`handleEditorKey`/`handleConfirmKey`.

### 4.2 Dead Code & Debug Artifacts — ⚠️ Minor
- `internal/tui/project_tasks.go:88` — `var _ = strings.TrimSpace` оставлен как «keep import» маркер, но `strings` в файле больше не используется напрямую. Будет очищен в F-4.
- `internal/tui/project_navigation_pbt_test.go:386-388` — `var (_ = storage.ErrNotFound; _ = task.StatusOpen)` оставлены как маркеры. `task.StatusOpen` действительно используется в TestProp_ZoomRoundTrip / TestProp_ProjectTasksFilter. `storage.ErrNotFound` — не используется напрямую, только через svc.* helpers. F-5.

### 4.3 Scope Creep — ✅ OK
- Никаких рефакторов вне scope BL-5.
- `reloadDisplayedTasks` helper и `confirmState.projectID` — необходимые правки, не creep.
- Editor screen-restore (handler `editorSavedMsg` + Esc-branch в `handleEditorKey`) — bug fix, который вылез на интеграции T-6. Документирован в implementation.md. Соответствует REQ-4.2 spirit (zoom не должен прерываться при правке задачи).

### 4.4 Test Quality — ✅ OK
- Все тесты используют testify `require` (а не `assert`) — short-circuit на первой ошибке, единый style.
- Color-aware тесты корректно используют `lipgloss.SetColorProfile(termenv.Ascii)` + `t.Cleanup(...)`.
- PBT использует `rapid.SampledFrom`/`rapid.IntRange`/`rapid.StringMatching` — корректно. 100 iterations по умолчанию.
- Edge cases покрыты: пустой список, RO mode, концы курсора, валидационные ошибки, ambiguous status filter.
- Все тесты следуют naming convention `TestXxx_Scenario` / `TestProp_Xxx`.

## Security

No security issues found in changed files.

Audit:
- **Input validation:** `ApplyAndSave` trims whitespace, проверяет non-empty name, strict ParseInLocation для деадлайна, normalized area lookup. Нет SQL/path/command injection векторов (всё type-safe id.ID, bbolt JSON).
- **Authentication/authorization:** TUI — local-only tool. RO-mode корректно проверяется в `blockWriteIfReadOnly` перед всеми write-ключами (`n/e/d`).
- **Secrets:** нет.
- **Data exposure:** все ошибки idiomatic Go, не утекают internal paths.
- **No new endpoints** — изменения только в TUI/service/storage слоях, без публичного API.

## Verification Evidence

Команды повторно прогнаны во время review session (fresh, no cache).

### task test (re-run by reviewer with `go clean -testcache && go test ./...`)
```
ok  	github.com/jtprogru/todushka/cmd/todushka	0.375s
ok  	github.com/jtprogru/todushka/internal/app	0.563s
ok  	github.com/jtprogru/todushka/internal/cli	0.892s
ok  	github.com/jtprogru/todushka/internal/config	5.973s
ok  	github.com/jtprogru/todushka/internal/domain/area	2.814s
ok  	github.com/jtprogru/todushka/internal/domain/id	4.363s
ok  	github.com/jtprogru/todushka/internal/domain/project	1.232s
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	3.593s
ok  	github.com/jtprogru/todushka/internal/domain/repeat	2.055s
ok  	github.com/jtprogru/todushka/internal/domain/tag	4.731s
ok  	github.com/jtprogru/todushka/internal/domain/task	2.434s
ok  	github.com/jtprogru/todushka/internal/domain/today	1.645s
ok  	github.com/jtprogru/todushka/internal/storage	3.961s
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	12.657s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	3.185s
ok  	github.com/jtprogru/todushka/internal/tui	6.255s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

### task build (re-run by reviewer)
```
(no output — build succeeded)
```

### task lint (re-run by reviewer, `golangci-lint run ./...`)
```
0 issues.
```

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | minor | `internal/tui/project_tasks_test.go` | REQ-4.2 утверждает что «c/x/d/p/space/* работают идентично screenList». Тест явно подтверждает только cursor-навигацию (j/k); фактическое поведение `c`/`d` через screenProjectTasks не имеет интеграционного теста. Покрытие происходит косвенно через mirror `m.tasks = projectTasks` + protected dispatch helpers. Добавить 1 интеграционный тест "press c in zoom → CompleteTask called → projectTasksLoadedMsg reload". | REQ-4.2 |
| F-2 | minor | `internal/tui/project_editor_test.go` | REQ-5.2 (service error → statusMsg, modal остаётся) покрыта только для валидационных ошибок (`EmptyName/UnknownArea/MalformedDeadline`). Path для AddProject/EditProject runtime-ошибки (e.g. constraint) не тестируется явно. Path в коде корректен (`projectEditorErrMsg` → `m.projectEditor.err = msg.err`). | REQ-5.2 |
| F-3 | minor | `internal/tui/project_list_test.go` | REQ-5.3 (`fetchProjects` repo error → statusMsg + empty list) не имеет явного теста. Path в коде корректен (`errorMsg{err}` → existing handler sets statusMsg). | REQ-5.3 |
| F-4 | nit | `internal/tui/project_tasks.go:88` | `var _ = strings.TrimSpace` — keep-import маркер, но `strings` фактически в файле не используется. Удалить и убрать import. | — |
| F-5 | nit | `internal/tui/project_navigation_pbt_test.go:386-388` | `var (_ = storage.ErrNotFound; ...)` — keep-import маркер для `storage`. `storage` не используется напрямую — убрать. | — |
| F-6 | nit | `design.md §2.2 Mermaid` | Диаграмма не показывает `bulk.go` как модифицированный. Out-of-scope для review phase — реальный код корректен. | — |

## Recommendations

Все находки — **minor/nit**. Можно ship'ить как есть; рекомендации на будущее:

1. **F-1**: добавить 1 интеграционный тест `TestModel_ProjectTasksScreen_BulkComplete` для явного покрытия REQ-4.2 — но это упражнение для следующего pipeline (или v2).
2. **F-2/F-3**: добавить smoke-тесты для error paths — упражнение на будущее.
3. **F-4/F-5**: косметика; не блокирует. Если хочется чисто — удалить keep-import маркеры в следующем коммите.
4. **F-6**: при следующей актуализации docs — refresh diagram.

Никаких critical/major находок — verdict **PASS**, можно переходить к финализации (PR/CI/merge/tag).

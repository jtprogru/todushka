# area-picker (BL-8) — Task Plan

**Work Type:** Pure feature — новый интерактивный picker. Изменения в
редакторах обслуживают фичу; неизменяемое поведение (резолв Project/Heading,
сохранение не-area полей) защищено preservation-тестами.

**Test Style Source:** Tier 2
- Соседние тесты: `internal/tui/editor_test.go`, `internal/tui/project_editor_test.go`,
  `internal/tui/project_navigation_pbt_test.go`.
- Паттерны: `stretchr/testify` (`require`), PBT через `pgregory.net/rapid`
  (`rapid.Check`, `rapid.SliceOf`, `rapid.StringMatching`); хелперы
  `newTestModelWithService(t)`, `setupModelWithInboxTasks(t, ...)`.
- Naming: `TestXxx` для unit, `TestProp_Xxx` для property-based.

**Commands:**
| Action | Command | Source |
|--------|---------|--------|
| Test | `task test` | Taskfile.yml |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |

## Coverage Matrix

| Requirement | Task(s) | Correctness Property |
|-------------|---------|----------------------|
| REQ-1.1 | T-4, T-6 | CP-14 (propagation) |
| REQ-1.2 | T-2, T-3 | CP-6 (equivalence) |
| REQ-1.3 | T-2, T-3 | CP-6 (equivalence) |
| REQ-2.1 | T-3, T-4 | CP-7 (propagation) |
| REQ-2.2 | T-3, T-5 | CP-7 (propagation) |
| REQ-3.1 | T-2, T-3 | CP-4 (propagation) |
| REQ-3.2 | T-2, T-3 | CP-5 (absence) |
| REQ-3.3 | T-2, T-3 | CP-1 (propagation) |
| REQ-3.4 | T-2, T-3 | CP-2 (propagation) |
| REQ-3.5 | T-2, T-3 | CP-3 (absence) |
| REQ-4.1 | T-2, T-3 | CP-8 (round-trip) |
| REQ-4.2 | T-2, T-3 | CP-8 (round-trip) |
| REQ-4.3 | T-2, T-3 | CP-9 (absence) |
| REQ-4.4 | T-2, T-3 | CP-10 (exclusion) |
| REQ-4.5 | T-2, T-3 | CP-15 (absence) |
| REQ-5.1 | T-2, T-3 | CP-6, CP-11 |
| REQ-5.2 | T-2, T-3 | CP-11 (absence) |
| REQ-6.1 | T-4, T-6 | CP-1, CP-12 |
| REQ-6.2 | T-4, T-6 | CP-2, CP-12 |
| REQ-7.1 | T-1, T-6 | CP-13 (equivalence) |

Каждое требование покрыто ≥1 задачей; каждое CP связано с ≥1 задачей.

---

## T-1 — Preservation tests для неизменяемого поведения редакторов
**Type: GREEN** · *_Requirements: 7.1_* · *_Complexity: standard_*

> GOAL: Зафиксировать поведение, которое НЕ должно сломаться от введения
> picker: резолв Project/Heading в task editor и сохранение не-area полей в
> обоих редакторах.

IMPORTANT: Эти тесты должны проходить (GREEN) на текущем коде ДО любых
изменений. Они же должны проходить после.
IMPORTANT: Следовать стилю из Test Style Source (`editor_test.go`).
DO NOT: трогать production-код в этой задаче.

Subtasks:
- [ ] 1. В `internal/tui/editor_test.go` добавить `TestEditor_ProjectHeadingResolveUnchanged`: создать project + heading через `svc`, задать имена в полях project/heading редактора, `ApplyAndSave`, проверить что `t.ProjectID`/`t.HeadingID` указывают на созданные сущности — `task test`.
- [ ] 2. В `internal/tui/editor_test.go` добавить `TestEditor_NonAreaFieldsSavePreserved`: задать title/notes/start/deadline/tags/when, `ApplyAndSave`, проверить что все эти поля сохранены без изменений — `task test`.
- [ ] 3. В `internal/tui/project_editor_test.go` добавить `TestProjectEditor_NonAreaFieldsSavePreserved`: задать name/notes/deadline/autoClose, `ApplyAndSave`, проверить сохранение — `task test`.
- [ ] 4. Запустить `task test`; все три теста должны быть GREEN на немодифицированном коде.

---

## T-2 — Реализовать компонент `areaPicker`
**Type: CODE** · *_Requirements: 1.2, 1.3, 2.1, 2.2, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2_* · *_Preservation: CP-12, CP-13_* · *_Complexity: complex_*

> GOAL: Создать изолированный, редактор-агностичный компонент picker согласно
> design §2.3/§2.5.

CRITICAL: Один subtask = один файл (`internal/tui/area_picker.go`).
IMPORTANT: Row-индексация: index `0` = no-area; `1..len(areas)` = `areas[i-1]`;
`len(areas)+1` = create row (присутствует только если `!readOnly`).
IMPORTANT: inline-create вызывает `svc.AddArea` синхронно; на
`storage.ErrAlreadyExists` резолвить существующую через
`svc.Repo().AreaFindByNormalized(ctx, area.Normalize(name))`.
DO NOT: вводить новые `screenKind`/msg-типы; не обращаться к конкретным
редакторам — общаться только через `pickerResult`.

Subtasks:
- [ ] 1. В `internal/tui/area_picker.go` объявить `pickerOutcome` (iota: `pickerNone`, `pickerCancel`, `pickerSelected`), `pickerResult{outcome, areaID *id.ID, areaName string}`, и `areaPicker` struct (`areas []area.Area`, `cursor int`, `noAreaLabel string`, `readOnly bool`, `creating bool`, `nameInput textinput.Model`, `err string`) — `task build`.
- [ ] 2. В `internal/tui/area_picker.go` реализовать `newAreaPicker(areas, current *id.ID, noAreaLabel string, readOnly bool) areaPicker`: установить cursor на строку, соответствующую `current` (на no-area row при `current == nil`) — реализует CP-4 — `task build`.
- [ ] 3. В `internal/tui/area_picker.go` реализовать helper `lastRowIndex()` и `rowAt(i)` (тип строки: no-area / area / create) с учётом `readOnly` — реализует CP-6 — `task build`.
- [ ] 4. В `internal/tui/area_picker.go` реализовать `Update` для навигации/выбора/отмены (не в режиме create): Up/Down с клампом в `[0, lastRowIndex]` (CP-5); Enter на area → `pickerSelected{&id, name}` (CP-1); Enter на no-area → `pickerSelected{nil, ""}` (CP-2); Esc → `pickerCancel` (CP-3); Enter на create row → `creating=true`, `pickerNone` (CP-8 шаг 1) — `task build`.
- [ ] 5. В `internal/tui/area_picker.go` дополнить `Update` режимом create: непустое имя + Enter → `svc.AddArea`; success → `pickerSelected{&newID, name}` (CP-8); `ErrAlreadyExists` → резолв существующей → `pickerSelected` (CP-10); пустое (после TrimSpace) → `err="name required"`, `pickerNone`, без `AddArea` (CP-9); прочая ошибка → `err=err.Error()`, `pickerNone` (CP-15); делегировать ввод символов в `nameInput.Update`. При `readOnly` ветка create недостижима (CP-11) — `task build`.
- [ ] 6. В `internal/tui/area_picker.go` реализовать `View(theme Theme, width int) string`: рендер строк (cursor через `theme.Selected`, остальное `theme.Dim`), create row только если `!readOnly`; в режиме create — `nameInput.View()` + подпись; `err` через `theme.StatusError`; обёртка `theme.Modal` — реализует CP-6, CP-7 — `task build`.

После всех subtasks: `task build` и `task lint` — без ошибок компиляции/стиля.

---

## T-3 — Тесты компонента `areaPicker`
**Type: GREEN** · *_Requirements: 1.2, 1.3, 2.1, 2.2, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2_* · *_Complexity: standard_*

> GOAL: Покрыть CP-1..CP-11 и CP-15 unit + property-based тестами.

IMPORTANT: Следовать Test Style Source. PBT через `rapid.Check`; генераторы
имён — `rapid.StringMatching`; списки — `rapid.SliceOf`. Хелпер `svc` —
`newTestModelWithService(t)`.
NOTE: CP-7 и CP-14 имеют фиксированное входное пространство — для CP-7 здесь
пишем targeted unit-тест с тегом `Property/7`.

Subtasks:
- [ ] 1. В новом `internal/tui/area_picker_test.go` unit-тесты: `TestAreaPicker_OpensCursorOnCurrent` (CP-4), `TestAreaPicker_NoAreaLabelContextual` (CP-7), `TestAreaPicker_RowsLayout` (CP-6), `TestAreaPicker_ReadOnlyHidesCreate` (CP-6/CP-11) — `task test`.
- [ ] 2. В `internal/tui/area_picker_test.go` unit-тесты выбора: `TestAreaPicker_EnterOnAreaSelects` (CP-1), `TestAreaPicker_EnterOnNoAreaClears` (CP-2), `TestAreaPicker_EscCancels` (CP-3) — `task test`.
- [ ] 3. В `internal/tui/area_picker_test.go` unit-тесты create: `TestAreaPicker_CreateEmptyNameErrors` (CP-9), `TestAreaPicker_CreateDuplicateSelectsExisting` (CP-10), `TestAreaPicker_CreateServiceErrorKeepsOpen` (CP-15, через read-only repo или форс-ошибку) — `task test`.
- [ ] 4. В `internal/tui/area_picker_test.go` PBT: `TestProp_SelectAreaSetsID` (CP-1), `TestProp_NoAreaClearsID` (CP-2), `TestProp_EscPreservesSelection` (CP-3), `TestProp_CursorOpensOnCurrent` (CP-4) — `task test`.
- [ ] 5. В `internal/tui/area_picker_test.go` PBT: `TestProp_CursorInBounds` (CP-5, случайная последовательность Up/Down), `TestProp_RowLayout` (CP-6), `TestProp_ReadOnlyNeverCreates` (CP-11) — `task test`.
- [ ] 6. В `internal/tui/area_picker_test.go` PBT: `TestProp_CreateNewNameRoundTrip` (CP-8), `TestProp_EmptyNameNoCreate` (CP-9), `TestProp_DuplicateNameNoDup` (CP-10), `TestProp_CreateErrorKeepsOpen` (CP-15) — `task test`.

После всех subtasks: `task test` — все новые тесты GREEN.

---

## T-4 — Интегрировать picker в task editor
**Type: CODE** · *_Requirements: 1.1, 2.1, 6.1, 6.2_* · *_Preservation: CP-12, CP-13_* · *_Complexity: complex_*

> GOAL: Заменить area-textinput в `EditorModel` на `areaID/areaName/picker`,
> открывать picker по Enter, писать выбор в `areaID`, упростить save.

CRITICAL: Один subtask = один файл.
IMPORTANT: После каждого subtask запускать `task test` — preservation-тесты
из T-1 не должны регрессировать.
IMPORTANT: `fieldArea` перестаёт быть textinput — ведёт себя как `fieldWhen`
(фокусируемая display-строка). Лейбл no-area в task editor = `"Inbox"`.
DO NOT: трогать резолв Project/Heading; не рефакторить несвязанный код.

Subtasks:
- [ ] 1. В `internal/tui/editor.go` `EditorModel`: убрать поле `area textinput.Model`, добавить `areaID *id.ID`, `areaName string`, `picker *areaPicker` — `task build`.
- [ ] 2. В `internal/tui/editor.go` `NewEditor`: убрать инициализацию `areaIn`; установить `m.areaID = t.AreaID` и при `t.AreaID != nil` — `m.areaName = AreaGet(...).Name` — `task build`.
- [ ] 3. В `internal/tui/editor.go` убрать `fieldArea` из `focusCurrent` (blur+focus) и из `UpdateForm` switch (как `fieldWhen` — без textinput-диспатча) — `task build`.
- [ ] 4. В `internal/tui/editor.go` `View`: рендер строки Area как display (`m.areaName` или `"Inbox"` при nil); если `m.picker != nil` — в начале `View` вернуть `m.picker.View(theme, width)` — `task build`.
- [ ] 5. В `internal/tui/editor.go` `ApplyAndSave`: заменить блок резолва area (`AreaFindByNormalized`) на `t.AreaID = m.areaID` (CP-12); резолв Project/Heading оставить без изменений (CP-13) — `task test`.
- [ ] 6. В `internal/tui/app.go` `handleEditorKey`: в начало добавить guard `if m.editor.picker != nil` (route Up/Down/Enter/Esc/символы в `picker.Update`; применить `pickerResult`: cancel→`picker=nil`, selected→set `areaID`/`areaName`+`picker=nil`); добавить кейс «`focus==fieldArea` и Enter» → открыть picker через `m.service.ListAreas` с `noAreaLabel="Inbox"`, `readOnly=m.readOnly` (CP-14); при ошибке `ListAreas` — `m.editor.err=...`, picker не открывать — `task test`.

После всех subtasks: `task build` и `task lint` — без ошибок.

---

## T-5 — Интегрировать picker в project editor
**Type: CODE** · *_Requirements: 2.2, 6.1, 6.2_* · *_Preservation: CP-12, CP-13_* · *_Complexity: standard_*

> GOAL: Тот же паттерн для `ProjectEditorModel`/`pefArea`; лейбл no-area =
> `"No area"`.

CRITICAL: Один subtask = один файл.
IMPORTANT: После каждого subtask — `task test` (preservation из T-1).
DO NOT: менять валидацию name или поведение AddProject/EditProject.

Subtasks:
- [ ] 1. В `internal/tui/project_editor.go` `ProjectEditorModel`: убрать `area textinput.Model`, добавить `areaID *id.ID`, `areaName string`, `picker *areaPicker` — `task build`.
- [ ] 2. В `internal/tui/project_editor.go` `newProjectEditor`: убрать `areaIn`; при edit с `p.AreaID != nil` установить `m.areaID = p.AreaID` и `m.areaName = AreaGet(...).Name` — `task build`.
- [ ] 3. В `internal/tui/project_editor.go` убрать `pefArea` из `focusCurrent` и `UpdateForm`; `View`: Area как display (`m.areaName` или `"No area"`); `if m.picker != nil` → вернуть `m.picker.View(...)` — `task build`.
- [ ] 4. В `internal/tui/project_editor.go` `ApplyAndSave`: заменить блок `AreaFindByNormalized` на `areaIDPtr = m.areaID` (CP-12) — `task test`.
- [ ] 5. В `internal/tui/app.go` `handleProjectEditorKey`: добавить guard `if m.projectEditor.picker != nil` (route + применить `pickerResult`) и кейс «`focus==pefArea` и Enter» → открыть picker с `noAreaLabel="No area"`, `readOnly=m.readOnly` (CP-14); ошибка `ListAreas` → `m.projectEditor.err` — `task test`.

После всех subtasks: `task build` и `task lint` — без ошибок.

---

## T-6 — Интеграционные тесты task editor
**Type: GREEN** · *_Requirements: 1.1, 6.1, 6.2, 7.1_* · *_Complexity: standard_*

> GOAL: Покрыть CP-12, CP-13, CP-14 на уровне task editor; обновить
> сломанные сменой структуры тесты.

IMPORTANT: Следовать Test Style Source. Удалить/переписать тесты, опиравшиеся
на area-textinput и ошибку «area not found» (поведение намеренно удалено).

Subtasks:
- [ ] 1. В `internal/tui/editor_test.go` обновить существующие тесты, ссылавшиеся на area-textinput / «area %q not found», под новый контракт (`areaID`) — `task test`.
- [ ] 2. В `internal/tui/editor_test.go` добавить `TestEditor_EnterOnAreaOpensPicker` (CP-14): фокус на `fieldArea`, послать Enter в `handleEditorKey`, проверить `m.editor.picker != nil` — `task test`.
- [ ] 3. В `internal/tui/editor_test.go` добавить `TestEditor_ApplyAndSaveUsesAreaID` (CP-12): задать `m.areaID` на существующую area, `ApplyAndSave`, проверить `t.AreaID == m.areaID`; и `TestEditor_ApplyAndSaveNilAreaID`: `areaID=nil` → `t.AreaID==nil` — `task test`.

После всех subtasks: `task test` — GREEN (включая preservation T-1: CP-13).

---

## T-7 — Интеграционные тесты project editor
**Type: GREEN** · *_Requirements: 2.2, 6.1, 6.2_* · *_Complexity: standard_*

> GOAL: Покрыть CP-12, CP-14 и контекстный лейбл на уровне project editor;
> обновить сломанные тесты.

IMPORTANT: Следовать Test Style Source.

Subtasks:
- [ ] 1. В `internal/tui/project_editor_test.go` обновить тесты, опиравшиеся на area-textinput, под `areaID` — `task test`.
- [ ] 2. В `internal/tui/project_editor_test.go` добавить `TestProjectEditor_EnterOnAreaOpensPicker` (CP-14) и `TestProjectEditor_ApplyAndSaveUsesAreaID` (CP-12) — `task test`.
- [ ] 3. В `internal/tui/project_editor_test.go` добавить `TestProjectEditor_NoAreaLabelIsNoArea`: `View` при `areaID==nil` содержит `"No area"` (REQ-2.2) — `task test`.

После всех subtasks: `task test` — GREEN.

---

## T-8 — Checkpoint: проверка полного покрытия
**Type: GATE** · *_Requirements: ALL_* · *_Complexity: standard_*

CRITICAL: Это ПОСЛЕДНЯЯ задача. Не выполнять, пока T-1..T-7 не завершены.

Instructions:
1. `task test` — 100% тестов GREEN.
2. `task test-race` — без гонок.
3. `task build` — без ошибок.
4. `task lint` — без нарушений.
5. Сверить Coverage Matrix: каждое требование имеет ≥1 проходящий тест.
6. Подтвердить отсутствие orphan-задач (всё трассируется к REQ).
7. При любом сбое — вернуться к соответствующей задаче, не закрывать checkpoint.

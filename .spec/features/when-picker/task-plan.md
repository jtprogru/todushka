# when-picker (BL-11) — Task Plan

**Work Type:** Pure feature (новый `whenPicker` + рефактор task-editor; домен не меняется).

**Test Style Source:** Tier 2
- Референсы: `internal/tui/area_picker_test.go`, `internal/tui/editor_test.go`.
- Паттерны: прямое конструирование пикера/редактора; `Update` с `tea.KeyMsg`; `lipgloss.SetColorProfile(termenv.Ascii)` для View; PBT через `pgregory.net/rapid`; даты относительно `time.Now()`.

**Commands:**
| Action | Command | Source |
|--------|---------|--------|
| Test | `task test` | Taskfile.yml |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |

## Coverage Matrix
| Requirement | Task(s) | Correctness Property |
|-------------|---------|----------------------|
| REQ-1.1 | T-3 | CP-6 |
| REQ-1.2 | T-1 | CP-6 |
| REQ-1.3 | T-1, T-3 | CP-4 |
| REQ-2.1 | T-1 | CP-1 |
| REQ-2.2 | T-1 | CP-1, CP-2 |
| REQ-2.3 | T-1 | CP-1 |
| REQ-2.4 | T-1 | CP-1 |
| REQ-3.1 | T-1 | CP-3 |
| REQ-3.2 | T-1 | CP-6 |
| REQ-4.1 | T-2 | CP-6 |
| REQ-4.2 | T-2 | CP-5 |
| REQ-5.1 | T-3 | CP-6 |
| REQ-6.1 | T-2, T-3 | CP-7 |
| REQ-6.2 | T-3 | CP-4 |
| REQ-7.1 | T-4 | CP-8 |
| REQ-7.2 | T-2 | CP-7 |

---

## T-1 — `whenPicker` + `whenDisplay` компонент `[GREEN→CODE]`
*_Requirements: 1.2, 1.3, 2.1, 2.2, 2.3, 2.4, 3.1, 3.2_*  *_Complexity: standard_*
GOAL: чистый overlay-компонент по образцу `areaPicker`.
1. `internal/tui/when_picker.go` `[NEW]`: типы `whenOutcome` (whenNone/whenCancel/whenSelected), `whenResult{outcome, startDate *task.Date, someday bool}`, struct `whenPicker{cursor, entering, dateInput textinput.Model, now time.Time, err string}`; `newWhenPicker(startDate, someday, now)` с initial cursor (someday→2; startDate==startOfDay(now)→0; startDate!=nil→1; иначе 3).
2. `internal/tui/when_picker.go`: `Update(msg) (whenPicker, whenResult)` — nav up/down (k/j), Esc(list)→whenCancel, Enter→selectCurrent (Today→{today,false}; PickDate→entering=true; Someday→{nil,true}; Anytime→{nil,false}); под-режим: Esc→list, Enter→`time.ParseInLocation("2006-01-02",...)`, err→err+whenNone, ok→{date,false}.
3. `internal/tui/when_picker.go`: `View(theme,width)` (список из 4 строк + под-режим даты, по образцу `areaPicker.View`); `whenDisplay(startDate, someday, now)` → "Someday"/"Anytime"/"Today"/"YYYY-MM-DD".
4. `internal/tui/when_picker_test.go` `[NEW]`: unit `TestWhenPicker_SelectToday/SelectSomeday/SelectAnytime/PickDateValid/PickDateInvalid/EscInDateReturnsToList/EscInListCancels/InitialCursor`, `TestWhenDisplay_Mapping` + PBT `TestProp_WhenChoiceMapping` (P1), `TestProp_WhenExclusivity` (P2), `TestProp_WhenInvalidDateInert` (P3), `TestProp_WhenDisplayMapping` (P5).
5. Запустить `task test` для тестов пикера → GREEN.

## T-2 — Рефактор `EditorModel` `[CODE→GREEN]`
*_Requirements: 4.1, 4.2, 6.1, 7.2_*  *_Preservation: CP-7 (прочие поля сохранены)_*  *_Complexity: complex_*
GOAL: заменить `Start`+тумблер на When-состояние.
1. `internal/tui/editor.go`: в struct убрать `start textinput.Model` и `when shellEditorWhen`; добавить `startDate *task.Date`, `someday bool`, `whenPicker *whenPicker`. Удалить enum `shellEditorWhen`/`whenAnytime`/`whenSomeday` и значение `fieldStart` из `editorField` (порядок: Title,Notes,Deadline,Area,Project,Heading,Tags,When; `fieldCount`=8).
2. `internal/tui/editor.go`: `NewEditor` — убрать `startIn` и блок `when := …`; задать `startDate: t.StartDate`, `someday: t.Someday`.
3. `internal/tui/editor.go`: `focusCurrent` и `UpdateForm` — убрать ветки `fieldStart`.
4. `internal/tui/editor.go`: `ApplyAndSave` — удалить парсинг `m.start`; заменить на `t.StartDate = m.startDate`; заменить `t.Someday = m.when==whenSomeday` на `t.Someday = m.someday`. Deadline-парсинг и прочее не трогать.
5. `internal/tui/editor.go`: `View` — убрать строку `field("Start", …)`; заменить `whenSection` на `field("When", whenDisplay(m.startDate, m.someday, time.Now()), m.focus==fieldWhen)`; в начале `View` рендерить `m.whenPicker.View(...)` когда `m.whenPicker != nil` (рядом с веткой area `picker`).
6. `internal/tui/editor_test.go`: обновить/убрать тесты, ссылающиеся на `fieldStart`, `whenAnytime/whenSomeday`, Space-toggle и поле "Start", чтобы пакет компилировался и проходил; добавить `TestEditor_NoStartField` (View без "Start") и `TestEditor_ApplyAndSave_WhenState` (startDate/someday персистятся, прочие поля сохранены) + `TestProp_WhenSavePreserves` (P7), `TestProp_WhenDisplayMapping` уже в T-1.
7. Запустить `task test` → GREEN.

## T-3 — Интеграция в `handleEditorKey` `[CODE→GREEN]`
*_Requirements: 1.1, 1.3, 5.1, 6.1, 6.2_*  *_Preservation: CP-4, area-picker routing не сломан_*  *_Complexity: standard_*
1. `internal/tui/app.go`: в `handleEditorKey` добавить (перед общим switch, рядом с `m.editor.picker`) ветку `if m.editor.whenPicker != nil`: `np, res := m.editor.whenPicker.Update(msg); m.editor.whenPicker = &np`; на `whenSelected` → `m.editor.startDate = res.startDate; m.editor.someday = res.someday; m.editor.whenPicker = nil`; на `whenCancel` → `m.editor.whenPicker = nil`; `return m, nil`.
2. `internal/tui/app.go`: заменить ветку `m.editor.focus == fieldWhen && msg.Type == tea.KeySpace` на `… msg.Type == tea.KeyEnter`: `wp := newWhenPicker(m.editor.startDate, m.editor.someday, time.Now()); m.editor.whenPicker = &wp; return m, nil`.
3. `internal/tui/when_picker_test.go`: `TestEditor_EnterOnWhenOpensPicker` (Enter на fieldWhen → `whenPicker!=nil`); PBT `TestProp_WhenOpenRoundTrip` (P4: random (startDate,someday) → initial cursor + whenDisplay совпадают), `TestProp_EditorNoStartInertCancel` (P6: нет "Start"; Esc-пути не меняют состояние; строки = 4).
4. Запустить `task test` → GREEN.

## T-4 — read-only + GATE `[VERIFY→GATE]`
*_Requirements: 7.1_*  *_Complexity: standard_*
1. `internal/tui/when_picker_test.go`: `TestProp_ReadOnlyBlocksWhenSave` (P8: read-only редактор → `saveEditor` не пишет; по образцу `readonly_pbt_test.go`).
2. Запустить `task test` (полный, `-count=1`) → всё зелёное.
3. Запустить `task lint` и `task build` → без ошибок.
4. Проверить `task fmt` (нет несформатированных файлов).

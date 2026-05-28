# when-picker (BL-11) — Requirements

**Status:** Draft
**Author:** Claude (spec-driven-dev)
**Date:** 2026-05-28

## Overview
Единый Things-подобный выбор «When» в редакторе задачи: overlay-пикер
со строками Today / Pick date… / Someday / Anytime, заменяющий текущее
free-text поле `Start` и тумблер Anytime/Someday. Пишет в существующие
`StartDate`/`Someday` (доменных изменений нет). Затрагивает
`internal/tui/when_picker.go` (NEW), `internal/tui/editor.go`,
`internal/tui/app.go` (`handleEditorKey`). This Evening и moon — вне
scope (BL-12).

## Glossary
| Term | Definition | Code Artifact |
|------|------------|---------------|
| whenPicker | Overlay выбора When-состояния (Today/дата/Someday/Anytime) | `internal/tui/when_picker.go` (NEW) |
| When-состояние | Одно из Today / конкретная дата / Someday / Anytime; проецируется на `StartDate`+`Someday` | `editor.go` `ApplyAndSave` |
| под-режим даты | Состояние пикера, где вводится `YYYY-MM-DD` (аналог `creating` в `areaPicker`) | `when_picker.go` |

## User Stories
- Как пользователь, я хочу одним контролом задать «когда начать»
  (Today / конкретная дата / Someday / Anytime), чтобы планировать
  задачи как в Things, без разрозненных полей.

## Requirements

### Группа 1 — вызов и состав пикера
**REQ-1.1** WHEN фокус на поле «When» в редакторе и нажат Enter, the system SHALL открыть overlay `whenPicker`.

**REQ-1.2** WHEN `whenPicker` открыт в режиме списка, the system SHALL показать строки в порядке: Today, Pick date…, Someday, Anytime.

**REQ-1.3** WHEN `whenPicker` открывается, the system SHALL установить курсор на строку, соответствующую текущему состоянию задачи (Someday→Someday; `StartDate==today`→Today; `StartDate!=nil` и не today→Pick date…; иначе Anytime).

### Группа 2 — семантика выбора
**REQ-2.1** WHEN в пикере выбрана строка Today, the system SHALL установить `StartDate=сегодня` и `Someday=false`.

**REQ-2.2** WHEN в пикере выбрана строка Someday, the system SHALL установить `Someday=true` и `StartDate=nil`.

**REQ-2.3** WHEN в пикере выбрана строка Anytime, the system SHALL установить `StartDate=nil` и `Someday=false`.

**REQ-2.4** WHEN выбрана Pick date… и введена корректная дата `YYYY-MM-DD`, the system SHALL установить `StartDate=<эта дата>` и `Someday=false`.

### Группа 3 — ввод даты
**REQ-3.1** WHEN в под-режиме даты введена строка, не парсящаяся как `YYYY-MM-DD` (через `time.ParseInLocation`), и нажат Enter, the system SHALL показать inline-ошибку и оставить пикер открытым, не меняя задачу.

**REQ-3.2** WHEN в под-режиме даты нажат Esc, the system SHALL вернуться к списку строк пикера без изменения состояния.

### Группа 4 — замена поля Start и отображение
**REQ-4.1** WHEN редактор задачи отрисован, the system SHALL не показывать отдельное free-text поле `Start` (его роль выполняет When-пикер).

**REQ-4.2** WHEN поле «When» не в фокусе и пикер закрыт, the system SHALL отображать текущее When-состояние текстом: `Today` / `YYYY-MM-DD` / `Someday` / `Anytime`.

### Группа 5 — закрытие
**REQ-5.1** WHEN `whenPicker` открыт в режиме списка и нажат Esc, the system SHALL закрыть пикер без изменения When-состояния.

### Группа 6 — сохранение и round-trip
**REQ-6.1** WHEN задача сохраняется (Ctrl+S) после выбора во When-пикере, the system SHALL персистить `StartDate`/`Someday` согласно выбору через `svc.EditTask`.

**REQ-6.2** WHEN редактор открывается для задачи с `StartDate==today`, the system SHALL изначально отражать состояние Today; для `Someday==true` — Someday (round-trip).

### Группа 7 — read-only и отсутствие регрессий
**REQ-7.1** WHEN редактор в read-only и пользователь пытается сохранить, the system SHALL не записывать изменения (существующее поведение сохраняется).

**REQ-7.2** WHEN задача сохраняется, the system SHALL сохранять прочие поля (Title, Notes, Deadline, Area, Project, Heading, Tags) без регрессий.

## Topological Order
REQ-1.* → REQ-2.*/REQ-3.* → REQ-4.* → REQ-6.*
Reason: пикер должен открываться и навигироваться (1) до того, как
проверяется семантика выбора (2/3); отображение/замена `Start` (4)
зависит от наличия When-состояния; персист/round-trip (6) — поверх.
REQ-5.* и REQ-7.* — сквозные.

## Open Design Questions
| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Где живёт `textinput` для ввода даты — поле в `whenPicker` (как `nameInput` в `areaPicker`) | Влияет на структуру компонента и переиспользование валидации | REQ-2.4, REQ-3.1 |
| Как `editor` хранит выбранное When-состояние между открытием пикера и Ctrl+S (новое поле модели vs производное от `StartDate/Someday`) | Определяет round-trip и `ApplyAndSave` | REQ-2.*, REQ-6.* |
| Курсор при `StartDate` в прошлом (не today) — Today или Pick date… | UX позиционирования | REQ-1.3 |

## Verification Commands
| Action | Command | Source |
|--------|---------|--------|
| Test | `task test` | Taskfile.yml |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |
| Fmt | `task fmt` | Taskfile.yml |

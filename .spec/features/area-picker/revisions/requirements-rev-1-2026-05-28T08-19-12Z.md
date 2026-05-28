# area-picker (BL-8) — Requirements

**Status:** In Review
**Author:** Claude (spec-driven-dev)
**Date:** 2026-05-28

## Overview

Заменяем free-text ввод поля Area на интерактивный picker в обоих редакторах
TUI — task editor (`internal/tui/editor.go`) и project editor
(`internal/tui/project_editor.go`). Picker показывает пункт «нет area»,
список существующих areas (`svc.ListAreas`) и строку создания новой area
(`svc.AddArea`, inline-create). Поле Area перестаёт хранить имя-строку и
резолвить его на save — вместо этого редактор хранит выбранный `*id.ID`.
Сервисный и доменный слои не меняются.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| Area picker | Overlay-список внутри редактора для выбора area, пункта «нет area» или создания новой | `internal/tui/` (новый компонент) |
| Пункт «нет area» | Строка в picker, выбор которой устанавливает `AreaID = nil` | `internal/tui/editor.go`, `internal/tui/project_editor.go` |
| Строка создания | Строка `+ New area…` в picker, запускающая inline-create | `app.AddArea` |
| Inline-create | Создание новой area из picker без выхода в CLI | `app.Service.AddArea` |
| Read-only | Режим запрета записи (`Model.readOnly`) | `internal/tui/app.go` |

## User Stories

- Как пользователь TUI, я хочу выбирать area из списка, чтобы не помнить
  точное имя и не получать ошибку «not found» при опечатке.
- Как пользователь TUI, я хочу создать новую area прямо из редактора, чтобы
  не прерываться на CLI.
- Как пользователь в read-only режиме, я хочу спокойно просматривать area без
  риска что-либо изменить или создать.

## Requirements

### Group 1 — Открытие picker

**REQ-1.1** WHEN фокус находится на поле Area и нажат Enter (в task editor или
project editor), the system SHALL открыть area picker вместо передачи Enter
сфокусированному виджету.

**REQ-1.2** WHEN area picker открывается, the system SHALL загрузить актуальный
список areas через `svc.ListAreas` и отобразить пункты в порядке: «нет area»,
затем все areas, затем (если режим не read-only) строку создания `+ New area…`.

**REQ-1.3** WHEN список areas пуст, the system SHALL всё равно отобразить пункт
«нет area» и (если режим не read-only) строку создания.

### Group 2 — Лейбл пункта «нет area»

**REQ-2.1** WHEN area picker открыт из task editor, the system SHALL подписать
пункт «нет area» как `Inbox`.

**REQ-2.2** WHEN area picker открыт из project editor, the system SHALL
подписать пункт «нет area» как `No area`.

### Group 3 — Навигация и выбор

**REQ-3.1** WHEN area picker открывается, the system SHALL установить курсор на
текущее значение: на area с `AreaID` редактируемой сущности, либо на пункт
«нет area», если `AreaID == nil`.

**REQ-3.2** WHEN в picker нажаты Up или Down, the system SHALL переместить
курсор на один пункт в соответствующем направлении, не выходя за границы
списка.

**REQ-3.3** WHEN курсор на пункте-area и нажат Enter, the system SHALL
установить `AreaID` редактируемой сущности равным ID выбранной area и закрыть
picker, вернувшись в форму редактора.

**REQ-3.4** WHEN курсор на пункте «нет area» и нажат Enter, the system SHALL
установить `AreaID = nil` и закрыть picker.

**REQ-3.5** WHEN в picker нажат Esc, the system SHALL закрыть picker без
изменения текущего `AreaID`.

### Group 4 — Inline-create

**REQ-4.1** WHEN курсор на строке создания и нажат Enter, the system SHALL
перевести picker в режим ввода имени новой area.

**REQ-4.2** WHEN в режиме ввода введено непустое (после `TrimSpace`) имя и
нажат Enter, the system SHALL создать area через `svc.AddArea`, установить
`AreaID` равным ID созданной area и закрыть picker.

**REQ-4.3** WHEN в режиме ввода имя пусто после `TrimSpace` и нажат Enter, the
system SHALL не вызывать `svc.AddArea`, показать сообщение об ошибке и оставить
picker открытым.

**REQ-4.4** WHEN `svc.AddArea` вернул `storage.ErrAlreadyExists` (нормализованное
имя уже занято), the system SHALL не создавать дубликат, выбрать существующую
одноимённую area (установить её `AreaID`) и закрыть picker.

**REQ-4.5** WHEN `svc.AddArea` вернул любую другую ошибку, the system SHALL
показать сообщение об ошибке в picker, оставить его открытым и не менять
текущий `AreaID`.

### Group 5 — Read-only

**REQ-5.1** WHEN редактор находится в read-only режиме, the system SHALL не
отображать строку создания `+ New area…` в picker.

**REQ-5.2** WHEN редактор находится в read-only режиме, the system SHALL не
вызывать `svc.AddArea` ни при каких действиях в picker.

### Group 6 — Контракт сохранения

**REQ-6.1** WHEN форма редактора сохраняется, the system SHALL использовать
`AreaID`, выбранный в picker, и SHALL NOT резолвить имя area через
`AreaFindByNormalized`.

**REQ-6.2** WHEN `AreaID` не выбран (`nil`), the system SHALL сохранить
сущность с `AreaID = nil`.

### Group 7 — Сохранение поведения вне scope

**REQ-7.1** WHEN форма task editor сохраняется, the system SHALL резолвить поля
Project и Heading прежним free-text способом, без изменения их поведения.

## Topological Order

```
REQ-1.* → REQ-3.* → REQ-6.*
Reason: picker должен открываться (Group 1) и грузить список прежде, чем
выбор (Group 3) может быть проверен; выбор формирует AreaID, который Group 6
использует на save.

REQ-4.* depends on REQ-1.* (строка создания появляется только в открытом picker).
REQ-2.* depends on REQ-1.2 (лейбл пункта «нет area» отображается в открытом picker).
REQ-5.* depends on REQ-1.2 / REQ-4.* (read-only влияет на видимость и поведение create).
REQ-7.1 (independent — изолированная проверка неизменности Project/Heading).
```

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Где хранить выбранный `AreaID` в модели редактора и как рендерить поле Area (display-only vs остаточный textinput) | Влияет на структуру `EditorModel`/`ProjectEditorModel` и pre-fill | REQ-3.3, REQ-3.4, REQ-6.1 |
| Picker как отдельный sub-state редактора (флаг + под-модель) vs отдельный `screenKind` | Влияет на маршрутизацию клавиш и композицию `View()` | REQ-1.1, REQ-3.2, REQ-3.5 |
| Общий компонент picker для обоих редакторов vs две интеграции | Влияет на переиспользование и объём кода | REQ-1.2, REQ-2.1, REQ-2.2 |

## Verification Commands

| Action | Command | Source |
|--------|---------|--------|
| Test | `task test` | Taskfile.yml |
| Test (race) | `task test-race` | Taskfile.yml |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |

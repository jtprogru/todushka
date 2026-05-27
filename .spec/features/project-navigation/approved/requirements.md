# Project Navigation — Requirements

**Status:** Draft
**Author:** Claude (assisted)
**Date:** 2026-05-27

## Overview

Добавить в TUI отдельный экран для просмотра, открытия, создания,
редактирования и удаления Проектов. Сейчас TUI работает только с задачами
(шесть GTD-списков); проекты доступны только через CLI и упомянуты как
поле в редакторе задачи. Этот feature вводит новый `screenKind`
(`screenProjects`), вложенный режим зума на задачи проекта
(`screenProjectTasks`), модалку редактирования проекта
(`ProjectEditorModel`) и новый сервисный метод `DeleteProject`
с очисткой `ProjectID` у задач (зеркалит `DeleteArea`).

Затронуты: `internal/tui/*`, `internal/app/service.go`,
`internal/app/queries.go`, `internal/storage/repository.go` (без
изменения контракта). Существующее поведение экранов screenList /
screenEditor / screenQuickEntry / screenHelp не меняется.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| `screenProjects` | Новый screenKind: входной режим, список всех проектов | `internal/tui/msgs.go` |
| `screenProjectTasks` | Подэкран: задачи выбранного проекта (zoom-in) | `internal/tui/msgs.go` |
| `ProjectEditorModel` | Модалка create/edit с полями Name/Notes/Area/Deadline/AutoClose | `internal/tui/project_editor.go` |
| `projectStatusFilter` | Состояние фильтра: Open (default) или All | `internal/tui/app.go` (Model) |
| `DeleteProject` | Сервисный метод: clear `ProjectID` у задач + soft-delete проекта | `internal/app/service.go` |
| `ListProjectsSorted` | Возврат `[]project.Project`, отсортированный по `Position` + `Name` (case-fold) | `internal/app/queries.go` |
| `modeProjects` | Новое значение `shellMode` для footer chip | `internal/tui/shell.go` |

## User Stories

- Как пользователь TUI, я хочу одной клавишей открыть список всех своих
  проектов, чтобы видеть общую картину, не выходя в CLI.
- Как пользователь, я хочу зайти в проект и увидеть только его задачи,
  чтобы сфокусироваться на одном направлении.
- Как пользователь, я хочу создавать, переименовывать и удалять проекты
  без перезапуска и без CLI.
- Как пользователь, я хочу, чтобы удаление проекта не теряло мои задачи
  — они должны вернуться в Inbox.
- Как пользователь в read-only режиме, я хочу спокойно просматривать
  проекты и их задачи, без возможности случайно их изменить.

## Requirements

### Группа 1 — Вход/выход в screenProjects

**REQ-1.1** WHEN пользователь в `screenList` нажимает клавишу `P`,
the system SHALL переключить `Model.screen` в `screenProjects` и
загрузить список проектов через service.

**REQ-1.2** WHEN пользователь в `screenProjects` нажимает `P` или `Esc`,
the system SHALL вернуть `Model.screen` в `screenList` без потери
позиции курсора в основной задачей-списке.

**REQ-1.3** WHEN `Model.screen == screenProjects`, the system SHALL
показать в footer chip метку `PROJECTS` через `shellMode = modeProjects`
с контекстными подсказками клавиш для этого режима.

**REQ-1.4** WHEN `Model.screen == screenProjects` и
`Model.confirm == nil` и `Model.filtering == false`, the system SHALL
блокировать обработку GTD-навигационных клавиш (`1..6`, `Tab`,
`Shift+Tab`, `n` quick-entry, `c/x/d/p` bulk-actions из screenList) —
эти клавиши не вызывают `m.switchList` и не диспатчат bulk-actions.

### Группа 2 — Рендеринг списка проектов

**REQ-2.1** WHEN рендерится `screenProjects` с непустым списком,
the system SHALL вывести по одной строке на проект в формате:
маркер курсора, short-id, имя проекта, имя Area (если есть), Status-иконка,
счётчик `[open/total]` задач, дата deadline (если есть).

**REQ-2.2** WHEN рендерится `screenProjects` и список проектов пуст,
the system SHALL вывести dim-плейсхолдер `(no projects)`.

**REQ-2.3** WHEN рендерится `screenProjects` и активный
`projectStatusFilter == Open`, the system SHALL включить в список
только проекты со `Status == StatusOpen`.

**REQ-2.4** WHEN рендерится `screenProjects` и активный
`projectStatusFilter == All`, the system SHALL включить в список
проекты любого Status, кроме soft-deleted (`DeletedAt != nil`).

**REQ-2.5** WHEN загружается список проектов, the system SHALL
отсортировать его по `Position` ASC, при равенстве — по `Name`
case-fold ASC (зеркалит `ListAreas`).

**REQ-2.6** WHEN пользователь нажимает `a` в `screenProjects`,
the system SHALL переключить `projectStatusFilter` между Open ↔ All
и перезагрузить список без потери позиции курсора в пределах
доступных строк.

**REQ-2.7** WHEN пользователь нажимает `j`/`↓` или `k`/`↑` в
`screenProjects`, the system SHALL переместить курсор соответственно,
ограничивая значениями `[0, len(projects)-1]`.

**REQ-2.8** WHEN пользователь нажимает `/` в `screenProjects`,
the system SHALL войти в filter-mode для проектов: применять
case-fold substring match по `Name` к отображаемому списку.

### Группа 3 — CRUD проектов

**REQ-3.1** WHEN пользователь в `screenProjects` нажимает `n` и
`readOnly == false`, the system SHALL открыть `ProjectEditorModel`
с пустыми полями для создания нового проекта.

**REQ-3.2** WHEN пользователь в `screenProjects` нажимает `e` на
непустом курсоре и `readOnly == false`, the system SHALL открыть
`ProjectEditorModel` с полями, предзаполненными из выбранного
проекта.

**REQ-3.3** WHEN пользователь в `ProjectEditorModel` нажимает
`Ctrl+S` и поле `Name` непусто, the system SHALL вызвать
`AddProject` (создание) или `EditProject` (редактирование) и
вернуться в `screenProjects` с обновлённым списком.

**REQ-3.4** WHEN пользователь в `ProjectEditorModel` нажимает
`Ctrl+S` и поле `Name` пусто (после `TrimSpace`), the system SHALL
показать ошибку валидации `name required` в модалке и НЕ закрыть её.

**REQ-3.5** WHEN пользователь в `ProjectEditorModel` указывает
непустой `Area` и такой Area не существует, the system SHALL
показать ошибку `area "<name>" not found` в модалке и НЕ закрыть её.

**REQ-3.6** WHEN пользователь в `ProjectEditorModel` указывает
`Deadline` в формате не `YYYY-MM-DD`, the system SHALL показать
ошибку парсинга в модалке и НЕ закрыть её.

**REQ-3.7** WHEN пользователь в `ProjectEditorModel` нажимает `Esc`,
the system SHALL закрыть модалку без сохранения, вернуться в
`screenProjects` с тем же списком.

**REQ-3.8** WHEN пользователь в `screenProjects` нажимает `d` на
непустом курсоре и `readOnly == false`, the system SHALL установить
`Model.confirm` для удаления выбранного проекта.

**REQ-3.9** WHEN активен `confirm` для удаления проекта и пользователь
нажимает `y`, the system SHALL вызвать `DeleteProject(ctx, pid,
confirm=true)`, очистить курсор и перезагрузить список проектов.

**REQ-3.10** WHEN активен `confirm` для удаления проекта и пользователь
нажимает любую другую клавишу, the system SHALL снять confirm без
удаления и оставить список без изменений.

### Группа 4 — Zoom-in на задачи проекта

**REQ-4.1** WHEN пользователь в `screenProjects` нажимает `Enter` на
непустом курсоре, the system SHALL переключить
`Model.screen` в `screenProjectTasks`, запомнить выбранный
`Model.activeProjectID = &project.ID` и загрузить задачи через
`TaskList` с фильтром `ProjectID = activeProjectID`.

**REQ-4.2** WHEN рендерится `screenProjectTasks` с непустым списком
задач, the system SHALL переиспользовать `viewList`/`viewDetails` так,
что все существующие bulk/cursor-actions (`c`, `x`, `d`, `p`, space,
`*`) работают идентично screenList. Изменения статуса фиксируются в
backing store и видимы после рефреша.

**REQ-4.3** WHEN рендерится `screenProjectTasks` и список задач
проекта пуст, the system SHALL вывести dim-плейсхолдер
`(no tasks in this project)`.

**REQ-4.4** WHEN пользователь в `screenProjectTasks` нажимает `Esc`,
the system SHALL вернуть `Model.screen` в `screenProjects`,
сохраняя позицию курсора по проекту.

**REQ-4.5** WHEN в `screenProjectTasks` для задачи задан `HeadingID`,
the system SHALL отобразить имя heading inline рядом с заголовком
задачи (бэйджем в скобках). Полноценная группировка по headings
не входит в скоуп v1.

**REQ-4.6** WHEN пользователь в `screenProjectTasks` нажимает `P` или
`Tab`/`1..6`, the system SHALL игнорировать эту клавишу. Возврат в
основные GTD-списки требует сначала `Esc` (→ screenProjects),
затем `Esc` или `P` (→ screenList).

### Группа 5 — Read-only и обработка ошибок

**REQ-5.1** WHEN `readOnly == true` и пользователь нажимает `n`, `e`
или `d` в `screenProjects`, the system SHALL установить
`statusMsg = "read-only mode: writes disabled"` и НЕ открывать
модалку / confirm. Просмотр (Enter, j/k, /, a, Esc) работает.

**REQ-5.2** WHEN сервисный вызов (`AddProject`, `EditProject`,
`DeleteProject`) возвращает ошибку, the system SHALL установить
`statusMsg` с текстом ошибки на `statusFadeDuration` и оставить
видимый экран без изменений (без потери незакрытой модалки).

**REQ-5.3** WHEN загрузка списка проектов (`ListProjectsSorted`) или
загрузка задач (`TaskList` по projectID) возвращает ошибку,
the system SHALL установить `statusMsg` с текстом ошибки и
показать пустой список без падения.

### Группа 6 — Сервисный контракт DeleteProject

**REQ-6.1** WHEN вызван `DeleteProject(ctx, pid, confirm=false)` и
у проекта есть хотя бы одна не-soft-deleted задача (через
`TaskList(ProjectID=pid)`), the system SHALL вернуть
`ErrProjectNotEmpty` и НЕ изменить ни задачи, ни проект.

**REQ-6.2** WHEN вызван `DeleteProject(ctx, pid, confirm=true)`, the
system SHALL для каждой задачи с `ProjectID == pid` обнулить
`ProjectID = nil` и `HeadingID = nil`, обновить `UpdatedAt = Now()` и
сохранить задачу через `TaskUpdate`.

**REQ-6.3** WHEN `DeleteProject(ctx, pid, confirm=true)` после очистки
задач, the system SHALL вызвать `repo.ProjectDelete(ctx, pid,
soft=true)`. Проект с `DeletedAt != nil` не попадает в
`ProjectList` без флага `IncludeDeleted`.

**REQ-6.4** WHEN `DeleteProject(ctx, pid, …)` получает `pid`, для
которого `ProjectGet` возвращает `ErrNotFound`, the system SHALL
вернуть mapped доменную ошибку (`ErrTaskNotFound` или новая
`ErrProjectNotFound`).

**REQ-6.5** WHEN read-only repository вызывает
`DeleteProject(ctx, pid, confirm=true)` и repo возвращает
`ErrReadOnly`, the system SHALL вернуть `ErrReadOnly` без частичной
очистки задач (либо все обновления прошли, либо ни одного — но при
ErrReadOnly первый же `TaskUpdate` упадёт и вернёт ошибку без
дальнейших попыток).

### Группа 7 — ListProjectsSorted

**REQ-7.1** WHEN вызван `ListProjectsSorted(ctx, areaID *id.ID,
includeAllStatuses bool)`, the system SHALL вернуть проекты по фильтру
`AreaID` и сортированные по `Position` ASC, потом `Name` case-fold ASC.

**REQ-7.2** WHEN `includeAllStatuses == false`, the system SHALL вернуть
только проекты со `Status == StatusOpen`. Soft-deleted проекты
исключаются всегда.

**REQ-7.3** WHEN `includeAllStatuses == true`, the system SHALL вернуть
все проекты (Open, Completed, Cancelled), но всё ещё исключая
soft-deleted.

## Topological Order

```
REQ-6.* → REQ-3.8, REQ-3.9   (DeleteProject должен быть до UI удаления)
REQ-7.* → REQ-2.5, REQ-2.6   (Sorted list нужен до UI рендера)
REQ-1.1, REQ-1.2 → REQ-2.*, REQ-3.*, REQ-4.* (вход/выход — фундамент)
REQ-2.* → REQ-3.*, REQ-4.* (рендер списка — основа для всех action'ов)
REQ-4.1 → REQ-4.2, REQ-4.3, REQ-4.4, REQ-4.5, REQ-4.6
REQ-5.* — кросс-секционные, верифицируются параллельно
```

Reason: сервисный контракт (Группа 6) и sort-helper (Группа 7) —
фундамент для UI; навигация (Группа 1) — фундамент для всего
остального TUI-кода.

## Conflict Priority

Нет прямых конфликтов. Косвенное напряжение между REQ-2.4 (All
показывает все статусы) и REQ-7.2 (default Open only): обрабатывается
через явный флаг `includeAllStatuses` в `ListProjectsSorted` — UI
вызывает с правильным флагом в зависимости от `projectStatusFilter`.

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Хранить ли `activeProjectID` в Model или в отдельной "view stack" структуре? | Влияет на расширяемость, если в v2 появятся другие zoom-views (по Area, по Tag) | REQ-4.1, REQ-4.4 |
| `ProjectEditorModel` — отдельный struct или generic-ная переиспользуемая модалка для разных сущностей? | Влияет на дублирование кода в v2 (heading editor, area editor) | REQ-3.1, REQ-3.2 |
| Где счёт `[open/total]` строится: каждый рендер или закэшировано в Model? | Влияет на производительность при большом количестве проектов | REQ-2.1 |
| `DeleteProject` транзакционный или per-task? | Влияет на консистентность при ошибках посередине | REQ-6.2, REQ-6.5 |

## Verification Commands

| Action   | Command                  | Source       |
|----------|--------------------------|--------------|
| Test     | `task test`              | Taskfile.yml |
| Test (race) | `task test-race`      | Taskfile.yml |
| Build    | `task build`             | Taskfile.yml |
| Lint     | `task lint`              | Taskfile.yml |
| Format   | `task fmt`               | Taskfile.yml |

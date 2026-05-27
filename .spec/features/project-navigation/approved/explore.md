# Exploration: Project Navigation (BL-5)

## Intent

Сейчас проекты доступны только косвенно: задача может ссылаться на `ProjectID`,
проект показывается в details pane (имя + status/due/notes благодаря v0.8.0),
редактор позволяет ввести имя проекта в текстовое поле. Но отдельного
**слоя/экрана/режима для самих проектов в TUI нет**: нельзя «посмотреть все
проекты», «открыть проект и увидеть его задачи», «создать/переименовать/удалить
проект, не вылезая в CLI». В CLI это есть (`project add|list|delete` в
`internal/cli/project.go`), в TUI — нет.

BL-5 в бэклоге: «отдельный layer/view/screen для просмотра/выбора/создания/
редактирования проектов. Похоже на Things 3 Areas & Projects sidebar.»
Это **greenfield**-добавление: новый `screenKind`, новые тесты, никаких
ломаемых контрактов в существующих экранах.

## Investigation

### Текущая модель экранов и навигации

`internal/tui/msgs.go:26-33` — `screenKind` enum:

```go
const (
    screenList screenKind = iota
    screenQuickEntry
    screenHelp
    screenEditor
)
```

`internal/tui/app.go:67-83` — Model инициализируется на `screenList`. Поля
для навигации по задачам: `screen`, `activeList`, `tasks`, `cursor`,
`selected`, `filterQuery`, `filtering`.

`internal/tui/msgs.go:57-66` — шесть GTD-списков (`listInbox`..`listLogbook`)
с ключами 1-6 (`internal/tui/keys.go:39-44`). Эти ключи и `Tab`/`Shift+Tab`
циклятся через `allLists` в `app.go:56`. Добавлять седьмой `listKind` для
проектов — некрасиво: он не имеет отношения к GTD-buckets и не должен
сидеть в заголовке `Inbox Today Upcoming ...`.

### Что доступно из storage / app слоёв

`internal/storage/repository.go:67-77` — методы для проектов и headings:

```go
ProjectCreate(ctx, project.Project) error
ProjectGet(ctx, id) (project.Project, error)
ProjectList(ctx, ProjectFilter{AreaID, Statuses, IncludeDeleted})
ProjectUpdate(ctx, project.Project) error
ProjectDelete(ctx, id, soft) error
ProjectFindByName(ctx, name) ([]project.Project, error)
HeadingCreate/Update/Delete/List(ctx, projectID)
```

Сервис `internal/app/service.go`:
- `AddProject(ctx, AddProjectInput{Name, Notes, AreaID, Deadline, AutoClose})`
- `EditProject(ctx, project.Project)` — с авто-закрытием через `maybeAutoCloseProject`
- `AddHeading(ctx, projectID, name)` — есть, для headings создания
- `DeleteArea(ctx, aid, confirm)` — паттерн: при подтверждении переписывает
  `AreaID=nil` у проектов и осиротевших задач.

Нет в сервисе:
- `DeleteProject(ctx, pid, ...)` — есть только сырой `Repo().ProjectDelete`.
- `EditHeading` — есть только Add/Delete на сторону headings.
- `MoveTask` (есть!) — для bulk-move задач между проектами уже готов.

`internal/app/queries.go:152-158`:

```go
ListProjects(ctx, areaID *id.ID) ([]project.Project, error)  // raw, не сортирует
GetProject(ctx, pid) (project.Project, error)
```

Сортировки нет — `ProjectList` в bbolt (`bbolt.go:455`) и в `InMemRepo`
возвращают в нестабильном порядке. `ListAreas` (`queries.go:138-150`)
сортирует по `Position` + `Name`. Для нового экрана нужна аналогичная
sort-helper: по `Position` + `Name`.

### Существующие паттерны рендеринга, которые переиспользуем

- `internal/tui/app.go:638-695` — `viewList` для задач (cursor, marker, short-id,
  title с переносом).
- `internal/tui/details.go:157-241` — `viewDetails` group-based.
- `internal/tui/bulk.go` — `confirmState` модалка с y/n.
- `internal/tui/editor.go` — полный редактор с `textinput`/`textarea` и
  Tab-навигацией. Уже умеет project picker через `ProjectFindByName`.
- `internal/tui/shell.go:21-31` — `shellMode` enum для footer chip.
  Понадобится `modeProjects`.

### Какие тесты придётся не сломать

- `internal/tui/app_test.go` — навигация, фильтр, селект, Tab-цикл.
- `internal/tui/details_redesign_test.go` (v0.8.0) — формат details.
- `internal/tui/list_render_polish_test.go` (v0.7.3) — wrap+strikethrough.
- `internal/tui/readonly_pbt_test.go` — RO-инварианты.

Все они должны остаться зелёными. Новый функционал — за отдельным экраном,
включается только по клавише.

### Build Tooling

- **Orchestrator:** Taskfile (`Taskfile.yml`)
- **Test:** `task test` (`go test ./...`)
- **Test (race):** `task test-race` (`go test -race ./...`)
- **Build:** `task build` (`go build -o bin/todushka ./cmd/todushka`)
- **Lint:** `task lint` (`golangci-lint run`)
- **Format:** `task fmt` (`go fmt ./... && goimports -w .`)
- **Source:** `Taskfile.yml`

## Options Considered

### Option A — Полноценный sidebar (Things 3 hierarchy)

Постоянная боковая колонка слева: Inbox / Today / Upcoming / ... / Areas
(раскрываются) / Projects под каждой Area. Текущий header превращается в
sidebar. Все списки задач — справа.

- Pros: каноничный Things-look, единая навигация через sidebar.
- Cons: переписывает `viewHeader`, ломает зрелую горизонтальную
  навигацию (1-6 keys, Tab/Shift+Tab), требует адаптации `paneWidths`
  (третья колонка), сильно расширяет scope BL-5. Велика вероятность
  регрессий в существующих тестах.
- Complexity: **большая**, 3-4 PR-а.

### Option B — Отдельный экран проектов + zoom-in на задачи проекта

Новый `screenProjects` — отдельный полноэкранный режим (как `screenEditor`),
включается по клавише (`P` или `7`). Внутри:

1. **Project list view** (входной режим): список проектов с курсором.
   Поля строки: marker, short-id, Name, Area name, Status icon, Open/Total
   task counts, Deadline (если есть).
2. **Project tasks view** (Enter на проекте): переиспользует `viewList` с
   задачами, отфильтрованными по `ProjectID`. Назад — `Esc`.
3. **Project edit modal** (`e` на проекте): полностью переиспользует
   паттерн `EditorModel`, но для полей `Name / Notes / Area / Deadline
   / AutoClose`. Новая отдельная форма `ProjectEditorModel`.
4. **Project create** (`n`): та же модалка с пустыми полями.
5. **Project delete** (`d`): через `confirmState`. На delete: clear
   `ProjectID` у всех связанных задач (паттерн `DeleteArea(confirm=true)`).
   Soft-delete (`DeletedAt`), не destructive.
6. **Выход**: `Esc`/`P` возвращает в `screenList`.

- Pros: не ломает существующую горизонтальную навигацию; переиспользует
  `viewList`, `confirmState`, `EditorModel`-паттерн; чёткая
  изолированная задача; сразу удовлетворяет BL-5 acceptance criteria
  (просмотр/выбор/CRUD проектов).
- Cons: не sidebar — это не «всегда видно», нужно нажать клавишу.
  Но BL-5 буквально написан как «отдельный layer/view/screen», поэтому
  совпадает с формулировкой.
- Complexity: **средняя**, один pipeline, ~5-7 файлов.

### Option C — Только project picker внутри editor

Расширить существующий `field("Project", ...)` в editor.go fuzzy-picker
из настоящего списка проектов вместо name-typing.

- Pros: минимум кода.
- Cons: не покрывает «просмотр/создание/редактирование/удаление» из
  BL-5. Слишком узко.

## Constraints & Risks

- **Контракт `nameCacheLoadedMsg`** уже несёт `projects map[id.ID]project.Project`
  (после v0.8.0). Новому экрану кэш не нужен — он сам загрузит проекты.
- **`ProjectList` не сортирует** — нужен новый sort-helper в сервисе
  (`ListProjectsSorted` или встроить сортировку в `ListProjects`).
- **`DeleteProject` отсутствует в сервисе** — есть только `Repo().ProjectDelete`,
  без обработки осиротевших задач. Это **новый сервисный метод**: minor
  bump по нашей cadence-конвенции.
- **Headings management** — расширять headings UI (create/rename/delete
  внутри проекта) — большой scope; реалистично deferred в v2.
- **AutoClose interaction** — `EditProject` уже умеет авто-закрытие,
  оно должно продолжать работать через новый UI без изменений.
- **Read-only mode** — все write-операции должны блокироваться (как в
  bulk.go `dispatch`). Helpers `blockWriteIfReadOnly` уже есть.
- **Filter mode** — `/` для фильтрации списка проектов по имени.
  Переиспользовать существующий `handleFilterKey`.
- **Keybinding конфликт**: ключ `P` свободен. Используем именно его, а не
  `7` — `7` дальше визуально продолжает 1-6 и читается как «седьмая
  GTD-категория», что неправда.
- **`ProjectStatus` фильтрация** — показывать ли completed/cancelled
  проекты? По дефолту — только Open; toggle (`a` — all) показывает все.
  Имитирует поведение Things 3 (Logbook отдельно).

## Recommended Direction

**Option B** — отдельный `screenProjects` с двумя под-режимами:

```
screenList ──[P]──► screenProjects (project list)
                      │
                      ├──[Enter]──► screenProjectTasks (task list filtered)
                      │              │
                      │              └──[Esc]──► back to project list
                      │
                      ├──[n]──► project create modal
                      ├──[e]──► project edit modal
                      ├──[d]──► confirm delete
                      ├──[/]──► filter by name
                      ├──[a]──► toggle Status filter (Open / All)
                      └──[Esc/P]──► back to screenList
```

Это лучший trade-off: даёт полный CRUD без переписывания основной
навигации, естественно ложится в существующую архитектуру Bubble Tea
(чистая `Update`, `View`, optimistic splice), и chunk-able — можно
сделать одним pipeline без накопления риска.

## Scope Boundaries

### Must-have (v1)

1. Новый `screenKind`: `screenProjects` (project list) и `screenProjectTasks`.
2. Service:
   - `ListProjectsSorted(ctx, ...)` или сортировка внутри `ListProjects` (Position+Name).
   - `DeleteProject(ctx, pid, confirm bool) error` — clear `ProjectID` у
     задач, soft-delete проекта.
3. TUI:
   - `viewProjectList` — рендер списка проектов.
   - `viewProjectTasks` — рендер задач, фильтр по `ProjectID`.
   - `ProjectEditorModel` — модалка create/edit с полями Name / Notes /
     Area / Deadline / AutoClose.
   - Keybindings: `P` enter/exit, `n` create, `e` edit, `d` delete,
     `Enter` zoom, `Esc` back, `/` filter, `a` toggle status, j/k.
   - `modeProjects` в shell.go для footer chip.
   - Read-only режим блокирует n/e/d.
4. Тесты: unit + property-tests, по образцу `app_test.go` / `editor_test.go`.

### Deferred (v2)

- **Headings management** — create/rename/delete headings внутри
  проекта. Сейчас headings есть в storage, но создаются только через
  `service.AddHeading` (нет UI). Расширяет screenProjectTasks
  отдельно — отдельный pipeline.
- **Bulk-move задач между проектами** — выделение в screenProjectTasks
  и `m` для move. `MoveTask` в service уже есть, но UX-flow «выбрать
  destination» нетривиален и тащит свой scope.
- **Area picker в ProjectEditor** — picker из существующих areas
  (а не name-typing).
- **Сортировка drag-n-drop по Position** — Position уже в модели,
  но UI для re-order не входит в v1.
- **Bulk-операции над проектами** (massowo complete/cancel/delete) — нишево.
- **Sidebar layout** (Option A) — если когда-нибудь захотим. Архитектурно
  ничего не блокирует.

### Needs spike

- **Поведение AutoClose при удалении последней open task в проекте через
  bulk delete из screenProjectTasks**: текущий `maybeAutoCloseProject`
  вызывается из `EditProject`, не из `DeleteTask`. Проверить, что
  пользовательский опыт не сюрпризный — если все задачи удалены, проект
  пустой, но не закрывается. Это **существующее** поведение, не
  регрессия от BL-5. В v1 не трогаем.

## Assumptions & Open Questions

- [ASSUMPTION: вход в screenProjects — клавиша `P` (uppercase Shift+p),
  чтобы не конфликтовать с `p` (Pin to Today) в screenList.]
- [ASSUMPTION: при удалении проекта осиротевшие задачи отправляются в
  Inbox (clear ProjectID), не удаляются. Зеркалит DeleteArea(confirm=true).]
- [ASSUMPTION: по умолчанию screenProjects показывает только Status=Open
  проекты; `a` toggle — все статусы.]
- [ASSUMPTION: project counts (Open/Total) считаем в кэше при загрузке
  списка, не на каждый ререндер. Использовать `TaskList` с `ProjectID`
  фильтром в одном проходе на всех проектах суммарно.]
- [ASSUMPTION: project edit модалка не имеет поля Status — closing
  проекта идёт через AutoClose + правка задач, либо future-pipeline.
  В v1 — только Name/Notes/Area/Deadline/AutoClose.]
- [ASSUMPTION: новый `DeleteProject(ctx, pid, confirm)` — `confirm=false`
  и есть привязанные задачи → возвращает `ErrProjectNotEmpty`;
  `confirm=true` → reassign задачи в Inbox + soft-delete. Зеркалит
  `DeleteArea`.]

**Open questions:**

1. Нужны ли headings в screenProjectTasks как визуальные группы (для
   отображения, не редактирования)? Сейчас задачи показываются плоско.
   Headings уже хранятся, но не отображаются. → Предлагаю: **нет в v1**,
   плоский список; headings рендерим как inline-badge `[heading]` рядом
   с задачей, если `HeadingID != nil`.
2. В screenProjectTasks применять ли все существующие bulk-actions
   (c/x/d/p)? → Предлагаю: **да**, переиспользуем `dispatch`. Это даёт
   мгновенно реальную пользу.
3. После delete проекта куда возвращается курсор? → Предлагаю: тот же
   `m.cursor` clamped, как в `deleteSelected` для задач.

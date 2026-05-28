# Exploration: area-picker (BL-8)

## Intent

Сейчас поле Area в обоих редакторах (task editor и project editor) — это
free-text `textinput`. Пользователь обязан помнить точное имя area; при
сохранении имя резолвится через `Repo.AreaFindByNormalized`, и если area не
найдена — save **падает с ошибкой** `area %q not found`. Создать новую area
можно только из CLI (`todushka area add`). Это плохой UX: невозможно увидеть
список доступных areas, легко ошибиться в имени, нельзя завести новую area не
выходя из TUI.

Цель — заменить free-text ввод на интерактивный **picker**: список
существующих areas + пункт «Inbox» (без area) + возможность создать новую area
прямо из пикера (inline-create). Триггер — backlog item BL-8, deferred-пункт из
pipeline `project-navigation` (v0.9.0).

## Investigation

Прочитаны: `internal/tui/editor.go`, `internal/tui/project_editor.go`,
`internal/tui/app.go` (editor dispatch), `internal/tui/project_list.go`
(референс list-рендера), `internal/app/queries.go`, `internal/app/service.go`,
`internal/domain/area/area.go`, `internal/storage/bbolt/bbolt.go` (Area CRUD),
а также артефакты прошлой фичи `editor-when-and-picker`.

### Сервисный слой — готов, новых методов не нужно
- `svc.ListAreas(ctx) ([]area.Area, error)` — `queries.go:138`. Возвращает
  non-deleted areas, отсортированные по `Position` ASC, затем `Name`
  case-fold ASC.
- `svc.AddArea(ctx, name) (area.Area, error)` — `service.go:279`. Валидирует
  непустое имя (`area.Validate` → `ErrEmptyName`); `Repo.AreaCreate`
  возвращает `storage.ErrAlreadyExists`, если нормализованное имя уже занято
  (`bbolt.go:577`).
- `area.Normalize(name)` = `strings.ToLower(strings.TrimSpace(name))` —
  единая форма для уникальности (`area.go:23`).
- `Repo.AreaGet(ctx, id)` — для отображения имени по `AreaID` (уже
  используется в обоих редакторах при pre-fill).

### Текущее поведение редакторов (что меняем)
- **Task editor** (`editor.go`): `fieldArea` — `textinput` с placeholder
  `"area name (empty = Inbox)"`. Pre-fill: если `t.AreaID != nil`, пишет
  `AreaGet(...).Name` в textinput (`editor.go:97-104`). Резолв на save:
  `ApplyAndSave` → `AreaFindByNormalized`, ошибка «not found»
  (`editor.go:254-267`).
- **Project editor** (`project_editor.go`): `pefArea` — идентичный textinput,
  placeholder `"area name (empty = none)"`. Тот же резолв на save
  (`project_editor.go:186-198`).
- Поток данных: сейчас в модели редактора хранится **имя** (string), а
  `AreaID` вычисляется только в момент сохранения. С пикером естественнее
  хранить выбранный **`*id.ID`** (+ имя для отображения) сразу в момент
  выбора.

### Существующие UI-паттерны для переиспользования
- **Нет** готового list/picker-компонента и **нет** настоящих overlay поверх
  модалок. Композиция экранов — через `screenKind` + sub-state флаги.
- `When` (task editor) и `Auto-close` (project editor) — это **не** textinput,
  а спец-секции: рендерятся вручную в `View()` и обрабатываются спец-кейсом в
  key-handler (`app.go:425`: при `focus == fieldWhen` и `Space` → toggle).
  Это рабочий образец того, как поле-не-textinput живёт внутри формы.
- `confirm`-модалки (`m.confirm != nil`) — образец sub-state внутри экрана:
  пока активна, перехватывает все клавиши (`handleConfirmKey`).
- `project_list.go` — образец рендера выбираемого списка с курсором
  (`viewProjectList`, `displayedProjects`, cursor up/down) — ближайший
  референс для отрисовки списка areas.
- `handleEditorKey` (`app.go:408`) и `handleProjectEditorKey` (`app.go:653`) —
  две точки, куда встроится перехват «picker активен → роутим клавиши в
  picker».

## Build Tooling
- **Orchestrator:** [Task](https://taskfile.dev) — `Taskfile.yml`.
- **Test:** `task test` (`go test ./...`); race — `task test-race`.
- **Build:** `task build` (`go build -o bin/todushka ./cmd/todushka`).
- **Lint:** `task lint` (`golangci-lint run`).
- **Format:** `task fmt` (`go fmt ./...` + `goimports -w .`).
- **Source:** `Taskfile.yml`.
- **Тест-стек:** `stretchr/testify` + property-based `pgregory.net/rapid`
  (используется в `*_pbt_test.go`).

## Options Considered

### Option A — Overlay-picker (sub-state внутри редактора) ⭐
Новый компонент `areaPickerModel`: список = `[Inbox] + areas + [+ New area…]`,
курсор вверх/вниз, Enter — выбрать, Esc — отмена (вернуться в форму). Внутри
редактора флаг `areaPicker *areaPickerModel`; пока не nil — `View()` рисует
пикер, а `handleEditorKey`/`handleProjectEditorKey` роутят клавиши в пикер
(по образцу `m.confirm != nil`). Inline-create — отдельная строка `+ New
area…`, которая открывает мини-textinput для имени → `svc.AddArea` → при
`ErrAlreadyExists` показываем ошибку и выбираем существующую.
- **Pros:** чистый UX (видно все areas), естественно ложится inline-create,
  переиспользует list-рендер `project_list.go` и паттерн confirm-модалки; один
  общий компонент на оба редактора; хранит `*id.ID` сразу.
- **Cons:** новый компонент + изменение модели редактора (хранить `areaID`
  вместо/в дополнение к textinput); нужно аккуратно прокинуть загрузку
  `ListAreas`.
- **Сложность:** средняя.

### Option B — Inline-cycle (как When/Auto-close)
Поле Area циклически переключается стрелками/Space по списку areas прямо в
форме, без отдельного списка.
- **Pros:** минимум нового кода, копия существующего When-паттерна.
- **Cons:** плохо масштабируется (>3–4 areas — мучительно листать); нет
  обзора списка; **некуда деть inline-create** естественно. Не отвечает
  требованию BL-8.
- **Сложность:** низкая, но не решает задачу.

### Option C — Free-text + автокомплит-дропдаун
Оставить textinput, под ним показывать совпадающие areas по мере ввода; Enter
— принять; если совпадений нет — предложить создать.
- **Pros:** клавиатурно-быстро для тех, кто помнит имя; сохраняет free-text.
- **Cons:** самый сложный рендер (выпадающий список под полем в модалке,
  композиция со скроллом), больше edge-cases (частичный матч, регистр,
  навигация по подсказкам). Избыточно для нашего объёма areas.
- **Сложность:** высокая.

## Constraints & Risks
- **Изменение контракта модели редактора:** переход с «area как имя-строка,
  резолв на save» на «area как `*id.ID`, выбранный в пикере». Затрагивает
  `NewEditor`/`newProjectEditor` (pre-fill), `ApplyAndSave` (больше не резолвит
  имя — берёт готовый `AreaID`), и тесты этих путей. `editor-when-and-picker`
  тесты (`TestEditor_FieldCount…`, save-roundtrip PBT) придётся обновить.
- **Read-only режим:** inline-create — это write. В read-only (`m.readOnly`)
  создание area надо запретить (как `saveEditor` уже делает для save,
  `app.go:439`). Выбор существующей area в read-only — безопасен (запись
  происходит только на save задачи/проекта, который и так заблокирован).
- **Inbox-семантика:** в task editor пусто = Inbox; в project editor пусто =
  none. Пункт списка для «нет area» желательно назвать единообразно (см.
  Open Questions).
- **`ErrAlreadyExists` при inline-create:** нужно дружелюбно обработать — не
  падать, а либо выбрать существующую одноимённую area, либо показать ошибку
  в пикере.
- **Зависимости:** новых внешних зависимостей не требуется — всё на уже
  используемых `bubbles/textinput` + `lipgloss`.
- **Регрессии:** оба редактора и их тесты — основная зона риска; CLI-путь
  (`area add`, save через CLI) не затрагивается.

## Recommended Direction

**Option A — overlay-picker как общий компонент** для обоих редакторов.
Триггер открытия: фокус на поле Area + Enter (по аналогии с тем, как `Space`
обрабатывается для When). Список: `[Inbox / none]`, затем areas из
`ListAreas`, затем строка `+ New area…` для inline-create. Хранить выбранный
`*id.ID` в модели редактора; `ApplyAndSave` перестаёт резолвить имя.

Причина: единственный вариант, который закрывает все три части требования
BL-8 (обзор списка + выбор + inline-create), переиспользует существующие
паттерны (list-рендер + confirm-sub-state) и не тянет внешних зависимостей.

## Scope Boundaries
- **Must-have (v1):**
  - Общий `areaPicker` компонент.
  - Интеграция в **оба** редактора (task + project).
  - Список существующих areas (`ListAreas`) + пункт «нет area» (Inbox/none).
  - Inline-create новой area (`AddArea`) с обработкой `ErrAlreadyExists`.
  - Запрет inline-create в read-only.
  - Юнит-тесты + хотя бы один PBT (по сложившейся практике проекта).
- **Deferred (v2):**
  - Аналогичные пикеры для Project / Heading (сейчас тоже free-text) —
    отдельный pipeline, если понадобится.
  - Фильтрация/поиск внутри пикера по мере набора (типизация-фильтр).
  - Inline-rename / delete area из пикера.
- **Needs spike:**
  - Нет. Все механизмы (list-рендер, sub-state, сервисные методы) уже
    существуют в кодовой базе.

## Assumptions & Open Questions

**Assumptions:**
- `[ASSUMPTION: триггер открытия пикера — Enter на сфокусированном поле Area]`
  (по аналогии со Space для When). Альтернатива — Space; уточнить в Open
  Questions.
- `[ASSUMPTION: free-text для area полностью убирается]` — пикер становится
  единственным способом задать area в обоих редакторах. (Project/Heading
  остаются free-text — вне scope.)
- `[ASSUMPTION: список areas грузится лениво при открытии пикера]` через
  `ListAreas`, а не при открытии редактора — чтобы не платить за загрузку,
  если пользователь не трогает area.
- `[ASSUMPTION: при inline-create с уже существующим (нормализованным) именем
  пикер не падает, а выбирает существующую area]` вместо показа ошибки.

**Open Questions (для requirements):**
1. **Клавиша открытия пикера:** Enter или Space на поле Area? (Tab/Shift+Tab
   уже заняты переключением полей.)
2. **Название пункта «нет area»:** единое «Inbox» в обоих редакторах, или
   контекстное — «Inbox» в task editor и «No area» в project editor?
3. **Поведение Esc внутри пикера:** отменяет выбор и возвращает прежнее
   значение, или закрывает редактор целиком? (Ожидание: только закрывает
   пикер, значение не меняется.)
4. **Inline-create в read-only:** показывать строку `+ New area…` задизейбленной
   или скрывать совсем?
5. **Подтверждение inline-create:** одношагово (ввёл имя → Enter → создано и
   выбрано) или с явным confirm? (Ожидание: одношагово.)

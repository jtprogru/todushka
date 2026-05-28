# Backlog

Список идей и доработок, ожидающих pipeline-планирования. Нумерация —
авторская; держим её стабильной чтобы можно было ссылаться (BL-1.1).

## UI / Layout

### BL-1 — Убрать поля start/due из списка задач ✅ done (list-rendering-polish)
Сейчас в `viewList` (`internal/tui/app.go`) каждая строка содержит
`start:YYYY-MM-DD due:YYYY-MM-DD` после заголовка — это перегружает
список.

#### BL-1.1 — Перенести даты в details pane ✅ done (details-pane-redesign)
Поля `start`, `due` (и потенциально остальные) лучше показывать в
секции подробностей. Каждое имя поля — явно выделено цветом и жирным;
между полями — пустая строка для воздуха.

### BL-2 — Уменьшить details pane до ≤ 40% ширины ✅ done (details-pane-redesign)
Сейчас правая панель занимает ~55% (config `list_pane_share: 0.45`).
Сделать `list_pane_share ≥ 0.6` либо инвертировать смысл field. Хочется
чтобы details ≤ 40%.

### BL-3 — Сделать разделители зон жирнее ✅ done (list-rendering-polish)
Текущий `renderSeparator` — одинарная `─` через `theme.Help`. Хочется
визуально более выраженных границ (двойная линия / контрастный цвет /
жирный символ).

### BL-4 — Перенос строки внутри колонки заголовка ✅ done (list-rendering-polish)
Сейчас длинный title переносится с начала экрана и ломает выравнивание
icon/short/dates. Перенос должен быть строго в пределах title-колонки,
индентация второй строки должна совпадать с началом title.

## Details pane

### BL-6 — В details показать инфо о проекте ✅ done (details-pane-redesign)
Сейчас в details pane нет данных о проекте задачи (только tags, area).
Добавить project name + heading. Возможно ещё что-то — нужен короткий
explore phase прежде чем фиксировать scope.

## Features

### BL-5 — Навигация по проектам (отдельный view) ✅ done (project-navigation)
Реализовано в v0.9.0: новый `screenProjects` + `screenProjectTasks`,
CRUD через `ProjectEditorModel`, service `DeleteProject` с reassign
задач в Inbox. Headings management / bulk-move задач между проектами /
area picker — deferred v2 (если понадобятся, отдельный pipeline).

### BL-8 — Area picker вместо free-text ввода ✅ done (area-picker)
Сейчас area задаётся свободным текстом в двух местах: task editor
(`fieldArea`, `internal/tui/editor.go`) и project editor (`pefArea`,
`internal/tui/project_editor.go`). Оба резолвят имя через
`AreaFindByNormalized` и **падают с ошибкой**, если area не найдена;
создать новую можно только через CLI (`internal/cli/area.go`). Нужен
интерактивный picker:

- **Охват**: оба редактора (task + project) — единый переиспользуемый
  компонент, т.к. паттерн идентичен.
- **Поведение**: список существующих areas (через `svc.ListAreas`) +
  пункт «Inbox» (пусто, `AreaID = nil`) + **inline create** нового area
  прямо из пикера (через `svc.AddArea`), без похода в CLI.
- **Сервис**: `ListAreas` и `AddArea` уже есть — новые методы не нужны.
- **UI**: готового list/picker-компонента в TUI нет; project-navigation
  писал свой кастомный список (`internal/tui/project_list.go`) — можно
  взять как референс или вынести общий помощник.
- **Открытые вопросы для explore/design phase**: как открывать picker
  из формы (Enter/Space на поле → overlay vs inline-cycle); сохранять
  ли free-text как fallback; UX подтверждения inline-create
  (нормализация имени, дубликаты). Нужен короткий explore прежде чем
  фиксировать дизайн.

## Bugs

### BL-7 — Курсор уезжает за пределы экрана при overflow
В `viewList` / `viewProjectList` / `viewProjectTasks` сейчас рендерятся
**все** задачи/проекты целиком, без учёта доступной высоты `bodyH`.
Когда строк больше чем влезает на экран, `View()` обрезает overflow
через lipgloss height-clamp, но курсор продолжает двигаться по списку
— пользователь не видит, что выбрано. Нужен viewport со
scroll-offset, который следит за курсором (edge-follow + scrolloff).

## Things 3 parity (роадмап)

Цель — приблизить TUI к ощущению Things 3. Ключевой факт: доменный слой
уже моделирует почти весь Things (`task.Task`: `AreaID`, `ProjectID`,
`HeadingID`, `Tags`, `StartDate`=When, `Deadline`, `Someday`,
`PinnedToday`, `Checklist`, `Repeat`; `project.Project`:
`Deadline`/`AutoClose`/`Headings`; все шесть списков в `internal/app`).
Поэтому работа почти целиком в presentation-слое (`internal/tui`), новых
схем почти не требуется.

Вне scope для TUI: drag «Magic Plus», слияние событий OS-календаря,
анимации. Capture через quick-entry (`#tag @today @project !date`) уже
закрывает сценарий быстрого ввода.

### Фаза A — визуальная идентичность (fast-track, без смены архитектуры)

#### BL-9 — Progress ring у проектов ✅ done (things-visual)
`projectCounts [open,total]` / `CountProjectTasks` уже есть; сейчас
показываются числом. Рендерить как кольцо прогресса (○ ◔ ◑ ◕ ●) рядом с
именем проекта в `internal/tui/project_list.go` (опц. агрегат по area
позже). Чистый рендер, без новых данных.

#### BL-10 — Визуальные маркеры Today ✅ done (things-visual)
Реализовано: жёлтая звезда `★` у today-задач в списке Anytime (через
`today.ComputeToday`), приглушение завершённых строк, стиль `Theme.Star`.
Иконка-луна This Evening **не вошла** — перенесена в BL-12 (избегаем dead
code до появления самого состояния This Evening).

### Фаза B — точность планирования (концептуальное ядро Things)

#### BL-11 — Богатый «When» picker (overlay) ✅ done (when-picker)
Реализован overlay `whenPicker` (`internal/tui/when_picker.go`): Today /
Pick date… / Someday / Anytime — заменил free-text поле `Start` и тумблер
Anytime/Someday в task-editor. «Today» → `StartDate=today`. Строка This
Evening **не вошла** — она требует нового состояния (BL-12).

#### BL-12 — Зона This Evening в Today
Новое состояние (флаг на задаче или производное от StartDate) +
разделитель в списке Today (дневные сверху, вечерние снизу под луной) +
toggle показа. Зависит от BL-11 (общий выбор состояния When).

#### BL-13 — Группировка Upcoming по дням и Logbook по датам
Сейчас списки плоские. Добавить заголовки-группы по дате в рендере
`ListUpcoming`/`ListLogbook`. Logbook — по дате завершения, различать
completed/cancelled визуально.

### Фаза C — структура и группировка

#### BL-14 — Headings: рендер по группам + CRUD
Отложенная половина BL-5. Группировать задачи проекта под headings в
`internal/tui/project_tasks.go`; CRUD heading (домен `Heading`,
`HeadingID`, кэш `headingNamesByID` уже есть). Drag-группы — N/A в TUI,
вместо этого move-под-heading клавишами.

#### BL-15 — Tag filter bar (чипы) + фильтр по тегам
Сейчас `/`-фильтр — это regex по заголовку (`internal/tui/filter.go`).
Добавить строку чипов активных тегов сверху списка + фильтрацию по
выбранному тегу (теги уже в домене и в details).

#### BL-16 — Inline-редактирование чеклиста
`task.Checklist []ChecklistItem` уже есть; добавить добавление /
переключение / удаление пунктов прямо в task editor
(`internal/tui/editor.go`).

### Фаза D — постоянный сайдбар (определяющий лейаут, крупнейший рефактор)

#### BL-17 — Persistent sidebar
Заменить header-tabs + отдельный `screenProjects` на левый сайдбар
(шесть списков + дерево Areas→Projects) | основной список | details.
Затрагивает монолитный `Model` (63 поля) и dispatch во `View`
(`internal/tui/app.go`). Делать последним; желательно после выделения
саб-компонентов. Высокий payoff, высокий риск.

## Группировка для планирования (предложение)

Можно сгруппировать в 3 будущих feature-pipeline'а:

1. ~~**list-rendering-polish** (BL-1, BL-3, BL-4)~~ ✅ done — рендеринг
   списка и разделителей. Один fast-track цикл.
2. ~~**details-pane-redesign** (BL-1.1, BL-2, BL-6)~~ ✅ done —
   переработка details pane: ширина, контент, стилизация. Один full
   pipeline (затрагивает layout + data fetching).
3. ~~**project-navigation** (BL-5)~~ ✅ done — отдельный полноценный
   pipeline.

Открытые пункты:

4. ~~**area-picker** (BL-8)~~ ✅ done — full pipeline. Общий
   `areaPicker` компонент + интеграция в оба редактора + inline-create.

Things 3 parity (новые pipeline'ы, по фазам — порядок = приоритет):

5. ~~**things-visual** (BL-9, BL-10) — фаза A, fast-track, без архитектуры.~~ ✅ done — кольцо прогресса + звезда «today» + faint завершённых; `Theme.Star` (релиз minor v0.11.0).
6. **things-scheduling** — фаза B. BL-11 ✅ done (отдельный pipeline
   `when-picker`, релиз minor); осталось BL-12 (This Evening + moon) и
   BL-13 (группировка Upcoming/Logbook).
7. **things-structure** (BL-14, BL-15, BL-16) — фаза C, можно дробить.
8. **things-sidebar** (BL-17) — фаза D, крупный рефактор, последним.

Решение по группировке остаётся за пользователем; список выше —
рекомендация для удобства.

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
   Бэклог снова пуст; ждём новых идей.

Решение по группировке остаётся за пользователем; список выше —
рекомендация для удобства.

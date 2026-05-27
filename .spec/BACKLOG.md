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
   pipeline. Все 6 пунктов бэклога закрыты.

Решение по группировке остаётся за пользователем; список выше —
рекомендация для удобства.

# Exploration: when-picker (BL-11)

## Intent
Дать задаче единый Things-подобный выбор «When» — overlay-пикер с
вариантами Today / выбрать дату / Someday / Anytime(очистить) — вместо
текущего разрозненного планирования. This Evening и moon-иконка
сознательно вне scope (BL-12). Brownfield: меняем редактор задачи,
домен не трогаем.

## Investigation
- **Сейчас планирование разорвано на 3 контролах:**
  1. `fieldStart` — free-text `YYYY-MM-DD` (`editor.go:73-78`, парсится в
     `ApplyAndSave` → `t.StartDate`).
  2. `fieldWhen` — тумблер Anytime↔Someday (Space в `handleEditorKey`,
     `app.go`); `ApplyAndSave` ставит `t.Someday`.
  3. «Today» вообще не в редакторе — это отдельный pin `p` в списке
     (`PinnedToday`, `app.go pinSelected`).
- **Движок Today** (`today.ComputeToday`, `today.go:21`): задача
  «сегодня», если `PinnedToday==today` ∨ `StartDate<=today` ∨
  (`StartDate==nil` ∧ `Deadline<=today+window`). Значит «Today» через
  `StartDate=today` корректно попадёт и в Today, и в Anytime (со
  звездой из v0.11.0).
- **Паттерн overlay уже есть** — `areaPicker` (`area_picker.go`):
  структура с `cursor`, `Update(msg, svc) → pickerResult{outcome,...}`,
  `View(theme,width)`; редактор держит `picker *areaPicker`, открывает по
  Enter на `fieldArea`, роутит ключи пока `picker != nil`
  (`handleEditorKey`, `app.go:408-421`), под-режим ввода (`creating` +
  `nameInput`) — точный аналог для «выбрать дату».
- **Тесты, которые затронем:** `editor_test.go` (ссылается на
  `fieldStart`, `fieldWhen`, `whenAnytime/whenSomeday`, Space-toggle),
  `project_navigation_pbt_test.go` (editor round-trip).

## Build Tooling
- **Orchestrator:** task (`Taskfile.yml`)
- **Test:** `task test` · **Build:** `task build` · **Lint:** `task lint` · **Fmt:** `task fmt`

## Options Considered
Центральный вопрос — как пикер соотносится с текущим free-text `Start` и
тумблером Anytime/Someday.

### Option A — единый When-пикер заменяет и `Start`, и тумблер
Поле «When» открывает overlay (Enter): строки **Today / Pick date… /
Someday / Anytime (clear)**. «Pick date…» переводит в под-режим ввода
даты (как `creating` у areaPicker). Пикер пишет `StartDate`
(Today=сегодня, date=выбранная), либо `Someday=true`, либо очищает
(`StartDate=nil, Someday=false`). Отдельное текстовое поле `Start`
удаляется.
- **Pros:** один контроль, Things-подобно, меньше визуального шума,
  единый источник истины для start.
- **Cons:** меняется набор полей редактора (удаляем `fieldStart`),
  правки в `editor_test.go`; ввод даты теперь вложен в пикер.
- **Complexity:** standard.

### Option B — пикер заменяет только тумблер, `Start` остаётся
Пикер: Today / Someday / Anytime. Free-text `Start` сохраняется.
- **Pros:** минимальная правка, поле `Start` для явных дат не трогаем.
- **Cons:** два контрола пишут в `StartDate` (пикер «Today» и поле
  `Start`) — раздвоённый источник истины, путаница, не Things-подобно.
- **Complexity:** mechanical, но архитектурно хуже.

## Constraints & Risks
- **Source of truth:** при Option B легко рассинхронить пикер и `Start`.
  Option A устраняет это by design.
- **«Today» семантика:** `StartDate=today` (Things-way) против
  `PinnedToday=today` (текущий pin). Рекомендую `StartDate=today`; pin
  `p` в списке остаётся для ad-hoc отметки без смены start-даты.
- **Read-only:** пикер только формирует локальный выбор; запись — на
  Ctrl+S (редактор уже блокирует save в RO). Доп. проверок не нужно.
- **Quick-entry** (`@today`, `!date`) — отдельный путь, не затрагивается.
- **Контракт:** доменных изменений нет (используем `StartDate`/`Someday`).
  Меняется UI-набор полей редактора → правки тестов редактора.

## Recommended Direction
**Option A.** Реализация по образцу `areaPicker`: новый `whenPicker`
overlay; «Pick date…» переиспользует `textinput` для `YYYY-MM-DD`
(валидируется тем же `time.ParseInLocation`, что сейчас в `ApplyAndSave`).
«Today» → `StartDate=today`. `Deadline` остаётся отдельным полем (это
другой концепт). Список рядов в BL-11: Today / Pick date… / Someday /
Anytime — строка This Evening добавится в BL-12.

## Scope Boundaries
- **Must-have (v1):** overlay `whenPicker` (Today / Pick date… / Someday /
  Anytime) на месте поля When; ввод даты в под-режиме; интеграция в
  task-editor; перенос логики `StartDate`/`Someday` из `Start`-поля и
  тумблера в пикер.
- **Deferred (v2):** строка This Evening + moon (BL-12); тот же пикер в
  project-editor (у проекта нет StartDate — отдельный вопрос); reminder
  со временем.
- **Needs spike:** нет (паттерн overlay уже валидирован в BL-8).

## Assumptions & Open Questions
- `[ASSUMPTION: Deadline остаётся отдельным текстовым полем, в пикер не входит]`
- `[ASSUMPTION: пикер только в task-editor; project-editor не трогаем (у проекта нет start-даты)]`
- **Q1 — решено:** Option A — единый When-пикер **заменяет** free-text `Start`.
- **Q2 — решено:** «Today» → `StartDate=today`; pin `p` (`PinnedToday`) остаётся отдельным механизмом.

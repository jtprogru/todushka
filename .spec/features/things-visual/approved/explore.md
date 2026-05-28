# Exploration: things-visual (Фаза A)

## Намерение
Приблизить вид TUI к Things 3 двумя дешёвыми визуальными приёмами,
полностью в presentation-слое, без изменения домена: **BL-9** — кольцо
прогресса у проектов вместо текстового `[open/total]`; **BL-10** —
жёлтая звезда «сегодня» у задач, видимых в списке Anytime (фирменный
маркер Things). Триггер — роадмап Things 3 parity (`.spec/BACKLOG.md`).
Brownfield: меняем существующий рендер, домен не трогаем.

## Исследование
- **BL-9.** Счётчик рендерится в `internal/tui/project_list.go:124-126`
  из `m.projectCounts[p.ID] = [2]int{open, total}` (заполняется в
  `fetchProjects`, через `svc.CountProjectTasks`). Прогресс выполнения =
  `(total-open)/total`. Глиф-кольцо можно поставить рядом с `icon`.
- **BL-10.** Классификация «сегодня» — чистый движок
  `today.ComputeToday` (`internal/domain/today/today.go:21`): задача
  «сегодняшняя», если open и (`PinnedToday==today` ∨ `StartDate<=today`
  ∨ (`StartDate==nil` ∧ `Deadline<=today+window`)). `ListAnytime`
  (`internal/app/queries.go:71`) включает open, не-Someday, с
  area/project, `StartDate` не в будущем — т.е. сегодняшние задачи с
  area/project одновременно попадают и в Anytime. Значит звезду в
  Anytime ставим тем строкам, что проходят предикат «today».
- **Рендер списка.** `viewList` (`internal/tui/app.go:1017`) знает
  активный список через `m.activeList`; глифы статуса ✓/✗ уже есть
  (`app.go:1050-1055`) со strikethrough. Отдельного стиля для звезды
  нет.
- **Тема.** `internal/tui/style.go`: `Warning` (жёлтый `#eed49f`)
  подходит под звезду; monochrome-тема существует и должна
  деградировать.

## Build Tooling
- **Orchestrator:** task (`Taskfile.yml`)
- **Test:** `task test` (и `task test-race`)
- **Build:** `task build`
- **Lint:** `task lint`
- **Fmt:** `task fmt`
- **Source:** `Taskfile.yml`

## Рассмотренные варианты (источник `now` для звезды)
- **A — предикат в рендере через `time.Now()`.** Просто, но `View`
  становится time-dependent: тесты придётся завязывать на дату.
- **B — предвычислять множество today-ID при загрузке Anytime** и
  хранить в `Model` (напр. `todayIDs map[id.ID]struct{}`). `View`
  остаётся чистым и детерминированным; +1 поле и шаг в загрузке.
  Соответствует курсу проекта на детерминизм (PBT, `-count=1`).

## Ограничения и риски
- **Monochrome / NO_COLOR:** кольцо и звезда должны деградировать
  (кольцо — ASCII-доля или прочерк, звезда — `*`).
- **Ширина / dual-pane:** глифы — одноширинные руны (lipgloss width 1),
  но добавляют символы в строку проекта; следить за переносом.
- **Регрессии:** не ломать `viewport_test.go`, `project_list_test.go`
  (они проверяют маркер курсора и границы окна).

## Рекомендуемое направление
**Решено с пользователем:**
- **Источник «today» для звезды — вариант A** (`time.Now()` в рендере).
  Чтобы не дублировать предикат, `viewList` при `activeList==Anytime`
  вызывает `today.ComputeToday(disp, time.Now(), 0)`, строит множество
  ID и помечает их звездой. Детерминизм тестов достигается тем, что
  тестовые задачи строятся относительно `time.Now()` (напр. `PinnedToday`
  = сегодня → звезда есть; `StartDate` в будущем → звезды нет).
- **Scope BL-10 — расширенный:** звезда **плюс** полировка глифов
  completed/cancelled и тюнинг акцентов/воздуха. Конкретные,
  проверяемые критерии этой полировки фиксируются в фазе Requirements
  (иначе «polish» невозможно верифицировать).

Moon-иконку This Evening по-прежнему НЕ добавляем — это BL-12.

## Границы scope
- **Must-have (v1):**
  1. кольцо ◯◔◑◕● у проектов в `project_list.go`;
  2. жёлтая звезда у today-задач в списке Anytime в `viewList`;
  3. полировка глифов completed/cancelled + тюнинг акцентов/воздуха
     (точные критерии — в Requirements).
- **Deferred (v2):** moon-иконка This Evening (→ BL-12); агрегатное
  кольцо прогресса по area.
- **Needs spike:** нет.

## Допущения и открытые вопросы
- `[ASSUMPTION: кольцо = доля выполненных (total-open)/total → ближайший
  из ◯◔◑◕●; пустой проект (total==0) → ◯]`
- `[ASSUMPTION: звезда видна только в активном списке Anytime; в самом
  Today звезда не нужна, т.к. там и так всё «сегодня»]`
- **Open Q (для Requirements):** какой именно набор изменений считать
  «полировкой глифов и воздуха» — предложу конкретный минимальный список
  в фазе Requirements на подтверждение.

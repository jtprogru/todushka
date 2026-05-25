# Exploration: Dual-Pane Layout for TUI

## Intent

Сейчас при `screen == screenList` экран занят одной колонкой — list of tasks. Чтобы посмотреть детали задачи (notes, checklist, tags, история completion), пользователь обязан жать `Enter` и открывать full-screen EditorModel — это смена контекста, плюс редактор сразу в режиме edit (нельзя просто "посмотреть и поскроллить").

Цель v1: при ширине терминала ≥ некоторого порога TUI автоматически переключается на **двухколоночный режим**:

- **Левая панель** — текущий список (полная функциональность filter/select/bulk сохранена).
- **Правая панель** — read-only details выделенной курсором задачи (полный title, notes с переносами, dates, tags, area/project linkage, completion history).
- При узком терминале (< threshold) — single-pane как сейчас (полная backward-compat).
- Курсорная навигация в списке (`j`/`k`) обновляет правую панель в реальном времени.

Это первый шаг к "приложение, а не CLI с экраном" — резко поднимает воспринимаемую "полноценность" продукта.

Greenfield-функциональность; backward compat для узких терминалов критичен.

## Investigation

### Существующая модель ширины

- **`app.go:36`**: `Model.width int` — обновляется в `Update` через `tea.WindowSizeMsg` (`app.go:66-68`).
- **`app.go:436-444`** `editorWidth()`: уже использует `m.width` для clamp'а editor pane к [60, 80]. Pattern для адаптивных width'ов установлен.
- **`viewList`** (`app.go:458-494`): рендерит линии через `lipgloss.JoinVertical`. Длина строк не обрезается под `width` — длинные titles на узком терминале просто переносятся wrapping'ом терминала. Это сейчас визуально неприятно но "работает".

### Существующие screens и их layout

- **screenList**: header + viewList + footer. Простая 3-row структура.
- **screenQuickEntry**: вызов `lipgloss.JoinVertical(viewList, viewQuickEntry)` — quick-entry рендерится поверх (вернее, под) списком.
- **screenEditor**: `m.editor.View(m.theme, m.editorWidth())` — full pane, заменяет viewList.
- **screenHelp**: `viewHelp()` — full pane.
- **Confirm modal**: `JoinVertical(viewList, modal)` — modal под списком.

Pattern: для overlays используется vertical join. Для split-screen нет prior art в проекте — это будет первый горизонтальный split.

### Lipgloss возможности

- `lipgloss.JoinHorizontal(lipgloss.Top, leftStr, rightStr)` — горизонтальный split.
- `lipgloss.NewStyle().Width(N)` — фиксированная ширина для pane.
- `lipgloss.Style.MaxHeight(N)` — clamp по высоте (для дисциплины при overflow).
- `lipgloss.Border()` — border между панелями (опционально).

### Task domain — что показывать в правой панели

Из `internal/domain/task/task.go` (изучено в feature #1):
- `Task.Title string`
- `Task.Notes string` — multi-line, до 4000 chars
- `Task.Status` (Open / Completed / Cancelled)
- `Task.StartDate *Date`, `Task.Deadline *Date`
- `Task.Tags []id.ID` — нужен lookup name через `Repo().TagGet` (как в `lookupTagNames` в editor.go)
- `Task.AreaID *id.ID`, `Task.ProjectID *id.ID`, `Task.HeadingID *id.ID` — тоже lookup
- `Task.Checklist []ChecklistItem` — массив `{ID, Text, Done bool}`
- `Task.Someday bool`
- `Task.Repeat *Rule`
- `Task.PinnedToday *Date`
- `Task.CompletedAt *time.Time`, `Task.CancelledAt *time.Time`
- `Task.CreatedAt`, `Task.UpdatedAt`

### Tag/Area/Project lookup performance

`lookupTagNames` (`app.go:245-255`) делает N сетевых вызовов через `Repo().TagGet`. Для каждой смены курсора (`j`/`k`) делать full resolve = пере-rendering на каждый keypress с N DB lookups. Для типичной задачи N ≤ 3-5 — приемлемо (~1ms каждый). Но если задача в Inbox без tags/area/project — 0 lookups.

Mitigation для perf: cache resolved names в Model по ID. Или ленивая lazy-resolve.

### Тесты и стиль

- `app_test.go` тестирует через `m.viewList()`, `m.viewFooter()`, etc. Можно тестировать `m.viewDetails()` тем же паттерном.
- Тесты width-зависимого поведения: установить `m.width = 120`, проверить что виден правый pane; установить `m.width = 70`, проверить single-pane.

## Build Tooling

- **Orchestrator:** Taskfile (`Taskfile.yml`)
- **Test:** `task test` / `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format:** `task fmt`
- **Source:** `Taskfile.yml`

## Options Considered

### Option A — Static threshold + lipgloss horizontal split

- Hardcoded `dualPaneMinWidth = 100` (or 120).
- `viewBody()` решает: если `m.width >= threshold` → `JoinHorizontal(viewList, viewDetails)`; иначе `viewList()`.
- Правая панель — `viewDetails(m)` функция, читает `m.selectedTask()`, форматирует поля задачи.
- Tag/area/project names resolve'ятся по запросу через `Repo()` (синхронно внутри View — но View должен быть pure, так что лучше pre-fetch на keypress).

**Pros:**
- Минимальный blast radius — одна новая функция + ветка во `View()`.
- Backward-compat автоматичен — при `m.width == 0` (тесты, начало сессии) применяется fallback к single-pane.

**Cons:**
- Tag/area lookup в View нельзя делать (View должен быть pure). Нужно резолвить в Model — либо при `tasksLoadedMsg`, либо при `cursor` change.
- Жёсткий threshold не учитывает font/zoom — на retina с маленьким шрифтом 100 cols узкие, на стандартном macOS Terminal — широкие.

**Complexity:** Medium (M ~ 2-3 дня)

### Option B — Configurable threshold + adaptive ratio

Как Option A, но:
- `TODUSHKA_DUAL_PANE_MIN_WIDTH` env / config-flag — переопределяет дефолт.
- `dualPaneRatio = 0.4` (40% список / 60% details) configurable.

**Pros:**
- Гибкость для разных setup'ов.

**Cons:**
- Усложнение API без явной потребности — YAGNI.

**Complexity:** Medium-Large (added config infrastructure).

### Option C — Manual toggle (key `v` для view modes)

- Пользователь явно переключает: 1-pane / 2-pane / details-only.
- Width не влияет автоматически.

**Pros:**
- Полный контроль пользователя.
- Не зависит от детектирования ширины.

**Cons:**
- Лишний шаг при каждом сеансе.
- Усложнение mental model.

**Complexity:** Medium.

### Option D — Modal "preview" вместо split

Вместо split — отдельный screen `screenPreview` который активируется `Tab` (или другая клавиша) и показывает details full-screen, как viewHelp.

**Pros:**
- Никакой horizontal split — проще layout.
- Работает на любой ширине.

**Cons:**
- Контекст списка теряется при просмотре деталей.
- Это в сущности read-only вариант editor'а — почему не открыть editor?

**Complexity:** Small. Но это не "двухпанельный layout", это другая фича.

## Constraints & Risks

- **View purity**: `View()` метод не должен делать I/O. Tag/area/project name resolution через `Repo()` — это IO. Нужно pre-resolve в Model.
  - Resolution на каждый cursor move — лишний overhead.
  - Mitigation: keep tag names cache `Model.tagNamesByID map[id.ID]string` который lazy-populate'ится при `tasksLoadedMsg` ОДНИМ batch lookup (или при первом обращении к unfamiliar ID).

- **Terminal width = 0 в тестах**: `newTestModel(t)` не получает `tea.WindowSizeMsg`, так что `m.width == 0`. Текущий код `editorWidth()` уже handle'ит это (clamp к min 60). Dual-pane должен использовать тот же подход — `m.width == 0` → single-pane (как сейчас).

- **Lipgloss wrapping**: длинные строки в notes могут разорвать pane. Использовать `lipgloss.Style.Width(N).MaxHeight(M)` для clamp'а deteails-pane.

- **Filter mode на узком терминале**: filter input занимает footer. На широком — где? Решение: filter input всегда в footer, не влияет на dual-pane (footer общий для обеих pane).

- **Confirm modal в dual-pane**: modal сейчас стекается под list через `JoinVertical(viewList, modal)`. В dual-pane вариантах: (a) overlay на список only, (b) overlay на весь body, (c) modal "поверх" с центрированным positioning через lipgloss.Place. Для v1 — variant (a): JoinVertical(splitBody, modal).

- **Editor screen в dual-pane**: при открытии editor (`Enter`) на широком терминале — editor по-прежнему full-pane (`screenEditor`), details-pane скрывается. Не вижу причин менять editor logic.

- **Empty state**: курсор на пустом списке (или filter narrow → 0). Правая панель показывает "Select a task to see details" или пустоту. Решение: показывать "—" или пустую панель с border'ом.

## Recommended Direction

**Option A — Static threshold + horizontal split**, со следующими решениями:

- `dualPaneMinWidth = 100` (8 cols header + 6 cols dividers + 86 useful).
- `dualPaneListShare = 0.45` (45% список, 55% details — детали обычно длиннее).
- Tag/area/project name resolution: pre-resolve in batch при `tasksLoadedMsg` через единый `Repo().TagList()` + `Repo().AreaList()` + `Repo().ProjectList()` → in-memory maps в Model. Cursor move → no IO.
- Single new file `internal/tui/details.go` — `viewDetails(m Model) string`.
- `viewBody()` (новая функция в app.go) — диспетчер 1-pane / 2-pane.

Threshold и share — пока константы; если в feedback потребуется конфиг — вернёмся к Option B.

## Scope Boundaries

### Must-have (v1)

- При `m.width >= 100` и `screen == screenList` (без filter/confirm/editor) — split: список слева, details справа.
- Details показывают: full Title (с переносом), Status, Notes (с переносом, до 8 строк max), StartDate/Deadline, Tags (по именам), Area/Project/Heading names, Someday flag, RepeatRule short description, CompletedAt/CancelledAt, ChecklistItem'ы с `[x]/[ ]`.
- При `m.width < 100` — single-pane (текущее поведение, без изменений).
- Курсор в списке (`j`/`k`/`Up`/`Down`) обновляет details pane (через recompute на каждый render).
- При пустом списке / cursor вне диапазона — details показывают placeholder `(no task selected)`.
- Filter, multi-select, bulk, confirm modal продолжают работать (left pane неизменно).
- Confirm modal стекается под весь body через JoinVertical.
- Tag/area/project names cached в Model — populate при `tasksLoadedMsg`.

### Deferred (v2)

- Configurable threshold + ratio через env / config-file.
- Manual toggle `v` для override (Option C).
- Scroll внутри details pane (для длинных notes).
- Inline edit одного поля прямо в details pane (без открытия editor'а).
- Подсветка changes в details (notes-diff с предыдущей версии).
- Анимация при cursor move (smooth highlight transition).

### Needs spike

- Поведение на очень узких терминалах < 60 cols — может потребоваться отдельный обработчик (truncation вместо wrapping).
- Markdown rendering в notes через glamour — отдельная фича.

## Assumptions & Open Questions

### Assumptions

- **[ASSUMPTION: 100 cols — реалистичный порог для dual-pane; стандартный 80-col терминал остаётся single-pane]** — большинство dev'ов работают на 120+ cols (iTerm/Terminal default).
- **[ASSUMPTION: lipgloss.JoinHorizontal + Width(N) гарантирует визуальное соблюдение ширины без overflow]** — стандартный pattern, документирован.
- **[ASSUMPTION: Pre-resolve tag/area/project names в Model — приемлемо по memory]** — даже 1000 tasks × 10 tags = 10k IDs × ~50b name = 500KB. Незаметно.
- **[ASSUMPTION: View() остаётся pure — никакого IO в render path]** — фундаментальный принцип Bubble Tea, нарушать нельзя.
- **[ASSUMPTION: Existing single-pane тесты остаются валидны — `m.width = 0` означает single-pane]** — backward compat.

### Open Questions

1. **Threshold = 100 или 120?** Trade-off: 100 покрывает больше пользователей (включая 100x30 split-window в tmux). 120 даёт details pane комфортнее. **Рекомендация: 100, можно подсмотреть в Things 3 / iTerm стандарт.**
2. **Какие из 10+ полей задачи показывать?** Все? Или приоритизировать (title, notes, dates, status — must; tags, area, project, checklist — nice; createdAt/updatedAt — debug-only)? **Рекомендация: всё кроме CreatedAt/UpdatedAt (это metadata, не пользовательские поля).**
3. **Refresh tag/area/project names при их редактировании?** Если пользователь переименовал tag через CLI в другой сессии, то TUI продолжит показывать старое имя до перезагрузки. **Рекомендация: re-fetch при `tasksLoadedMsg` — это уже happens после bulk-операций.** Если нужен polling — defer to v2.
4. **Border между панелями — да или нет?** Border — это lipgloss `BorderLeft(true)` на details. Эстетично, но "съедает" 1 кол. **Рекомендация: тонкий border `│` через `Style.Border(lipgloss.NormalBorder()).BorderLeft(true)`.**
5. **Что показывать в нижнем углу details — счётчик типа `3/12 tasks`?** Можно. **Рекомендация: defer to v2 — footer уже это показывает.**

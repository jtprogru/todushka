# Exploration: Search & Multi-Select for TUI

## Intent

В текущем TUI любая операция — `complete`/`cancel`/`delete`/`pin` — работает только над курсором (одна задача за раз). Это болезненно при разборе Inbox после нескольких дней или при еженедельном обзоре, когда нужно за раз пометить 5–10 однотипных задач. Дополнительно: нет способа найти задачу в текущем списке по тексту — если Inbox разросся до 30+ позиций, поиск превращается в визуальное сканирование.

Цель v1: добавить две ортогональные возможности, которые усиливают друг друга:

1. **Inline-фильтр текущего списка** — `/` запускает live-фильтрацию, обновление при каждом keypress.
2. **Multi-select** — `Space` тоггл выделения на строке, существующие шорткаты (`c`/`x`/`d`/`p`) автоматически работают над выделенными при непустом наборе, иначе — как раньше, над курсором.

Greenfield-функциональность, никакая существующая логика не меняется. Триггер — собственный UX-фидбек (см. предложение от 2026-05-25 в Roadmap-обсуждении).

## Investigation

### TUI architecture (`internal/tui/`)

- **`app.go:23-37`**: `Model` хранит `screen screenKind`, `activeList listKind`, `tasks []task.Task`, `cursor int`, `statusMsg string`. Нет места для фильтр-state или selection-state — Model придётся расширять.
- **`app.go:104-167`**: `handleKey` ветвится сначала по `screen` (quickEntry/editor имеют свои хэндлеры), потом по key bindings. Текущий screenList не имеет sub-states — фильтрация потребует либо нового `screenSearch`, либо boolean `filterActive` без выделенного screen'а.
- **`app.go:299-357`**: per-action хелперы (`completeSelected`/`cancelSelected`/`deleteSelected`/`pinSelected`) одинаково вытягивают `selectedTask()` и зовут service. Все четыре можно рефакторить под "над `tid` или slice IDs".
- **`app.go:398-421`**: `viewList` — прямой `strings.Join` без bubbles/list виджета. Добавить чекбокс-префикс `[x]` / `[ ]` — тривиально.
- **`keys.go`**: `/` и `Space` сейчас не забиндены в list-режиме. Конфликтов нет.
- **`app.go:264-290`**: `loadCurrentList()` загружает полный список — это означает, что **фильтрация может быть чисто in-memory, без обращений к storage**.

### Storage layer (`internal/storage/repository.go`)

- **`TaskFilter` (lines 30-43)**: поля для Status/Area/Project/Tags/даты — но **нет** `TitleContains` или подобного. Текстовый поиск через storage в репозитории отсутствует.
- **`TaskMatchShort`**: единственный пример "fuzzy lookup" — по префиксу короткого ID. Используется только в CLI для разрешения IDs.
- bbolt-имплементация делает full scan при `TaskList` (видно из `app/queries.go:42` — ListToday загружает все open tasks и фильтрует in-memory). Это значит: добавлять text-search в `TaskFilter` сейчас не даст perf-выигрыша — будет тот же full scan плюс substring match.

### Service layer (`internal/app/`)

- **`queries.go:19-112`**: каждый `ListXxx(ctx)` возвращает уже отсортированный slice. TUI вызывает их и получает финальные данные.
- **`service.go:99-164`**: `CompleteTask`/`CancelTask`/`DeleteTask` принимают один `taskID id.ID`. Bulk-операций нет.
- **Service не транзакционный**: каждый Update — отдельный `repo.TaskUpdate`. Для v1 это ок: bulk-операция в TUI = цикл по выделенным с непрерывающимся прогрессом.

### Тесты и стиль

- `internal/tui/app_test.go` тестирует Update через прямые вызовы `Update(tea.KeyMsg{...})` и проверяет переходы Model. Те же паттерны подходят для тестов filter/select.
- Тестовый фреймворк: `testify` + `pgregory.net/rapid` (property-based для repeat-движка). Конвенции — table-driven tests.
- Никаких screenshot/VHS-тестов TUI пока нет.

## Build Tooling

- **Orchestrator:** Taskfile (`Taskfile.yml`)
- **Test:** `task test` (`go test ./...`) или `task test-race` (`go test -race ./...`)
- **Build:** `task build` (`go build -o bin/todushka ./cmd/todushka`)
- **Lint:** `task lint` (`golangci-lint run`)
- **Format:** `task fmt`
- **Run:** `task run`
- **Source:** `Taskfile.yml` + `.golangci.yml`

## Options Considered

### Option A — Client-side filter + in-Model selection

Фильтрация и выделение живут целиком в TUI Model:

- `Model.filterQuery string` + `Model.filtering bool`
- `Model.selected map[id.ID]struct{}` (set-семантика)
- `displayedTasks()` метод вычисляет видимый срез `tasks` через substring-match по `filterQuery`
- Bulk: `completeSelected()` итерирует по `selected`, генерит `tea.Batch` команд

**Pros:**

- Zero изменений в storage / service интерфейсах
- Мгновенный отклик (filter on keypress, no I/O)
- Backward-compatible: пустое выделение = старое поведение

**Cons:**

- Поиск только в **текущем списке** — нельзя искать "купить молоко" сразу везде
- Если в будущем datasets перейдут за 100k задач, full-scan на каждый keypress начнёт тормозить (но это далеко)

**Complexity:** Low (M ~= 1 день: 200-300 строк + тесты)

### Option B — Service-level `SearchTasks` + TaskFilter.TitleContains

Поиск становится первоклассным запросом к service:

- `TaskFilter.TitleContains string` (новое поле)
- `Service.SearchTasks(ctx, query, scope)` где scope = current list или "all open"
- CLI бонус: `todushka search "молоко"` работает через тот же путь

**Pros:**

- Глобальный поиск (across all lists, not just active)
- Переиспользуемо в CLI и потенциальном HTTP API
- "Правильнее" по слоям

**Cons:**

- Меняет storage interface (нужно пробрасывать TitleContains в bbolt, fakes, repository)
- Performance idентичен Option A (full scan)
- Большая поверхность изменений на не-самый-большой выигрыш в v1

**Complexity:** Medium (M ~= 2-3 дня: ~600 строк, включая обновление repository contract)

### Option C — Hybrid: Option A для v1, Option B как deferred

- Сейчас: client-side filter в текущем списке + multi-select (Option A целиком)
- Когда понадобится глобальный поиск — добавим service-level `SearchTasks` и bind `g/` или `Ctrl+F` под отдельный "All open" pseudo-список

**Pros:**

- Быстро дойти до пользовательской ценности
- Не создавать сейчас интерфейсы под use-case, которого ещё нет
- Не закрывает дверь к Option B

**Cons:**

- Risk: позже придётся два пути держать (in-list filter ≠ global search) — но это норм, Things 3 так же делает (cmd+F локально vs Quick Find глобально).

**Complexity:** Low (= Option A); v2 расширение — Medium позже.

## Constraints & Risks

- **Конфликт клавиш**: `/` и `Space` в screenList сейчас свободны. `Space` уже используется в screenEditor (`app.go:218` — toggle Someday), но это другой screen — конфликта нет. Перед коммитом проверить, что `Space` не съест keymap helper'ов bubbles.
- **Визуальный layout**: префикс `[x]` (4 символа) сдвинет содержимое строки. На терминале <60 cols длинные title начнут обрезаться раньше. Mitigation: показывать `[x]` только в "select-mode" (когда selected != ∅), иначе сохранять текущий вид.
- **Selection invariants**: при `loadCurrentList()` после bulk-операции `selected` должно очищаться (иначе ID несуществующих задач остаются). Также — при `switchList` (Tab/1-6) и при выходе из фильтра, если задача "скрылась" фильтром, но осталась выделенной.
- **Backward compat**: текущие пользователи привыкли что `c` complete'ит **строку под курсором**. Новое поведение: `c` complete'ит **выделенные, если selected ≠ ∅; иначе строку под курсором**. Это additive — не ломает старый воркфлоу.
- **Filter case-sensitivity**: `strings.Contains(strings.ToLower(t.Title), strings.ToLower(query))` — но Russian text без `strings.ToLower` нормализуется некорректно (Turkish-i проблема). Использовать `golang.org/x/text/cases` для unicode-safe lower. **[ASSUMPTION: substring fold-case достаточно для v1; fuzzy/Levenshtein → v2]**.
- **Bulk error handling**: если из 10 выделенных complete'ится 7 а 3 фейлятся — что показать? Решение для v1: продолжать до конца, показать `"completed 7/10, see logs for errors"` в status bar. **[ASSUMPTION: best-effort с агрегированным сообщением приемлемо]**.
- **Bubble Tea Cmd batching**: bulk через `tea.Batch(...)` — все Cmds стартуют параллельно. Для bbolt это последовательный fsync — параллелизм не выиграет, но не сломается (bbolt thread-safe). Можно начать с последовательного варианта (один Cmd, цикл внутри).

## Recommended Direction

**Option C — Hybrid (Option A для v1)**.

Доставляет 90% UX-ценности с минимальным blast radius:

- Никаких изменений в `Repository` / `Service` интерфейсах
- ~250-350 строк нового кода в `internal/tui/` (фильтр, выделение, view-обновления)
- ~150-200 строк тестов
- Backward-compatible: каждая существующая клавиша работает как раньше при пустом выделении

Глобальный поиск (Option B) оставляем на v2 — он понадобится **после** того как Areas/Projects заполнятся реальными данными и появится потребность в Quick Find через всё.

Бонус: Option A создаёт инфраструктуру (Model.selected, displayedTasks), на которой v2 (target picker для bulk Move, fuzzy autocomplete) построится естественно.

## Scope Boundaries

### Must-have (v1)

- `/` в screenList переключает в filter-mode: status-line снизу становится input'ом, ввод сужает список в реальном времени
- `Esc` или пустая строка + Enter — выход из фильтра, очистка query
- `Space` тоггл выделения на строке под курсором; в строке появляется `[x]` / `[ ]` (только когда selected ≠ ∅)
- Пути ввода Multi-select: ручной (`Space` по одной) и "Select All Visible" — клавиша `*` или `Ctrl+A`
- Bulk-операции через существующие шорткаты (`c`/`x`/`d`/`p`), которые в наличии selected действуют над всем selected, иначе — над курсором
- `Esc` (или `Ctrl+\`) полностью сбрасывает selection
- Status bar показывает `Filter: <q>` и/или `Selected: N` когда активно
- Selection очищается при `switchList` и после bulk-операции
- Filter — substring case-insensitive по `Title` только

### Deferred (v2)

- Глобальный поиск (Option B) с CLI-командой `todushka search`
- Фильтр по Notes / Tag / Area / Project в `/` синтаксисе (`@area:work`, `#tag`)
- Fuzzy / Levenshtein matching
- Highlight matched substring в title
- Bulk Move (target picker)
- Bulk Tag (tag picker)
- Регистр-чувствительный режим через флаг

### Needs spike

- Если корпус > 100k задач — нужен индекс. Bleve / SQLite FTS5 / Tantivy. Сейчас далеко.

## Assumptions & Open Questions

### Assumptions

- **[ASSUMPTION: substring case-insensitive matching по Title достаточно для v1 — не fuzzy, не regex, не searches Notes]**
- **[ASSUMPTION: scope выделения = активный список; смена списка (Tab/1-6) очищает selection]**
- **[ASSUMPTION: bulk-операция best-effort: один failure не прерывает остальное, агрегированное сообщение в status bar]**
- **[ASSUMPTION: backward compat критичен — пустое selected → старое поведение per-cursor]**
- **[ASSUMPTION: `[x]`/`[ ]` префикс показывается только когда selected ≠ ∅ или активен select-mode — чтобы не загромождать обычный вид]**
- **[ASSUMPTION: filter и selection — независимые состояния: можно отфильтровать, выделить видимое, очистить фильтр — selection сохраняется по ID]**

### Open Questions

1. **Что делать с выделенной задачей, если она "уходит" из фильтра?** Например: выделил, потом сузил фильтр, и задача больше не отображается. Варианты: (a) оставить в selection, но не показывать; (b) "выпадает" из selection. — Рекомендую (a): selection — это семантическое множество, не visual.
2. **Bulk-confirm для destructive операций?** Сейчас `d` (delete) выполняется без подтверждения. При bulk delete 10 задач — спросить confirm? — Рекомендую да, отдельная Y/N модалка для bulk-delete (только для delete, не для complete/cancel/pin).
3. **Подсветка match'а в title?** — Deferred в v2; добавит strings операций на каждый render.

# Исследование: TUI Todo (todushka) в стиле Things 3

## Цель

Построить терминальный todo-менеджер, повторяющий ключевую модель Things 3 (Inbox → Today → Upcoming → Anytime → Someday + Areas/Projects + теги + заметки + повторы) с упором на скоростной keyboard-driven UX и красивый, но не перегруженный интерфейс. Локальное хранилище, без сетевой синхронизации в v1, всё работает офлайн в одном бинаре.

Драйверы:

- У пользователя уже есть `go.mod` для `github.com/jtprogru/todushka` — это greenfield в Go-репозитории.
- Пользователь явно открыт к выбору языка: «можно взять и Rust».
- Текущая модель Things 3 — про *жизненный workflow*, а не просто список задач (Areas/Projects/Today-Engine с auto-scheduling по start date и deadline).

## Исследование

Кодовая база:

- `go.mod` — модуль `github.com/jtprogru/todushka`, Go `1.26.3`. Зависимостей нет.
- `.gitignore` — стандартный шаблон Go (бинари, `*.test`, `vendor/`, `*.out`).
- Других файлов нет: ни исходников, ни `Makefile`, ни тестов, ни CI.

Это полноценный greenfield. Никаких brownfield-ограничений (нет behaviour, который нужно сохранить, нет существующих тестов).

Эталон функциональности — Things 3 (по публичной документации Cultured Code):

- **Inbox** — быстрый сбор задач без классификации.
- **Today** — задачи на сегодня (с разделением Morning / Evening при желании).
- **Upcoming** — таймлайн задач со start date в будущем + календарные события (в v1 без интеграции с календарём).
- **Anytime** — пул задач без start date, но активных (с проектом или Area).
- **Someday** — отложенные «когда-нибудь».
- **Logbook** — выполненные задачи (архив).
- **Trash** — корзина.
- **Areas of Responsibility** — долгоживущие сферы (Work, Health, Home).
- **Projects** — конечные начинания со списком задач, опционально с deadline.
- **Headings** внутри проектов — группировка задач.
- **Checklist items** внутри задачи — подпункты.
- **Tags** — иерархические; одна задача может иметь несколько тегов.
- **Start date** / **Deadline** — две независимые даты.
- **Repeat** — повторяющиеся задачи (daily, weekly, every N days, on weekdays).
- **Quick Entry** — глобальное окно ввода (в терминале → быстрая модалка из любого экрана).
- **Magic Plus** — добавление задачи в произвольную позицию списка.

## Build Tooling

Будет создано в фазе implementation (сейчас файлов нет). Целевая конфигурация для рекомендуемого варианта (Go):

- **Orchestrator:** `Taskfile.yml` (go-task) — пользователь, судя по структуре `~/.claude/skills`, активный пользователь Taskfile; альтернативно `Makefile`.
- **Test:** `go test ./...`
- **Build:** `go build -o bin/todushka ./cmd/todushka`
- **Lint:** `golangci-lint run`
- **Generate:** не требуется в v1 (без proto / без моков, тестируем через интерфейсы напрямую). Если возьмём `sqlc`, будет `sqlc generate`.
- **Source:** `Taskfile.yml` в корне (будет добавлен на implementation phase).

## Options Considered

### Вариант A: Go + Bubble Tea + bbolt (рекомендуемый)

- **Описание:** TUI на [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm/MVU-архитектура) + [Bubbles](https://github.com/charmbracelet/bubbles) (готовые компоненты: textinput, list, viewport, table) + [Lipgloss](https://github.com/charmbracelet/lipgloss) для стилизации. Хранилище — embedded KV [bbolt](https://github.com/etcd-io/bbolt) (форк BoltDB), один файл `~/.local/share/todushka/db`.
- **Pros:**
  - Существующий `go.mod` — нулевая миграция языка.
  - Bubble Tea — фактически индустриальный стандарт для современных TUI на Go (gh CLI, glow, soft-serve, lazygit-подобные проекты). MVU-модель отлично ложится на todo: одно состояние, чистые редьюсеры, легко тестировать.
  - bbolt — pure-Go, без CGO, кроссплатформенная сборка тривиальна (`GOOS=windows go build` работает «из коробки»).
  - Single static binary — для пользовательского CLI это идеально.
  - У вас под рукой skills `golang-pro`, `golang-cli`, `golang-troubleshooting`, `golang-benchmark` — toolchain поддержки агента сильнее.
- **Cons:**
  - Bubble Tea «дольше» Rust-аналога на тяжёлых перерисовках (для todo это неактуально — десятки/сотни строк, не тысячи).
  - bbolt — KV, а не SQL: запросы «дай все задачи на сегодня по нескольким тегам» придётся писать вручную через индексы; для текущего scope это плюс (простота), но если позже понадобятся аналитика/отчёты — придётся либо мигрировать на SQLite, либо городить индексы.
- **Complexity:** средняя. MVU-архитектура требует дисциплины (явные `Msg` / `Cmd`), но позволяет писать unit-тесты на чистые редьюсеры без TUI.

### Вариант B: Go + Bubble Tea + SQLite (modernc.org/sqlite)

- **Описание:** То же, что A, но хранилище — SQLite через [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure-Go порт, без CGO). Опционально `sqlc` для типобезопасных запросов.
- **Pros vs A:** SQL-запросы (фильтры по start date, дедлайнам, тегам пишутся декларативно). Готовая база для будущих фич: smart lists, поиск, статистика.
- **Cons vs A:** `modernc.org/sqlite` — тяжёлый (большой size бинаря, ~10MB), холодный старт чуть медленнее. С CGO-вариантом `mattn/go-sqlite3` придётся настраивать кросс-компиляцию.
- **Complexity:** выше A на старте (миграции, схема, sqlc), но ниже A на v2-функционале (smart lists делаются SQL-запросом).

### Вариант C: Rust + ratatui + sqlx + SQLite

- **Описание:** Переписать стартовый `go.mod` на Cargo workspace, [ratatui](https://github.com/ratatui-org/ratatui) (immediate-mode), [crossterm](https://github.com/crossterm-rs/crossterm) для backend, [sqlx](https://github.com/launchbadge/sqlx) + SQLite.
- **Pros:**
  - Производительность и память лучше Go на нагрузке (для todo нерелевантно).
  - Очень мощная типизация — компилятор реально ловит ошибки в моделях.
  - Async-первый стек подходит, если в v2 захочется sync с iCloud/CalDAV.
- **Cons:**
  - Полный отказ от существующего `go.mod` (мелкое, но всё-таки).
  - У агента нет специализированных Rust-skill'ов под рукой — golang-стек поддержан гораздо лучше; это напрямую увеличит длительность фазы implementation.
  - ratatui — immediate-mode (рисуем каждый кадр), требует ручного управления state vs MVU подходом Bubble Tea; для долгоживущего todo это менее естественно.
  - Время компиляции и onboarding-кривая выше.
- **Complexity:** высокая. Особенно учитывая, что вы оставили выбор открытым, а не «давайте на Rust».

### Вариант D: Минималистичный CLI без TUI

- **Описание:** Просто CLI команды (`todushka add`, `todushka today`, `todushka list inbox`).
- **Pros:** очень просто.
- **Cons:** не отвечает запросу — нужен именно TUI как Things 3.
- **Отброшен.**

## Constraints & Risks

- **Совместимость с терминалами:** True-color ANSI работает в iTerm2/Alacritty/Kitty/WezTerm/Ghostty, но в стандартном macOS Terminal — частично. Bubble Tea / ratatui оба корректно деградируют в 256-цветный режим.
- **Хранилище и формат:** локальный файл; нужно подумать о миграциях схемы (даже для bbolt — версионирование bucket'ов). Без миграционного механизма апгрейды сломают пользовательские БД.
- **Конкурентность:** TUI-приложение в одном процессе, без многопроцессного доступа к БД. Но если пользователь запустит второй экземпляр — bbolt лочит файл, нужно отдавать читаемую ошибку.
- **Производительность:** при тысячах задач список Today/Inbox должен рендериться без лагов. Bubble Tea `list` справляется, но потребуется виртуализация (она в `bubbles/list` встроена).
- **Безопасность:** локальный файл; шифрования не предусматриваем в v1. Заметки пользователя — открытым текстом на диске (документировать это явно).
- **Backup / экспорт:** нужно решить с v1 — даже простой `todushka export --json` снижает риск потери данных и упрощает миграции.
- **Зависимости (для рекомендованного варианта A):** `bubbletea`, `bubbles`, `lipgloss`, `bbolt`, `lo` (или `samber/lo`) для утилит, `cobra` или `urfave/cli` для CLI-обвязки. Аккуратно: не тащить лишнее.
- **Repeating tasks:** требует cron-подобной логики `next_occurrence`. Лучше иметь хорошо протестированную чистую функцию.

## Recommended Direction

**Вариант A: Go + Bubble Tea + bbolt.**

Обоснование:

1. У вас уже Go-репозиторий — переход на Rust = искусственное препятствие без выгоды для текущего scope.
2. Bubble Tea + bubbles покрывают всё, что нужно для UX Things 3 (list, viewport, textinput, modal через `tea.WindowSize`), а MVU-модель упрощает unit-тестирование редьюсеров без TUI.
3. bbolt + явные индексы (bucket'ы по start_date, tag → []task_id) достаточно для v1. Если в v2 понадобятся сложные смарт-списки — миграция на SQLite не сломает контракты domain-слоя, если домен изолирован от storage (что мы заложим в design phase).
4. Single static binary без CGO → дистрибуция через `go install`, Homebrew, prebuilt релизы — тривиально.

Альтернатива — Вариант B (SQLite вместо bbolt) — если на фазе requirements выяснится, что smart lists и сложные фильтры обязательны в v1, переключаемся на B на дизайн-фазе. Архитектурно это вопрос реализации `Repository` интерфейса.

Rust (Вариант C) предлагаю отклонить *в этом проекте*, но это решение пользователя — если есть стратегическая причина (изучить Rust, синк с уже существующей Rust-кодовой базой и т.п.), переключаемся.

## Scope Boundaries

### Must-have (v1)

- Inbox / Today / Upcoming / Anytime / Someday / Logbook / Trash — все 7 системных списков.
- Areas (плоский список) и Projects (с заголовками, deadline и опциональным Area-родителем).
- Task с полями: title, notes (multiline), start date, deadline, tags (multi), checklist items (опционально, простой плоский список подзадач), state (`open` / `completed` / `cancelled`).
- Tags (плоские в v1; иерархические — v2).
- Repeat: `daily`, `every N days`, `weekly on <weekdays>`, `monthly on <day>`. Без полноценного cron-выражения.
- Quick Entry (модалка по горячей клавише из любого экрана; парсинг inline-синтаксиса: `купить хлеб #дом @today !2026-05-30`).
- Keyboard-first навигация (vim-style по умолчанию: `j/k`, `gg/G`, `/` для поиска, `:` для команд; альтернативные стрелки тоже работают).
- Локальное хранилище в `$XDG_DATA_HOME/todushka/db` (fallback `~/.local/share/todushka/db`).
- Экспорт всей БД в JSON (`todushka export`) и импорт (`todushka import`).
- Базовое цветовое оформление + тёмная тема по умолчанию, автодетект fallback на 256 цветов.
- CLI команды для скриптинга: `todushka add "..."`, `todushka today`, `todushka complete <id>` (даже без запуска TUI).

### Deferred (v2)

- Hierarchical tags (вложенные теги Things 3).
- Repeating с произвольным cron-выражением.
- Темы оформления / кастомные цвета через config.
- Глобальный fuzzy search по всему контенту с подсветкой.
- Smart Lists (пользовательские сохранённые фильтры).
- Календарная интеграция (.ics import, CalDAV).
- Sync (iCloud-style end-to-end или просто файл в Dropbox/Syncthing).
- Mobile companion / web view.
- Markdown в notes (рендеринг через glamour).
- Mouse-поддержка.
- Внутренние ссылки между задачами (`[[task-uuid]]`).

### Needs spike

- **Repeating tasks семантика:** «every weekday», «каждое 31 число месяца» — граничные случаи требуют отдельной чистой функции `NextOccurrence(rule, after time.Time) time.Time` с покрытием table-driven тестами до старта основной реализации.
- **Quick Entry inline-парсер:** синтаксис `купить хлеб #дом @today !2026-05-30` — нужен формальный грамматический разбор (regex-pipeline или go-participle). Стоит сделать минимальный prototype до фиксации API.
- **Today engine:** что именно считается «задачей на сегодня»? Things 3 сочетает: start date ≤ today + due date ≤ today + явно перемещённые в Today. Нужно зафиксировать формальные правила на requirements-фазе.

## Assumptions & Open Questions

Допущения, на которых базируется рекомендация:

- `[ASSUMPTION: пользователь работает преимущественно один — не нужен multi-user / multi-device sync в v1]`
- `[ASSUMPTION: macOS — основная целевая платформа (исходя из контекста: Things 3 — macOS-приложение); Linux считаем равноценной целью; Windows — best effort, без целенаправленного тестирования в v1]`
- `[ASSUMPTION: True-color терминал доступен у пользователя; для 16-цветных терминалов поддерживаем безопасный fallback, но не оптимизируем UI под него]`
- `[ASSUMPTION: производительность при ≤ 10 000 активных задач достаточна без специальной оптимизации (bbolt + in-memory cache индексов)]`
- `[ASSUMPTION: notes хранятся plaintext; шифрование БД не требуется в v1]`
- `[ASSUMPTION: один экземпляр приложения за раз — конкурентный доступ к файлу БД из второго процесса корректно отклоняется с понятной ошибкой]`
- `[ASSUMPTION: «как Things 3» означает воспроизвести модель и keyboard-workflow, а не пиксельно копировать визуал — TUI имеет свои выразительные средства]`

Открытые вопросы:

1. **Выбор языка:** подтвердить Go или предпочитаете Rust по нефункциональным причинам (обучение, унификация со стеком)?
2. **Объём v1 «Things 3»:** считаем ли мы Logbook (история выполненного) обязательной частью v1 или это deferred? Сейчас включён в must-have.
3. **Дистрибуция:** только `go install`, или сразу планируем Homebrew formula / GitHub Releases с prebuilt binaries в v1?
4. **Импорт из Things 3:** есть ли необходимость импортировать существующие задачи из Things 3 (через CSV/JSON export)? Это меняет приоритеты v1.
5. **Repeating tasks в v1:** какой минимум правил повторения вы реально используете в Things 3? Если только «daily» и «weekly on weekdays», можно радикально упростить.
6. **Sync в горизонте 6 месяцев:** влияет на выбор хранилища (SQLite + CRDT vs bbolt) и на чистоту domain-слоя. Готовы зафиксировать «локальное только» хотя бы на v1?
7. **Системные интеграции:** notifications на macOS (через `osascript`/`terminal-notifier`)? В Things 3 они есть для дедлайнов.

# Exploration: TUI Shell (Full-Screen + Header Counts + Status Line + Config)

## Intent

Сейчас TUI имеет несколько эргономических дыр:

1. **Незанятое пространство:** `View()` рендерит header + body + footer без явного clamping'а к `m.height`. На высоком терминале (60+ rows) после нескольких задач остаётся пустота снизу. На очень узком — line wrap может растянуть TUI за пределы видимой области.
2. **Header без подсказок:** строка `Inbox Today Upcoming Anytime Someday Logbook` не показывает (a) клавиши для прыжка (1-6), (b) количество задач в каждом списке. Пользователь должен помнить keymap и переключаться "вслепую".
3. **Footer как сплошной текст:** один длинный hint line `?: help  ⇥: next view  /: filter  ...` — не контекстный, не визуально выделяет важные клавиши, не показывает текущий "mode".
4. **Нет конфигурации:** все настройки (порог dual-pane, threshold confirm modal, max notes lines, тема) — hardcoded константы. Пользователь не может настроить под себя.

**Reference UI: zellij** — multiplexer с очень читаемым layout'ом:
- **Top:** tabs с hotkey letter, выделенной фоном (`Tab Ctrl+t`).
- **Bottom:** mode indicator (`-- NORMAL --`, `-- RESIZE --`) + контекстные key hints в формате `<key bg> action`.
- Цветовое кодирование различает (a) hotkey, (b) текст label, (c) value/status.

Цель v1: довести todushka TUI до этого уровня "discoverability" + добавить app-level config.

## Investigation

### Текущий рендер layout

- **`app.go:38` Model.width** — единственное измерение терминала. **`m.height` не отслеживается.**
- **`app.go:74-77`**: `tea.WindowSizeMsg` обновляет только `m.width`, ignore'ит `msg.Height`. Это позволяет фиксу №1 быть простым (добавить `height` field, читать `msg.Height`).
- **`app.go:458-475`** `View()`: `lipgloss.JoinVertical(viewHeader, body, viewFooter)`. Никаких `Height(N)` clamp'ов — выход растягивается естественной высотой контента, терминал рисует остальное empty.

### Текущий header

- **`app.go:446-456`** `viewHeader`: `for _, l := range allLists` → активный `theme.Header.Render(l.String())`, остальные `theme.HeaderDim.Render(l.String())`. Без digit prefix, без counts.
- Текущий вид: `Inbox Today Upcoming Anytime Someday Logbook` (с padding в каждом styled segment).
- **Активный сегмент** уже визуально выделен через theme.Header (inverse background).

### Текущий footer

- **`app.go:515-540`** `viewFooter`: один text hint + опциональный status + опциональный `Selected: N`.
- Mode-сwitch logic: editor → "Tab: field Ctrl+S: save Esc: cancel"; filter mode → "Filter: <query>_ ..."; default — общий long hint.
- Нет mode indicator (`NORMAL`/`FILTER`/`SELECT-MODE`).

### List counts source

- Service-level methods: `ListInbox`, `ListToday`, `ListUpcoming`, `ListAnytime`, `ListSomeday`, `ListLogbook` (в `internal/app/queries.go`). Каждый делает full TaskList + in-memory filter — O(N) per call.
- Для 1000 задач это ~1-2ms × 6 lists = ~10ms total. Безопасно делать на каждый `tasksLoadedMsg`.
- Альтернатива: одна `TaskList(open=true)` + локальный filter through all 6 buckets. Сложнее, но 6x быстрее.

### Существующая поддержка config

- **`internal/config/paths.go`** — только XDG dirs (`DataDir`, `StateDir`, `LogPath`). Никакого app-level config файла.
- Env vars: `TODUSHKA_THEME` (только для theme select в `style.go:96`). Никаких других.
- Hardcoded константы которые хочется configurable: `dualPaneMinWidth=100`, `listPaneShare=0.45`, `detailsNotesMaxLines=8`, `bulkConfirmThreshold=5`, `statusFadeDuration=5s`.

### Lipgloss capabilities для full-screen

- `lipgloss.NewStyle().Width(N).Height(M)` — фиксированная ширина и высота, с padding до этих размеров.
- `lipgloss.PlaceVertical(height, lipgloss.Top, content)` — выравнивание с phantom padding.
- `lipgloss.Style.MaxHeight(M)` — ограничение по высоте (truncate если overflow).

Для full-screen паттерн:
```go
headerH := lipgloss.Height(viewHeader)
footerH := lipgloss.Height(viewFooter)
bodyH := m.height - headerH - footerH
body := lipgloss.NewStyle().Height(bodyH).Render(viewBody())
return JoinVertical(viewHeader, body, viewFooter)
```

### Config-file ecosystem in Go

- **YAML:** самый идиоматичный для CLI tools (kubectl, helm, etc.). Уже есть transitive dep через `gopkg.in/yaml.v3` в Bubble Tea-зависимостях? Проверить.
- **TOML:** более строгий, проще парсится. Используется в Cargo/Hugo. Меньше распространён в Go.
- **JSON:** работает но user-unfriendly для редактирования.

Голосую за **YAML** — стандарт для todo/CLI-инструментов, легко читается.

## Build Tooling

- **Orchestrator:** Taskfile.yml
- **Test:** `task test` / `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`
- **Format:** `task fmt`
- **Source:** `Taskfile.yml`

## Options Considered

### Option A — All four pieces in one feature ("tui-shell")

Реализовать всё за один pipeline:
- m.height tracking + full-screen clamping
- Header: digit prefix + counts + colour coding
- Footer: zellij-style mode indicator + context-aware keys
- Config: `~/.config/todushka/config.yaml` + env var overrides for ~5 settings

**Pros:**
- Всё что user попросил — в одном релизе (v0.4.0).
- Cohesive — все 4 элемента касаются "shell" вокруг content.

**Cons:**
- Большой scope. ~50+ subtasks, 2-3 дня implementation.
- Если что-то блокируется (например конфиг парсинг) — задерживает всё.

**Complexity:** Large (L — 2-3 дня).

### Option B — Split в две фичи (v0.4.0 + v0.5.0)

- **v0.4.0 "tui-chrome":** full-screen + header counts + status line.
- **v0.5.0 "config-system":** YAML config + env overrides для всех hardcoded констант.

**Pros:**
- Каждый PR обозримее. Faster ship of v0.4.0.
- Config-system может потом покрыть НЕ только TUI настройки (например defaults для CLI flags).

**Cons:**
- v0.4.0 всё ещё с hardcoded thresholds. Пользователь не сможет настроить пока v0.5.0 не выйдет.
- Дублирование работы в pipeline overhead (2 explore + 2 requirements + ...).

**Complexity:** Medium × 2.

### Option C — Minimum viable shell + config-via-env-only

Только env-vars в v1 (skip YAML config file). Конфиг придёт в v2 если будет реальная потребность.

**Pros:**
- Простейшая реализация — несколько `os.Getenv` calls.
- Быстро.

**Cons:**
- Env vars неудобны для нескольких настроек одновременно. Пользователь не запомнит 5+ `TODUSHKA_X` переменных.
- Не масштабируется на per-theme custom palette.

**Complexity:** Small для config-части; остальное как Option A.

## Constraints & Risks

- **Backward compat:** существующие тесты тестируют `viewList`, `viewFooter`, `viewHeader` через прямые вызовы. Изменения в этих функциях ломают N≥20 unit-тестов. Нужна стратегия: либо обновить тесты под новый формат, либо оставить старые методы как pure-content и завернуть в shell-renderer.

- **Height pollution:** на маленьких терминалах (40x10) текущий TUI scroll'ится верх; с full-screen clamping body будет обрезаться. Нужен min-height check и fallback к scrollable list если рядов задач больше чем bodyH.

- **Count consistency:** если user удаляет task в одной list, count в header должен обновиться. Re-fetch всех 6 counts через `loadCounts` Cmd при `tasksLoadedMsg` после bulk операций — достаточно. Стоимость: 6× O(N) per refresh. На типичных N≤1000 — незаметно.

- **Config schema versioning:** если позже добавим новые ключи — нужна toleration к unknown fields (`yaml.Decoder.KnownFields(false)` — это default). Forward-compat ok.

- **Theme name validation:** если в config указан несуществующий theme name, упасть с error или fallback к default? Решение: warn в log + fallback.

- **Env var precedence:** `env > config file > defaults`. Стандартный 12-factor pattern.

- **Lipgloss border накопление:** dual-pane уже использует BorderLeft на details. С header/footer styling нужно проверить что не накапливается двойной border на pane boundaries.

- **Header overflow:** на узком терминале (60 cols) header `(1) Inbox [24]  (2) Today [5]  ...` может не поместиться. Mitigation: truncate сегменты (показывать только digit + first letter `(1)I[24]`) ИЛИ просто wrap'нуть в 2 строки — но это меняет header height. Решение для v1: wrap на 2 строки если нужно; height учитывается dynamically.

- **Test helpers:** `bareTestModel()`, `newTestModel()` придётся обновить если Model приобретает новые fields (height, config, counts). Backward-compatibility через `make(map...)` инициализации в NewModel.

## Recommended Direction

**Option A — All four pieces in one feature.**

Аргументы за:
- Эти 4 части когезивны: full-screen без новых header/footer бессмысленен; config без use-case (= hardcoded → configurable) — преждевременная абстракция; counts без зрелого header — некрасиво. Bundling даёт целостный UX-uplift в одной версии.
- Размер scope — большой, но управляемый. ~7-8 top-level tasks (примерно как dual-pane).
- Не нужно дважды проходить explore/requirements для пересекающихся вопросов (theme handling, render integration).

Аргументы против — преимущественно про "одна большая PR". Mitigation: внутренние atomic commits + pipeline tasks дают review-granularity.

### Config-format decision

**YAML** через `gopkg.in/yaml.v3`. Файл: `$XDG_CONFIG_HOME/todushka/config.yaml` (fallback `$HOME/.config/todushka/config.yaml`).

Пример:
```yaml
theme: macchiato        # macchiato | latte | mono | <custom-name>
dual_pane_min_width: 100
list_pane_share: 0.45
bulk_confirm_threshold: 5
notes_max_lines: 8
status_fade_seconds: 5
```

Env vars override file через mapping:
- `TODUSHKA_THEME` → `theme`
- `TODUSHKA_DUAL_PANE_MIN_WIDTH` → `dual_pane_min_width`
- ...

### Header design

Каждый list segment: `(N) Name [Count]` где:
- `(N)` — digit prefix, цвет accent
- `Name` — label, text color
- `[Count]` — count, subtext color
- **Активный сегмент** — все три части на accent background (inverted).
- На узком терминале (< 80 cols) — компактный mode: `(N)Initial[C]` (`(1)I[24]`).

### Footer (zellij-style)

Структура: `[MODE] | key1: action1 | key2: action2 | ...`

Modes:
- `NORMAL` (default — list view)
- `FILTER` (filter input active)
- `SELECT` (≥1 task selected)
- `CONFIRM` (confirm modal active)
- `EDITOR` (editor screen)
- `HELP` (help screen)

Per-mode key sets:
- NORMAL: `/: filter`, `space: select`, `n: quick`, `↵: edit`, `c: complete`, `?: help`, `q: quit`
- FILTER: `↵: save`, `esc: cancel`
- SELECT: `c/x/d/p: bulk`, `*: all`, `esc: clear`
- CONFIRM: `y: yes`, `any: cancel`
- EDITOR: `Tab: next field`, `Shift+Tab: prev`, `Ctrl+S: save`, `esc: cancel`
- HELP: `?: close`

Mode indicator слева — coloured chip (background = accent + mode name).

## Scope Boundaries

### Must-have (v1)

- Track `m.height` from `tea.WindowSizeMsg`.
- `View()` рендерит full-screen: `header_h + body_h + footer_h == m.height`. Body clamped через `lipgloss.NewStyle().Height(bodyH)`.
- Header: каждый segment как `(N) Name [Count]` с цветовой дифференциацией; активный — inverted bg.
- List counts: cached в Model (`map[listKind]int`); populate'ятся при `tasksLoadedMsg` через batch fetch (6 ListXxx calls в одном Cmd).
- Footer: mode chip + context-aware key hints, разделённые `│` или space + colour.
- Modes: NORMAL / FILTER / SELECT / CONFIRM / EDITOR / HELP — детектируются из Model state без новых полей.
- Config file: `$XDG_CONFIG_HOME/todushka/config.yaml`, парсится через `yaml.Unmarshal`. Ошибки декодирования → warn в log, продолжить с defaults.
- Env var overrides для 5 settings: `theme`, `dual_pane_min_width`, `list_pane_share`, `bulk_confirm_threshold`, `notes_max_lines`.
- Config injection в Model: новое поле `Model.config Config`. NewModel принимает `cfg Config` argument.
- На узком терминале (< 80 cols) — header компактный mode: digit + initial letter + count.
- На маленькой высоте (< 10 rows) — fallback к non-clamped legacy rendering (визуально лучше чем cut-off content).

### Deferred (v2)

- Custom theme palette через config (custom colors).
- Per-keymap rebinding через config.
- Theme hot-reload без рестарта.
- TOML/JSON config alternative formats.
- Window resize debouncing (если flicker заметен).

### Needs spike

- Если выяснится что `gopkg.in/yaml.v3` не подключён транзитивно — добавить direct dep.
- Если lipgloss `Height(bodyH)` ломает confirm/quick-entry stacking — придётся обновить modal layout.

## Assumptions & Open Questions

### Assumptions

- **[ASSUMPTION: YAML — наиболее идиоматичный config-формат для CLI-инструментов в Go ecosystem]**
- **[ASSUMPTION: env > config > defaults — стандартный 12-factor precedence]**
- **[ASSUMPTION: 6 ListXxx calls в одном Cmd для count refresh — приемлемо по perf на N≤1000 tasks]**
- **[ASSUMPTION: на терминалах < 80×10 — fallback к legacy без full-screen clamp; пользователь скорее всего использует широкий window]**
- **[ASSUMPTION: header compact mode (`(1)I[24]`) — допустимый trade-off на 60-79 cols; <60 cols — fall to multiline wrap]**

### Open Questions

1. **Config path:** `$XDG_CONFIG_HOME/todushka/config.yaml` (стандарт) или `$HOME/.todushka.yaml` (проще для пользователей)? **Рекомендую XDG.**
2. **Counts refresh strategy:** на каждый `tasksLoadedMsg` (после load list) или только после bulk операций? **Рекомендую: при tasksLoadedMsg — гарантирует свежесть, дёшево.**
3. **Empty list count:** показывать `[0]` или скрывать? **Рекомендую: показывать `[0]`** — visual consistency.
4. **Mode chip background:** один цвет на все modes (accent) или per-mode разный? **Рекомендую один accent для v1, разные — в v2.**
5. **Header compact threshold:** 80 cols (стандарт)? 60? **Рекомендую 80.**
6. **Если config файл не существует — создавать с defaults или просто использовать defaults?** **Рекомендую: НЕ создавать, тихо использовать defaults.** (Меньше неожиданных файлов на disk.)
7. **Поддержать ли `--config <path>` CLI флаг для override config file location?** **Рекомендую да — стандарт для CLI tools, тривиально через cobra `PersistentFlags`.**

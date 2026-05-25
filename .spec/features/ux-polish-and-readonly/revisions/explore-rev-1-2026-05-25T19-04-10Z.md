# Exploration: UX Polish, Read-Only Mode, and Editor Improvements

## Intent

Пять разнокалиберных улучшений, накопившихся из пользовательского опыта v0.4.0:

1. **`theme: auto/system`** — приложение должно автоматически следовать OS-настройке (dark/light) вместо явного выбора пользователем.
2. **Refresh после редактирования** — после `Ctrl+S` в editor'е дисплей не обновляется немедленно; пользователь должен нажать Tab для refresh. Это bug или perceived bug.
3. **Anytime toggle в editor** — `Someday` есть, симметричного `Anytime` нет. Пользователь не может явно поставить задачу в Anytime-bucket.
4. **Read-only мониторинг при second instance** — сейчас вторая копия падает с `storage: database is locked`. Нужно либо явное сообщение, либо открывать read-only.
5. **Visual section borders** — заголовки разделов (Inbox, Today, …) слишком близко к остальному UI; нужны явные границы между header / body / footer / panes.

## Investigation

### 1. Theme auto-detection

Сейчас `internal/tui/style.go:SelectTheme(env)`:
- `NO_COLOR` → monochrome
- `TODUSHKA_THEME=light|latte` → Latte (light)
- Default → Macchiato (dark)

После v0.4.0 config layer добавил `cfg.Theme` со значениями `macchiato | latte | mono | <empty=fallback>`. Нет автодетекта.

**OS dark/light detection:**
- **macOS:** `defaults read -g AppleInterfaceStyle` → `"Dark"` если dark, error (`exit 1`) если light. Shell-out, ~20-100ms.
- **Linux:** разнообразие. `gsettings get org.gnome.desktop.interface color-scheme` для GNOME. `kreadconfig5 --group "General" --key "ColorScheme"` для KDE. `$GTK_THEME` env hint. Не универсально.
- **Windows:** registry `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize\AppsUseLightTheme`. Не наш приоритет.

Cross-platform Go packages:
- `github.com/Codehardt/go-darkmode` — macOS-only, archived.
- `github.com/diamondburned/gotk4/pkg/gtk/v4` — too heavy.
- **Манчaaal approach:** платформо-специфичные `// +build darwin` файлы с шеллауту.

**Verdict:** Implement `theme: auto` via platform-specific code:
- `internal/tui/detect_dark_darwin.go` (macOS shell-out).
- `internal/tui/detect_dark_linux.go` (gsettings + fallback к env var).
- `internal/tui/detect_dark_other.go` (default to dark).

### 2. UI refresh after edit

Текущий flow (`app.go:157-159`):
```go
case editorSavedMsg:
    m.screen = screenList
    return m, m.loadCurrentList()
```

`loadCurrentList` Cmd:
1. Calls `svc.List<ActiveList>(ctx)`.
2. Returns `tasksLoadedMsg{tasks}`.
3. Update for `tasksLoadedMsg` updates `m.tasks` + fires `tea.Batch(fetchNameCache, fetchListCounts)`.

**Theory:** Это ASYNC. Между `editorSavedMsg` (screen=list) и `tasksLoadedMsg` (tasks updated) — небольшое окно. На быстрых машинах ~ms; пользователь не должен заметить. Но если bbolt ListInbox медленный (>100ms), View рендерится с устаревшими `m.tasks`.

**Возможные причины user-perceived stale:**
- a) Курсор остаётся на той же позиции в списке; задача не изменила позицию (sort by CreatedAt стабилен). Пользователь визуально не видит "что-то изменилось". Edit titulли уже отображается.
- b) Tag names cache. Если изменили tags, они резолвятся через nameCacheLoadedMsg (через async Cmd после tasksLoadedMsg). До этого — short-ID fallback в details pane.
- c) Counts. Если перенесли задачу в другой bucket (теоретически — Someday/Anytime), counts в header показывают старые значения до countsLoadedMsg.
- d) Dual-pane details. После save, cursor на той же позиции — детали показывают обновлённую задачу. Но если у пользователя cursor был "выше" нужной задачи, edit пройдёт но details покажут другую задачу.

**Mitigation (best UX):** При получении `editorSavedMsg{updated: t}`:
- Найти в `m.tasks` задачу с `t.ID` и заменить inline (мгновенный visual update).
- Параллельно fire `loadCurrentList` для refresh sort / counts / cache.

Это даёт мгновенный визуальный отклик независимо от async Cmd latency.

### 3. Anytime toggle в editor

Сейчас editor (`internal/tui/editor.go:218-226`) имеет Someday checkbox (Space toggles). Полей Area/Project/Anytime нет.

Anytime — derived bucket (queries.go `ListAnytime`):
```
t.AreaID != nil OR t.ProjectID != nil  (has area or project)
AND NOT t.Someday
AND (t.StartDate == nil OR t.StartDate <= today)
```

Так что "Anytime" — это вычисляемое состояние. Нет `task.Anytime bool`.

**Options:**

a. **Add `task.Anytime bool` domain field.** Breaking schema change. Слишком много для этой проблемы.

b. **Replace `Someday` checkbox with 2-state `When` toggle:** `[Anytime] [Someday]`. Space cycles between Anytime↔Someday.
   - Internally: Anytime selected → `task.Someday = false`; Someday selected → `task.Someday = true`.
   - Note: если у задачи нет Area/Project, "Anytime" toggle переведёт её в Inbox (по queries.go logic). Показать hint.

c. **Tri-state `When` picker:** `[Inbox] [Anytime] [Someday]`. Если task имеет Area/Project — показывается `Anytime`; если нет — `Inbox`. Someday — это override.
   - Сложнее: что-то надо делать с Area/Project в editor'е (его сейчас нет).

**Verdict:** Option (b). Symmetric UI rename из Someday в `When: [Anytime] [Someday]`. Простая семантика. Helper text если applicable.

### 4. Read-only mode для second instance

bbolt mechanism:
- `bolt.Open(path, mode, options)` с `options.ReadOnly = true` — shared lock, multiple readers OK, no writer.
- Без `ReadOnly` — exclusive lock; conflict → `bolterrors.ErrTimeout` after `Timeout`.

Текущее: `bbolt/bbolt.go:59` — `bolt.Open(path, 0o600, &bolt.Options{Timeout: openTimeout})`. На lock conflict возвращает `storage.ErrDatabaseLocked`.

**Implementation:**

1. **Storage layer:**
   - `bbolt.Open` пробует write mode; на `ErrTimeout` → `bbolt.OpenReadOnly(path)` пробует ReadOnly.
   - `Repo` хранит `readOnly bool` field; expose via `Repo.ReadOnly() bool`.
   - Все write methods (`TaskCreate`/`TaskUpdate`/etc.) check `r.readOnly` → return `storage.ErrReadOnly`.

2. **Repository interface:**
   - Add `ReadOnly() bool` method.
   - New error `storage.ErrReadOnly`.

3. **Service layer:**
   - No changes; service просто пробрасывает storage errors.

4. **TUI:**
   - При construction Model: detect через `svc.Repo().ReadOnly()`.
   - Mode chip extended: when read-only, show `-- READ-ONLY --` prefix или indicator.
   - Block write keymaps (`c`, `x`, `d`, `p`, `n`, `Enter` в смысле editor save, `Space` selection toggle — нет, selection это in-memory).
   - Actually: selection allowed (in-memory only); bulk operations blocked; editor open allowed but save → error message.

**Alternative simpler approach:** просто более явное error message при lock conflict ("Another todushka instance is running. Close it first, or run with `--readonly` to monitor."). Этот fallback proще, но user wants the actual read-only mode.

**Verdict:** Implement full read-only mode. Большой scope. См. Recommendations о split.

### 5. Visual section borders

Сейчас (`app.go:View`) — `lipgloss.JoinVertical(header, body, footer)`. Никаких visual separators между разделами. Header — это просто строка с inverted active segment и regular other segments. Body и footer прижаты к header без бордера.

**Options:**

a. **lipgloss `Border` на каждом блоке.** Может выглядеть громоздко на 6 segments header'а.

b. **Single horizontal rule** `────────────` между header/body и body/footer.

c. **Background tint** для header/footer (отличный от body bg).

d. **Padding/spacing.** Добавить пустую строку между header и body, и между body и footer. Простейшее.

Combining (b) + (d) даёт zellij-style: chip + separator. Уже у нас есть chip; добавляется ── separator под header'ом и над footer'ом.

**Verdict:** Add `theme.Separator` style + render `lipgloss.PlaceHorizontal(width, lipgloss.Left, strings.Repeat("─", width))` between sections. Subtle accent color.

### Существующая архитектура — что задеваем

- `internal/tui/style.go` — Theme + SelectTheme. Add: `auto` resolution.
- `internal/tui/editor.go` — replace Someday checkbox с When toggle.
- `internal/tui/app.go` — editorSavedMsg refresh logic; View borders.
- `internal/tui/shell.go` — mode chip extension if readonly.
- `internal/storage/repository.go` — Repository interface +ReadOnly(); +ErrReadOnly.
- `internal/storage/bbolt/bbolt.go` — Open fallback to read-only.
- `internal/storage/fakes/repository.go` — implement ReadOnly() bool.
- `internal/config/app.go` — `theme: auto` allowed value; validate.
- `cmd/todushka/main.go` — wire detect_dark + readonly.

## Build Tooling

- **Orchestrator:** Taskfile.yml — unchanged.
- **Test:** `task test` / `task test-race`.
- **Build:** `task build`.
- **Lint:** `task lint`.

## Options Considered

### Option A — Bundle all 5 в одну фичу

**Pros:**
- Cohesive UX-polish release (v0.5.0).
- Один pipeline вместо двух.

**Cons:**
- **Большой scope:** ~10-12 tasks, 70+ subtasks, 2-3 дня implementation.
- Read-only mode (#4) — самый большой; задерживает остальное.

**Complexity:** Large.

### Option B — Split на 2 фичи

- **v0.5.0 `ux-polish`:** items 1, 2, 3, 5 (theme auto, refresh, anytime, borders). Medium scope.
- **v0.6.0 `readonly-mode`:** item 4. Medium-Large scope (storage + repo + TUI).

**Pros:**
- Каждый PR обозримее.
- v0.5.0 ships faster.
- Read-only as standalone — отдельная атомарная единица для review.

**Cons:**
- 2 pipeline runs vs 1.

**Complexity:** Medium × 2.

### Option C — Только items 1, 2, 3, 5; #4 fallback к error message

Вместо полноценного read-only — просто **clearer error message** на lock conflict:
```
Error: another todushka instance is running on the same database.
       Close it before launching a second copy.
       (Read-only monitoring mode is not yet implemented — see v0.6.0.)
```

**Pros:**
- Малый scope.
- Решает immediate pain (явное сообщение).

**Cons:**
- Не даёт read-only мониторинг, которое user явно попросил.

**Complexity:** Small.

## Constraints & Risks

- **Theme auto detection performance:** macOS shell-out — 20-100ms на старте. На каждом запуске. Кешировать в config (sticky after first detect)? Или просто принять latency? Решение: один-shot на startup, без cache.

- **Theme auto fallback chain:** `cfg.Theme == "auto"` → run detector → if detector fails (e.g., Linux no gsettings, Windows) → fall back к `dark`. Без error.

- **Refresh inline splice:** `editorSavedMsg{updated: t}` — найти `t.ID` в `m.tasks` и заменить. Если задача не в текущем active list (например, перенесли в Someday) — она "пропадёт" из текущего списка. Но это нормальный UX — пользователь видит что задача больше не здесь. И `loadCurrentList` всё равно последует.

- **Anytime + no Area/Project:** Пользователь снимает Someday, у задачи нет area/project — она появляется в Inbox, не Anytime. Это правильно по semantics queries, но может удивить. Helper hint в editor: "Anytime требует Area/Project (currently: Inbox)".

- **Read-only TUI indicator:** в footer mode chip как `[RO] -- NORMAL --` или отдельный chip. И полный re-disable write keys.

- **Read-only error feedback:** если пользователь нажмёт `c` (complete) в read-only mode — нужно сообщить через status bar: "Read-only mode: writes disabled".

- **Borders взаимодействие:** existing dual-pane уже использует `║` border между pane'ами. New header/footer separators — должны быть совместимы. Lipgloss border'ы могут накапливаться.

- **Backward compat tests:** existing 174 тестов могут ломаться по двум причинам:
  - (a) viewHeader/viewFooter format changes (new separators) — нужно обновить assertion'ы.
  - (b) editor someday → "When" change — `TestTUI_EditorTabCyclesFields` может потребовать обновления (field order/count).
  - (c) ReadOnly — Repository interface change ломает fakes/bbolt — нужно адаптировать.

## Recommended Direction

**Option B — Split на 2 фичи.**

- **v0.5.0 `ux-polish`** (items 1, 2, 3, 5): theme auto, refresh after edit, anytime toggle, section borders.
- **v0.6.0 `readonly-mode`** (item 4): full read-only mode с TUI indicator.

Аргументы:
- Read-only mode — большой scope с storage interface changes; review его отдельно даёт чище signals.
- Items 1-3, 5 — все TUI-only, могут разрабатываться cohesively.
- Пользователь получает 4/5 фиксов быстрее, full read-only — следующим релизом.

Если предпочитают Option A (один bundle) — реализую с явным scope ~10 tasks.

## Scope Boundaries (для Option B v0.5.0 `ux-polish`)

### Must-have (v0.5.0)

- **Theme auto:** `cfg.Theme == "auto"` или `"system"` валидно. Platform-specific detector. Fallback к `dark` при недоступности.
- **Refresh after edit:** `editorSavedMsg` inline-splices updated task в `m.tasks` (immediate UI update) + параллельный `loadCurrentList` + `fetchListCounts` для refresh sort/counts.
- **Anytime in editor:** Замена Someday checkbox на `When: [Anytime] [Someday]` 2-state toggle. Internal mapping: Anytime → Someday=false, Someday → Someday=true. Inline hint про area/project requirement.
- **Section borders:** Horizontal `────` separator между header/body и body/footer; subtle accent color через `theme.Help` foreground.
- **Backward compat:** все 174 existing tests passing (с adjustment'ами на assertion text для header/footer).

### Deferred (v0.6.0)

- **Read-only mode:** полноценная реализация (см. отдельную фичу).

### Needs spike

- **Linux dark-mode detection portability** — gsettings/kreadconfig/env vars. Maybe just check env `TODUSHKA_DARK_MODE` или fallback к dark.

## Assumptions & Open Questions

### Assumptions

- **[ASSUMPTION: Bundle items 1, 2, 3, 5 в v0.5.0; item 4 как отдельная v0.6.0]**
- **[ASSUMPTION: macOS — primary target для theme auto-detect; Linux best-effort через gsettings; Windows — fallback dark]**
- **[ASSUMPTION: editor refresh issue реален; inline splice + async reload — комбинированный fix]**
- **[ASSUMPTION: Anytime requires Area/Project — visible hint в editor сообщает об этом]**
- **[ASSUMPTION: Horizontal `─` separator между sections достаточен; vertical borders на dual-pane уже есть]**

### Open Questions

1. **Bundle vs split?** Recommend split (Option B). User decides.
2. **Theme value:** `auto` или `system`? Оба или один? **Рекомендую: оба алиаса для одного поведения.**
3. **Read-only TUI flow при write attempt:** silent ignore vs status bar message vs modal? **Рекомендую status bar message.** (defer to v0.6.0)
4. **Editor When toggle: какие labels?** `[Anytime] [Someday]` или `When: Anytime|Someday`? **Рекомендую: чекбокс-стиль с двумя строками: `[•] Anytime / [ ] Someday` и `[ ] Anytime / [•] Someday`, Space переключает.**
5. **Section separator: жирный `━` или тонкий `─`?** **Рекомендую тонкий `─`** — менее назойливо.
6. **Refresh inline splice fallback:** если задача перенесена в другой bucket (например, теперь Someday) — оставить inline AND let next loadCurrentList убрать её? Да, естественная UX.

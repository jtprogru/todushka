# Exploration: Editor "When" Context-Aware Label + Area/Project Picker

## Intent

В v0.5.0 editor показывает `[•] Anytime / [ ] Someday` toggle, но это вводит в заблуждение когда задача без Area/Project: пользователь видит "Anytime", а задача висит в Inbox. Inline-hint `(will appear in Inbox without Area/Project)` объяснил *почему*, но **label всё ещё врёт**.

Решение состоит из двух частей:

- **Part A — Context-aware label (v0.6.0 quick fix):** Editor показывает `[•] Inbox` или `[•] Anytime` в зависимости от наличия `AreaID`/`ProjectID`. Hint удаляется (label теперь честен).
- **Part B — Area/Project picker (v0.6.0 feature):** Editor получает поля Area и Project. Пользователь может сам переместить задачу из Inbox в Anytime через назначение Area.

Оба объединены в одну фичу, так как Part A — это переходное состояние без Part B (пользователь видит Inbox label, но не может изменить bucket из editor'а).

## Investigation

### Текущий editor (`internal/tui/editor.go`)

- **`EditorModel` struct** (lines 32-44): поля `title`, `notes`, `start`, `deadline`, `tags`, `when shellEditorWhen`, `focus editorField`, `err string`. **Нет полей Area/Project.**
- **`editorField` enum** (lines 26-31): `fieldTitle`, `fieldNotes`, `fieldStart`, `fieldDeadline`, `fieldTags`, `fieldWhen`, `fieldCount=6`. После Part B `fieldCount` станет 8 (+ Area, + Project).
- **`shellEditorWhen` enum** (lines 18-23): `whenAnytime`, `whenSomeday`. Может расшириться `whenInbox` или label стать context-aware.
- **`ApplyAndSave`** (lines 153-200): обновляет `t.Someday`, `t.StartDate`, `t.Deadline`, `t.Tags`. AreaID/ProjectID НЕ меняются (используется `m.original.AreaID` неявно — но т.к. в editor нет полей, оригинальный AreaID сохраняется).

### Service / domain capabilities

- **`task.Task`** (lines 67-69): уже имеет `AreaID *id.ID`, `ProjectID *id.ID`, `HeadingID *id.ID`. Все опциональные.
- **`Service.ListAreas(ctx)`** (queries.go:138-150): возвращает все области, отсортированные.
- **`Service.ListProjects(ctx, areaID *id.ID)`** (queries.go:152-154): возвращает проекты, опционально фильтрованные по area.
- **`Service.MoveTask(ctx, taskID, target MoveTarget)`** (service.go): уже позволяет переместить задачу. Но в editor мы используем `EditTask` который читает full task — достаточно установить `t.AreaID = ...; svc.EditTask(...)`.

### Anytime semantics

`queries.ListAnytime`:
```go
if t.Someday { continue }            // someday excludes
if t.AreaID == nil && t.ProjectID == nil { continue }  // need area or project
if t.StartDate != nil && t.StartDate.Time.After(todayStart) { continue }  // future-start excludes
```

Inbox:
```go
if t.AreaID != nil || t.ProjectID != nil { continue }  // having area/project excludes
if t.Someday { continue }
if t.StartDate != nil { continue }
```

**Implication for Part A label:** "Inbox" если `(AreaID == nil AND ProjectID == nil AND !Someday)`; "Anytime" если `(AreaID != nil OR ProjectID != nil) AND !Someday`; "Someday" если `Someday`.

### Existing TUI infrastructure

- `Model.areaNamesByID`, `projectNamesByID`, `headingNamesByID` — name caches (from dual-pane feature). Уже populated в `tasksLoadedMsg` через `fetchNameCache`. Используются в details pane.
- `details.go:viewDetails` уже отображает `Area: <name>`, `Project: <name>` через resolveName из caches.
- Bubble Tea pattern для pickers: либо textinput с autocomplete (как для tags), либо отдельный modal screen (как для editor). Чистый picker = новый screen `screenAreaPicker` + `screenProjectPicker`.

### Что нужно для Part B

Сложности:
1. **Список areas/projects:** загружать при открытии editor (через `Service.ListAreas`/`ListProjects`).
2. **UI выбора:** options:
   - Textinput с autocomplete по именам (наименьший change).
   - Modal screen со списком + cursor (как для list switching).
   - Inline radio expandable (mini picker внутри editor).
3. **Project depends on Area:** если выбран Area X, projects должны фильтроваться по X. Или показывать "all projects" с пометкой area name.
4. **Clearing:** убрать area/project (move to Inbox).
5. **Heading:** на v1 пропустить — Heading требует picker внутри project. Defer to v2.

## Build Tooling

Unchanged: `task test`, `task test-race`, `task build`, `task lint`, `task fmt`.

## Options Considered

### Option A1 — Part A only (quick label fix; v0.5.1 patch)

Только context-aware label. `Area/Project picker` отдельной фичей позже.

**Pros:**
- Минимальный scope (~30 min implementation).
- Решает иммедиативный "label врёт" pain.
- Не блокирует другие фичи.

**Cons:**
- Не решает root cause: пользователь видит "Inbox" в editor, но не может переместить.
- Семантически "застрял" — у пользователя нет UI инструмента вне editor'а тоже (CLI имеет `area add`, `project add`, но привязка задачи к area через CLI отдельная команда).

### Option C — Both в одной фиче (v0.6.0; user-selected)

**Pros:**
- Cohesive UX update — label честен AND пользователь может изменить.
- Один release, одна история коммитов.

**Cons:**
- Больше scope для одной фичи (~3-4 часа vs 30 min).

**Complexity:** Medium-Large (M-L).

### Option D — Part B with simpler picker (textinput autocomplete)

Вместо modal screen — textinput field в editor (рядом с Tags). User набирает area name; on save, `UpsertArea` или существующий ID matches.

**Pros:**
- Не требует нового screen.
- Consistent с tags pattern.

**Cons:**
- Нет visual list — пользователь должен помнить areas/projects.

### Option E — Part B with modal picker screen

Tab to field → Enter → opens area picker → arrow up/down → Enter selects → back to editor. Same for project.

**Pros:**
- Visual list — discovery.
- Чище separation of concerns.

**Cons:**
- Больше moving parts (2 new screens, state).

## Recommended Direction

**Option C (Part A + Part B) bundled в v0.6.0. Picker через Option D (textinput с simple typing).**

Аргументы:
- User explicit asked Option C.
- Option D (textinput) — proven pattern (tags уже так), меньше scope.
- Part A — 30 min внутри Part B (один и тот же editor.go).
- Inline editor пользователь уже привык; modal screens — overkill для скромного количества areas (типично <20) и projects (<100).

### Конкретный план

**Part A — Context-aware "When" label:**
- В `editor.go View()`: compute `whenLabel` based on `(m.original.AreaID, m.original.ProjectID)`:
  - Если оба nil → `"Inbox"`
  - Иначе → `"Anytime"`
- Radio show: `[•] <whenLabel> / [ ] Someday`
- Hint удаляется (label теперь честен).
- Internal mapping unchanged: `whenAnytime → Someday=false; whenSomeday → Someday=true`.
- **BUT** при назначении Area/Project через Part B picker — задача автоматически становится Anytime (label обновится при следующем re-render).

**Part B — Area/Project picker (textinput-style):**
- Добавить 2 новых textinput field в `EditorModel`: `area`, `project`.
- Pre-fill: если task has AreaID, lookup name via `m.original.AreaID → svc.AreaGet → name`. Same для project.
- Editor поле `fieldArea`, `fieldProject` — добавить в editorField enum (между fieldTags и fieldWhen). `fieldCount` становится 8.
- `ApplyAndSave`: парсит values:
  - Empty string → clear `t.AreaID = nil` / `t.ProjectID = nil`.
  - Non-empty → lookup существующий area/project by name (`Repo.AreaFindByNormalized`/`ProjectFindByName`). Если existed → use ID. Если нет → return validation error (рекомендуем CLI command для creation, since editor isn't best place to create taxonomies).
- View: добавить 2 new field blocks между Tags и When sections.

### Подробности

- **Area picker:** Tab → fieldArea → user types name → ApplyAndSave validates.
- **Project picker:** Tab → fieldProject → user types name → ApplyAndSave validates.
- **Heading picker:** Defer to v2 (требует знания project ID и его headings).
- **Tab cycle:** Title → Notes → Start → Deadline → Tags → Area → Project → When (fieldCount=8).

## Scope Boundaries

### Must-have (v1 of this feature)

- **Part A:** Context-aware label `[•] Inbox / [ ] Someday` or `[•] Anytime / [ ] Someday` based on AreaID/ProjectID at editor open time.
- **Part A:** Hint `(will appear in Inbox without Area/Project)` удалён.
- **Part B:** New textinput fields `fieldArea`, `fieldProject` в editor (между Tags и When).
- **Part B:** `ApplyAndSave` парсит area/project names, lookup existing IDs, или return error если name не найден.
- **Part B:** Empty area/project = clear field (move to Inbox).
- **Part B:** Editor `fieldCount = 8` (title, notes, start, deadline, tags, area, project, when).
- **Backward compat:** existing 213 tests passing.

### Deferred (v0.7+)

- **Modal picker screen** (Option E): scrollable list of areas/projects with arrow nav.
- **Heading picker:** depend на project ID, scrolling sub-list.
- **Inline create:** "if area X не существует, создать?" prompt.
- **Tag autocomplete:** existing tags suggestion (similar pattern, separately).

### Needs spike

- **Name uniqueness:** `Repo.AreaFindByNormalized` returns single area or none. `Repo.ProjectFindByName` — returns slice (one or more). Что делать при дубликатах project name? Defer: при `len(matches) > 1`, return error "ambiguous project name, use CLI". Or pick first.

## Constraints & Risks

- **Editor field count change:** `fieldCount` going from 6 → 8 ломает `TestEditor_FieldCountIsSix`. Тест нужно обновить или удалить.

- **Tab navigation tests:** `TestTUI_EditorTabCyclesFields` cycles fieldTitle → fieldNotes → ... → fieldWhen. С новыми fields порядок изменится. Test should still pass IF assertion is about title-to-notes (which it is per current code).

- **textinput для area/project names:** Чувствительность регистра, normalization. `Repo.AreaFindByNormalized` уже handles lowercase comparison. Project — `ProjectFindByName` — нужно проверить семантику.

- **Editor open performance:** Load area/project names via `Repo` calls at open. Latency unmeasurable for typical sizes (< 100 entries).

- **Save errors:** if user types non-existent area, error "area 'foo' not found" должен быть understandable. Editor uses `m.err` field to display errors — pattern уже есть.

- **Clear vs unchanged:** user opens editor, sees pre-filled area, leaves alone → AreaID unchanged. User clears field → AreaID=nil. Differentiating "untouched" vs "explicitly cleared" — needed?
  - If pre-filled value == empty string after Trim → clear.
  - If pre-filled value matches existing name → look up ID.
  - Edge case: user changes area name in field but type-typo — error.

## Recommended Direction Recap

**Bundle Part A + Part B в v0.6.0** через Option D (textinput-based picker). 2 new fields в editor + context-aware label + simple name-lookup at save.

## Open Questions

1. **Project name ambiguity:** при `len(ProjectFindByName(name)) > 1` — error (recommend) или first-match?
2. **Field order in editor:** `Tags → Area → Project → When` vs `Area → Project → Tags → When`? Эстетически: Area/Project — мета (контекст), Tags — атрибут. **Recommend: Area → Project → Tags → When.** Изменяется существующий Tab cycle (но не сломает существующие тесты, которые проверяют только Title → Notes).
3. **Heading picker:** Defer to v0.7? **Recommend: yes.**
4. **Auto-redirect to Anytime on Area save:** если user добавил Area в Inbox задачу, label моментально становится Anytime после save? **Recommend: yes** (editor closes на save, next time opened — Area present, label = Anytime).
5. **Read-only mode (v0.6.0 original target):** в этом цикле скипаем — пушим в v0.7? **Recommend: yes** (Area picker — quick win; read-only — отдельная архитектурная задача).

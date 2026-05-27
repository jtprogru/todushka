# Exploration: Details Pane Redesign

## Intent

Переработка правой панели подробностей (`viewDetails` в `internal/tui/details.go`). Бэклог содержит три пункта:

- **BL-1.1** — лейблы полей (`Start:`/`Due:`/`Area:`/`Project:`/`Heading:`/`Tags:`/`Pinned:`/`Status:`) сейчас рендерятся как обычный текст. Хочется явное визуальное выделение: жирный + цветной label, и пустая строка между полями для воздуха.
- **BL-2** — текущее распределение пространства: `list_pane_share = 0.45` → details занимает ~55% ширины. Хочется details ≤ 40% (list ≥ 60%).
- **BL-6** — расширить инфо о проекте: сейчас показывается только `Project: <name>`. Бэклог явно отмечает, что нужен короткий explore: «возможно, что-то ещё стоит туда добавить».

Это **brownfield**-изменение: `viewDetails` уже существует, нужно не сломать существующее поведение (тесты `TestViewDetails_*` в `internal/tui/details_test.go`), но переписать стилизацию и шире отобразить project info.

## Investigation

### Текущее состояние

`internal/tui/details.go:140-181` — `viewDetails(m Model, width int) string`:

```
Title (theme.Title, bold accent)
Status: <Open|Completed|Cancelled>
""                                       ← пустая строка ДО notes (если есть)
<Notes wrapped>
Start:  <YYYY-MM-DD>
Due:    <YYYY-MM-DD>
Pinned: <YYYY-MM-DD>
Area:    <name>
Project: <name>
Heading: <name>
Tags: <comma-separated>
Someday                                  ← dim, если задача Someday
```

Лейблы — это сырые строки с фиксированной шириной отступа (`"Start:  "`, `"Due:    "` — выровнены пробелами). Цвет и bold к лейблам не применяются. Значения тоже без стиля.

Пустая строка добавляется только перед notes-блоком (line 149). Между остальными полями вертикальной воздушности нет.

### Конфиг и ширина

`internal/config/app.go:18`:
- `DualPaneMinWidth = 100`
- `ListPaneShare = 0.45`

`internal/tui/details.go:30-37` — `paneWidths(m)`:
```go
list := int(float64(m.width-1) * m.config.ListPaneShare)
details := m.width - 1 - list
```

При `ListPaneShare = 0.45` и `m.width = 100` → list = 44, details = 55. То есть details = 55%, list = 45%. (имена переменных в коде "ListPaneShare" — это **доля list pane**, не details.) Чтобы details ≤ 40%, нужно `ListPaneShare ≥ 0.60`.

### Project domain

`internal/domain/project/project.go:21-34` — Project содержит:
- `Name`, `Notes`, `AreaID`, `Deadline` (`*task.Date`), `Status`, `AutoClose`, `Position`, timestamps.

Сейчас в `fetchNameCache` (`details.go:48-89`) для проектов извлекается **только `Name`**. Полная сущность не загружается. Чтобы показать project info шире (notes/deadline/status) — нужно изменить fetching: либо хранить полные `Project` объекты в кэше (новое поле `projectsByID map[id.ID]project.Project`), либо отдельный fetch при позиционировании на task с проектом.

`internal/storage/repository.go:69` — `ProjectGet(ctx, id) (project.Project, error)` есть, легко вызвать.

### Heading

NOTE из `details.go:44-47`: HeadingGet нет, только `HeadingList(ctx, projectID)`. Имя heading через short-ID fallback. Можно либо добавить HeadingGet в Repository, либо игнорировать в этом feature (heading info не первая ценность).

### Tests существующие

`internal/tui/details_test.go` содержит ~20 тестов на `viewDetails` (status, notes, dates, area, project, tags, someday, short-ID fallback, narrow width). Все они проверяют:
- наличие подстрок (`"Status:"`, `"Start:"`, `"Due:"`, etc.) — **survive** при добавлении стилизации, потому что lipgloss обёртки оставляют исходный текст внутри ANSI escapes.
- НО: если мы перепишем лейбл с пробельным выравниванием (`"Start:  "` → `"Start:"`), тесты с `Contains` всё ещё пройдут. То есть стилевая обёртка совместима с тестами.

`internal/tui/details_test.go:573` — `setupRapidModel` уже есть для PBT.

`internal/config/app_test.go:14` — закрепляет дефолт `ListPaneShare: 0.45`. Если меняем дефолт — этот тест надо обновить.

## Build Tooling

- **Orchestrator:** Taskfile.yml
- **Test:** `task test`, `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`

## Options Considered

### Option A: Стилизация через `theme.Label` + ручное расположение полей

`Theme` уже содержит стиль `Label = lipgloss.NewStyle().Bold(true).Foreground(t.Subtext)` (`style.go:143`). Использовать его (или ввести новый, более акцентный — `LabelStrong = Bold + Foreground(Accent)`) для лейблов. Пустую строку вставлять программно между полями.

- **Pros:** минимальные изменения архитектуры; стиль централизован в `Theme`; легко тестируется ANSI-substring проверкой.
- **Cons:** надо договориться о палитре (использовать Subtext / Accent / отдельный цвет). Каждое условие `if t.X != nil` нужно дополнить вставкой пустой строки — лёгкая мультипликация кода.

### Option B: Декларативный список полей + унифицированный рендерер

Ввести структуру `detailField { label string; value string; visible bool }` и собрать список — потом единый рендер с пустыми строками между. Чище для добавления полей, проще тестировать.

- **Pros:** legibility, проще добавлять project info (BL-6).
- **Cons:** больше нового кода для cosmetic change; усложняет diff.

### Option C: Расширение `nameCacheLoadedMsg` для хранения полных Project (BL-6 dependency)

Сейчас кэшируется только `map[id.ID]string` для projects. Чтобы показать project.Deadline / Status / Notes — нужно либо кэшировать полные сущности, либо при наведении курсора отдельно фетчить project.

- **Pros:** один универсальный кэш, нет N+1 при навигации.
- **Cons:** меняет интерфейс `nameCacheLoadedMsg` и `fetchNameCache` — требует обновления тестов.

**Recommended:** **Option A + Option C**, без B (B — over-engineering для текущего объёма):
- A — стилизация и пустые строки добавляются inline в `viewDetails` с минимальной мутацией контракта.
- C — кэш `projectsByID map[id.ID]project.Project` заменяет `projectNamesByID`, или дополняет его. Я склоняюсь к **замене**: один источник истины. В рендере `resolveName` для project станет `resolveProjectName(...)` или просто читает `.Name` из `Project` в новом кэше.

## Constraints & Risks

- **Backwards compatibility default `ListPaneShare`:** изменение с 0.45 → 0.60 **сломает test** `TestDefaults_AppConfig` (`internal/config/app_test.go:14`). Это видимый legitимный update — пользовательских конфигов пока в репо нет, но default действительно меняется. Решение: обновить тест и явно отметить в commit message.
- **Поведение при `width < ~30`:** при очень узком details pane стилизованный label + пустая строка не должны вызвать переполнение. Текущий `wrapAndTruncate` ограничивает notes; для других полей wrap не применяется (короткие значения). При сильно зажатом details (< 20 char) values могут обрезаться терминалом — приемлемо для edge-case.
- **Тесты `details_test.go`:** проверяют `Contains` по plain text (`"Start:"`). Стилизация не ломает — ANSI escapes окружают подстроку, не разрывают её. Но: если ввести новый кэш `projectsByID`, тесты `TestViewDetails_RelationsAndTags` (использует `m.projectNamesByID`) сломаются. Решение: либо сохранить старое поле как deprecated synonym, либо явно мигрировать тесты.
- **Heading info:** `HeadingGet` отсутствует в Repository — heading name всё ещё через short-ID fallback. Это **деферр** (Deferred v2), в текущем scope не входит.
- **Project Notes длинный:** если показать `Project Notes`, его тоже надо wrap-truncate как task notes. Простая reuse `wrapAndTruncate`. Что показывать дополнительно — открытый вопрос для interview (см. ниже).

## Recommended Direction

**Скоуп для v1:**

1. **BL-1.1 — стилизованные лейблы + воздух:**
   - Лейблы `Status:`/`Start:`/`Due:`/`Pinned:`/`Area:`/`Project:`/`Heading:`/`Tags:` рендерятся через **`theme.Label`** (Bold + Subtext) — или новый **`theme.LabelStrong`** (Bold + Accent) для большего контраста. **[ASSUMPTION: использовать существующий `theme.Label` — он уже Bold+Subtext, достаточно для отбивки от значений. Если выглядит блёкло — переключим на Accent на этапе implementation]**.
   - Между каждой группой полей вставляется пустая строка для воздуха. Group = одна строка вывода (label+value). Notes остаётся блоком (может занимать несколько строк) — отделяется пустой строкой сверху и снизу.

2. **BL-2 — details pane ≤ 40%:**
   - Изменить дефолт `ListPaneShare` с `0.45` на `0.60` в `internal/config/app.go:18`.
   - Обновить `TestDefaults_AppConfig` в `internal/config/app_test.go`.
   - Документировать новый дефолт (commit message + потенциально README, если он рассказывает о настройке).

3. **BL-6 — расширенное project info в details:**
   - Заменить `projectNamesByID map[id.ID]string` на `projectsByID map[id.ID]project.Project` в Model.
   - `fetchNameCache` теперь кэширует `project.Project` целиком; `nameCacheLoadedMsg.projects` меняет тип.
   - В `viewDetails` отображать (если поле задано):
     - `Project: <Name>` (уже есть)
     - `  Status: <project.Status>` (новое — статус проекта если != Open)
     - `  Deadline: <YYYY-MM-DD>` (новое — отдельный project.Deadline, не путать с task.Deadline)
     - `  Notes: <wrapped, max 3 lines>` (новое — короткий project.Notes блок)
   - Лейблы для sub-полей с отступом в 2 пробела, чтобы визуально показать иерархию.

## Scope Boundaries

- **Must-have (v1):**
  - Стилизация лейблов + пустые строки между полями (BL-1.1).
  - Default `ListPaneShare = 0.60` (BL-2).
  - Project sub-fields: Status, Deadline, Notes (BL-6).
  - Кэш `projectsByID` вместо `projectNamesByID`.
- **Deferred (v2):**
  - `HeadingGet(ctx, id)` в Repository + heading name resolution.
  - Конфигурируемый цвет labels (например, через `theme.LabelStrong`).
  - Project Area name display (currently project.AreaID не используется в details, можно показать «Project Area: X»).
  - Pretty-print дат через relative (today/yesterday/in 3 days).
- **Needs spike:** none — все необходимые контракты в Repository уже есть.

## Assumptions & Open Questions

- **[ASSUMPTION: `theme.Label` (Bold+Subtext) достаточен для визуального отличия лейблов от значений]**. Альтернатива — Accent. Решение можно поправить во время implementation после визуального теста.
- **[ASSUMPTION: пользователь принимает breaking change дефолта `ListPaneShare` (0.45 → 0.60). Существующие YAML-конфиги остаются работать; меняется только built-in default]**.
- **[ASSUMPTION: показ Project Notes уместен в details pane (а не только Project Name). Если нет — уберём Notes из v1 и оставим только Status+Deadline]**.
- **[ASSUMPTION: replace, а не extend для `projectNamesByID → projectsByID`. Тесты, которые сейчас задают `m.projectNamesByID[pid] = "..."`, будут переписаны на `m.projectsByID[pid] = project.Project{Name: "..."}` ]**.

**Open Questions** (на interview в requirements phase):

1. Каким цветом стилизовать лейблы? Subtext (мягко) или Accent (контрастно)?
2. Какие именно project sub-fields показывать в details? Status, Deadline, Notes — все три, или только подмножество?
3. Если задача внутри проекта **и** есть heading — где показывать heading относительно project? (`Project: Foo / Heading: Bar` или две строки?)

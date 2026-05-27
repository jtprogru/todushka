# Details Pane Redesign — Task Plan

**Status:** Draft
**Work Type:** Pure feature (extends existing `viewDetails`, replaces project cache, adds new Theme field, changes config default).
**Date:** 2026-05-27

## Test Style Source

**Test Style Source:** Tier 2
- Reference test files: `internal/tui/details_test.go`, `internal/tui/app_test.go`, `internal/tui/shell_test.go`, `internal/config/app_test.go`.
- Key patterns:
  - `testify/require` для unit; `pgregory.net/rapid` для property-based.
  - Helpers: `newTestModel(t)`, `newTestModelWithService(t)`, `setupRapidModel(rt, titles...)`.
  - ANSI-sensitive tests: `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup(... Ascii)`.
  - Тесты широко мутируют поля Model прямо (`m.areaNamesByID[id] = "..."`).

## Commands

| Action       | Command           | Source       |
|--------------|-------------------|--------------|
| Test         | `task test`       | Taskfile.yml |
| Test (race)  | `task test-race`  | Taskfile.yml |
| Build        | `task build`      | Taskfile.yml |
| Lint         | `task lint`       | Taskfile.yml |

## Coverage Matrix

| Requirement | Task(s)      | Correctness Property |
|-------------|--------------|----------------------|
| REQ-1.1     | T-1, T-2     | CP-1, CP-2           |
| REQ-1.2     | T-1, T-5     | CP-3                 |
| REQ-1.3     | T-1, T-5     | CP-4                 |
| REQ-1.4     | T-1, T-5     | CP-4                 |
| REQ-2.1     | T-1, T-3     | CP-5                 |
| REQ-2.2     | T-1, T-3     | CP-5                 |
| REQ-2.3     | T-1, T-3     | CP-6                 |
| REQ-3.1     | T-1, T-4     | CP-7                 |
| REQ-3.2     | T-1, T-4     | CP-7                 |
| REQ-3.3     | T-1, T-5     | CP-8                 |
| REQ-3.4     | T-1, T-5     | CP-8                 |
| REQ-3.5     | T-1, T-5     | CP-8                 |
| REQ-3.6     | T-1, T-5     | CP-9                 |
| REQ-3.7     | T-1, T-5     | CP-10                |
| REQ-4.1     | T-1, T-3     | CP-6                 |
| REQ-4.2     | T-1, T-5     | CP-11                |

---

## T-1 — Написать новые тесты (GREEN-stubs)

***_Requirements:_*** REQ-1.1, REQ-1.2, REQ-1.3, REQ-1.4, REQ-2.1, REQ-2.2, REQ-2.3, REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4, REQ-3.5, REQ-3.6, REQ-3.7, REQ-4.1, REQ-4.2
***_Complexity:_*** standard
***_Test_Style:_*** `internal/tui/details_test.go`, `internal/config/app_test.go`

GOAL: записать наблюдаемое поведение всех новых REQ. Из-за миграции `projectNamesByID → projectsByID` тесты в новом файле НЕ скомпилируются до T-4. Это допустимый RED-state для pure feature: compile error = «поведение ещё не реализовано».

### T-1.1 — Создать файл `internal/tui/details_redesign_test.go` с unit-тестами на стиль

CRITICAL: тест использует `m.theme.DetailLabel` — поле ещё не существует. Compile fail до T-2.

Тесты:
- `TestTheme_DetailLabelColorTheme`: `t := NewTheme(); out := t.DetailLabel.Render("X"); require.Contains(t, out, "\x1b[1")` (bold) AND `require.Regexp` на foreground ANSI (38;2 или 38;5).
- `TestTheme_DetailLabelMonochrome`: `t := NewMonochromeTheme(); out := t.DetailLabel.Render("X"); require.Contains(t, out, "\x1b[1"); require.NotContains(t, out, "\x1b[38;")`.
- `TestProp_DetailLabelHasBold` (rapid): сгенерировать выбор color/mono; assert bold.

### T-1.2 — Добавить unit-тесты на стилизацию лейблов в `viewDetails`

CRITICAL: эти тесты упадут до T-5 (стилизация не реализована).

- `TestViewDetails_LabelsStyled`: `lipgloss.SetColorProfile(termenv.TrueColor) + t.Cleanup`; настроить task со всеми основными полями (Start, Due, Area, Project, Tags); render `viewDetails(m, 80)`; для каждого лейбла `"Status:"`, `"Start:"`, `"Due:"`, `"Area:"`, `"Project:"`, `"Tags:"` assert `require.Contains(out, m.theme.DetailLabel.Render(label))`.

### T-1.3 — Добавить unit-тесты на spacing (blank lines)

- `TestViewDetails_NoOrphanBlankLines`: minimal task (только Title + Status), render `viewDetails`; assert `require.NotContains(t, out, "\n\n\n")`.
- `TestViewDetails_EmptyLineBetweenGroups`: task с Status + Start + Area + Tags; assert между группами Status и Dates ровно одна `"\n"` (т.е. наличие подстроки `"<status row>\n\n<start row>"` с одной пустой строкой). Использовать `strings.Split` и проверять, что blank-line присутствует между группами.
- `TestProp_NoConsecutiveBlankLines` (rapid): рандомная подмножество полей; assert no `"\n\n\n"`.

### T-1.4 — Добавить unit-тесты на ширину details (config)

В `internal/config/app_test.go` ДОПОЛНИТЬ (не заменять) существующий `TestDefaults_AppConfig` — он будет обновлён в T-3. В `internal/tui/details_redesign_test.go` добавить:
- `TestConfig_DefaultsListPaneShareIs06`: `require.InDelta(t, 0.60, config.Defaults().ListPaneShare, 1e-9)`.
- `TestPaneWidths_DetailsAtMost40Percent`: для `m.width ∈ {100, 120, 150, 200}` (`m.config = config.Defaults()`) assert `details, _ := paneWidths(m)... wait, signature returns (list, details)`; `require.LessOrEqual(t, float64(details)/float64(m.width), 0.40)`.
- `TestConfig_ValidatePreservesValidListPaneShare`: для значений 0.20, 0.30, 0.45, 0.50, 0.80 — `AppConfig{ListPaneShare: v}.Validate()` возвращает `v`.
- `TestProp_DetailsLeq40Percent` (rapid): `m.width ∈ [100..400]`; assert ≤ 0.40.
- `TestProp_ListPaneShareRoundtrip` (rapid): `v ∈ Float64Range(0.01, 0.99)`; assert `Validate(AppConfig{ListPaneShare: v}).ListPaneShare == v`.

NOTE: `paneWidths` возвращает `(list, details)` (см. `details.go:33`). Использовать ту же подпись.

### T-1.5 — Добавить unit-тесты на project cache flow (Model + msg)

CRITICAL: эти тесты используют `m.projectsByID` — compile fail до T-4.

- `TestNameCache_FetchEmitsFullProject`: добавить Project через `svc.AddProject(...)` со всеми полями (Name, Notes, Status, Deadline); создать task с этим ProjectID; вызвать `fetchNameCache(svc, []task.Task{tk})()`; cast в `nameCacheLoadedMsg`; assert `res.projects[pid].Name == "..."`, `Notes == "..."`, `Status == ...`, `Deadline == ...`.
- `TestNameCache_UpdateStoresFullProject`: `msg := nameCacheLoadedMsg{projects: map[id.ID]project.Project{pid: {Name:"foo", Status: project.StatusCompleted}}}`; `m2, _ := m.Update(msg)`; assert `m2.(Model).projectsByID[pid].Status == project.StatusCompleted`.
- `TestProp_ProjectFlowsEndToEnd` (rapid): random Name/Status/Deadline/Notes, через svc; assert msg.

NOTE: проверить, есть ли `svc.AddProject(...)` метод в app.Service. Если нет — использовать `svc.Repo().ProjectPut(ctx, p)` напрямую.

### T-1.6 — Добавить unit-тесты на project sub-fields в `viewDetails`

CRITICAL: используют `m.projectsByID` — compile fail до T-4; render-логика fail до T-5.

- `TestViewDetails_ProjectStatusSubField`: task с ProjectID; `m.projectsByID[pid] = project.Project{Name:"foo", Status: project.StatusCompleted}`; assert `out` содержит `"Project status:"` and `"Completed"`.
- `TestViewDetails_ProjectDeadlineSubField`: project с `Deadline: &date`; assert `out` содержит `"Project due:"` and `"2026-05-26"` (или whatever date).
- `TestViewDetails_ProjectNotesSubField`: project с `Notes: "important notes"`; assert `out` содержит `"Project notes:"` and `"important notes"`.
- `TestViewDetails_ProjectSubFieldsHiddenWhenOpenAndEmpty`: project с Status=Open, Deadline=nil, Notes=""; assert `out` содержит `"Project:"` но НЕ `"Project status:"`/`"Project due:"`/`"Project notes:"`.
- `TestViewDetails_ProjectFallbackOnCacheMiss`: task с ProjectID; пустой `m.projectsByID`; assert `out` содержит `"Project: <id.Short(pid)>"` и НЕ `"Project status:"` и других sub-fields.
- `TestViewDetails_ProjectAndHeadingSeparateLines`: task с обоими ProjectID + HeadingID, оба в кэше; assert никакая строка вывода не содержит одновременно `"Project:"` и `"Heading:"`.
- `TestViewDetails_RegressionContains`: task со всеми полями set; assert все: `"Status:"`, `"Start:"`, `"Due:"`, `"Area:"`, `"Project:"`, `"Tags:"`.
- `TestProp_ProjectSubFieldsVisibilityMatchesCache` (rapid): random combinations.
- `TestProp_ProjectFallbackOnMissing` (rapid): random ProjectIDs, empty cache.
- `TestProp_AllLabelsStyled` (rapid): random subset fields, all labels ANSI-wrapped.

---

## T-2 — Реализовать `Theme.DetailLabel` (CODE)

***_Requirements:_*** REQ-1.1
***_Preservation:_*** CP-11 (existing Theme usage unaffected — editor.go still uses theme.Label which stays Subtext)
***_Complexity:_*** mechanical

### T-2.1 — Добавить поле `DetailLabel lipgloss.Style` в struct `Theme` в `internal/tui/style.go`

В `internal/tui/style.go:11-40` (struct `Theme`): добавить строку `DetailLabel lipgloss.Style` после `Label`.

### T-2.2 — Инициализировать `DetailLabel` в `newColorTheme`

В `internal/tui/style.go:103-145` (`newColorTheme`): добавить после `t.Label = ...`:

```go
t.DetailLabel = lipgloss.NewStyle().Bold(true).Foreground(t.Accent)
```

### T-2.3 — Инициализировать `DetailLabel` в `NewMonochromeTheme`

В `internal/tui/style.go:150-172` (`NewMonochromeTheme`): добавить в struct-literal:

```go
DetailLabel: bold,
```

(использовать existing `bold := lipgloss.NewStyle().Bold(true)`).

### T-2.4 — Прогнать тесты темы

```
task test -- -run "TestTheme_DetailLabel|TestProp_DetailLabelHasBold"
```

CRITICAL: 3 теста должны пройти.

---

## T-3 — Реализовать `ListPaneShare` default = 0.60 (CODE)

***_Requirements:_*** REQ-2.1, REQ-2.2, REQ-2.3, REQ-4.1
***_Preservation:_*** CP-6 (Validate preserves user-set values)
***_Complexity:_*** mechanical

### T-3.1 — Изменить дефолт в `internal/config/app.go`

В `internal/config/app.go:18`: `ListPaneShare: 0.45` → `ListPaneShare: 0.60`.

### T-3.2 — Обновить существующий `TestDefaults_AppConfig` в `internal/config/app_test.go`

В `internal/config/app_test.go:14`: `require.InDelta(t, 0.45, c.ListPaneShare, 1e-9)` → `require.InDelta(t, 0.60, c.ListPaneShare, 1e-9)`.

### T-3.3 — Обновить связанные тесты, если есть

Поискать другие места, где ожидается 0.45: `grep -rn "0.45" internal/config/ internal/tui/`. Если есть — обновить только если ожидание совпадает с дефолтом (НЕ обновлять, если ожидание — про корректную обработку custom value 0.45).

### T-3.4 — Прогнать config + pane-width тесты

```
task test -- -run "TestConfig_DefaultsListPaneShareIs06|TestPaneWidths_DetailsAtMost40Percent|TestConfig_ValidatePreservesValidListPaneShare|TestDefaults_AppConfig|TestProp_DetailsLeq40Percent|TestProp_ListPaneShareRoundtrip"
```

CRITICAL: все 6 тестов должны пройти.

---

## T-4 — Миграция кэша на `projectsByID` (CODE)

***_Requirements:_*** REQ-3.1, REQ-3.2
***_Preservation:_*** CP-11 (existing substring tests in `details_test.go` continue to pass after migration), CP-7
***_Complexity:_*** standard

### T-4.1 — Изменить тип в `internal/tui/msgs.go`

В `internal/tui/msgs.go:87-92` (`nameCacheLoadedMsg`):
- Импорт: добавить `"github.com/jtprogru/todushka/internal/domain/project"`.
- Поле `projects` тип: `map[id.ID]string` → `map[id.ID]project.Project`.

### T-4.2 — Изменить поле Model в `internal/tui/app.go`

В `internal/tui/app.go:42-50` (struct Model):
- Импорт: добавить `"github.com/jtprogru/todushka/internal/domain/project"`.
- Поле: `projectNamesByID map[id.ID]string` → `projectsByID map[id.ID]project.Project`.

В `internal/tui/app.go:70-80` (`NewModel`):
- `projectNamesByID: make(map[id.ID]string)` → `projectsByID: make(map[id.ID]project.Project)`.

В `internal/tui/app.go:120-133` (`case nameCacheLoadedMsg:`):
- Заменить:
  ```go
  for k, v := range msg.projects {
      m.projectNamesByID[k] = v
  }
  ```
- На:
  ```go
  for k, v := range msg.projects {
      m.projectsByID[k] = v
  }
  ```

### T-4.3 — Обновить `fetchNameCache` в `internal/tui/details.go`

В `internal/tui/details.go:48-89`:
- Импорт: добавить `"github.com/jtprogru/todushka/internal/domain/project"`.
- Локальная переменная: `projects := make(map[id.ID]string)` → `projects := make(map[id.ID]project.Project)`.
- Цикл `for pid := range projectSet`: вместо `projects[pid] = p.Name` записать `projects[pid] = p` (полная сущность).

### T-4.4 — Обновить вызов в `viewDetails` в `internal/tui/details.go`

В `internal/tui/details.go:164-166`:
- Заменить `lines = append(lines, "Project: "+resolveName(m.projectNamesByID, *t.ProjectID))` на временный stub:
  ```go
  lines = append(lines, "Project: "+resolveProjectName(m.projectsByID, *t.ProjectID))
  ```
- Добавить новый helper в `details.go` (рядом с `resolveName`):
  ```go
  func resolveProjectName(cache map[id.ID]project.Project, pid id.ID) string {
      if p, ok := cache[pid]; ok && p.Name != "" {
          return p.Name
      }
      return id.Short(pid)
  }
  ```

NOTE: полную перезапись `viewDetails` (sub-fields + spacing) делает T-5. Здесь только миграция compile.

### T-4.5 — Мигрировать существующие тесты под новый тип

В `internal/tui/app_test.go:862` (`TestNameCache_LoadedMsgPopulatesModel`):
- `msg.projects = map[id.ID]string{pid: "todushka"}` → `msg.projects = map[id.ID]project.Project{pid: {Name: "todushka"}}`.
- `require.Equal(t, "todushka", mm.projectNamesByID[pid])` → `require.Equal(t, "todushka", mm.projectsByID[pid].Name)`.

В `internal/tui/details_test.go` поискать `projectNamesByID`:
- `grep -n "projectNamesByID" internal/tui/details_test.go` — обновить все callsites на `projectsByID[pid] = project.Project{Name: "..."}`.

### T-4.6 — Прогнать compile + cache flow тесты

```
task build
task test -- -run "TestNameCache_FetchEmitsFullProject|TestNameCache_UpdateStoresFullProject|TestProp_ProjectFlowsEndToEnd|TestNameCache_FetchCmdEmitsMsg|TestNameCache_LoadedMsgPopulatesModel"
```

CRITICAL: build green, все 5 тестов pass.

---

## T-5 — Перезаписать `viewDetails` (CODE)

***_Requirements:_*** REQ-1.2, REQ-1.3, REQ-1.4, REQ-3.3, REQ-3.4, REQ-3.5, REQ-3.6, REQ-3.7, REQ-4.2
***_Preservation:_*** CP-7, CP-11 (все существующие substring-assertions в `details_test.go`)
***_Complexity:_*** complex

### T-5.1 — Спроектировать новую структуру `viewDetails`

В `internal/tui/details.go:140-181` — переписать `viewDetails`. Структура:

```go
func viewDetails(m Model, width int) string {
    t := cursorTask(m)
    if t == nil {
        return m.theme.Dim.Render("(no task selected)")
    }
    var groups [][]string  // each group = list of rendered lines

    // Group 1: Title (always present)
    groups = append(groups, []string{m.theme.Title.Render(wrapAndTruncate(t.Title, width, 4))})

    // Group 2: Status
    groups = append(groups, []string{m.theme.DetailLabel.Render("Status:") + " " + statusLabel(t.Status)})

    // Group 3: Notes
    if t.Notes != "" {
        groups = append(groups, []string{wrapAndTruncate(t.Notes, width, m.config.NotesMaxLines)})
    }

    // Group 4: Dates
    var dates []string
    if t.StartDate != nil {
        dates = append(dates, m.theme.DetailLabel.Render("Start:")+"  "+t.StartDate.Format("2006-01-02"))
    }
    if t.Deadline != nil {
        dates = append(dates, m.theme.DetailLabel.Render("Due:")+"    "+t.Deadline.Format("2006-01-02"))
    }
    if t.PinnedToday != nil {
        dates = append(dates, m.theme.DetailLabel.Render("Pinned:")+" "+t.PinnedToday.Format("2006-01-02"))
    }
    if len(dates) > 0 {
        groups = append(groups, dates)
    }

    // Group 5: Relations (Area, Project + sub-fields, Heading) — Project and Heading
    // are on SEPARATE lines per REQ-3.7. Sub-fields appear between Project: line
    // and Heading: line.
    var relations []string
    if t.AreaID != nil {
        relations = append(relations, m.theme.DetailLabel.Render("Area:")+"    "+resolveName(m.areaNamesByID, *t.AreaID))
    }
    if t.ProjectID != nil {
        relations = append(relations, m.theme.DetailLabel.Render("Project:")+" "+resolveProjectName(m.projectsByID, *t.ProjectID))
        // Sub-fields (REQ-3.3..3.6): only when project is in cache.
        if p, ok := m.projectsByID[*t.ProjectID]; ok && p.Name != "" {
            if p.Status != project.StatusOpen && p.Status != "" {
                relations = append(relations, "  "+m.theme.DetailLabel.Render("Project status:")+" "+string(p.Status))
            }
            if p.Deadline != nil {
                relations = append(relations, "  "+m.theme.DetailLabel.Render("Project due:")+" "+p.Deadline.Format("2006-01-02"))
            }
            if p.Notes != "" {
                relations = append(relations, "  "+m.theme.DetailLabel.Render("Project notes:")+" "+wrapAndTruncate(p.Notes, width-2, 3))
            }
        }
    }
    if t.HeadingID != nil {
        relations = append(relations, m.theme.DetailLabel.Render("Heading:")+" "+resolveName(m.headingNamesByID, *t.HeadingID))
    }
    if len(relations) > 0 {
        groups = append(groups, relations)
    }

    // Group 6: Tags
    if len(t.Tags) > 0 {
        names := make([]string, 0, len(t.Tags))
        for _, tg := range t.Tags {
            names = append(names, resolveName(m.tagNamesByID, tg))
        }
        groups = append(groups, []string{m.theme.DetailLabel.Render("Tags:") + " " + strings.Join(names, ", ")})
    }

    // Group 7: Someday
    if t.Someday {
        groups = append(groups, []string{m.theme.Dim.Render("Someday")})
    }

    // Join groups with blank line between non-empty groups.
    var out []string
    for i, g := range groups {
        if i > 0 {
            out = append(out, "")
        }
        out = append(out, g...)
    }
    return strings.Join(out, "\n")
}
```

CRITICAL:
- **REQ-3.7 (Heading на отдельной строке)** обеспечивается тем, что Heading appended as separate element `relations[last]`, после всех project sub-fields.
- **REQ-1.4 (no orphan blank lines)** обеспечивается тем, что blank line inserted ТОЛЬКО между группами и ТОЛЬКО если предыдущая группа была непустой (что гарантирует условие `i > 0`, потому что мы добавляем только непустые группы).
- **REQ-4.2 (regression)** — substring `"Status:"`, `"Start:"`, `"Due:"`, `"Area:"`, `"Project:"`, `"Tags:"` остаются точно теми же подстроками (DetailLabel.Render оборачивает ANSI, не разрывая текст).

IMPORTANT: импорт `"github.com/jtprogru/todushka/internal/domain/project"` нужен для `project.StatusOpen`.

### T-5.2 — Удалить старый код `viewDetails` и применить новый

Заменить тело `viewDetails` целиком на код из T-5.1.

### T-5.3 — Прогнать существующие details тесты (regression)

```
task test -- -run "TestViewDetails_" -count=1
```

CRITICAL: ВСЕ существующие `TestViewDetails_*` тесты (из `details_test.go`) должны пройти. Это включает:
- `TestViewDetails_RelationsAndTags`
- `TestViewDetails_ShortIDFallback`
- и все остальные, проверяющие substring `"Status:"`, `"Start:"`, etc.

Если упали — fix регрессию ДО продолжения. Не модифицировать старые тесты (они — regression-lock).

### T-5.4 — Прогнать новые тесты на стилизацию + spacing + project sub-fields

```
task test -- -run "TestViewDetails_LabelsStyled|TestViewDetails_NoOrphanBlankLines|TestViewDetails_EmptyLineBetweenGroups|TestViewDetails_ProjectStatusSubField|TestViewDetails_ProjectDeadlineSubField|TestViewDetails_ProjectNotesSubField|TestViewDetails_ProjectSubFieldsHiddenWhenOpenAndEmpty|TestViewDetails_ProjectFallbackOnCacheMiss|TestViewDetails_ProjectAndHeadingSeparateLines|TestViewDetails_RegressionContains|TestProp_NoConsecutiveBlankLines|TestProp_AllLabelsStyled|TestProp_ProjectSubFieldsVisibilityMatchesCache|TestProp_ProjectFallbackOnMissing"
```

CRITICAL: все 14 тестов pass.

---

## T-6 — VERIFY + GATE (полный прогон + lint + build)

***_Requirements:_*** все (REQ-1.1..4.2)
***_Complexity:_*** mechanical

### T-6.1 — Полный прогон тестов

```
task test
task test-race
```

CRITICAL: оба зелёные. Особое внимание `internal/tui/` и `internal/config/`.

### T-6.2 — Lint

```
task lint
```

CRITICAL: 0 issues.

### T-6.3 — Build

```
task build
```

CRITICAL: бинарь собран.

### T-6.4 — Финальный traceability check

Прочитать coverage matrix и убедиться, что каждый REQ закрыт ≥1 прошедшим тестом:
- REQ-1.x (стиль/spacing): TestTheme_DetailLabel*, TestViewDetails_LabelsStyled, TestViewDetails_*BlankLines, TestProp_DetailLabelHasBold, TestProp_NoConsecutiveBlankLines, TestProp_AllLabelsStyled.
- REQ-2.x (ширина): TestConfig_DefaultsListPaneShareIs06, TestPaneWidths_DetailsAtMost40Percent, TestConfig_ValidatePreservesValidListPaneShare, TestProp_DetailsLeq40Percent, TestProp_ListPaneShareRoundtrip.
- REQ-3.x (project info): TestNameCache_*, TestViewDetails_Project*, TestProp_ProjectFlowsEndToEnd, TestProp_ProjectSubFieldsVisibilityMatchesCache, TestProp_ProjectFallbackOnMissing.
- REQ-4.x (back-compat): TestConfig_ValidatePreservesValidListPaneShare, TestViewDetails_RegressionContains + all existing `TestViewDetails_*` tests.

### T-6.5 — Smoke test в TUI (manual, optional)

Если есть тестовая бд — запустить `task run` и визуально оценить:
- Лейблы в details выглядят как Bold + синий (Accent в Macchiato).
- Между группами видна пустая строка.
- Details pane занимает не больше ~40% экрана при `width=120`.
- Project sub-fields (если есть тестовый project с notes/deadline/status) видны с отступом.

NOTE: manual smoke не gating требование — automation покрывает все CP. Но визуальная проверка ловит палитру/контраст rough edges.

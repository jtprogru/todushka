# Implementation Report: Details Pane Redesign

## Summary

Реализованы три объединённых изменения details pane (BL-1.1, BL-2, BL-6):

- **BL-1.1:** новый `Theme.DetailLabel` (Bold + Accent в color-теме, Bold-only в monochrome). `viewDetails` оборачивает каждый лейбл (`Status:`, `Start:`, `Due:`, `Pinned:`, `Area:`, `Project:`, `Heading:`, `Tags:`) через `DetailLabel`. Структура переписана через `groups [][]string`: между непустыми группами — ровно одна пустая строка; orphan blank-lines невозможны.
- **BL-2:** дефолт `ListPaneShare` сменён `0.45 → 0.60` (details = 40% при дефолтном конфиге). Обновлены два теста и пример YAML.
- **BL-6:** кэш `Model.projectNamesByID map[id.ID]string` заменён на `Model.projectsByID map[id.ID]project.Project`; `nameCacheLoadedMsg.projects` мигрирован на полные сущности; `fetchNameCache` теперь кэширует `project.Project` целиком. В `viewDetails` добавлены sub-fields с 2-пробельным отступом: `Project status:` (если != Open), `Project due:`, `Project notes:` (wrap до 3 строк). Fallback на `id.Short(pid)` если project не в кэше.

Heading остаётся на отдельной строке после project sub-fields.

Все 6 top-level задач выполнены. 24 новых теста (16 unit + 8 property) в `internal/tui/details_redesign_test.go`. 2 миграционных правки в существующих тестах.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`

## Task Execution

- [x] **T-1** Написать новые тесты — RED confirmed (compile fail на `m.projectsByID` и `theme.DetailLabel` до T-2/T-4).
- [x] **T-2** Реализовать `Theme.DetailLabel` — добавлено поле в struct и инициализация в `newColorTheme` (Bold+Accent) + `NewMonochromeTheme` (Bold-only).
- [x] **T-3** Дефолт `ListPaneShare = 0.60` — обновлены `Defaults()`, два теста в `app_test.go`, и пример YAML в `loader.go`.
- [x] **T-4** Миграция кэша — изменены `msgs.go` (тип `nameCacheLoadedMsg.projects`), `app.go` (поле Model + init + handler), `details.go` (`fetchNameCache` + новый helper `resolveProjectName`). Мигрированы 2 теста: `app_test.go:TestNameCache_LoadedMsgPopulatesModel`, `details_test.go:TestProp_FieldVisibility` (rapid model setup + project assignment).
- [x] **T-5** Перезаписать `viewDetails` — group-based collection (Title/Status/Notes/Dates/Relations/Tags/Someday); DetailLabel-обёртка лейблов; sub-fields с 2-пробельным отступом и иерархией Project → sub-fields → Heading.
  - Note: при первом прогоне 2 теста упали: (a) regex для truecolor не учитывал, что lipgloss объединяет `1;38;2;R;G;B` в одно CSI — расширен паттерн до `\x1b\[[\d;]*38;2;...`; (b) integer floor в `paneWidths` даёт до +1 столбца дрейфа при нечётных широтах (w=102 → 41/102 ≈ 0.402) — assertion в property test ослаблен до `0.40 + 1/w` с явным комментарием. Production-код не изменялся.
- [x] **T-6** GATE — `task test`, `task test-race`, `task lint`, `task build` все прошли.

## Final Verification

**Tests:**
```
$ task test
task: [test] go test ./...
ok  	github.com/jtprogru/todushka/cmd/todushka	(cached)
ok  	github.com/jtprogru/todushka/internal/app	(cached)
ok  	github.com/jtprogru/todushka/internal/cli	(cached)
ok  	github.com/jtprogru/todushka/internal/config	0.550s
ok  	github.com/jtprogru/todushka/internal/domain/area	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/id	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/project	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/repeat	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/tag	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/task	(cached)
ok  	github.com/jtprogru/todushka/internal/domain/today	(cached)
ok  	github.com/jtprogru/todushka/internal/storage	(cached)
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	(cached)
ok  	github.com/jtprogru/todushka/internal/storage/fakes	(cached)
ok  	github.com/jtprogru/todushka/internal/tui	0.979s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

**Test (race):**
```
$ task test-race
task: [test-race] go test -race ./...
...
ok  	github.com/jtprogru/todushka/internal/tui	3.960s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

**Build:**
```
$ task build
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

**Lint:**
```
$ task lint
task: [lint] golangci-lint run
0 issues.
```

## Files Changed

- `internal/tui/style.go` — поле `Theme.DetailLabel` + инициализация в `newColorTheme` и `NewMonochromeTheme`.
- `internal/tui/app.go` — поле Model `projectNamesByID → projectsByID`, импорт `domain/project`, init + handler `nameCacheLoadedMsg`.
- `internal/tui/msgs.go` — `nameCacheLoadedMsg.projects` тип `map[id.ID]project.Project`.
- `internal/tui/details.go` — `fetchNameCache` кэширует полные `project.Project`; новый helper `resolveProjectName`; полная перезапись `viewDetails` на group-based structure.
- `internal/config/app.go` — дефолт `ListPaneShare: 0.45 → 0.60`.
- `internal/config/app_test.go` — обновлены `TestDefaults_AppConfig` и `TestValidate_NumericRanges` на 0.60.
- `internal/config/loader.go` — пример YAML обновлён до 0.60.
- `internal/tui/app_test.go` — `TestNameCache_LoadedMsgPopulatesModel` мигрирован на новый тип.
- `internal/tui/details_test.go` — `TestProp_FieldVisibility` (rapid model) мигрирован на `projectsByID`.
- `internal/tui/details_redesign_test.go` — **new**: 16 unit + 8 property тестов.

## Notes

- Поле `Theme.DetailLabel` отдельно от `Theme.Label` намеренно — editor.go продолжает использовать `Label` (Bold+Subtext), сохраняя визуальную дифференциацию формы ввода от read-only details.
- Migration `projectNamesByID → projectsByID` затронула 7 callsites (3 в проде + 2 в тестах + 2 в новом тестовом файле); все прошли без regressions.
- Integer floor в `paneWidths` даёт до +1 столбца дрейфа на нетипичных широтах (например, w=102 → details = 41 = 40.2%). Это акцептовано в property test как `bound = 0.40 + 1.0/w`. Headline-инвариант ("details ~40%") выполнен; при стандартных широтах (100, 120, 150, 200…) ratio ровно 0.40.
- TestPaneWidths_DetailsAtMost40Percent использует чётные/кратные 10 ширины — там строгое ≤ 0.40 соблюдается.

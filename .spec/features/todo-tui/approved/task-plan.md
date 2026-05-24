# todushka — Implementation Task Plan

**Status:** Draft
**Author:** Claude (spec-driven-dev pipeline)
**Date:** 2026-05-25
**Feature:** todo-tui
**Work Type:** **Pure feature** — greenfield, нет существующего поведения, которое нужно сохранять; вся область — новая.

---

## Preamble

### Test Style Source

**Test Style Source:** Tier 3
- Evidence: репозиторий содержит только `go.mod`, `.gitignore`, `.git/`. Adjacent тестов нет, родительские директории — личный workspace пользователя, не source-of-truth для конвенций. CI-конфигурации (`.github/workflows`, `.gitlab-ci.yml`) отсутствуют.
- Конвенции (Tier 3, выбраны в design §2.8):
  - Stdlib `testing` + `t.Run` table-driven subtests.
  - `github.com/stretchr/testify/require` (fail-fast) + `assert` локально, где нужны множественные проверки.
  - PBT: `pgregory.net/rapid` (pure Go).
  - Тесты в том же пакете (`package foo`); whitebox по умолчанию, blackbox (`package foo_test`) — только для public API surface.
  - Naming: `func TestType_Scenario(t *testing.T)`, subtest names на английском кратко.
  - `t.Helper()` обязателен у вспомогательных функций.
  - bbolt-тесты — через `t.TempDir()`.
  - Mocking: рукописные fakes; для `storage.Repository` — `internal/storage/fakes/InMemRepo`.

### Commands

| Action       | Command                                          | Source                                  |
|--------------|--------------------------------------------------|-----------------------------------------|
| Test         | `go test ./...`                                  | design §2.8 (Project Commands)          |
| Test (race)  | `go test -race ./...`                            | design §2.8                             |
| Build        | `go build -o bin/todushka ./cmd/todushka`        | design §2.8                             |
| Lint         | `golangci-lint run`                              | design §2.8                             |

Generate-команда отсутствует — codegen в v1 не используется (REQ-9 implementation note).

---

## Coverage Matrix

| Requirement | Task(s)        | Correctness Property |
|-------------|----------------|----------------------|
| REQ-1.1     | T-9            | (UI; покрыт unit-тестом TUI hotkey) |
| REQ-1.2     | T-6, T-4       | CP-1, CP-8 |
| REQ-1.3     | T-6, T-4       | CP-5 |
| REQ-1.4     | T-4, T-6       | CP-6 |
| REQ-1.5     | T-4, T-6       | CP-8 |
| REQ-1.6     | T-4, T-6       | CP-25 |
| REQ-1.7     | T-4            | CP-8 (через valid date positive path) |
| REQ-1.8     | T-4, T-6       | CP-7 |
| REQ-2.1     | T-6            | CP-1 |
| REQ-2.2     | T-6, T-3       | CP-2, CP-3, CP-4 |
| REQ-2.3     | T-6            | (unit: ListUpcoming sort) |
| REQ-2.4     | T-6            | (unit: ListAnytime filter) |
| REQ-2.5     | T-6            | (unit: ListSomeday filter) |
| REQ-2.6     | T-6            | (unit: ListLogbook sort) |
| REQ-2.7     | T-6            | (unit: ListTrash compose) |
| REQ-3.1     | T-2, T-5, T-6  | CP-9 (для области имён tags применима аналогия; areas — unit-тест uniqueness) |
| REQ-3.2     | T-5, T-6       | (unit: AddArea collision) |
| REQ-3.3     | T-6            | CP-20 |
| REQ-3.4     | T-6            | (unit: ListAreaContent) |
| REQ-4.1     | T-2, T-5, T-6  | (unit: CreateProject happy path) |
| REQ-4.2     | T-2, T-5, T-6  | (unit: AddHeading) |
| REQ-4.3     | T-6            | (unit: project auto-close behavior) |
| REQ-4.4     | T-6, T-9       | (unit: deadline overdue flag in service view) |
| REQ-4.5     | T-6            | (unit: MoveTask preserves fields) |
| REQ-5.1     | T-2, T-6, T-9  | CP-14 |
| REQ-5.2     | T-2, T-6       | (unit: Validate empty title) |
| REQ-5.3     | T-6            | CP-21 |
| REQ-5.4     | T-2, T-6       | (unit: checklist add) |
| REQ-5.5     | T-6            | (unit: complete checklist does not close task) |
| REQ-5.6     | T-6            | CP-12, CP-23 |
| REQ-5.7     | T-6            | CP-13, CP-23 |
| REQ-6.1     | T-2, T-5, T-6  | CP-9 |
| REQ-6.2     | T-6            | CP-19 |
| REQ-6.3     | T-5            | (unit: TaskFilter TagsAll AND-semantics) |
| REQ-6.4     | T-6            | CP-9 |
| REQ-7.1     | T-3            | CP-11 |
| REQ-7.2     | T-3, T-6       | CP-10, CP-12 |
| REQ-7.3     | T-6            | CP-13 |
| REQ-7.4     | T-3            | CP-11 |
| REQ-7.5     | T-3            | CP-11 |
| REQ-8.1     | T-3            | CP-2 |
| REQ-8.2     | T-3            | CP-2 |
| REQ-8.3     | T-3, T-6       | CP-2 |
| REQ-8.4     | T-3            | CP-3 |
| REQ-8.5     | T-3            | CP-3 |
| REQ-9.1     | T-1, T-5       | (unit: config.DataDir + bbolt open creates dir) |
| REQ-9.2     | T-5            | CP-17 |
| REQ-9.3     | T-5            | CP-22 |
| REQ-9.4     | T-5            | CP-17 |
| REQ-9.5     | T-5            | CP-18 |
| REQ-9.6     | T-5            | CP-14, CP-24 |
| REQ-10.1    | T-6            | CP-15 |
| REQ-10.2    | T-6            | CP-15 |
| REQ-10.3    | T-6            | CP-16 |
| REQ-11.1    | T-7            | (unit: CLI add command) |
| REQ-11.2    | T-7            | (unit: CLI today JSON output) |
| REQ-11.3    | T-7            | (unit: CLI complete not-found / ambiguous) |
| REQ-11.4    | T-7            | (unit: CLI no-args launches TUI — assert via test seam) |
| REQ-12.1    | T-9            | (unit: TUI app keymap dispatch) |
| REQ-12.2    | T-9            | (unit: list cursor j/k bounded) |
| REQ-12.3    | T-9            | (unit: app handles q / Ctrl+C → tea.Quit) |
| REQ-12.4    | T-9            | (unit: help toggle) |
| REQ-13.1    | T-9            | (unit: ErrorOccurred Msg → status bar) |
| REQ-13.2    | T-7, T-9       | (unit: recover in main + tui Program) |
| REQ-14.1    | T-9            | (unit: style returns colored ANSI when supported) |
| REQ-14.2    | T-9            | (unit: NO_COLOR → bold/underline only) |
| REQ-14.3    | T-1            | (gate: cross-compile matrix in Taskfile) |

Все 14 групп требований и все 25 Correctness Properties (CP-1..CP-25) покрыты как минимум одной задачей.

---

## Task Order (Pure Feature, bottom-up)

```
T-1 Bootstrap
  → T-2 Domain types
    → T-3 Repeat + Today
      → T-4 Quick Entry parser
        → T-5 Storage layer (Repository + bbolt + fakes)
          → T-6 Application Service (commands + queries + export/import)
            → T-7 CLI front-end (Cobra)
            → T-8 TUI infrastructure (root model, keymap, styles)
              → T-9 TUI screens (lists, editor, quick entry)
                → T-10 GATE checkpoint
```

T-7 и T-8 могут идти параллельно после T-6 — оба зависят только от Service.

---

## T-1 — Project bootstrap

*_Requirements: REQ-9.1, REQ-14.3_*
*_Complexity: mechanical_*

GOAL: подготовить инфраструктуру проекта так, чтобы `task test` / `task build` / `task lint` работали на пустой `cmd/todushka/main.go` с `package main; func main(){}`.

NOTE: Это setup-задача без бизнес-логики. Тесты на этой стадии — smoke: проверка что `package main` собирается под все 4 целевые платформы и линтер не падает на пустом проекте.

Subtasks:
- [ ] 1. Создать `Taskfile.yml` в корне с таргетами `test`, `test-race`, `build`, `lint`, `fmt`, `tidy`, `run`, `cross-compile`; `cross-compile` запускает `GOOS=linux/darwin × GOARCH=amd64/arm64 go build -o bin/todushka-$GOOS-$GOARCH ./cmd/todushka`. Запустить `task --list-all` — все 7 таргетов в списке.
- [ ] 2. Создать `.golangci.yml` с включёнными линтерами: `govet`, `staticcheck`, `errcheck`, `gosec`, `gocritic`, `revive`, `gofmt`, `goimports`, `unused`, `ineffassign`. `output.formats: colored-line-number`. Запустить `golangci-lint run` — должен пройти без ошибок (репо ещё пустой).
- [ ] 3. Создать `cmd/todushka/main.go` с `package main\nfunc main() {}`. Запустить `go build -o bin/todushka ./cmd/todushka` — артефакт создаётся.
- [ ] 4. Добавить зависимости через `go get`: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles`, `github.com/charmbracelet/lipgloss`, `go.etcd.io/bbolt`, `github.com/spf13/cobra`, `github.com/oklog/ulid/v2`, `github.com/stretchr/testify`, `pgregory.net/rapid`. Запустить `go mod tidy` — `go.sum` создан, `go.mod` обновлён.
- [ ] 5. Создать `README.md` (минимальный — 1 страница: install через `go install github.com/jtprogru/todushka/cmd/todushka@latest`, быстрый старт, локация БД, набор горячих клавиш заполняется в T-9). Запустить `go test ./...` — выводит «no test files», exit 0.
- [ ] 6. Создать `internal/config/paths.go` с экспортируемыми функциями `DataDir() (string, error)`, `StateDir() (string, error)`, `LogPath() (string, error)`, реализующими XDG fallback (REQ-9.1). Создать `internal/config/paths_test.go` с тестами `TestPaths_XDGFallback` и `TestPaths_RespectsXDGOverride` (см. design §2.8). Запустить `go test ./internal/config/...` — оба теста проходят.

After all subtasks: запустить `task cross-compile`. Подтвердить наличие 4 бинарей `bin/todushka-{linux,darwin}-{amd64,arm64}`.

---

## T-2 — Domain core types

*_Requirements: REQ-3.1, REQ-3.4, REQ-4.1, REQ-4.2, REQ-5.1, REQ-5.2, REQ-5.4, REQ-6.1_*
*_Preservation: CP-9, CP-23_*
*_Complexity: standard_*

GOAL: реализовать все pure domain-структуры с валидацией. Без I/O, без storage, без TUI. Каждый тип — отдельный пакет.

IMPORTANT: тесты пишутся вместе с кодом в одном subtask (RED→GREEN внутри subtask). Это допустимо для pure greenfield: ожидаемое поведение известно из design §2.5, тестируется в момент написания.

DO NOT: добавлять методы, которые не указаны в design §2.5 (например, никакой `Task.Move()` — перемещение происходит на уровне Service).

DO NOT: импортировать `internal/storage`, `internal/app`, `internal/tui` или `internal/cli` из domain-пакетов. Это нарушит слоистую архитектуру (design §2.1).

Subtasks:
- [ ] 1. Создать `internal/domain/id/id.go` с типом `ID string` и функциями `New() ID` (через `github.com/oklog/ulid/v2.MustNew(ulid.Now(), entropy)`), `Parse(s string) (ID, error)`, `Short(i ID) string`. Создать `internal/domain/id/id_test.go` с `TestID_NewIsULID` (генерирует 26-char Crockford base32, парсится через `ulid.Parse`), `TestID_ShortStable` (Short детерминирован, длина 6), `TestID_ParseRejectsInvalid` (пустая строка, мусор, длина ≠ 26). Запустить `go test ./internal/domain/id/...` — все 3 теста проходят.
- [ ] 2. Создать `internal/domain/task/task.go` с типами `Status`, `Date`, `Task`, `ChecklistItem` ровно как в design §2.5; реализовать `func (t Task) Validate() error` с проверками: непустой Title (REQ-5.2 → `ErrEmptyTitle`), Status ∈ {open, completed, cancelled}, инварианты completed_at/cancelled_at по REQ-5.6/5.7 (CP-23). Создать `internal/domain/task/task_test.go` с тестами `TestTask_ValidateEmptyTitle`, `TestTask_ValidateStatusInvariants` (см. design §2.8). Запустить `go test ./internal/domain/task/...` — оба прохода зелёные.
- [ ] 3. Создать `internal/domain/area/area.go` с типом `Area` (design §2.5) и `func (a Area) Validate() error` (непустой Name). Создать `internal/domain/area/area_test.go` с тестом `TestArea_ValidateEmptyName`. Запустить `go test ./internal/domain/area/...` — тест проходит.
- [ ] 4. Создать `internal/domain/project/project.go` с типами `ProjectStatus`, `Project`, `Heading`, `func (p Project) Validate() error` (непустой Name, Status ∈ enum), `func (h Heading) Validate() error` (непустой Name, непустой ProjectID). Создать `internal/domain/project/project_test.go` с table-driven `TestProject_ValidateScenarios` и `TestHeading_Validate`. Запустить `go test ./internal/domain/project/...` — оба прохода.
- [ ] 5. Создать `internal/domain/tag/tag.go` с типом `Tag` и экспортируемой функцией `Normalize(name string) string` (`strings.ToLower(strings.TrimSpace(name))`). Создать `internal/domain/tag/tag_test.go` с тестом `TestTag_NormalizeLowercaseAndTrim` и `TestTag_NormalizeIdempotent` (`Normalize(Normalize(x)) == Normalize(x)`). Запустить `go test ./internal/domain/tag/...` — оба прохода.

After all subtasks: запустить `go test ./internal/domain/...` — все доменные тесты зелёные. Запустить `golangci-lint run ./internal/domain/...` — без ошибок.

---

## T-3 — Repeat rule + Today engine (pure functions)

*_Requirements: REQ-7.1, REQ-7.2, REQ-7.4, REQ-7.5, REQ-8.1, REQ-8.2, REQ-8.3, REQ-8.4, REQ-8.5_*
*_Preservation: CP-2, CP-3, CP-4, CP-10, CP-11_*
*_Complexity: complex_*

GOAL: две независимые pure-функции — `repeat.Validate(r)`, `repeat.NextOccurrence(r, after)` и `today.ComputeToday(tasks, now, deadlineWindow)`. Высокая ценность property-based тестов.

CRITICAL: `NextOccurrence` — needs spike из exploration. Покрытие edge cases (понедельник→понедельник если выполнено в понедельник; «every weekday» через `weekly_weekdays={Mon..Fri}`; monthly day=15 при текущей дате 20-го числа) обязательно.

CRITICAL: тесты `prop_NextOccurrenceMonotonic` (CP-10) и `prop_RepeatInvalidNoMutation` (CP-11) пишутся в этой задаче и должны проходить на 1000 итерациях `rapid.Check` каждый.

IMPORTANT: `ComputeToday` не должен читать `time.Now()` сам — параметр `now` инжектируется. Это проверяется в `prop_TodayDeterminism` (CP-4).

Subtasks:
- [ ] 1. Создать `internal/domain/repeat/repeat.go` с типами `Kind`, `Rule` (design §2.5) и функцией `Validate(r Rule) error` с проверками: REQ-7.1 (valid kind), REQ-7.5 (N≥1, weekdays non-empty), REQ-7.4 (monthly day ∈ [1, 28]). Возвращаемые ошибки: `ErrInvalidRepeatRule`, `ErrMonthlyDayUnsupportedInV1`. Создать `internal/domain/repeat/repeat_test.go` с table-driven `TestRepeat_ValidateScenarios` (минимум 8 случаев: daily/everyN-OK/everyN-zero/weekly-OK/weekly-empty/monthly-OK/monthly-day-29/unknown-kind). Запустить `go test ./internal/domain/repeat/...` — все проходят.
- [ ] 2. Дополнить `internal/domain/repeat/repeat.go` функцией `NextOccurrence(r Rule, after time.Time) (time.Time, error)`: для `daily` → `after.AddDate(0,0,1).Truncate(day)`; для `every_n_days` → `after.AddDate(0,0,r.N).Truncate(day)`; для `weekly_weekdays` → ближайший день недели из `r.Weekdays`, строго после `after.Date()` (если `after.Date()` тоже подходит — берём следующий, чтобы соблюсти `result > after`); для `monthly_day` → ближайшая дата `r.Day`-го числа строго после `after.Date()`. Дополнить тестовый файл сценариями `TestRepeat_NextOccurrenceDaily`, `TestRepeat_NextOccurrenceEveryNDays`, `TestRepeat_NextOccurrenceWeeklyWeekdays` (включая late-completion edge case), `TestRepeat_NextOccurrenceMonthly`. Запустить `go test ./internal/domain/repeat/...` — все проходят.
- [ ] 3. Дополнить `internal/domain/repeat/repeat_test.go` PBT-тестами `prop_NextOccurrenceMonotonic` (CP-10, 1000 итераций — `rapid.Check` с генератором валидных Rule, проверка `next > after`) и `prop_RepeatInvalidNoMutation` (CP-11, генератор невалидных Rule, проверка `Validate` возвращает ошибку для каждого). Запустить `go test ./internal/domain/repeat/... -run 'prop_'` — оба прохода.
- [ ] 4. Создать `internal/domain/today/today.go` с функцией `ComputeToday(tasks []task.Task, now time.Time, deadlineWindowDays int) []task.Task`, реализующей правила REQ-8.1..8.5: open + (start_date ≤ now.Date() ∨ (start_date == nil ∧ deadline ≤ now.Date()+window) ∨ pinned_today.Date() == now.Date()) ∧ deleted_at == nil. NOTE: `ComputeToday` не должен импортировать ничего кроме `internal/domain/task` и stdlib `time`. Создать `internal/domain/today/today_test.go` с table-driven `TestComputeToday_Scenarios` (минимум 8 кейсов: start=today, start=tomorrow, start=nil+deadline=today, start=nil+deadline=today+1, pinned=today, pinned=yesterday, completed-excluded, deleted-excluded). Запустить `go test ./internal/domain/today/...` — все проходят.
- [ ] 5. Дополнить `internal/domain/today/today_test.go` PBT-тестами `prop_TodayInclusion` (CP-2), `prop_TodayExclusion` (CP-3), `prop_TodayDeterminism` (CP-4) — каждый 1000 итераций. Генератор задач — в `internal/testgen/task.go` (новый файл, экспортирует `rapid.Generator[task.Task]`); создать его в этом же subtask. Запустить `go test ./internal/domain/today/... -run 'prop_'` — все проходят.

After all subtasks: `go test ./internal/domain/...` зелёный. `golangci-lint run ./internal/domain/...` без ошибок.

---

## T-4 — Quick Entry parser

*_Requirements: REQ-1.4, REQ-1.5, REQ-1.6, REQ-1.7, REQ-1.8_*
*_Preservation: CP-5, CP-6, CP-7, CP-8_*
*_Complexity: standard_*

GOAL: чистый парсер `internal/domain/quickentry` с интерфейсом `Parse(input string) (Parsed, error)`. На этом уровне разрешение `@<project>`-имени в ID не выполняется (это работа Service); парсер возвращает только `ProjectRef *string`.

IMPORTANT: токены могут идти в любом порядке: `"buy milk @today #shop !2026-06-01"` == `"#shop !2026-06-01 buy milk @today"`. Title — это всё, что не является токеном, склеенное пробелами в исходном порядке.

DO NOT: делать парсер регулярным выражением «в одну строку» — он покрывает 5 типов токенов и должен давать понятные ошибки с позицией. Использовать пошаговый сплит по whitespace + token-classifier.

Subtasks:
- [ ] 1. Создать `internal/domain/quickentry/parsed.go` с типом `Parsed` (design §2.3) и типом `ParseError` (поля: `Position int`, `Token string`, `Reason string`). Запустить `go build ./internal/domain/quickentry/...` — компилируется.
- [ ] 2. Создать `internal/domain/quickentry/parser.go` с функцией `Parse(input string) (Parsed, error)`: split по whitespace, классификация каждого токена (`#name`, `@today`, `@<name>`, `!YYYY-MM-DD`, всё остальное — фрагмент title). Невалидный `!YYYY-MM-DD` → `ParseError{Position: i, Token: t, Reason: "invalid date"}`. Пустое имя тега (одинокий `#`) → `ParseError{Reason: "empty tag"}`. Пустой `@` → `ParseError{Reason: "empty mention"}`. Создать `internal/domain/quickentry/parser_test.go` с тестами `TestQuickEntry_BareTitle`, `TestQuickEntry_TitleWithTags`, `TestQuickEntry_TitleWithTodayAndDeadline`, `TestQuickEntry_InvalidDateRejected`, `TestQuickEntry_EmptyTagIgnored`, `TestQuickEntry_TokenOrderIndependent` (два вызова с одинаковыми токенами в разном порядке → equal `Parsed`). Запустить `go test ./internal/domain/quickentry/...` — все 6 проходят.
- [ ] 3. Дополнить `internal/domain/quickentry/parser_test.go` PBT-тестом `prop_QuickEntryInvalidAbsence` (CP-7, генератор строк с инжектированным невалидным токеном `!`-даты → ожидаем `ParseError` всегда). Запустить — проходит.
- [ ] 4. Создать `internal/domain/quickentry/quickentry.go` с экспортной функцией `IsEmpty(input string) bool` (`strings.TrimSpace(input) == ""`). Дополнить тестами `TestQuickEntry_IsEmpty` (table-driven: `""`, `"   "`, `"\t\n"`, `"x"`). Запустить — проходит.

After all subtasks: `go test ./internal/domain/quickentry/...` зелёный. `golangci-lint run ./internal/domain/quickentry/...` без ошибок.

---

## T-5 — Storage layer: Repository interface + bbolt impl + in-memory fake

*_Requirements: REQ-3.1, REQ-3.2, REQ-4.1, REQ-4.2, REQ-6.1, REQ-6.3, REQ-9.1, REQ-9.2, REQ-9.3, REQ-9.4, REQ-9.5, REQ-9.6_*
*_Preservation: CP-14, CP-17, CP-18, CP-19, CP-22, CP-24_*
*_Complexity: complex_*

GOAL: интерфейс `Repository` (design §2.3) + реализация на bbolt с миграциями и индексами + in-memory реализация для service-тестов. После этой задачи все persistence-инварианты валидируются property-тестами.

CRITICAL: bbolt-реализация должна выдерживать concurrent open от второго процесса (REQ-9.3 → `ErrDatabaseLocked`). Реализуется через `bbolt.Options{Timeout: 200 * time.Millisecond}`; если timeout — возвращаем доменную ошибку.

IMPORTANT: миграции — список структур `[]Migration{Version int, Apply func(tx *bbolt.Tx) error}` в `migrations.go`. v1: создать все buckets из design §2.5, записать `meta.schema_version = 1`. Применение — в одной транзакции для всей последовательности; если упало — bbolt rollback.

IMPORTANT: каждая запись Task через `TaskUpdate` ОБЯЗАНА синхронно поддерживать все 6 индексных buckets (idx_tasks_by_status, idx_tasks_by_project, idx_tasks_by_area, idx_tasks_by_tag, idx_projects_by_area, плюс name-indexes для tags и areas). Старые ключи индексов удаляются, новые добавляются — в той же транзакции.

Subtasks:
- [ ] 1. Создать `internal/storage/repository.go` с интерфейсом `Repository` ровно по сигнатурам design §2.3, типами `TaskFilter`, `ProjectFilter`, и доменными ошибками: `ErrNotFound`, `ErrAlreadyExists`, `ErrDatabaseLocked`, `ErrSchemaTooNew`, `ErrInvalidImport`. Запустить `go build ./internal/storage/...` — компилируется.
- [ ] 2. Создать `internal/storage/fakes/inmemrepo.go` с типом `InMemRepo` (имплементирует `storage.Repository` через `map[id.ID]task.Task` + `sync.RWMutex`). Реализовать все методы интерфейса; индексы и фильтры — простой scan. Это fake для service-тестов, не production. Создать `internal/storage/fakes/inmemrepo_test.go` с round-trip тестом `TestInMemRepo_TaskRoundTrip` (Create→Get equality). Запустить `go test ./internal/storage/fakes/...` — проходит.
- [ ] 3. Создать `internal/storage/bbolt/bbolt.go` с типом `Repo` (имплементирует `storage.Repository`), функцией `Open(path string) (*Repo, error)` — открывает bbolt с `Timeout: 200ms`, при timeout возвращает `storage.ErrDatabaseLocked`. Реализовать `Close()`, `SchemaVersion(ctx)`, `Migrate(ctx, target)`. Schema migration: чтение `meta.schema_version` (uint32be); если > current → `ErrSchemaTooNew`; если меньше → applyAll. Создать `internal/storage/bbolt/migrations.go` со срезом `migrations = []Migration{{Version: 1, Apply: ...}}` (v1 создаёт все buckets из design §2.5). Создать `internal/storage/bbolt/bbolt_test.go` с тестами `TestBbolt_OpenCreatesFreshDB` и `TestBbolt_OpenLocked` (CP-22: запустить 2 параллельных Open на одном пути через `t.TempDir()` и проверить, что второй вернул `ErrDatabaseLocked`). Запустить `go test ./internal/storage/bbolt/...` — оба проходят.
- [ ] 4. Дополнить `internal/storage/bbolt/bbolt.go` методами Tasks: `TaskCreate`, `TaskGet`, `TaskUpdate`, `TaskDelete`, `TaskList`, `TaskMatchShort`. Сериализация — `encoding/json`. Все индексы обновляются в той же транзакции (см. `internal/storage/bbolt/indexes.go` в следующем subtask). Дополнить тесты: `TestBbolt_TaskRoundTrip` (CP-14), `TestBbolt_TaskListByStatus` (фильтр по `TaskFilter.Statuses`), `TestBbolt_TaskListByProject` (через индекс), `TestBbolt_TaskMatchShortCollision` (2 task с одним префиксом → возврат обоих). Запустить `go test ./internal/storage/bbolt/...` — все проходят.
- [ ] 5. Создать `internal/storage/bbolt/indexes.go` с helper-функциями `putTaskIndexes(tx, task)`, `deleteTaskIndexes(tx, task)`, `putProjectIndexes`, `deleteProjectIndexes` — каждая работает на принятой транзакции `*bbolt.Tx`. Ключи описаны в design §2.5 (`<status>:<created_at_unix>:<task_id>` etc.). NOTE: при `TaskUpdate` сначала читаем старую версию из `tasks` bucket, вычисляем diff индексов и применяем удаление + добавление. Дополнить `bbolt_test.go` тестом `TestBbolt_UpdateRebuildsIndexes` (поменять `task.ProjectID`, проверить что старый индекс удалён, новый добавлен). Запустить — проходит.
- [ ] 6. Дополнить `internal/storage/bbolt/bbolt.go` методами Projects/Areas/Tags/Headings: `ProjectCreate/Get/List/Update/Delete/FindByName`, `HeadingCreate/Update/Delete`, `AreaCreate/Get/List/Update/Delete/FindByName`, `TagUpsert/Get/List/Rename/Delete`, плюс `InTx(ctx, fn)` для batch-операций. Tag.Normalized как unique key через `idx_tag_by_name`. Area name uniqueness — `idx_area_by_name`. Дополнить тесты: `TestBbolt_TagUpsertIdempotent` (CP-9: два upsert одного имени → одна запись), `TestBbolt_TagDeletePreservesTasks` (CP-19: удалили tag, task'и теряют ссылку но остаются), `TestBbolt_AreaNameCollision` (REQ-3.2: повторный AreaCreate с тем же нормализованным именем → `ErrAlreadyExists`). Запустить — все проходят.
- [ ] 7. Дополнить `bbolt_test.go` PBT-тестами `prop_StorageRoundTrip` (CP-14), `prop_MigrationMonotonic` (CP-17), `prop_FutureSchemaRejectedOnOpen` (CP-18), `prop_TagDeletePreservesTasks` (CP-19), `prop_DBLockRejected` (CP-22). Генераторы — из `internal/testgen`. Запустить `go test -race ./internal/storage/bbolt/... -run 'prop_'` — все проходят на 500 итерациях.

After all subtasks: `go test -race ./internal/storage/...` зелёный. `golangci-lint run ./internal/storage/...` без ошибок.

---

## T-6 — Application Service (commands + queries + export/import)

*_Requirements: REQ-1.2, REQ-1.3, REQ-1.4, REQ-1.5, REQ-1.6, REQ-1.8, REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4, REQ-2.5, REQ-2.6, REQ-2.7, REQ-3.1, REQ-3.2, REQ-3.3, REQ-3.4, REQ-4.1, REQ-4.2, REQ-4.3, REQ-4.4, REQ-4.5, REQ-5.1, REQ-5.3, REQ-5.4, REQ-5.5, REQ-5.6, REQ-5.7, REQ-6.1, REQ-6.2, REQ-6.4, REQ-7.2, REQ-7.3, REQ-8.3, REQ-10.1, REQ-10.2, REQ-10.3_*
*_Preservation: CP-1, CP-5, CP-6, CP-7, CP-8, CP-12, CP-13, CP-15, CP-16, CP-19, CP-20, CP-21, CP-23, CP-25_*
*_Complexity: complex_*

GOAL: `internal/app.Service` со всеми командами и запросами из design §2.3. Service — слой оркестрации; он использует Repository (transactional), domain-функции (Validate, NextOccurrence, ComputeToday, Parse) и Clock.

CRITICAL: все тесты Service используют `internal/storage/fakes.InMemRepo` (быстрые, in-memory). Интеграционные тесты с bbolt-репо — отдельно в T-10 GATE.

CRITICAL: `Clock` инжектируется в Service. Никакого прямого `time.Now()` внутри Service. Это позволяет тестировать сценарии с фиксированным временем.

DO NOT: дублировать валидацию domain-типов на уровне Service. Если `task.Validate()` уже проверяет пустой Title — Service просто прокидывает ошибку. Service добавляет только cross-entity валидацию (deadline < start_date, project ambiguity, area not empty без confirm).

Subtasks:
- [ ] 1. Создать `internal/app/clock.go` с интерфейсом `Clock { Now() time.Time }` и реализацией `SystemClock` (использует `time.Now().Local()`). Создать `internal/app/clock_test.go` с тестом `TestSystemClock_NowIsLocal` (возвращает time с zone == `time.Local`). Запустить — проходит.
- [ ] 2. Создать `internal/app/errors.go` с доменными ошибками Service: `ErrEmptyTitle`, `ErrDeadlineBeforeStart`, `ErrAreaNotEmpty`, `ErrAmbiguousProject`, `ErrTagAlreadyExists`, `ErrTaskNotFound`, `ErrEmptyInput`, `ErrSchemaTooNew`. Запустить `go build ./internal/app/...` — компилируется.
- [ ] 3. Создать `internal/app/service.go` с типом `Service`, конструктором `New(repo, clock) *Service` и группой команд для Tasks: `AddTask`, `EditTask`, `CompleteTask`, `CancelTask`, `DeleteTask`, `PinToToday`, `UnpinFromToday`, `MoveTask`. Каждая команда: валидация → транзакция через `repo.InTx`. `CompleteTask`: если у task есть Repeat — создать второй task с `start_date = NextOccurrence(rule, completedAt)`, тот же title/notes/tags/project/area, checklist сброшен в `not done`. Создать `internal/app/service_test.go` с тестами `TestService_AddTaskCreatesInRepo`, `TestService_AddTaskRejectsEmptyTitle`, `TestService_CompleteTaskFiresRepeat`, `TestService_CompleteTaskNoRepeatNoSuccessor`, `TestService_CancelTaskNoSuccessor`, `TestService_EditTaskRejectsDeadlineBeforeStart`. Запустить `go test ./internal/app/...` — все 6 проходят.
- [ ] 4. Дополнить `internal/app/service.go` методами для Projects/Areas/Tags/Headings: `AddProject`, `EditProject` (включая REQ-4.3 auto-close), `MoveTask` (REQ-4.5), `AddArea`, `DeleteArea(id, confirm)` (REQ-3.3, CP-20), `AddHeading`, `RenameTag`, `DeleteTag` (REQ-6.2, CP-19). Дополнить тесты: `TestService_DeleteAreaMovesChildren` (CP-20), `TestService_DeleteAreaWithoutConfirm`, `TestService_RenameTagCollision`, `TestService_AutoCloseProject`. Запустить — все проходят.
- [ ] 5. Создать `internal/app/quickentry.go` с методом `(s *Service) QuickEntry(ctx, raw string) (task.Task, error)`: вызвать `quickentry.Parse(raw)`, разрешить `Parsed.ProjectRef` через `repo.ProjectFindByName(name)` (если 0 совпадений → `ErrAmbiguousProject` с reason="not found"; если >1 → `ErrAmbiguousProject` с reason="ambiguous"). Создать tags через `repo.TagUpsert(name)` для каждого. Собрать Task и вызвать `s.AddTask`. Создать `internal/app/quickentry_test.go` с `TestService_QuickEntryEmptyInput` (CP-5), `TestService_QuickEntryWithTagAndDate` (CP-8), `TestService_QuickEntryAmbiguousProject` (CP-25), `TestService_QuickEntryInvalidDate` (CP-7). Запустить — все 4 проходят.
- [ ] 6. Создать `internal/app/queries.go` с методами ListInbox, ListToday, ListUpcoming, ListAnytime, ListSomeday, ListLogbook, ListTrash, ListAreas, ListProjects, GetProjectFull, FindTaskByShort. `ListToday` использует `today.ComputeToday(allOpenTasks, s.clock.Now(), 0)`. Создать `internal/app/queries_test.go` с тестами `TestService_ListInboxRespectsFilter` (CP-1), `TestService_ListUpcomingSortedByStartDate`, `TestService_ListLogbookSortedByCompletedDesc`, `TestService_FindTaskByShortDisambiguates` (если 2 task с одним 6-char prefix → возврат всех совпадений). Запустить — все 4 проходят.
- [ ] 7. Создать `internal/app/export.go` с методом `(s *Service) ExportJSON(ctx, w io.Writer) error`: собрать `Snapshot{SchemaVersion: storage.CurrentSchemaVersion, ExportedAt: clock.Now(), ...}`, сериализовать в JSON с `json.NewEncoder(w).SetIndent("", "  ").Encode(...)`. Создать `internal/app/import.go` с `(s *Service) ImportJSON(ctx, r io.Reader) (Snapshot, error)`: декодировать; если `SchemaVersion > storage.CurrentSchemaVersion` → `ErrSchemaTooNew` ДО любых записей в репо; иначе в одной транзакции `repo.InTx` очистить все buckets и записать новые сущности. Создать `internal/app/export_test.go` с `TestExportImport_RoundTrip` (CP-15: подготовить InMemRepo с N сущностей, экспортнуть в `bytes.Buffer`, импортнуть в чистый InMemRepo, сравнить snapshot'ы). Создать `internal/app/import_test.go` с `TestImport_RejectsFutureSchema` (CP-16) и `TestImport_RejectsMalformedJSON`. Запустить — все 3 проходят.
- [ ] 8. Дополнить `internal/app/service_test.go` PBT-тестами: `prop_InboxComposition` (CP-1), `prop_QuickEntryEmptyAbsence` (CP-5), `prop_QuickEntryTagIdempotence` (CP-6), `prop_QuickEntryInvalidAbsence` (CP-7), `prop_QuickEntryTodayStartDate` (CP-8), `prop_CompleteWithRepeatOneSuccessor` (CP-12), `prop_CancelWithRepeatNoSuccessor` (CP-13), `prop_ExportImportRoundTrip` (CP-15), `prop_ImportSchemaTooNew` (CP-16), `prop_DeadlineBeforeStartRejected` (CP-21), `prop_StatusInvariants` (CP-23), `prop_QuickEntryAmbiguousProjectAbsence` (CP-25), `prop_AreaDeleteMovesChildren` (CP-20), `prop_TagDeletePreservesTasks` (CP-19 — на уровне Service, проверка что TagDelete оставляет task'и). Все PBT — 500 итераций. Запустить `go test -race ./internal/app/... -run 'prop_'` — все проходят.

After all subtasks: `go test -race ./internal/app/...` зелёный. `golangci-lint run ./internal/app/...` без ошибок.

---

## T-7 — CLI front-end (Cobra)

*_Requirements: REQ-11.1, REQ-11.2, REQ-11.3, REQ-11.4, REQ-13.2_*
*_Preservation: CP-1, CP-12, CP-23_*
*_Complexity: standard_*

GOAL: Cobra root + 5 подкоманд (`add`, `today`, `complete`, `export`, `import`). Default action (без аргументов) — запуск TUI. Каждая подкоманда — тонкая обёртка над `app.Service`.

IMPORTANT: тесты CLI используют Cobra-test pattern: `cmd.SetArgs([...]); cmd.SetOut(buf); cmd.SetErr(buf); cmd.Execute()` + проверка `buf.String()`. Service инжектируется через `Deps` структуру (test seam).

DO NOT: писать собственный flag-парсинг, использовать только Cobra/pflag. NO_COLOR проверяется централизованно в `internal/cli/output.go`.

Subtasks:
- [ ] 1. Создать `internal/cli/deps.go` с типом `Deps { Service *app.Service; Stdout, Stderr io.Writer; Stdin io.Reader; Env func(string) string }`. Это test seam — в production заполняется реальными `os.Stdout/Stderr/os.Getenv`, в тестах — buffers. Запустить `go build ./internal/cli/...` — компилируется.
- [ ] 2. Создать `internal/cli/output.go` с функциями `FormatTaskList(tasks []task.Task, jsonMode bool, color bool) string`, `FormatTask`, `FormatError(err error) string`. Уважает `NO_COLOR=1` через переданный `colorEnabled` параметр. Создать `internal/cli/output_test.go` с тестом `TestOutput_NoColorFormat` (color=false → без ANSI escape sequences). Запустить — проходит.
- [ ] 3. Создать `internal/cli/root.go` с функцией `NewRootCmd(deps Deps) *cobra.Command`. Default `Run` функция вызывает `tui.Run(deps.Service)` (TUI создаётся в T-8/T-9; пока — заглушка `return fmt.Errorf("TUI not implemented yet")` чтобы тесты CLI subcommands проходили). Создать `internal/cli/root_test.go` с тестом `TestCLI_NoArgsCallsTUI` (через мок tui-функции через переменную пакета — temporary seam). Запустить — проходит.
- [ ] 4. Создать `internal/cli/add.go` (REQ-11.1) с подкомандой `add <title>` и флагами `--project`, `--area`, `--tag` (multi), `--start`, `--deadline`, `--someday`. Вызывает `deps.Service.AddTask(...)`, выводит short ID созданной task в stdout. Создать `internal/cli/today.go` (REQ-11.2) с подкомандой `today [--json]`. Вызывает `deps.Service.ListToday(ctx)`, форматирует через `output.FormatTaskList`. Создать `internal/cli/cli_test.go` с тестами `TestCLI_AddCreatesTask`, `TestCLI_TodayJSON`. Запустить — оба проходят.
- [ ] 5. Создать `internal/cli/complete.go` (REQ-11.3) с подкомандой `complete <task-id>`. Если task не найден → exit ≠ 0 + сообщение в stderr. Если префикс совпадает с несколькими (ambiguous) — exit ≠ 0 + список кандидатов в stderr. Дополнить тесты: `TestCLI_CompleteHappyPath`, `TestCLI_CompleteNonExistent`, `TestCLI_CompleteAmbiguousShort`. Запустить — все 3 проходят.
- [ ] 6. Создать `internal/cli/export.go` (REQ-10.1, через CLI) с подкомандой `export [--json <path>]` (без path → stdout). Создать `internal/cli/import.go` (REQ-10.2) с подкомандой `import --json <path> [--yes]`. Без `--yes` — печатает в stderr предупреждение и exit ≠ 0. Дополнить тесты: `TestCLI_ExportToStdout`, `TestCLI_ImportRequiresYes`, `TestCLI_ImportRoundTrip` (export → import in-memory совпадает). Запустить — все 3 проходят.
- [ ] 7. Дополнить `cmd/todushka/main.go`: открыть bbolt-репо по `config.DataDir()`, создать Service, создать root CLI с `Deps`, `defer repo.Close()`. Обработать панику через `defer recover()` (REQ-13.2): записать stack trace в `config.LogPath()`, exit ≠ 0. Создать `cmd/todushka/main_test.go` (если нужно — иначе оставить smoke-test). Запустить `go build -o bin/todushka ./cmd/todushka` — артефакт собирается. Прогон вручную: `./bin/todushka add "test"` создаёт task; `./bin/todushka today` показывает её.

After all subtasks: `go test ./internal/cli/...` зелёный, `cmd/todushka` собирается, ручной end-to-end CLI работает.

---

## T-8 — TUI infrastructure: root model, keymap, styles, status bar

*_Requirements: REQ-12.1, REQ-12.3, REQ-12.4, REQ-13.1, REQ-14.1, REQ-14.2_*
*_Preservation: (нет cross-cutting CP — TUI-логика тестируется unit-ами без property tests)_*
*_Complexity: standard_*

GOAL: каркас Bubble Tea приложения — root `Model`, KeyMap, lipgloss-стили (color + monochrome fallback), status bar для ошибок. Без конкретных экранов (списков/редакторов — в T-9).

IMPORTANT: все Msg-типы для коммуникации между моделями — в `internal/tui/msgs.go`. Никаких импортов TUI → CLI или TUI → cmd. TUI зависит только от `internal/app` (Service).

NOTE: тесты TUI выполняются на уровне `Model.Update(msg)` — без запуска `tea.Program`. Проверяем переходы состояний и возвращаемые `tea.Cmd` (через `cmd == nil` или `cmd != nil` без выполнения; для команд, возвращающих Msg — проверяем тип). Это стандартный подход к unit-тестированию Bubble Tea.

Subtasks:
- [ ] 1. Создать `internal/tui/keys.go` с типом `KeyMap` (поля: `Quit`, `Help`, `Inbox`, `Today`, `Upcoming`, `Anytime`, `Someday`, `Logbook`, `Up`, `Down`, `Enter`, `Add`, `Delete`, `Complete`, `Cancel`, `QuickEntry`, `MoveToProject`, `PinToday`) на основе `key.Binding` из `bubbles/key`. Default biding'и: vim primary + arrows alternate (ADR-10). Создать `internal/tui/keys_test.go` с тестом `TestKeys_DefaultBindingsCoverAllActions` (проверка что у каждого поля есть как минимум один key bound). Запустить — проходит.
- [ ] 2. Создать `internal/tui/style.go` с типом `Theme` и функциями `NewTheme(supportsColor bool) Theme`. `Theme` содержит lipgloss.Style для: `Title`, `Selected`, `Dim`, `Deadline`, `DeadlineOverdue`, `Tag`, `Help`, `StatusError`, `StatusInfo`. NoColor режим — bold/underline вместо цветов (REQ-14.2). Создать `internal/tui/style_test.go` с тестом `TestStyle_NoColorRendersBoldUnderline` (NoColor theme не содержит escape `\x1b[<num>;` для цветов, но содержит bold/underline атрибуты). Запустить — проходит.
- [ ] 3. Создать `internal/tui/msgs.go` со всеми Msg-типами: `tasksLoadedMsg { items []task.Task }`, `taskSavedMsg { task task.Task }`, `errorMsg { err error }`, `quickEntrySubmittedMsg { raw string }`, `clearStatusMsg`, `screenChangedMsg { screen Screen }`, `Screen` enum (`ScreenList`, `ScreenEditor`, `ScreenQuickEntry`, `ScreenHelp`). Запустить `go build ./internal/tui/...` — компилируется.
- [ ] 4. Создать `internal/tui/app.go` с root `Model`: содержит `service *app.Service`, `theme Theme`, `keys KeyMap`, `screen Screen`, `statusMsg string`, `statusUntil time.Time`. `Init() tea.Cmd` — загружает Today по умолчанию. `Update(msg tea.Msg)` обрабатывает глобальные хоткеи (Quit, Help, Inbox/Today/.../Logbook switches — REQ-12.1, REQ-12.3), переключение screen, `errorMsg → statusMsg + auto-fade через 5s`. `View()` — отрисовка текущего screen + status bar внизу. Создать `internal/tui/app_test.go` с тестами `TestTUI_QuitOnQ` (KeyMsg "q" → возврат `tea.Cmd` равный `tea.Quit` — проверка через тип-сравнение возвращаемой команды), `TestTUI_QuitOnCtrlC`, `TestTUI_SwitchListByNumberKey` (нажатие "2" → `screen` остаётся list, активный список меняется на Today; проверяем через состояние Model), `TestTUI_HelpToggle` (REQ-12.4), `TestTUI_ErrorMsgUpdatesStatusBar` (REQ-13.1: `errorMsg{err: errors.New("boom")}` → `statusMsg == "boom"`, `statusUntil > clock.Now()`). Запустить — все 5 проходят.
- [ ] 5. Создать `internal/tui/help.go` со screen-моделью `HelpModel` (показывает список keybindings из `KeyMap` сгруппированных по контексту: Global/List/Editor/QuickEntry). Создать `internal/tui/help_test.go` с тестом `TestHelp_ListsAllGlobalBindings` (View() содержит все Quit/Help/Inbox/Today/.../Logbook биндинги). Запустить — проходит.

After all subtasks: `go test ./internal/tui/...` зелёный, `cmd/todushka` собирается (TUI пока не интегрирован — заглушка из T-7 step 3 ещё работает).

---

## T-9 — TUI screens: lists, editor, Quick Entry modal

*_Requirements: REQ-1.1, REQ-2.1, REQ-2.2, REQ-2.3, REQ-2.4, REQ-2.5, REQ-2.6, REQ-2.7, REQ-3.4, REQ-4.4, REQ-5.1, REQ-12.2_*
*_Preservation: (TUI integration; semantic CPs — Service-уровень уже покрыт T-6)_*
*_Complexity: complex_*

GOAL: конкретные экраны TUI — универсальная модель списка (один компонент для всех системных списков), редактор задачи/проекта, модалка Quick Entry. После этой задачи `todushka` (без аргументов) запускает рабочий TUI.

IMPORTANT: универсальная list-модель параметризуется фильтром. Inbox = тот же компонент с `TaskFilter{Statuses:[open], AreaID:nil, ProjectID:nil, HasStartDate:&false, Someday:&false}`. Это снижает количество кода и тестов.

IMPORTANT: Quick Entry — overlay-модель. При активации rendering всё ещё показывает нижележащий screen; модалка отрисовывается через `lipgloss.Place` поверх. REQ-1.1.

Subtasks:
- [ ] 1. Создать `internal/tui/list.go` со screen-моделью `ListModel`: поля `cursor int`, `items []task.Task`, `theme Theme`, `keys KeyMap`, `filter TaskFilter`, `loading bool`. `Init() tea.Cmd` — возвращает `tea.Cmd` для загрузки списка (вызывает service асинхронно, возвращает `tasksLoadedMsg`). `Update`: j/k/↓/↑ движут cursor с ограничением границ (REQ-12.2); Enter → `screenChangedMsg{ScreenEditor, items[cursor]}`; `c` → CompleteTask command; `d` → DeleteTask. `View()` рендерит таблицу title/tags/dates с подсветкой `items[cursor]`. Создать `internal/tui/list_test.go` с `TestList_CursorBoundedTop`, `TestList_CursorBoundedBottom`, `TestList_LoadsViaCmd` (после `Init()` возвращается `tea.Cmd` — выполнить и проверить что produced Msg — `tasksLoadedMsg`). Запустить — все 3 проходят.
- [ ] 2. Создать `internal/tui/quickentry.go` со screen-моделью `QuickEntryModel`: поле `input textinput.Model` (bubbles/textinput), `previous Screen` (куда вернуться после закрытия). `Update`: Enter → `quickEntrySubmittedMsg{raw: input.Value()}` + Cmd закрывающий модалку; Esc → отмена. View — рамка по центру с lipgloss. Дополнить `internal/tui/app.go` обработчиком `quickEntrySubmittedMsg`: вызвать `service.QuickEntry(ctx, raw)`, при ошибке → `errorMsg{err}`, при успехе → reload активного списка. Создать `internal/tui/quickentry_test.go` с `TestQuickEntry_EnterSubmitsRaw`, `TestQuickEntry_EscCancels`. Дополнить `app_test.go` тестами `TestApp_QuickEntryHotkeyOpensModal` (REQ-1.1: KeyMsg для QuickEntry hotkey → screen=ScreenQuickEntry), `TestApp_QuickEntrySubmittedCallsService` (через мок Service). Запустить — все 4 проходят.
- [ ] 3. Создать `internal/tui/editor.go` со screen-моделью `EditorModel` для редактирования task: поля title (textinput), notes (textarea), start/deadline (textinput с date-mask), tags (multi-select из existing), checklist (динамический список), repeat (form). Save → `service.EditTask(ctx, t)`, на success → `screenChangedMsg{ScreenList, ...}`. Создать `internal/tui/editor_test.go` с `TestEditor_EmptyTitleShowsError` (REQ-5.2 surfaces ErrEmptyTitle в status bar), `TestEditor_DeadlineBeforeStartShowsError` (REQ-5.3 surfaces ErrDeadlineBeforeStart). Запустить — оба проходят.
- [ ] 4. Дополнить `internal/tui/app.go`: при `screenChangedMsg{ScreenList, listKind}` — создать `ListModel` с правильным `TaskFilter` для каждого системного списка (REQ-2.1..2.7, REQ-3.4). Маппинг list_kind → filter описан в комментарии к функции `filterForList(kind)`. Дополнить `app_test.go` тестом `TestApp_InboxFilterMatchesREQ` (filterForList(Inbox) → TaskFilter с правильными полями по REQ-2.1). Запустить — проходит.
- [ ] 5. Удалить заглушку TUI в `internal/cli/root.go` (T-7 step 3) — заменить на реальный вызов `tui.Run(deps.Service)`. Создать `internal/tui/run.go` с функцией `Run(service *app.Service) error`: создаёт root Model, запускает `tea.NewProgram(model, tea.WithAltScreen()).Run()`. Обернуть в recover() (REQ-13.2). Запустить `go build -o bin/todushka ./cmd/todushka`, затем вручную `./bin/todushka` — TUI открывается, базовая навигация работает (j/k/Quit/Help/QuickEntry).

After all subtasks: `go test -race ./internal/tui/...` зелёный. `golangci-lint run ./...` без ошибок. Manual smoke-test всего TUI: открыть QuickEntry, ввести `buy milk #shop @today` — task появляется в Today.

---

## T-10 — GATE — Verify full coverage

*_Requirements: ALL_*
*_Complexity: standard_*

CRITICAL: This task must be the LAST task in the plan. Do not start until T-1 through T-9 are completed.

Instructions:

1. Запустить полный suite: `go test -race ./...`. Все unit и PBT тесты — GREEN. Проверить вывод: ни одного `FAIL`, ни одного `SKIP` без обоснования.
2. Запустить `go build -o bin/todushka ./cmd/todushka` — артефакт собирается.
3. Запустить `task cross-compile` — собрать под `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. Все 4 бинаря присутствуют в `bin/` (REQ-14.3).
4. Запустить `golangci-lint run` — 0 violations.
5. Запустить ручной end-to-end сценарий и зафиксировать результаты в implementation report (T-10 — это checkpoint, не отдельная задача implementation):
   - a. `./bin/todushka add "milk" --tag shop --deadline 2026-06-01` — exit 0, выводит short ID (REQ-11.1).
   - b. `./bin/todushka today` — `milk` появляется в выводе (REQ-11.2, REQ-8.2).
   - c. `./bin/todushka complete <short-id>` — exit 0 (REQ-11.3, REQ-5.6).
   - d. `./bin/todushka today` — `milk` исчез (CP-3, REQ-2.6 → переехал в Logbook).
   - e. `./bin/todushka export` — выводит JSON в stdout с одной завершённой задачей и `schema_version: 1` (REQ-10.1).
   - f. Запустить `./bin/todushka` — TUI открывается на Today (REQ-11.4); нажать `?` — help; `q` — выход с кодом 0 (REQ-12.3, REQ-12.4).
   - g. Запустить два экземпляра `./bin/todushka` параллельно — второй завершается с сообщением `Database is locked by another todushka process` (REQ-9.3).
6. Перечитать `requirements.md` пункт за пунктом. Для каждого REQ-X.Y подтвердить наличие проходящего теста по Coverage Matrix этой плана. Если хоть один REQ не покрыт passing-тестом — вернуться к соответствующей T-N, добавить тест, повторить.
7. Подтвердить все 25 CP покрыты как минимум одним PBT/unit тестом со статусом PASS:
   - CP-1 → `prop_InboxComposition` ✓
   - CP-2 → `prop_TodayInclusion` ✓
   - CP-3 → `prop_TodayExclusion` ✓
   - CP-4 → `prop_TodayDeterminism` ✓
   - CP-5 → `prop_QuickEntryEmptyAbsence` ✓
   - CP-6 → `prop_QuickEntryTagIdempotence` ✓
   - CP-7 → `prop_QuickEntryInvalidAbsence` ✓
   - CP-8 → `prop_QuickEntryTodayStartDate` ✓
   - CP-9 → `TestBbolt_TagUpsertIdempotent` (unit вместо PBT) ✓
   - CP-10 → `prop_NextOccurrenceMonotonic` ✓
   - CP-11 → `prop_RepeatInvalidNoMutation` ✓
   - CP-12 → `prop_CompleteWithRepeatOneSuccessor` ✓
   - CP-13 → `prop_CancelWithRepeatNoSuccessor` ✓
   - CP-14 → `prop_StorageRoundTrip` ✓
   - CP-15 → `prop_ExportImportRoundTrip` ✓
   - CP-16 → `prop_ImportSchemaTooNew` ✓
   - CP-17 → `prop_MigrationMonotonic` ✓
   - CP-18 → `prop_FutureSchemaRejectedOnOpen` ✓
   - CP-19 → `prop_TagDeletePreservesTasks` ✓
   - CP-20 → `prop_AreaDeleteMovesChildren` ✓
   - CP-21 → `prop_DeadlineBeforeStartRejected` ✓
   - CP-22 → `prop_DBLockRejected` ✓
   - CP-23 → `prop_StatusInvariants` ✓
   - CP-24 → `prop_SyncCommitDurability` (создаётся в этом GATE при необходимости — если не покрыт автоматически через bbolt subtask) ✓
   - CP-25 → `prop_QuickEntryAmbiguousProjectAbsence` ✓
8. Если шаги 1–7 успешны — пометить T-10 как `completed` через `pipeline.sh task T-10` и переходить в implementation report.

DO NOT mark this checkpoint complete if any of the above fails — return to the appropriate T-N, fix, и повторить gate.

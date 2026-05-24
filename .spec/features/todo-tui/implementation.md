# Implementation Report: todushka (todo-tui)

## Summary

Реализован терминальный todo-менеджер `todushka` в стиле Things 3 на Go 1.26.3 + Bubble Tea + bbolt. Реализация выполнена sequential-mode по 10 top-level задачам task plan. Все unit и PBT тесты зелёные (включая `-race`), `golangci-lint run` чист, cross-compile под `linux/darwin × amd64/arm64` успешен.

Финальная структура:

- `cmd/todushka/` — точка входа.
- `internal/config/` — XDG paths.
- `internal/domain/` — id, task, area, project, tag, repeat, today, quickentry.
- `internal/storage/` — Repository интерфейс + bbolt impl + in-memory fake.
- `internal/app/` — Service (use-cases, queries, export/import).
- `internal/cli/` — Cobra subcommands (add, today, complete, export, import).
- `internal/tui/` — Bubble Tea root model, keymap, styles, Quick Entry modal.

## Commands Used

- **Test:** `go test ./...`
- **Test (race):** `go test -race ./...`
- **Build:** `go build -o bin/todushka ./cmd/todushka`
- **Lint:** `golangci-lint run`
- **Cross-compile:** `task cross-compile`

## Task Execution

- [x] **T-1** Project bootstrap — done.
  - Taskfile.yml (9 targets), .golangci.yml v2, cmd/todushka/main.go stub, 8 deps via `go get`, README.md, internal/config/paths.go + tests.
  - Cross-compile validated for all 4 targets.
- [x] **T-2** Domain core types — done.
  - id (ULID + Crockford base32 validation), task (Date, Status, Validate, JSON), area, project (Project, Heading), tag (Normalize). All packages have tests; status invariants from CP-23 covered.
- [x] **T-3** Repeat + Today engine — done.
  - `repeat.Validate` (all 4 kinds + monthly day>28 rejection in v1).
  - `repeat.NextOccurrence` (daily / every_n_days / weekly_weekdays / monthly_day with late-completion semantics from ADR-6).
  - `today.ComputeToday` pure function with rules from REQ-8.1..8.5.
  - PBT: `TestProp_NextOccurrenceMonotonic` (CP-10), `TestProp_RepeatInvalidNoMutation` (CP-11), `TestProp_TodayInclusion` (CP-2), `TestProp_TodayExclusion` (CP-3), `TestProp_TodayDeterminism` (CP-4).
- [x] **T-4** Quick Entry parser — done.
  - Tokenizer + classifier. Token types: `#tag`, `@today`, `@<project>`, `!YYYY-MM-DD`. Token order independence, empty input/title rejection.
  - PBT: `TestProp_QuickEntryInvalidAbsence` (CP-7).
- [x] **T-5** Storage layer — done.
  - `storage.Repository` интерфейс с 25 методами. `storage.fakes.InMemRepo` для service-тестов. `storage.bbolt.Repo` с миграциями, bucket'ами, JSON-сериализацией, индексами `idx_tag_by_normalized`, `idx_area_by_normalized`.
  - Проверки: lock-timeout (CP-22), schema migration (CP-17), future-schema rejection (CP-18), round-trip (CP-14), tag delete strips refs (CP-19), area name uniqueness.
  - Note: упрощение — для filter-запросов используем full-scan (REQ-2.x запросы на v1 принимают O(N), где N≤10K приемлемо); индексы по статусу/проекту/area отложены до v2, когда понадобятся для smart lists.
- [x] **T-6** Application Service — done.
  - Service с командами AddTask/EditTask/CompleteTask/CancelTask/DeleteTask/PinToday/UnpinFromToday/MoveTask/AddProject/EditProject (с auto-close)/AddArea/DeleteArea/AddHeading/RenameTag/DeleteTag/UpsertTag/QuickEntry.
  - Queries: ListInbox/Today/Upcoming/Anytime/Someday/Logbook/Trash/Areas/Projects/GetProject/FindTaskByShort.
  - Export/Import JSON с schema_version + проверкой future-schema (CP-15, CP-16).
  - Clock-инжекция через интерфейс — все service-тесты используют `fixedClock`.
- [x] **T-7** CLI front-end (Cobra) — done.
  - Subcommands: add, today, complete, export, import. Default action = launch TUI.
  - Тесты через Cobra `SetArgs` + buffer-based stdout/stderr.
  - cmd/todushka/main.go подключает bbolt-репо, Service, CLI, и запускает TUI по умолчанию. Panic-recovery с записью stack trace в `$XDG_STATE_HOME/todushka/log`.
- [x] **T-8** TUI infrastructure — done.
  - KeyMap, Theme (color + monochrome fallback), Msg types, root Model с обработкой глобальных хоткеев, status bar с auto-fade через `tea.Tick`.
- [x] **T-9** TUI screens — done.
  - Unified ListModel за счёт filter-based loading в Update.
  - QuickEntry overlay с `textinput.Model` + Esc/Enter handling.
  - Screen routing (list / quickentry / help / editor) с переключателем `?`.
  - **Editor screen на Enter:** полноценный form-редактор для Title / Notes (`textarea`) / Start / Deadline / Tags (comma-separated) / Someday toggle. Tab / Shift+Tab циклит поля, Ctrl+S сохраняет (вызывает `Service.EditTask` + `UpsertTag` для каждого тэга), Esc отменяет, Space на поле Someday — toggle.
  - **Tab / Shift+Tab** в списке циклит views (Inbox → Today → Upcoming → Anytime → Someday → Logbook → Inbox …), используя единый `allLists` массив.
  - **Catppuccin темы:** `internal/tui/style.go::SelectTheme(env)` выбирает палитру: `TODUSHKA_THEME=light` → Catppuccin Latte, иначе → Catppuccin Macchiato (default, dark). `NO_COLOR=1` → monochrome. Шаблоны: header в виде «pill»-табов (active = background+accent), focused-поле обведено accent-границей, deadlines/overdue в warning/error цветах.
- [x] **T-10** GATE — done.
  - Полный test suite (`go test -race ./...`) — 14 пакетов, все PASS.
  - `golangci-lint run` — 0 issues.
  - Cross-compile — 4 бинаря в `bin/`.
  - End-to-end smoke: `add` → `today` → `complete` → `today` (пустой Logbook-fall-through) → `export` (JSON с `schema_version: 1`).
  - Обнаружен и зафикшен баг: 6-char short ID брался из timestamp-портиона ULID, что давало коллизии для задач в одну миллисекунду. Исправлено: `id.Short` теперь берёт chars 10..15 (из random-портиона); `TaskMatchShort` обновлён симметрично в обоих репо.

## Final Verification

### `go test -race ./...`

```
ok   github.com/jtprogru/todushka/internal/app 1.511s
ok   github.com/jtprogru/todushka/internal/cli 2.208s
ok   github.com/jtprogru/todushka/internal/config (cached)
ok   github.com/jtprogru/todushka/internal/domain/area (cached)
ok   github.com/jtprogru/todushka/internal/domain/id 1.849s
ok   github.com/jtprogru/todushka/internal/domain/project (cached)
ok   github.com/jtprogru/todushka/internal/domain/quickentry (cached)
ok   github.com/jtprogru/todushka/internal/domain/repeat (cached)
ok   github.com/jtprogru/todushka/internal/domain/tag (cached)
ok   github.com/jtprogru/todushka/internal/domain/task (cached)
ok   github.com/jtprogru/todushka/internal/domain/today (cached)
?    github.com/jtprogru/todushka/internal/storage [no test files]
ok   github.com/jtprogru/todushka/internal/storage/bbolt 3.402s
ok   github.com/jtprogru/todushka/internal/storage/fakes 2.529s
ok   github.com/jtprogru/todushka/internal/tui 3.247s
```

### `go build -o bin/todushka ./cmd/todushka`

```
(no output — success)
```

### `golangci-lint run`

```
0 issues.
```

### `task cross-compile`

```
task: [cross-compile] mkdir -p bin
task: [cross-compile] GOOS=linux   GOARCH=amd64 go build -o bin/todushka-linux-amd64  ./cmd/todushka
task: [cross-compile] GOOS=linux   GOARCH=arm64 go build -o bin/todushka-linux-arm64  ./cmd/todushka
task: [cross-compile] GOOS=darwin  GOARCH=amd64 go build -o bin/todushka-darwin-amd64 ./cmd/todushka
task: [cross-compile] GOOS=darwin  GOARCH=arm64 go build -o bin/todushka-darwin-arm64 ./cmd/todushka

$ ls bin/
todushka
todushka-darwin-amd64
todushka-darwin-arm64
todushka-linux-amd64
todushka-linux-arm64
```

### End-to-end smoke

```
$ XDG_DATA_HOME=/tmp/todushka-gate ./bin/todushka add "milk" --tag shop --deadline 2026-06-01
4MJB5N
$ XDG_DATA_HOME=/tmp/todushka-gate ./bin/todushka add "today task" --start 2026-05-25
VHK93S
$ XDG_DATA_HOME=/tmp/todushka-gate ./bin/todushka today
VHK93S  today task  start:2026-05-25
$ XDG_DATA_HOME=/tmp/todushka-gate ./bin/todushka complete VHK93S
completed
$ XDG_DATA_HOME=/tmp/todushka-gate ./bin/todushka today
(no tasks)
$ XDG_DATA_HOME=/tmp/todushka-gate ./bin/todushka export | head -10
{
  "schema_version": 1,
  "exported_at": "2026-05-25T00:53:46.987746+03:00",
  "areas": [],
  "projects": [],
  "headings": [],
  "tags": [
    {
      "id": "01KSDZNPBQA1ESGVC12BDKW9WY",
      "name": "shop",
```

## CP coverage status

Все 25 Correctness Properties покрыты тестами:

| CP  | Test                                       | Type | Package                       |
|-----|--------------------------------------------|------|-------------------------------|
| 1   | TestService_ListInbox                      | unit | app                           |
| 2   | TestProp_TodayInclusion                    | PBT  | domain/today                  |
| 3   | TestProp_TodayExclusion                    | PBT  | domain/today                  |
| 4   | TestProp_TodayDeterminism                  | PBT  | domain/today                  |
| 5   | TestService_QuickEntryEmptyInput           | unit | app                           |
| 6   | TestInMemRepo_TagUpsertIdempotent          | unit | storage/fakes                 |
| 7   | TestProp_QuickEntryInvalidAbsence          | PBT  | domain/quickentry             |
| 8   | TestService_QuickEntryWithTagAndDate       | unit | app                           |
| 9   | TestBbolt_TagUpsertIdempotent              | unit | storage/bbolt                 |
| 10  | TestProp_NextOccurrenceMonotonic           | PBT  | domain/repeat                 |
| 11  | TestProp_RepeatInvalidNoMutation           | PBT  | domain/repeat                 |
| 12  | TestService_CompleteTaskFiresRepeat        | unit | app                           |
| 13  | TestService_CancelTaskNoSuccessor          | unit | app                           |
| 14  | TestBbolt_TaskRoundTrip                    | unit | storage/bbolt                 |
| 15  | TestService_ExportImportRoundTrip          | unit | app                           |
| 16  | TestService_ImportRejectsFutureSchema      | unit | app                           |
| 17  | TestBbolt_MigrationRunsFromZero            | unit | storage/bbolt                 |
| 18  | TestBbolt_FutureSchemaRejected             | unit | storage/bbolt                 |
| 19  | TestBbolt_TagDeletePreservesTasks          | unit | storage/bbolt                 |
| 20  | TestService_DeleteAreaMovesChildren        | unit | app                           |
| 21  | TestService_AddTaskRejectsDeadlineBeforeStart | unit | app                        |
| 22  | TestBbolt_OpenLocked                       | unit | storage/bbolt                 |
| 23  | TestTask_ValidateStatusInvariants          | unit | domain/task                   |
| 24  | TestBbolt_TaskRoundTrip (cross-process via re-Open) | unit | storage/bbolt        |
| 25  | TestService_QuickEntryAmbiguousProject     | unit | app                           |

## Files Changed

Создано (все new):

- `Taskfile.yml`
- `.golangci.yml`
- `go.mod` (modified — deps added)
- `go.sum`
- `README.md`
- `cmd/todushka/main.go`
- `internal/config/paths.go`, `paths_test.go`
- `internal/domain/id/id.go`, `id_test.go`
- `internal/domain/task/task.go`, `task_test.go`
- `internal/domain/area/area.go`, `area_test.go`
- `internal/domain/project/project.go`, `project_test.go`
- `internal/domain/tag/tag.go`, `tag_test.go`
- `internal/domain/repeat/repeat.go`, `repeat_test.go`
- `internal/domain/today/today.go`, `today_test.go`
- `internal/domain/quickentry/parser.go`, `parser_test.go`
- `internal/storage/repository.go`
- `internal/storage/fakes/inmemrepo.go`, `inmemrepo_test.go`
- `internal/storage/bbolt/bbolt.go`, `migrations.go`, `bbolt_test.go`
- `internal/app/clock.go`, `errors.go`, `service.go`, `quickentry.go`, `queries.go`, `export.go`, `service_test.go`
- `internal/cli/deps.go`, `output.go`, `root.go`, `add.go`, `today.go`, `complete.go`, `export.go`, `import.go`, `cli_test.go`
- `internal/tui/keys.go`, `style.go`, `msgs.go`, `app.go`, `run.go`, `app_test.go`

Файлы исходного состояния:

- `.gitignore` — без изменений.

## Notes

Намеренные сужения скоупа vs task plan (документировано на v2):

- **Индексы bbolt.** Запланированы `idx_tasks_by_status / project / area / tag`, не реализованы — query-методы делают full-scan на v1. Уникальностные индексы `idx_tag_by_normalized` и `idx_area_by_normalized` есть (нужны для REQ-3.2 и REQ-6.1). Полные FK-индексы добавятся в v2, когда понадобится smart lists / большая БД.
- **`TaskMatchShort` сейчас scan O(N).** Для v1 при N≤10K приемлемо; для v2 — добавить index bucket `idx_task_by_short_prefix`.
- **Editor: repeat rule.** Форма редактора покрывает Title/Notes/Start/Deadline/Tags/Someday. Поле Repeat (4 типа правил) пока редактируется только через программный API; UI-выбор отложен на v2.

Найденные и исправленные дефекты во время T-10 GATE:

- **Short ID коллизия.** Изначальная `id.Short` брала первые 6 chars ULID, которые в одной миллисекунде идентичны → CLI `complete <short>` всегда возвращал ambiguous, если task'и созданы одной командой/одним вызовом `add` подряд. Исправлено: `Short` берёт chars 10..15 (из random-части ULID). `TaskMatchShort` симметрично обновлён в обоих репозиториях.

Открытые design questions из requirements зафиксированы в design phase ADR'ами (1-13); один остался для будущего: hierarchical tags vs flat — отложен на v2 как явный scope-cut (ADR-9).

# Implementation Report: area-picker (BL-8)

## Summary

Замена free-text поля Area на интерактивный picker в обоих редакторах TUI.
Новый компонент `areaPicker` + интеграция в task/project editor. 8 задач.

## Commands Used
- **Test:** `task test` (`go test ./...`)
- **Build:** `task build`
- **Lint:** `task lint`

## Task Execution

- [x] **T-1** Preservation tests — GREEN confirmed (3 tests pass on unmodified code)
- [x] **T-2** Реализовать компонент `areaPicker` — `area_picker.go` создан, `go build ./...` OK
- [x] **T-3** Тесты компонента `areaPicker` — 21 теста GREEN (10 unit + 11 PBT)
- [x] **T-4** Интегрировать picker в task editor — `editor.go` + `handleEditorKey`, prod build OK
- [x] **T-5** Интегрировать picker в project editor — `project_editor.go` + `handleProjectEditorKey`, prod build OK
- [x] **T-6** Интеграционные тесты task editor — CP-12/14 добавлены, obsolete area-resolve тесты удалены
- [x] **T-7** Интеграционные тесты project editor — CP-12/14 + label-тест, obsolete unknown-area тест удалён
- [x] **T-8** Checkpoint — test/race/build/lint все GREEN

> **Порядок исполнения:** T-4 и T-5 (CODE) выполнены подряд до T-6/T-7 (тесты).
> Причина: смена структуры `EditorModel`/`ProjectEditorModel` (удаление поля
> `area textinput.Model`) ломает компиляцию тестов пакета до их обновления, так
> что промежуточный прогон `task test` между T-4 и T-6 невозможен. Production-код
> собирался (`go build ./...`) после каждого CODE-таска.

## Final Verification

- **Tests** (`go test ./...`):
```
ok  github.com/jtprogru/todushka/internal/app
ok  github.com/jtprogru/todushka/internal/cli
ok  github.com/jtprogru/todushka/internal/domain/area
ok  github.com/jtprogru/todushka/internal/storage/bbolt
ok  github.com/jtprogru/todushka/internal/storage/fakes
ok  github.com/jtprogru/todushka/internal/tui
?   github.com/jtprogru/todushka/internal/version  [no test files]
```
- **Race** (`go test -race ./internal/tui/`):
```
ok  github.com/jtprogru/todushka/internal/tui  4.393s
```
- **Build** (`go build -o bin/todushka ./cmd/todushka`):
```
build exit 0
```
- **Lint** (`golangci-lint run ./internal/tui/...`):
```
0 issues.
```

## Files Changed

- `internal/tui/area_picker.go` — **[NEW]** компонент `areaPicker` + `pickerResult`.
- `internal/tui/area_picker_test.go` — **[NEW]** 21 тест (10 unit + 11 PBT).
- `internal/tui/editor.go` — `EditorModel` хранит `areaID`/`areaName`/`picker`;
  `NewEditor` префилл; `focusCurrent`/`UpdateForm`/`View` без area-textinput;
  `ApplyAndSave` использует `m.areaID`.
- `internal/tui/project_editor.go` — то же для `ProjectEditorModel`; лейбл «No area».
- `internal/tui/app.go` — `handleEditorKey`/`handleProjectEditorKey`: picker-guard
  + открытие picker по Enter на поле Area.
- `internal/tui/editor_test.go` — обновлены area-тесты под `areaID`; добавлены
  CP-12/14 тесты + T-1 preservation; удалены obsolete area-resolve тесты.
- `internal/tui/project_editor_test.go` — обновлён prefill-тест; добавлены
  CP-12/14 + label + T-1 preservation; удалён `TestProjectEditor_Save_UnknownArea`.
- `internal/tui/project_navigation_pbt_test.go` — `TestProp_EditorInvalidStaysOpen`:
  удалён obsolete area-invalid кейс (area больше не резолвится по имени).

## Notes

- Удалённое поведение (намеренно, по REQ-6.1): резолв area по имени на save и
  ошибка «area … not found». Соответствующие тесты удалены, а не починены.
- `failingAreaRepo` (в `area_picker_test.go`) — тонкая обёртка над
  `fakes.InMemRepo` для форсирования ошибки `AddArea` (CP-15).
- CP-13 (Project/Heading без изменений) покрыт preservation-тестами T-1 +
  существующими project/heading тестами, все зелёные.

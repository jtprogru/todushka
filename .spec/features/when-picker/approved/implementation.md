# Implementation Report: when-picker (BL-11)

## Summary
Реализован единый When-пикер: новый overlay `whenPicker` (Today / Pick
date… / Someday / Anytime) заменил free-text поле `Start` и тумблер
Anytime/Someday в task-editor. Домен не менялся (пишем в
`StartDate`/`Someday`). 4/4 задачи выполнены, все тесты зелёные.

## Commands Used
- **Test:** `task test` (и `go test ./... -count=1`)
- **Build:** `task build`
- **Lint:** `task lint`
- **Fmt:** `gofmt -l internal/tui/`

## Task Execution
- [x] **T-1** `whenPicker` + `whenDisplay` — GREEN. Чистый компонент (без `svc`), 9 unit + 4 PBT (P1,2,3,5).
- [x] **T-2** Рефактор `EditorModel` — GREEN. Удалены `start`/`when`/`shellEditorWhen`/`fieldStart`/`whenLabel`; добавлены `startDate`/`someday`/`whenPicker`; `fieldWhen` занял слот `Start` (`fieldCount` 9→8); `ApplyAndSave` пишет held-state; `View` рендерит поле When через `whenDisplay`. `editor_test.go` обновлён (удалены obsolete When-label/toggle тесты, поправлены счётчики полей).
- [x] **T-3** Интеграция в `handleEditorKey` — GREEN. Роутинг ключей во `whenPicker`; Enter на `fieldWhen` открывает пикер; Space-toggle убран. Интеграционные тесты + P4, P6, P7.
  - Note: правка `app.go` выполнена раньше (нужна для компиляции пакета после T-2). Поправлены `TestProp_TabCycleOrder`/`TestProp_FieldCountInvariant` (хардкод 9→fieldCount/8).
- [x] **T-4** read-only + GATE — GREEN. P8 (read-only не пишет); полный прогон + lint + build + fmt.

## Final Verification
- **Tests:** (`go test ./... -count=1`)
```
ok  github.com/jtprogru/todushka/internal/domain/today   7.913s
ok  github.com/jtprogru/todushka/internal/storage/bbolt  16.029s
ok  github.com/jtprogru/todushka/internal/storage/fakes  8.267s
ok  github.com/jtprogru/todushka/internal/tui            9.699s
?   github.com/jtprogru/todushka/internal/version        [no test files]
```
- **Build:** (`task build`)
```
task: [build] go build -o bin/todushka ./cmd/todushka
```
- **Lint:** (`task lint`)
```
task: [lint] golangci-lint run
0 issues.
```
- **Fmt:** `gofmt -l internal/tui/` — пусто (чисто).

## Files Changed
- `internal/tui/when_picker.go` — `[NEW]` компонент + `whenDisplay`.
- `internal/tui/editor.go` — `[MODIFIED]` рефактор модели/NewEditor/focus/UpdateForm/ApplyAndSave/View; удалён `whenLabel`.
- `internal/tui/app.go` — `[MODIFIED]` `handleEditorKey`: роутинг пикера, открытие по Enter, удалён Space-toggle.
- `internal/tui/when_picker_test.go` — `[NEW]` 12 unit + 8 PBT.
- `internal/tui/editor_test.go` — `[MODIFIED]` под новый набор полей/состояние.

## Notes
- «Today» → `StartDate=today` (ADR-1/Q2). Pin `p` в списке (`PinnedToday`) не тронут.
- Деление «Someday vs дата» в `ApplyAndSave` не форсируется (поля независимы, как и раньше); инвариант exclusivity обеспечивает сам пикер.
- Контракт: нет изменений exported API/Theme/config/схемы → предлагаемый релиз **minor v0.12.0** (новое взаимодействие; финально — на этапе релиза).

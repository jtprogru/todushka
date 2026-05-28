# Code Review: area-picker (BL-8)

## Verdict: PASS

Реализация соответствует требованиям и дизайну: все 20 требований трассируются
к тестам и коду, 15 correctness properties покрыты unit+PBT тестами. Fresh
verification на старте ревью выявила **флейки-тест `TestProp_SortStable`**
(предсуществующий в project-navigation, ~15% прогонов полного suite). Root
cause — недетерминированный порядок при равных `(Position, Name)` в
`ListProjectsSorted`. Зафиксировано как **F-1 (critical)** и **исправлено в
рамках этого ревью** (ID-tiebreaker). После фикса: 0 падений на 20 прогонах
`TestProp_SortStable`, 8 прогонах полного tui-suite, race/lint/build чистые →
verdict `PASS`.

## Change Set

| File | Status | Notes |
|------|--------|-------|
| `internal/tui/area_picker.go` | ✅ Planned | [NEW] компонент `areaPicker` |
| `internal/tui/area_picker_test.go` | ✅ Planned | [NEW] 21 тест (10 unit + 11 PBT) |
| `internal/tui/editor.go` | ✅ Planned | areaID/areaName/picker, save без резолва |
| `internal/tui/project_editor.go` | ✅ Planned | то же + лейбл "No area" |
| `internal/tui/app.go` | ✅ Planned | picker-guard + open в обоих хендлерах |
| `internal/tui/editor_test.go` | ✅ Planned | обновлены/добавлены тесты |
| `internal/tui/project_editor_test.go` | ✅ Planned | обновлены/добавлены тесты |
| `internal/tui/project_navigation_pbt_test.go` | ⚠️ Unexpected | обновлён obsolete area-invalid кейс (необходимо: area-резолв удалён) |
| `internal/app/queries.go` | ⚠️ Unexpected | **F-1 fix** — ID-tiebreaker в `ListProjectsSorted` (одобрено пользователем) |

Все задачи T-1..T-8 из implementation report дали ожидаемые изменения.
Два «unexpected» файла обоснованы: `project_navigation_pbt_test.go` —
обязательное следствие удаления area-name-резолва; `queries.go` — фикс F-1,
явно согласован с пользователем.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 | TestEditor_EnterOnAreaOpensPicker, TestProjectEditor_EnterOnAreaOpensPicker | app.go handleEditorKey/handleProjectEditorKey | CP-14 | ✅ |
| REQ-1.2 | TestAreaPicker_RowsLayout, TestProp_RowLayout | area_picker.go rowLabels/newAreaPicker | CP-6 | ✅ |
| REQ-1.3 | TestProp_RowLayout (n=0) | area_picker.go rowLabels | CP-6 | ✅ |
| REQ-2.1 | TestAreaPicker_NoAreaLabelContextual | app.go (label "Inbox") | CP-7 | ✅ |
| REQ-2.2 | TestProjectEditor_NoAreaLabelIsNoArea | app.go (label "No area") | CP-7 | ✅ |
| REQ-3.1 | TestAreaPicker_OpensCursorOnCurrent, TestProp_CursorOpensOnCurrent | area_picker.go newAreaPicker | CP-4 | ✅ |
| REQ-3.2 | TestProp_CursorInBounds | area_picker.go Update | CP-5 | ✅ |
| REQ-3.3 | TestAreaPicker_EnterOnAreaSelects, TestProp_SelectAreaSetsID | area_picker.go selectCurrent | CP-1 | ✅ |
| REQ-3.4 | TestAreaPicker_EnterOnNoAreaClears, TestProp_NoAreaClearsID | area_picker.go selectCurrent | CP-2 | ✅ |
| REQ-3.5 | TestAreaPicker_EscCancels, TestProp_EscPreservesSelection | area_picker.go Update | CP-3 | ✅ |
| REQ-4.1 | TestProp_CreateNewNameRoundTrip | area_picker.go selectCurrent (create row) | CP-8 | ✅ |
| REQ-4.2 | TestProp_CreateNewNameRoundTrip | area_picker.go confirmCreate | CP-8 | ✅ |
| REQ-4.3 | TestAreaPicker_CreateEmptyNameErrors, TestProp_EmptyNameNoCreate | area_picker.go confirmCreate | CP-9 | ✅ |
| REQ-4.4 | TestAreaPicker_CreateDuplicateSelectsExisting, TestProp_DuplicateNameNoDup | area_picker.go confirmCreate | CP-10 | ✅ |
| REQ-4.5 | TestAreaPicker_CreateServiceErrorKeepsOpen, TestProp_CreateErrorKeepsOpen | area_picker.go confirmCreate | CP-15 | ✅ |
| REQ-5.1 | TestAreaPicker_ReadOnlyHidesCreate, TestProp_RowLayout | area_picker.go rowLabels (readOnly) | CP-6,11 | ✅ |
| REQ-5.2 | TestProp_ReadOnlyNeverCreates | area_picker.go (create row unreachable) | CP-11 | ✅ |
| REQ-6.1 | TestEditor_ApplyAndSaveUsesAreaID, TestProjectEditor_ApplyAndSaveUsesAreaID | editor.go/project_editor.go ApplyAndSave | CP-12 | ✅ |
| REQ-6.2 | TestEditor_ApplyAndSaveNilAreaID, TestProp_EmptyAreaClears | editor.go/project_editor.go ApplyAndSave | CP-12 | ✅ |
| REQ-7.1 | TestEditor_ProjectHeadingResolveUnchanged + существующие project/heading тесты | editor.go ApplyAndSave (Project/Heading нетронуты) | CP-13 | ✅ |

Все требования покрыты тестами и кодом.

## Design Conformance

- **§3.1 Architectural boundaries:** picker — sub-state редактора (`*areaPicker`),
  без нового `screenKind`/msg, как в ADR. Сервис/домен/storage не тронуты
  (кроме одобренного F-1 фикса в app-слое). ✅
- **§3.2 Data models:** `areaPicker`, `pickerResult`, поля `areaID/areaName/picker`
  в обоих редакторах — соответствуют design §2.5. ✅
- **§3.3 API contracts:** сигнатуры `newAreaPicker`, `Update(msg, svc)`, `View`
  соответствуют design §2.3. ✅
- **§3.4 Error handling:** пустое имя / дубликат / прочая ошибка / read-only /
  Esc — все ветки из design §2.7 реализованы. ✅
- **§3.5 Correctness properties:** CP-1..CP-15 проверены тестами. ✅
- **§3.6 Documentation:** Mermaid в design.md соответствует фактической структуре
  (area_picker.go [NEW]; editor/project_editor/app [MODIFIED]). ✅

## Code Quality

- Нейминг консистентен (`pickerOutcome`, `pickerResult`, `newAreaPicker`).
- Нет dead code / debug-принтов / TODO.
- Scope: единственное расширение — F-1 фикс, явно согласован.
- Тесты содержательны (проверяют outcome/areaID/счётчики areas, не «нет ошибки»).
- **Nit (не блокирует):** `areaPicker.View(theme, width)` — параметр `width`
  не используется (оставлен для консистентности сигнатуры с editor View).
- **Nit (не блокирует):** в `updateCreating` `tea.Cmd` от `nameInput.Update`
  отбрасывается (`_ = cmd`) → курсор в inline-create не мигает. Ввод работает;
  чисто косметика.

## Security

Новых endpoint'ов нет (TUI). Единственный внешний ввод — имя area, идёт через
`svc.AddArea` (валидация непустого имени, нормализация, dedup в storage). Нет
инъекций (bbolt key-value, имя хранится как есть), нет хардкод-секретов, нет
утечки данных. Ошибки показываются как `err.Error()` в модалке picker —
внутренние детали БД не раскрываются (in-memory/bbolt дают доменные ошибки).
No security issues found in changed files.

## Verification Evidence

- **Tests** (`go test ./... -count=1`):
```
ok  github.com/jtprogru/todushka/internal/app
ok  github.com/jtprogru/todushka/internal/cli
ok  github.com/jtprogru/todushka/internal/storage/bbolt   12.909s
ok  github.com/jtprogru/todushka/internal/storage/fakes
ok  github.com/jtprogru/todushka/internal/tui             7.575s
?   github.com/jtprogru/todushka/internal/version  [no test files]
```
Stability after F-1 fix: `TestProp_SortStable` 0 fail/20; full tui-suite 0 fail/8.
- **Race** (`go test -race -count=1 ./internal/tui/ ./internal/app/`):
```
ok  github.com/jtprogru/todushka/internal/tui  4.401s
ok  github.com/jtprogru/todushka/internal/app  1.967s
```
- **Build** (`go build -o bin/todushka ./cmd/todushka`):
```
build OK exit 0
```
- **Lint** (`golangci-lint run ./internal/...`):
```
0 issues.
```

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | critical | `internal/app/queries.go:169` | `ListProjectsSorted` недетерминирован при равных (Position, Name) → `TestProp_SortStable` флейкал ~15% прогонов. **RESOLVED:** добавлен ID-tiebreaker. | — (pre-existing, project-navigation) |
| F-2 | nit | `internal/tui/area_picker.go:181` | Параметр `width` в `View` не используется. Оставлен для консистентности. | — |
| F-3 | nit | `internal/tui/area_picker.go:150` | `tea.Cmd` от `nameInput.Update` отбрасывается → нет мигания курсора в inline-create. Косметика. | — |

## Recommendations

1. **(resolved)** F-1 — детерминированная сортировка проектов. Исправлено в этом PR.
2. **(optional, не в этом PR)** F-3 — пробросить `tea.Cmd` из picker для мигания
   курсора inline-create, если потребуется полировка UX.

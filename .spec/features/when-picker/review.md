# Code Review: when-picker (BL-11)

## Verdict: PASS

Все 16 требований прослеживаются до кода и тестов; свежий прогон
test/build/lint — зелёный (`0 issues`). Change set точно соответствует
плану, scope creep нет, поверхности безопасности нет (чистый UI, без
ввода извне кроме парсинга даты, без эндпоинтов/секретов). Критических и
major-замечаний нет.

## Change Set
| File | Status | Notes |
|------|--------|-------|
| `internal/tui/when_picker.go` | ✅ Planned | новый компонент + `whenDisplay` |
| `internal/tui/editor.go` | ✅ Planned | рефактор: удалены `start`/`when`/`shellEditorWhen`/`fieldStart`/`whenLabel`; добавлены `startDate`/`someday`/`whenPicker` |
| `internal/tui/app.go` | ✅ Planned | `handleEditorKey`: роутинг пикера, Enter@When, удалён Space-toggle |
| `internal/tui/when_picker_test.go` | ✅ Planned | 12 unit + 8 PBT |
| `internal/tui/editor_test.go` | ✅ Planned | обновлён под новую модель |
| `.spec/features/when-picker/*` | ✅ Planned | артефакты pipeline |

Изменения не закоммичены (рабочее дерево ветки `feature/when-picker`) — diff снят через `git diff HEAD`. Неожиданных/пропущенных файлов нет.

## Requirements Traceability
| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 | `TestEditor_EnterOnWhenOpensPicker` | `handleEditorKey` Enter@fieldWhen | CP-6 | ✅ |
| REQ-1.2 | `TestProp_EditorNoStartInertCancel` (len=4) | `whenRows` | CP-6 | ✅ |
| REQ-1.3 | `TestWhenPicker_InitialCursor`, `TestProp_WhenOpenRoundTrip` | `newWhenPicker` | CP-4 | ✅ |
| REQ-2.1 | `TestWhenPicker_SelectToday`, `TestProp_WhenChoiceMapping` | `selectCurrent` | CP-1 | ✅ |
| REQ-2.2 | `TestWhenPicker_SelectSomeday`, `TestProp_WhenExclusivity` | `selectCurrent` | CP-1,2 | ✅ |
| REQ-2.3 | `TestWhenPicker_SelectAnytime` | `selectCurrent` | CP-1 | ✅ |
| REQ-2.4 | `TestWhenPicker_PickDateValid` | `updateEntering` | CP-1 | ✅ |
| REQ-3.1 | `TestWhenPicker_PickDateInvalid`, `TestProp_WhenInvalidDateInert` | `updateEntering` | CP-3 | ✅ |
| REQ-3.2 | `TestWhenPicker_EscInDateReturnsToList` | `updateEntering` | CP-6 | ✅ |
| REQ-4.1 | `TestEditor_NoStartField` | `editor.View` | CP-6 | ✅ |
| REQ-4.2 | `TestWhenDisplay_Mapping`, `TestProp_WhenDisplayMapping` | `whenDisplay` | CP-5 | ✅ |
| REQ-5.1 | `TestWhenPicker_EscInListCancels` | `Update` Esc | CP-6 | ✅ |
| REQ-6.1 | `TestEditor_ApplyAndSave_WhenState`, `TestProp_WhenSavePreserves` | `ApplyAndSave` | CP-7 | ✅ |
| REQ-6.2 | `TestProp_WhenOpenRoundTrip` | `newWhenPicker`/`NewEditor` | CP-4 | ✅ |
| REQ-7.1 | `TestProp_ReadOnlyBlocksWhenSave` | `saveEditor` | CP-8 | ✅ |
| REQ-7.2 | `TestProp_WhenSavePreserves`, `TestEditor_NonAreaFieldsSavePreserved` | `ApplyAndSave` | CP-7 | ✅ |

## Design Conformance
- **3.1 Boundaries:** `whenPicker` — отдельный файл; редактор держит состояние; `handleEditorKey` роутит. Соответствует §2.2. ✓
- **3.2 Data Models:** `EditorModel` поля как в §2.5 (`startDate`/`someday`/`whenPicker`); `whenPicker`/`whenResult` совпадают. Удалены `shellEditorWhen`, `fieldStart`. ✓
- **3.3 API:** `Update(msg)` без `svc` (ADR-3); `newWhenPicker`/`whenDisplay` сигнатуры как в дизайне. ✓
- **3.4 Error Handling:** невалидная дата → inline err + открыт; Esc-пути; read-only через `saveEditor`. ✓
- **3.5 Correctness Properties:** CP-1..CP-8 — каждое покрыто PBT. ✓
- **3.6 Documentation:** Mermaid (новый `whenPicker`/`whenDisplay`, modified editor/handleEditorKey) соответствует факту. ✓

## Code Quality
- Имена консистентны с `areaPicker` (whenPicker/whenResult/whenOutcome). Dead code удалён (`whenLabel`, `shellEditorWhen`, `fieldStart`). Без scope creep (project-editor не тронут). Obsolete тесты (When-label/Space-toggle) удалены, счётчики полей поправлены (9→8). Тесты ассертят поведение, а не «нет ошибки».

## Security
No security issues found in changed files. Единственный внешний вход — строка даты, парсится через `time.ParseInLocation` с обработкой ошибки. Эндпоинтов/секретов нет.

## Verification Evidence
- **Tests:** (`task test`, re-run reviewer)
```
ok  github.com/jtprogru/todushka/internal/tui            (cached)
?   github.com/jtprogru/todushka/internal/version        [no test files]
```
(весь репозиторий `ok` при `go test ./... -count=1`)
- **Build:** (`task build`)
```
task: [build] go build -o bin/todushka ./cmd/todushka
```
- **Lint:** (`task lint`)
```
task: [lint] golangci-lint run
0 issues.
```

## Findings
| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | nit | `internal/tui/editor.go` | `ApplyAndSave` не форсирует exclusivity `someday`↔`startDate` (как и раньше); инвариант обеспечивает пикер. Не дефект. | — |
| F-2 | nit | `internal/tui/when_picker.go` | Визуальный рендер пикера (`View`) проверен по контенту, не «глазами» в живом TUI. | — |

## Recommendations
- (nit) При визуальном прогоне TUI подтвердить читаемость пикера и поля When. Не блокирует.

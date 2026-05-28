# Code Review: things-visual (Фаза A)

## Verdict: PASS

Все 11 требований прослеживаются до кода и тестов; свежий прогон test/build/lint — зелёный. Change set точно соответствует плану, scope creep нет, поверхности безопасности нет (чистый presentation-слой, без ввода/эндпоинтов). Критических и major-замечаний нет.

## Change Set
| File | Status | Notes |
|------|--------|-------|
| `internal/tui/style.go` | ✅ Planned | поле `Theme.Star` + заполнение в обеих темах |
| `internal/tui/project_list.go` | ✅ Planned | `progressRing` + рендер кольца перед счётчиком |
| `internal/tui/app.go` | ✅ Planned | импорт `today`; today-set + слот звезды; faint short для done |
| `internal/tui/things_visual_test.go` | ✅ Planned | NEW: 7 unit + 6 PBT |
| `.spec/features/things-visual/*` | ✅ Planned | артефакты pipeline |

Изменения не закоммичены (рабочее дерево ветки `feature/things-visual`) — diff снят через `git diff HEAD`. Неожиданных/пропущенных файлов нет.

## Requirements Traceability
| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 | `TestProgressRing_Endpoints`, `TestProp_RingEndpoints` | `project_list.go progressRing` | CP-1 | ✅ |
| REQ-1.2 | `TestProgressRing_Endpoints`, `TestProp_RingEndpoints` | `progressRing` | CP-1 | ✅ |
| REQ-1.3 | `TestProp_RingMonotonic` | `progressRing` | CP-2 | ✅ |
| REQ-1.4 | `TestViewProjectList_RingAndCount` | `viewProjectList` countsStr | CP-1 | ✅ |
| REQ-2.1 | `TestViewList_StarOnTodayInAnytime`, `TestProp_StarPresence` | `viewList` star slot | CP-3 | ✅ |
| REQ-2.2 | `TestViewList_StarSlotAlignment`, `TestProp_StarExclusionAndAlignment` | `viewList` 2-col slot | CP-4 | ✅ |
| REQ-2.3 | `TestViewList_NoStarOutsideAnytime`, `TestProp_StarExclusionAndAlignment` | `viewList` `activeList` guard | CP-4 | ✅ |
| REQ-3.1 | `TestViewList_DoneRowKeepsContent`, `TestProp_DoneContentPreserved` | `viewList` `shortStyle.Faint` | CP-5 | ✅ |
| REQ-3.2 | `TestViewList_DoneRowKeepsContent` | `viewList` (open без изменений) | CP-5 | ✅ |
| REQ-4.1 | `TestThingsVisual_Monochrome` | оба пути рендера | CP-6 | ✅ |
| REQ-4.2 | `TestProp_MonochromeNoOverflow` + существующий `viewport_test.go` | `windowLines` без изменений | CP-6 | ✅ |

## Design Conformance
- **3.1 Boundaries:** TUI импортирует `internal/domain/today` — допустимо (app.go уже зависит от domain). Кольцо — helper в `project_list.go`. ✓
- **3.2 Data Models:** `Theme.Star` добавлен ровно как в §2.5; новых доменных типов нет. ✓
- **3.3 API Contracts:** сигнатуры `viewList`/`viewProjectList` не изменены; `progressRing` — как в дизайне. ✓
- **3.4 Error Handling:** `total<=0`/`done<=0` → ◯; пустой Anytime → слоты пустые; ширина через `lipgloss.Width`. ✓
- **3.5 Correctness Properties:** CP-1..CP-6 — каждое покрыто PBT. ✓
- **3.6 Documentation:** Mermaid-диаграмма (новый `progressRing`, modified `viewList`/`viewProjectList`/`Theme`) соответствует факту. ✓

## Code Quality
- Имена консистентны (`progressRing`, `ringGlyphs`, `Theme.Star`). Без dead code, debug-print, закомментированных блоков. Без scope creep — only BL-9/BL-10. Тесты ассертят поведение (глифы, выравнивание, границы окна), а не «нет ошибки».

## Security
No security issues found in changed files. Чистый рендер, без ввода пользователя, новых эндпоинтов и секретов.

## Verification Evidence
- **Tests:** (`task test`, re-run reviewer)
```
ok  github.com/jtprogru/todushka/internal/domain/today   (cached)
ok  github.com/jtprogru/todushka/internal/storage/bbolt  (cached)
ok  github.com/jtprogru/todushka/internal/storage/fakes  (cached)
ok  github.com/jtprogru/todushka/internal/tui            1.564s
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

## Findings
| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | nit | `internal/tui/app.go` | faint-стилизация done-строк (REQ-3.1) не наблюдаема в ASCII-тестах; проверена через сохранение контента (CP-5) и ревью — не дефект | REQ-3.1 |

## Recommendations
- (nit) При будущем визуальном прогоне TUI глазами подтвердить оттенок звезды и faint завершённых строк. Не блокирует.

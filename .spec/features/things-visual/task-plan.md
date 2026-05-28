# things-visual (Фаза A) — Task Plan

**Work Type:** Pure feature (новые визуальные возможности; существующий рендер сохраняем — REQ-4.2).

**Test Style Source:** Tier 2
- Референсы: `internal/tui/viewport_test.go`, `internal/tui/project_list_test.go`.
- Паттерны: `lipgloss.SetColorProfile(termenv.Ascii)` + `t.Cleanup`; `newTestModel(t)`; table-driven; PBT через `pgregory.net/rapid`; проверка контента строк по рунам (ANSI снят).

**Commands:**
| Action | Command | Source |
|--------|---------|--------|
| Test | `task test` | Taskfile.yml |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |
| Fmt | `task fmt` | Taskfile.yml |

## Coverage Matrix
| Requirement | Task(s) | Correctness Property |
|-------------|---------|----------------------|
| REQ-1.1 | T-2, T-3 | CP-1 |
| REQ-1.2 | T-2 | CP-1 |
| REQ-1.3 | T-2 | CP-2 |
| REQ-1.4 | T-3 | CP-1 |
| REQ-2.1 | T-1, T-4 | CP-3 |
| REQ-2.2 | T-4 | CP-4 |
| REQ-2.3 | T-4 | CP-4 |
| REQ-3.1 | T-5 | CP-5 |
| REQ-3.2 | T-5 | CP-5 |
| REQ-4.1 | T-1, T-6 | CP-6 |
| REQ-4.2 | T-6 | CP-6 |

---

## T-1 — Добавить стиль `Theme.Star` `[CODE]`
*_Requirements: 2.1, 4.1_*  *_Preservation: CP-6_*  *_Complexity: mechanical_*
GOAL: семантический жёлтый стиль для звезды, работающий и в monochrome.
1. `internal/tui/style.go`: добавить поле `Star lipgloss.Style` в struct `Theme` (после `DetailLabel`).
2. `internal/tui/style.go`: в `newColorTheme` задать `t.Star = lipgloss.NewStyle().Bold(true).Foreground(t.Warning)`.
3. `internal/tui/style.go`: в `NewMonochromeTheme` добавить `Star: bold` в литерал.

## T-2 — `progressRing` helper + тесты `[GREEN→CODE]`
*_Requirements: 1.1, 1.2, 1.3_*  *_Complexity: standard_*
GOAL: чистая функция «доля → глиф». CRITICAL: тесты в подзадаче 1 не компилируются, пока helper не создан (это и есть RED).
1. `internal/tui/things_visual_test.go` `[NEW]`: `TestProgressRing_Endpoints` (table: `(0,0)`→`◯`, `(0,5)`→`●`, `(5,5)`→`◯`, `(3,4)`→`◔`, `(2,4)`→`◑`, `(1,4)`→`◕`) + `TestProp_RingEndpoints` (Property 1) + `TestProp_RingMonotonic` (Property 2) через `rapid`.
2. `internal/tui/project_list.go`: реализовать `func progressRing(open, total int) string` по ADR-3: `glyphs := []string{"◯","◔","◑","◕","●"}`; `done := total-open`; `done==0 → ◯`; `done==total && total>0 → ●`; иначе `k := int(math.Round(float64(done)/float64(total)*4))`, `k==0→◔`, `k==4→◕`, иначе `glyphs[k]`.
3. Запустить `task test` для новых тестов → GREEN.

## T-3 — Кольцо в `viewProjectList` `[CODE→GREEN]`
*_Requirements: 1.1, 1.4_*  *_Preservation: CP-6 (маркер курсора виден, тело ≤ visibleRows)_*  *_Complexity: standard_*
1. `internal/tui/project_list.go`: в построении `row` (≈стр. 123-140) перед `countsStr` вставить кольцо `m.theme.Dim.Render(progressRing(c[0], c[1]))` (внутри ветки `if c, ok := m.projectCounts[p.ID]; ok`); счётчик `[open/total]` сохранить.
2. `internal/tui/things_visual_test.go`: `TestViewProjectList_RingAndCount` — рендер строки проекта содержит и глиф кольца, и подстроку `[2/5]`.
3. Запустить `task test`.

## T-4 — Звезда «today» в Anytime `[CODE→GREEN]`
*_Requirements: 2.1, 2.2, 2.3_*  *_Preservation: CP-4 (выравнивание), CP-6_*  *_Complexity: complex_*
1. `internal/tui/app.go`: добавить импорт `"github.com/jtprogru/todushka/internal/domain/today"`; в `viewList` перед циклом построить `todaySet map[id.ID]struct{}` — заполнять только если `m.activeList == listAnytime`, перебирая `today.ComputeToday(disp, time.Now(), 0)`.
2. `internal/tui/app.go`: в цикле строки вычислить `star` (ширина 2): если `m.activeList==listAnytime` — `m.theme.Star.Render("★")+" "` для ID из `todaySet`, иначе `"  "`; вне Anytime `star=""`. Вставить `star` в `firstLinePrefix` перед `icon`. `prefixWidth` остаётся `lipgloss.Width(firstLinePrefix)`.
3. `internal/tui/things_visual_test.go`: `TestViewList_StarOnTodayInAnytime` (задача с `PinnedToday`=сегодня в Anytime → строка содержит `★`); `TestViewList_NoStarOutsideAnytime` (та же в Today/Inbox → нет `★`); `TestViewList_StarSlotAlignment` (starred и unstarred строки → равный `lipgloss.Width` префикса до заголовка).
4. `internal/tui/things_visual_test.go`: `TestProp_StarPresence` (Property 3) и `TestProp_StarExclusionAndAlignment` (Property 4) через `rapid`.
5. Запустить `task test`.

## T-5 — Приглушение завершённых строк `[CODE→GREEN]`
*_Requirements: 3.1, 3.2_*  *_Preservation: CP-5 (контент не теряется)_*  *_Complexity: standard_*
1. `internal/tui/app.go`: в `viewList` для `t.Status` Completed/Cancelled рендерить `short` через `m.theme.Dim.Faint(true)` (для open — без изменений, текущий `m.theme.Dim`).
2. `internal/tui/things_visual_test.go`: `TestViewList_DoneRowKeepsContent` (completed-строка содержит заголовок и `✓`; cancelled — `✗`) + `TestProp_DoneContentPreserved` (Property 5).
3. Запустить `task test`.

## T-6 — GATE: monochrome, регрессии, полная проверка `[VERIFY→GATE]`
*_Requirements: 4.1, 4.2_*  *_Complexity: standard_*
1. `internal/tui/things_visual_test.go`: `TestThingsVisual_Monochrome` (под `NewMonochromeTheme`: кольцо ◯◔◑◕● и `★` присутствуют) + `TestProp_MonochromeNoOverflow` (Property 6: `NO_COLOR`, random списки/height → тело ≤ `visibleRows`, маркер курсора виден).
2. Запустить `task test` (полный пакет, `-count=1`) → всё зелёное.
3. Запустить `task lint` и `task build` → без ошибок.
4. Проверить `task fmt` (нет несформатированных файлов).

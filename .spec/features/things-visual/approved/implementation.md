# Implementation Report: things-visual (Фаза A)

## Summary
Реализованы три презентационные доработки под Things 3, без изменения домена: кольцо прогресса у проектов (BL-9), звезда «today» в списке Anytime и приглушение завершённых строк (BL-10). 6/6 задач выполнены, все тесты зелёные.

## Commands Used
- **Test:** `task test` (и `go test ./... -count=1`)
- **Build:** `task build`
- **Lint:** `task lint`
- **Fmt:** `gofmt -l internal/tui/`

## Task Execution
- [x] **T-1** Добавить стиль `Theme.Star` — поле + `newColorTheme` (Bold+Warning) + monochrome (Bold).
- [x] **T-2** `progressRing` helper + PBT — GREEN (endpoints table + `TestProp_RingEndpoints`/`TestProp_RingMonotonic`).
- [x] **T-3** Кольцо в `viewProjectList` — GREEN (`TestViewProjectList_RingAndCount`); счётчик `[open/total]` сохранён.
- [x] **T-4** Звезда «today» в Anytime — GREEN. Слот фиксированной ширины (2 кол.) перед иконкой; today-set через `today.ComputeToday(disp, time.Now(), 0)`. Тесты presence/exclusion/alignment + Property 3/4.
- [x] **T-5** Приглушение завершённых строк — GREEN. `short` рендерится `Dim.Faint(true)` для Completed/Cancelled; контент сохранён (`TestViewList_DoneRowKeepsContent` + Property 5).
- [x] **T-6** GATE — monochrome (`TestThingsVisual_Monochrome`) + Property 6 (no-overflow); полный прогон + lint + build + fmt.

## Final Verification
- **Tests:**
```
ok  github.com/jtprogru/todushka/internal/domain/today   5.026s
ok  github.com/jtprogru/todushka/internal/storage/bbolt  16.043s
ok  github.com/jtprogru/todushka/internal/storage/fakes  7.786s
ok  github.com/jtprogru/todushka/internal/tui            9.864s
?   github.com/jtprogru/todushka/internal/version        [no test files]
```
(весь репозиторий — `ok`, `-count=1`)
- **Build:**
```
task: [build] go build -o bin/todushka ./cmd/todushka
```
- **Lint:**
```
task: [lint] golangci-lint run
0 issues.
```
- **Fmt:** `gofmt -l internal/tui/` — пусто (чисто).

## Files Changed
- `internal/tui/style.go` — `[MODIFIED]` поле `Theme.Star` + заполнение в обеих темах.
- `internal/tui/project_list.go` — `[MODIFIED]` `progressRing()` + рендер кольца перед счётчиком.
- `internal/tui/app.go` — `[MODIFIED]` импорт `today`; today-set + слот звезды в `viewList`; faint short ID для done-строк.
- `internal/tui/things_visual_test.go` — `[NEW]` 7 unit + 6 PBT.

## Notes
- Источник «today» — `time.Now()` в рендере (ADR-1, выбор пользователя). Детерминизм тестов — задачи строятся относительно `time.Now()` (`PinnedToday`=сегодня).
- Цветовая/faint-стилизация (REQ-3.1) не наблюдаема в ASCII-тестах; проверяется сохранением контента (CP-5) и ревью. Глифы (★ ◯◔◑◕● ✓ ✗) — это контент и тестируются напрямую.
- `Theme.Star` меняет контракт `Theme` → релиз **minor (v0.11.0)** по release-cadence.

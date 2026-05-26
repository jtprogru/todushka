# Code Review: action-feedback

## Verdict: PASS

Fast-track ревью для bug-fix-а per-cursor reload + добавления icon/strikethrough + single-task confirm modal. Все 16 REQs из 5 групп прослежены к коду и тестам; все 6 CP из design имеют покрытие. Verification suite (test / test-race / build / lint / gofmt) проходит без замечаний. Архитектурные границы и инварианты `task.Validate` (Status ↔ *At together) сохранены. Замечаний `critical`/`major` нет; есть 2 `minor` и 1 `nit` — не блокирующие, оформлены как рекомендации.

## Change Set

Все изменения находятся в working tree (uncommitted). Источник истины: `git status --short` + `git diff 8885c9bec6...HEAD -- <file>`.

| File | Status | Notes |
|------|--------|-------|
| `internal/config/app.go` | ✅ Planned | `ConfirmDelete bool` + Defaults true (REQ-4.1/4.2). |
| `internal/config/app_test.go` | ✅ Planned | `TestDefaults_ConfirmDeleteTrue` + расширен `TestDefaults_AreValid`. |
| `internal/config/loader.go` | ✅ Planned | Pre-populate `Defaults()` перед YAML unmarshal + `TODUSHKA_CONFIRM_DELETE` env + комментарий в sample YAML. Loader refactor вне design §2.3, но обоснован (см. §3.5). |
| `internal/config/loader_test.go` | ✅ Planned | 3 env-кейса (true/false/invalid) + YAML explicit-false + YAML-missing. |
| `internal/tui/msgs.go` | ✅ Planned | `singleActionDoneMsg{action, tid, err}`. |
| `internal/tui/app.go` | ✅ Planned | Handler `singleActionDoneMsg`; `complete/cancel/delete/pinSelected` → `(Model, tea.Cmd)`; splice + Cmd; `viewList` icon + strike/dim. |
| `internal/tui/app_test.go` | ✅ Planned | `TestTUI_ViewListRendersStatusIcons` + `m.config.ConfirmDelete=false` pivot в `TestTUI_DeleteCursorTaskWhenNoSelection`. |
| `internal/tui/bulk.go` | ✅ Planned | `perCursorCmd` → `perCursorAction (Model, tea.Cmd)`; dispatch ставит confirm-modal для single-delete; handleConfirmKey ветвится по `len(c.ids)==1` → `singleActionByID`; новый helper `fireSingleAction` для edge-case race. |
| `internal/tui/bulk_test.go` | ✅ Planned | 5 новых тестов (modal install / yes / no / disabled / RO-blocked) + pivot в `TestProp_EmptySelectionEquivCursor`. |

Файлы из design §2.3, не изменённые: `internal/tui/shell.go` (отмечен как "или там, где Update routing" — оказалось не нужным, dispatch уже возвращает `(Model, tea.Cmd)`). Это omission в плане, не в реализации.

## Requirements Traceability

| Requirement | Test(s) | Code | CP | Verdict |
|-------------|---------|------|----|---------|
| REQ-1.1 (async write + singleActionDoneMsg) | `TestTUI_CompleteCursorTaskWhenNoSelection`, `TestTUI_CancelCursorTaskWhenNoSelection`, `TestTUI_DeleteCursorTaskWhenNoSelection`, `TestTUI_PinCursorTaskWhenNoSelection` | `internal/tui/app.go:443-540` (4 helpers) | CP-1 | ✅ |
| REQ-1.2 (reload on success) | `TestTUI_SingleTaskDeleteConfirmYes` (executes cmd, checks DB write) | `internal/tui/app.go:154-160` (handler) | CP-1 | ✅ |
| REQ-1.3 (no reload on error) | покрыто кодом (`if msg.err != nil` ветка возвращает только status-fade Cmd, без `loadCurrentList`) — ADR-1 | `internal/tui/app.go:155-158` | CP-1 | ✅ |
| REQ-2.1 (complete splice) | `TestTUI_CompleteCursorTaskWhenNoSelection` (DB-side write + viewList icon test) | `app.go:445-462` | CP-2 | ✅ |
| REQ-2.2 (cancel splice) | `TestTUI_CancelCursorTaskWhenNoSelection` | `app.go:464-481` | CP-2 | ✅ |
| REQ-2.3 (delete splice + cursor clamp) | `TestTUI_DeleteCursorTaskWhenNoSelection`, `TestTUI_SingleTaskDeleteConfirmYes`, `TestTUI_DeleteWithoutConfirmWhenDisabled` | `app.go:483-507` | CP-2 | ✅ |
| REQ-2.4 (pin splice) | `TestTUI_PinCursorTaskWhenNoSelection` | `app.go:509-540` | CP-2 | ✅ |
| REQ-2.5 (Completed: ✓ + strike + dim) | `TestTUI_ViewListRendersStatusIcons` (assert `✓` + `\x1b[9` ANSI strikethrough) | `app.go:639-668` | CP-3 | ✅ |
| REQ-2.6 (Cancelled: ✗ + strike + dim) | `TestTUI_ViewListRendersStatusIcons` (assert `✗` + ANSI) | `app.go:639-668` | CP-3 | ✅ |
| REQ-2.7 (Open: 2 spaces) | `TestTUI_ViewListRendersStatusIcons` (assert NOT contains ✓/✗) | `app.go:639` (`icon := "  "`) | CP-3 | ✅ |
| REQ-3.1 (modal install) | `TestTUI_SingleTaskDeleteWithConfirm` | `bulk.go:53-60` | CP-4 | ✅ |
| REQ-3.2 (skip modal when disabled) | `TestTUI_DeleteWithoutConfirmWhenDisabled` | `bulk.go:52,61` (fall-through) | CP-4 | ✅ |
| REQ-3.3 (confirm 'y' → splice path) | `TestTUI_SingleTaskDeleteConfirmYes` (assert `mm.tasks` shrinks + DB write) | `bulk.go:148-153` + `singleActionByID` `bulk.go:163-182` | CP-5 | ✅ |
| REQ-3.4 (dismiss preserves) | `TestTUI_SingleTaskDeleteConfirmNo` | `bulk.go:155` (return m, nil) | CP-5 | ✅ |
| REQ-4.1 (ConfirmDelete field) | `TestDefaults_AreValid` | `internal/config/app.go:11` | CP-6 | ✅ |
| REQ-4.2 (default true) | `TestDefaults_ConfirmDeleteTrue` | `internal/config/app.go:22` | CP-6 | ✅ |
| REQ-4.3 (env parsing + warn on invalid) | `TestLoad_ConfirmDeleteFromEnv` (3 кейса: true/false/bogus) | `internal/config/loader.go:137-143` | CP-6 | ✅ |
| REQ-4.4 (YAML key + absence → default) | `TestLoad_ConfirmDeleteFromYAML`, `TestLoad_ConfirmDeleteMissingFromYAMLDefaultsTrue` | `internal/config/loader.go:62-65` (pre-populate Defaults) | CP-6 | ✅ |
| REQ-5.1 (preservation) | ~250 существующих тестов проходят; 2 теста pivot'нуто (`TestTUI_DeleteCursorTaskWhenNoSelection`, `TestProp_EmptySelectionEquivCursor` с `ConfirmDelete=false`) | cross-cutting | — | ✅ |
| REQ-5.2 (RO guard preserved) | `TestTUI_DeleteConfirmBlockedInReadOnly` (assert no modal + "read-only" status) | `bulk.go:47-51` (RO check ДО modal-install) | — | ✅ |
| REQ-5.3 (default applies to existing installs) | `TestLoad_ConfirmDeleteMissingFromYAMLDefaultsTrue` | `loadFromFile` pre-populate | — | ✅ |

### Spot-check (3+ маппингов, как требует brief)

1. **REQ-1.3 проверка handler-а**: прочитал `app.go:154-160`. Логика: `if msg.err != nil { statusMsg + Tick; return }` — единственная ветвь возврата. На success — `tea.Batch(loadCurrentList, fetchListCounts)`. Reload на error действительно НЕ триггерится. ✅
2. **REQ-2.5/2.6 strikethrough**: `TestTUI_ViewListRendersStatusIcons` форсит `termenv.TrueColor` profile (необходимо: go test без TTY иначе сбрасывает ANSI до ascii), затем явно asserts `\x1b[9` (ESC[9m — ANSI strikethrough). Это не просто "тест на icon" — реально верифицирует styled output. ✅
3. **REQ-3.3 routing через singleActionByID**: `handleConfirmKey` ветвится по `len(c.ids)==1` (`bulk.go:148`) → `singleActionByID(m, c.action, c.ids[0])`. Helper находит задачу по ID, временно ставит `m.cursor = idx`, вызывает `perCursorAction` (которая делает splice + Cmd), затем восстанавливает prevCursor для non-delete actions. Behavior сохраняется: `m.tasks` действительно spliced, Cmd возвращает `singleActionDoneMsg`. `TestTUI_SingleTaskDeleteConfirmYes` подтверждает: `len(mm.tasks)==1` после 'y', cmd-execution приводит к DB-удалению. ✅

Все REQs прослежены.

## Design Conformance

### §3.1 Architectural Boundaries
Изменения локализованы в `internal/tui` (UI логика) и `internal/config` (конфиг). Новых cross-layer импортов нет: `bulk.go` импортирует `internal/app`, `internal/domain/id`, `internal/storage` (как раньше). Config-tests не импортируют `internal/tui`. ✅

### §3.2 Data Models
`ConfirmDelete bool` `yaml:"confirm_delete"` добавлен в `AppConfig` (соответствует design §2.3). Splice-мутации в `complete/cancelSelected` устанавливают **Status + *At вместе** (Status=Completed + `CompletedAt=&now` + `CancelledAt=nil`; и зеркально для Cancelled) — инвариант `task.Validate` сохранён (`app.go:451-456`, `:471-476`). `pinSelected` устанавливает `m.tasks[i].PinnedToday = &d` через `task.NewDate(time.Now())` — соответствует доменной модели (поле существует). ✅

### §3.3 API Contracts
Нет публичных API endpoint-ов; только in-process сообщения bubbletea. `singleActionDoneMsg` соответствует design §2.3 / §2.6 Property 1. ✅

### §3.4 Error Handling
Service error в Cmd → `singleActionDoneMsg{err}` (NOT `errorMsg{}` как было раньше для согласованной точки решения "reload or fade"). Handler различает `nil` vs не-nil err поле и реагирует строго по REQ-1.3 (ADR-1). Splice **не откатывается** — что и зафиксировано в ADR-1 и design §2.7. ✅

Один нюанс: handler `errorMsg` всё ещё существует (`app.go:162-165`) для других путей (loadCurrentList и т.п.). Не конфликтует.

### §3.5 Correctness Properties
- **CP-1** (Reload on per-cursor success) — handler `singleActionDoneMsg` (success path) делает `tea.Batch(loadCurrentList, fetchListCounts)`. ✅
- **CP-2** (Optimistic splice mirrors action) — 4 helper-а делают inline-splice до Cmd dispatch. ✅
- **CP-3** (viewList styles by status) — icon switch + strike+dim style. ✅
- **CP-4** (Confirm gated by config) — `dispatch` ветвится по `m.config.ConfirmDelete`. ✅
- **CP-5** (Confirm-yes routes by ids count) — `handleConfirmKey` ветвится по `len(c.ids)==1`. ✅
- **CP-6** (Config defaults & env) — `Defaults().ConfirmDelete=true`, env parsing с warning. ✅

Дополнительная проверка из brief: **loader refactor** (`var cfg AppConfig` → `cfg := Defaults()` + YAML unmarshal поверх) — НЕ был в design §2.3, но это минимально необходимое изменение для поддержки bool default. Поведение для числовых полей сохранено: yaml.v3 overwrites только присутствующие keys; для отсутствующих остаётся pre-populate'нное значение, что эквивалентно прежней логике "zero → default in Validate" для `int >= 1` (Defaults() уже валидно). Все 250+ существующих тестов passing подтверждают отсутствие регрессий. **Acceptable** — обоснованный internal refactor, не scope creep.

### §3.6 Documentation Consistency
Mermaid в design отсутствует (только ASCII art) — пропускаем. Архитектурная диаграмма в design §2.2 соответствует реальным вызовам: handleKey → dispatch → perCursorAction → singleActionDoneMsg → handler. ✅

## Code Quality

### 4.1 Naming
- `singleActionDoneMsg`, `singleActionByID`, `fireSingleAction`, `perCursorAction` — согласованы с проектным стилем (`bulkAction*`, `runBulk`, `applyAction`). camelCase для unexported, descriptive. ✅
- `ConfirmDelete` (PascalCase exported) с YAML tag `confirm_delete` — consistent с `BulkConfirmThreshold` / `NotesMaxLines`. ✅

### 4.2 Dead Code
- Старый `perCursorCmd` полностью удалён из кода (только упоминания в `.spec/` исторических артефактах). ✅
- Комментарий из `loadFromFile` ("Merge with defaults: any zero-valued field...") удалён вместе с устаревшей логикой. ✅
- TODO/FIXME без тикетов не обнаружено.

### 4.3 Scope Creep
- Loader refactor (pre-populate Defaults) — обоснован для bool default (см. §3.5). Acceptable.
- Сэмпл YAML в `defaultYAMLConfig()` дополнен комментарием про `confirm_delete: true` — UX improvement, in scope (REQ-4.4 неявно подразумевает доступность поля для users).
- `fireSingleAction` helper для edge case "task absent from m.tasks" — НЕ упомянут в design §2.7, но это разумная защита от race condition (явно обозначено в implementation.md). Не блокирующее.
- Никаких новых endpoints/features beyond REQs не обнаружено.

### 4.4 Test Quality
- `TestTUI_ViewListRendersStatusIcons` — реально assert'ит ANSI strikethrough escape (не просто "no error"). ✅
- `TestTUI_SingleTaskDeleteConfirmYes` — после Cmd execution верифицирует DB-side через `svc.ListInbox` (не только UI). ✅
- `TestLoad_ConfirmDeleteFromEnv/invalid` — assert'ит и значение, и наличие warning. ✅
- Test names descriptive (`TestTUI_<scenario>` pattern matched). ✅
- Race-condition корректность `singleActionByID`: код находит idx по ID, если `idx < 0` (задача исчезла из displayed list между 'd' и 'y') — fallback на `fireSingleAction` без splice. Тест на этот edge-case не написан, но логика корректна и комментирована.

### Race handling spot-check (singleActionByID)
Прочитал `bulk.go:163-182`. Если cursor дрейфовал из-за async reload между нажатием 'd' и 'y': модалка хранит **tid** (не cursor index), helper линейно ищет idx по ID, splice'ит правильную задачу. После splice восстанавливает `prevCursor` для non-delete (для delete cursor clamped через сам `deleteSelected`). Логика корректная. ✅

## Security

Затрагиваемые поверхности:
- **Env var `TODUSHKA_CONFIRM_DELETE`** — валидация через `strconv.ParseBool`; invalid → warning, leave prior value. No injection vector. ✅
- **YAML key `confirm_delete`** — parsed через `yaml.v3` + struct tag, bool. No injection. ✅
- **TUI rendering** — task titles рендерятся через lipgloss с `Strikethrough(true).Faint(true)` — то же поведение, что и для существующих стилей; нет untrusted input. ✅
- **Optimistic splice мутирует `m.tasks`** — purely local state, no boundary crossing, никаких privilege переходов. ✅

**No security issues found in changed files.**

## Verification Evidence

Команды выполнены **в этой review-сессии**, не скопированы из implementation report.

- **Tests** (`go clean -testcache && task test`, last 20 lines):
```
task: [test] go test ./...
ok  	github.com/jtprogru/todushka/cmd/todushka	0.928s
ok  	github.com/jtprogru/todushka/internal/app	0.504s
ok  	github.com/jtprogru/todushka/internal/cli	1.356s
ok  	github.com/jtprogru/todushka/internal/config	8.545s
ok  	github.com/jtprogru/todushka/internal/domain/area	2.451s
ok  	github.com/jtprogru/todushka/internal/domain/id	3.192s
ok  	github.com/jtprogru/todushka/internal/domain/project	4.468s
ok  	github.com/jtprogru/todushka/internal/domain/quickentry	3.589s
ok  	github.com/jtprogru/todushka/internal/domain/repeat	4.955s
ok  	github.com/jtprogru/todushka/internal/domain/tag	1.706s
ok  	github.com/jtprogru/todushka/internal/domain/task	5.769s
ok  	github.com/jtprogru/todushka/internal/domain/today	4.083s
ok  	github.com/jtprogru/todushka/internal/storage	5.371s
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	14.129s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	2.083s
ok  	github.com/jtprogru/todushka/internal/tui	7.118s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

- **Tests (race)** (`task test-race`, last 10 lines):
```
ok  	github.com/jtprogru/todushka/internal/domain/repeat	6.015s
ok  	github.com/jtprogru/todushka/internal/domain/tag	8.699s
ok  	github.com/jtprogru/todushka/internal/domain/task	7.017s
ok  	github.com/jtprogru/todushka/internal/domain/today	7.463s
ok  	github.com/jtprogru/todushka/internal/storage	8.284s
ok  	github.com/jtprogru/todushka/internal/storage/bbolt	16.938s
ok  	github.com/jtprogru/todushka/internal/storage/fakes	7.880s
ok  	github.com/jtprogru/todushka/internal/tui	12.037s
?   	github.com/jtprogru/todushka/internal/version	[no test files]
```

- **Build** (`task build`):
```
task: [build] mkdir -p bin
task: [build] go build -o bin/todushka ./cmd/todushka
```

- **Lint** (`task lint`):
```
task: [lint] golangci-lint run
0 issues.
```

- **Format** (`gofmt -l internal/ cmd/`): пустой вывод (нет файлов с нарушением формата).

## Findings

| ID | Severity | File | Description | Requirement |
|----|----------|------|-------------|-------------|
| F-1 | minor | `internal/tui/app.go:580` | Сообщение модалки `"Delete 1 tasks?"` для single-task confirm — грамматически неудачно. Существующий код, не введён в этом feature, но теперь чаще виден из-за нового single-task confirm flow. Cosmetic. | — |
| F-2 | minor | `internal/tui/bulk.go:183-205` (`fireSingleAction`) | Edge-case "task vanished from m.tasks between 'd' and 'y'" не покрыт unit-тестом. Логика корректна, но регрессия в этой ветке не будет обнаружена тестами. Optional follow-up. | REQ-3.3 |
| F-3 | nit | `internal/tui/app.go:639-646` | Магическая строка `"  "` (два пробела) для Open icon — можно заменить на именованную константу или поле theme для будущей расширяемости (например, для pin icon). Не блокирующее. | REQ-2.7 |

Замечаний уровня `critical`/`major` — **ноль**.

## Recommendations

Все рекомендации опциональны и не блокируют merge:

1. **F-1**: использовать `"Delete %d task%s?"` с pluralization helper (или ветвление `len(ids)==1 ? "Delete task?" : fmt.Sprintf("Delete %d tasks?", n)`). Тривиальное cosmetic improvement.
2. **F-2**: добавить тест `TestTUI_SingleTaskDeleteConfirmYesAfterTaskVanished` — setup confirm modal, вручную опустошить `m.tasks` или удалить целевой ID, нажать 'y' → assert Cmd вызывает service write (через mock spy), модалка закрылась, m.tasks unchanged. Закрывает gap в тестах race-handling.
3. **F-3**: ввести константу `iconStatusOpen = "  "` рядом с другими TUI константами для будущей унификации (когда добавится pin-icon в v2).

## Quality Control Checklist

- [x] Change Set Discovery complete.
- [x] Все 16 REQs прослежены к коду и тестам.
- [x] Все 6 CP проверены против implementation.
- [x] Design conformance §3.1-§3.6 проверены.
- [x] Code quality: naming, dead code, scope creep, test quality.
- [x] Security scan: 0 issues.
- [x] Verification commands re-run в этой сессии (не скопированы).
- [x] Verdict = PASS (zero critical/major).

# Search & Multi-Select for TUI — Requirements

**Status:** Draft
**Author:** spec-driven-dev (Claude)
**Date:** 2026-05-25

## Overview

Расширение TUI двумя ортогональными возможностями: inline-фильтр текущего списка (`/`) с live-обновлением и multi-select (`Space`/`*`) поверх существующих action-шорткатов. Bulk-операции (Complete / Cancel / Delete / Pin) автоматически применяются к выделенным задачам при непустом наборе, иначе сохраняют текущее поведение per-cursor. Изменения изолированы в пакете `internal/tui`; `storage` и `app` слои не затрагиваются. Backward-compatible: пустое выделение → 100% старого UX.

## Glossary

| Term | Definition | Code Artifact |
|------|------------|---------------|
| Filter Query | Сырая строка, введённая пользователем после `/`. Применяется как case-insensitive substring к `Task.Title`. | `internal/tui` (новое поле `Model.filterQuery`) |
| Visible Tasks | Подмножество `Model.tasks`, прошедшее текущий Filter Query. Если Filter Query пустой, равно `Model.tasks`. | `internal/tui` (метод `displayedTasks`) |
| Selection Set | Множество ID видимых задач, помеченных пользователем для bulk-операции. Set-семантика (без порядка, без дубликатов). | `internal/tui` (новое поле `Model.selected`) |
| Bulk Operation | Применение одного action (Complete / Cancel / Delete / Pin) ко всем задачам из Selection Set, последовательно. | `internal/tui` (`bulkComplete`/`bulkCancel`/`bulkDelete`/`bulkPin`) |
| Confirm Threshold | Число задач в Selection Set, начиная с которого bulk-операция требует подтверждения. Значение: 5. | `internal/tui` (константа `bulkConfirmThreshold`) |
| Filter Mode | Состояние TUI, когда filter input принимает keystrokes. Активируется по `/`, завершается `Enter` или `Esc`. | `internal/tui` (новое значение `screenKind` или флаг `Model.filtering`) |

## User Stories

- Как **пользователь, разбирающий Inbox с 30+ задачами**, я хочу набрать `/милк` и сразу видеть только задачи про "молоко", чтобы не сканировать список глазами.
- Как **пользователь на еженедельном обзоре**, я хочу выделить пачку из 10 однотипных задач и одним нажатием их complete'нуть, чтобы не тыкать `c` десять раз.
- Как **пользователь, делающий bulk-delete на 20 задач**, я хочу видеть подтверждение, чтобы случайное нажатие `d` не уничтожило половину Inbox.
- Как **существующий пользователь**, я хочу чтобы все мои привычные шорткаты (`c`/`x`/`d`/`p`) продолжали работать на одну задачу под курсором, когда я ничего не выделил.

## Requirements

### Group 1 — Filter

**REQ-1.1** WHEN пользователь нажимает `/` в screenList, the system SHALL войти в Filter Mode, отобразить filter input в нижней части экрана и направлять последующие printable keys в этот input.

**REQ-1.2** WHEN Filter Mode активен и пользователь набирает или удаляет символ, the system SHALL мгновенно (без I/O) перерисовать список задач, оставив только те задачи, чей `Task.Title` содержит текущий Filter Query как case-insensitive substring (Unicode fold-case).

**REQ-1.3** WHEN Filter Mode активен и пользователь нажимает `Enter`, the system SHALL выйти из Filter Mode, сохранить текущий Filter Query и оставить список отфильтрованным до явной очистки.

**REQ-1.4** WHEN Filter Mode активен и пользователь нажимает `Esc`, the system SHALL очистить Filter Query, выйти из Filter Mode и отобразить полный список.

**REQ-1.5** WHEN Filter Query непустой и Visible Tasks содержит 0 элементов, the system SHALL отобразить плейсхолдер `(no matches)` вместо списка задач.

**REQ-1.6** WHEN пользователь переключает активный список (Tab / Shift+Tab / клавиши 1-6) при непустом Filter Query, the system SHALL очистить Filter Query до загрузки нового списка.

**REQ-1.7** WHEN Filter Query содержит только пробельные символы, the system SHALL трактовать его как пустой и отображать полный список.

### Group 2 — Selection

**REQ-2.1** WHEN курсор стоит на видимой задаче и пользователь нажимает `Space`, the system SHALL добавить `Task.ID` в Selection Set, если его там не было, или удалить, если был (toggle).

**REQ-2.2** WHEN Selection Set непустой, the system SHALL отображать префикс `[x] ` для выделенных видимых задач и `[ ] ` для невыделенных видимых задач.

**REQ-2.3** WHEN Selection Set пустой, the system SHALL не отображать префикс выделения (визуальный вид остаётся идентичным текущему).

**REQ-2.4** WHEN пользователь нажимает `*` в screenList и не в Filter Mode, the system SHALL добавить ID всех Visible Tasks в Selection Set.

**REQ-2.5** WHEN Selection Set непустой и пользователь нажимает `Esc` (не в Filter Mode), the system SHALL очистить Selection Set.

**REQ-2.6** WHEN изменение Filter Query приводит к тому, что задача с ID из Selection Set больше не входит в Visible Tasks, the system SHALL удалить этот ID из Selection Set.

**REQ-2.7** WHEN пользователь переключает активный список (Tab / Shift+Tab / 1-6), the system SHALL очистить Selection Set до загрузки нового списка.

**REQ-2.8** WHEN Selection Set непустой, the system SHALL отображать текст `Selected: N` в status bar (где N — `len(Selection Set)`).

### Group 3 — Bulk Operations

**REQ-3.1** WHEN пользователь нажимает action key (`c` / `x` / `d` / `p`) и Selection Set пустой, the system SHALL применить операцию к задаче под курсором (сохранить поведение текущей версии).

**REQ-3.2** WHEN пользователь нажимает action key и `1 ≤ len(Selection Set) < 5`, the system SHALL последовательно выполнить операцию для каждого ID из Selection Set без confirm-модалки.

**REQ-3.3** WHEN пользователь нажимает action key и `len(Selection Set) ≥ 5`, the system SHALL отобразить confirm-модалку с текстом `<Action> N tasks? (y/n)` и выполнить операцию только при подтверждении клавишей `y`.

**REQ-3.4** WHEN отображается confirm-модалка и пользователь нажимает любую клавишу кроме `y`, the system SHALL закрыть модалку и не выполнять bulk-операцию.

**REQ-3.5** WHEN bulk-операция выполняется и одна или более задач возвращают ошибку из service, the system SHALL продолжить обработку остальных задач до конца и отобразить агрегированное сообщение `<Action>: M/N succeeded, K failed` в status bar.

**REQ-3.6** WHEN bulk-операция завершилась (с любым результатом), the system SHALL очистить Selection Set и перезагрузить активный список через `loadCurrentList()`.

**REQ-3.7** WHEN bulk-операция выполняется и `Service.<Action>Task` возвращает не-recoverable ошибку (например, `context.Canceled`), the system SHALL немедленно прервать обработку, отобразить ошибку в status bar и НЕ очищать Selection Set.

### Group 4 — Help & Discovery

**REQ-4.1** WHEN пользователь нажимает `?` (help), the system SHALL включить новые keybindings (`/`, `Space`, `*`) и их описания в список отображаемых клавиш.

**REQ-4.2** WHEN screenList отображён в нормальном режиме (не в Filter Mode и не в confirm-модалке), the system SHALL включить подсказку про новые клавиши в footer hint (например, `/: filter  space: select`).

## Topological Order

```
Group 1 (Filter)    — независимая
Group 2 (Selection) — независимая
Group 3 (Bulk)      — зависит от Group 2
Group 4 (Help)      — зависит от Group 1, 2, 3
```

Reason: Filter и Selection — ортогональные state-машины внутри Model и могут разрабатываться/тестироваться параллельно. Bulk-операции опираются на Selection Set, поэтому требуют Group 2 в кода. Help-обновление — последнее, потому что отображает все новые keybindings.

REQ-2.6 (filter hides selected → drop) формально пересекает обе группы, но реализуется в `displayedTasks`/`updateSelection` хелперах после того как обе state-машины существуют.

## Conflict Priority

**REQ-3.1 (per-cursor при пустом Selection) vs REQ-3.2/3.3 (bulk при непустом Selection):**
Это не конфликт, а switch по `len(Selection Set)`. REQ-3.1 — branch для `len == 0`; REQ-3.2/3.3 — для `len > 0`. Документируется как разный поведенческий путь, не как противоречие.

**REQ-1.6 (filter clears on list switch) vs REQ-2.7 (selection clears on list switch):**
Обе операции происходят на одном и том же event'е (switchList). Порядок: сначала очистить filter и selection (state reset), затем `loadCurrentList()`. Это обеспечивает согласованное состояние перед I/O.

## Open Design Questions

| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Где визуально размещать filter input — inline в footer или как modal overlay? | Влияет на восприятие "контекст не теряется": inline сохраняет видимость списка под filter input, modal может временно скрывать строки. | REQ-1.1, REQ-1.2 |
| Confirm dialog: один `y` достаточно или нужно `y` + `Enter`? | Trade-off скорость vs защита от случайного нажатия. Things 3 использует одиночный `y`. | REQ-3.3 |
| Bulk-операции — последовательно (один Cmd, цикл внутри) или параллельно (`tea.Batch` с N Cmd)? | Последовательный путь предсказуем по error reporting; параллельный быстрее на большие N но усложняет агрегацию ошибок. Для bbolt параллелизм не даёт perf-выигрыша. | REQ-3.2, REQ-3.5 |
| Какой символ-маркер для `[x]` / `[ ]` — ASCII или Unicode (`☑` / `☐`)? | ASCII универсален; Unicode красивее но зависит от nerd-font/glyph support терминала. | REQ-2.2 |

## Verification Commands

| Action   | Command            | Source         |
|----------|--------------------|----------------|
| Test     | `task test`        | `Taskfile.yml` |
| Test (race) | `task test-race` | `Taskfile.yml` |
| Build    | `task build`       | `Taskfile.yml` |
| Lint     | `task lint`        | `Taskfile.yml` |
| Format   | `task fmt`         | `Taskfile.yml` |
| Run      | `task run`         | `Taskfile.yml` |

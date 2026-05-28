# things-visual (Фаза A) — Requirements

**Status:** Draft
**Author:** Claude (spec-driven-dev)
**Date:** 2026-05-28

## Overview
Презентационная доработка TUI под Things 3, без изменения домена: кольцо
прогресса у проектов (BL-9), фирменная жёлтая звезда «today» у задач в
списке Anytime и приглушение завершённых строк (BL-9/BL-10). Затрагивает
`internal/tui/project_list.go`, `internal/tui/app.go` (`viewList`) и
`internal/tui/style.go`. Источник «today» — `today.ComputeToday(...,
time.Now(), 0)`, вызываемый в рендере.

## Glossary
| Term | Definition | Code Artifact |
|------|------------|---------------|
| кольцо прогресса | Глиф ◯◔◑◕●, кодирующий долю выполненных задач проекта | `viewProjectList` в `internal/tui/project_list.go` |
| звезда «today» | Маркер ★ у задачи, попадающей в Today, при просмотре в Anytime | `viewList` в `internal/tui/app.go` |
| предикат today | `today.ComputeToday` — задача open и (pinned==today ∨ start≤today ∨ deadline в окне) | `internal/domain/today/today.go` |

## User Stories
- Как пользователь, я хочу видеть кольцо прогресса у проектов, чтобы
  оценивать готовность, не открывая проект.
- Как пользователь, я хочу видеть звезду у сегодняшних задач в Anytime,
  чтобы отличать приоритетные дела (как в Things).

## Requirements

### Группа 1 — кольцо прогресса (BL-9)
**REQ-1.1** WHEN строка проекта рендерится и доля выполненных задач равна 0 (включая `total==0`), the system SHALL показать глиф ◯.

**REQ-1.2** WHEN у проекта `total>0` и `open==0`, the system SHALL показать глиф ●.

**REQ-1.3** WHEN проект выполнен частично (0 < доля < 100%), the system SHALL показать один из ◔/◑/◕, монотонно возрастающий с долей выполненных (◯ и ● зарезервированы только под 0% и 100%).

**REQ-1.4** WHEN строка проекта рендерится, the system SHALL сохранить существующий текстовый счётчик `[open/total]` рядом с кольцом.

### Группа 2 — звезда «today» в Anytime (BL-10)
**REQ-2.1** WHEN активный список — Anytime и задача проходит предикат today, the system SHALL показать маркер-звезду ★ слева от строки.

**REQ-2.2** WHEN активный список — Anytime и задача НЕ проходит предикат today, the system SHALL не показывать звезду и сохранить выравнивание колонок за счёт слота звезды фиксированной ширины.

**REQ-2.3** WHEN активный список — любой кроме Anytime, the system SHALL не показывать звезду ни у одной строки.

### Группа 3 — приглушение завершённых строк (BL-10 polish)
**REQ-3.1** WHEN строка задачи имеет статус Completed или Cancelled, the system SHALL приглушить (faint) всю строку, включая короткий ID, а не только заголовок.

**REQ-3.2** WHEN строка задачи имеет статус Open, the system SHALL не применять faint/strikethrough к заголовку и короткому ID (поведение open-строк не меняется).

### Группа 4 — деградация и отсутствие регрессий
**REQ-4.1** WHEN активна monochrome-тема (`NO_COLOR`), the system SHALL отрисовать кольцо и звезду теми же рунами без цветового стиля, не нарушая вывод.

**REQ-4.2** WHEN запускается существующий набор тестов рендера (`viewport_test.go`, `project_list_test.go`), the system SHALL сохранять их прохождение (маркер курсора виден; тело списка не превышает `visibleRows`).

## Topological Order
Группы 1, 2, 3 независимы и могут реализовываться параллельно.
REQ-4.1 проверяется после 1–3; REQ-4.2 — сквозная, валидируется в конце.

## Open Design Questions
| Question | Why It Matters | Impacted Requirements |
|----------|---------------|----------------------|
| Точные пороги долей для ◔/◑/◕ | Определяет визуальный отклик кольца | REQ-1.3 |
| Ширина слота звезды и его место относительно cursor-маркера / `[x]`-префикса мультивыбора | Влияет на выравнивание и на `prefixWidth` при переносе заголовка | REQ-2.1, REQ-2.2 |
| Ambiguous-width рун (★, ◯, ●): измерять слот через `lipgloss.Width`, а не `len` | Иначе возможен сдвиг колонок в части терминалов | REQ-2.2, REQ-4.1 |

## Verification Commands
| Action | Command | Source |
|--------|---------|--------|
| Test | `task test` | Taskfile.yml |
| Test (race) | `task test-race` | Taskfile.yml |
| Build | `task build` | Taskfile.yml |
| Lint | `task lint` | Taskfile.yml |
| Fmt | `task fmt` | Taskfile.yml |

# Viewport Scroll — Requirements

**Status:** Draft
**Date:** 2026-05-27

## Overview

Bug fix: курсор в `viewList` / `viewProjectList` / `viewProjectTasks`
выходит за пределы видимой области когда список длиннее доступной
высоты. Добавляем scroll-offset с edge-follow + scrolloff=3.

## Requirements

**REQ-1.1** WHEN пользователь нажимает `j`/`Down` или `k`/`Up` в
screenList / screenProjects / screenProjectTasks, the system SHALL
поддерживать `cursor` (или `projectCursor`) и соответствующий
`scrollOffset` так, что отрисованный кадр содержит строку курсора
с минимум 3 строками контекста сверху и снизу (если они существуют
в списке и помещаются в доступную высоту `visibleRows`).

**REQ-1.2** WHEN `visibleRows < totalRows` (список длиннее экрана),
the system SHALL отрисовать ровно `visibleRows` строк подряд из
`disp`, начиная с индекса `scrollOffset`, причём `scrollOffset`
clamp'ится в `[0, max(0, totalRows - visibleRows)]` — в кадре
никогда не появляются пустые строки или out-of-bounds индексы.

**REQ-1.3** WHEN список перезагружается
(`tasksLoadedMsg` / `projectsLoadedMsg` / `projectTasksLoadedMsg`)
и новый `len(disp)` меньше предыдущего, the system SHALL
clamp'нуть `scrollOffset` так, что он остаётся валидным
(`scrollOffset <= max(0, totalRows - visibleRows)`).

## Verification Commands

| Action     | Command          | Source        |
|------------|------------------|---------------|
| Test       | `task test`      | Taskfile.yml  |
| Test (race)| `task test-race` | Taskfile.yml  |
| Build      | `task build`     | Taskfile.yml  |
| Lint       | `task lint`      | Taskfile.yml  |

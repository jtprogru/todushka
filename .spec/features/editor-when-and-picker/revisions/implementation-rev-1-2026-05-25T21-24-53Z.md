# Implementation Report: Editor When + Picker (v0.6.0)

## Summary

Реализованы Part A + Part B как одна cohesive feature:

- **Part A — Context-aware When label:** `whenLabel(t task.Task) string` helper выбирает "Inbox" или "Anytime" на основе `(AreaID, ProjectID)`. View() заменил hardcoded "Anytime" на context-aware label. Hint `(will appear in Inbox without Area/Project)` удалён.
- **Part B — Area/Project/Heading picker:** Editor получил 3 новых textinput поля между Deadline и Tags. NewEditor signature изменён: `(ctx, t, svc)`. Pre-fill через `Repo.AreaGet`/`ProjectGet`/`HeadingList`. ApplyAndSave делает sequential resolve: Area → Project → Heading, первая error abort. Heading требует Project; auto-clear heading при смене Project.

Все 8 tasks выполнены. 26 REQs покрыты unit + property тестами; 13 CPs покрыты PBT.

## Commands Used

- **Test:** `task test`
- **Test (race):** `task test-race`
- **Build:** `task build`
- **Lint:** `task lint`

## Task Execution

- [x] **T-1** Baseline — 213 tests passing
- [x] **T-2** enum extension + struct fields + focus/cycle — `fieldArea`, `fieldProject`, `fieldHeading` constants; `area`/`project`/`heading textinput.Model` fields; `focusCurrent`/`UpdateForm` extended; `TestEditor_FieldCountIsSix` → `TestEditor_FieldCountIsNine`
- [x] **T-3** NewEditor pre-fill — `NewEditor(ctx, t, svc)` signature change; pre-fill через `Repo.AreaGet`/`ProjectGet`/`HeadingList`; 4 new tests + all call sites updated
- [x] **T-4** whenLabel + context-aware View — Part A done; obsolete v0.5.0 hint tests заменены на 6 new label tests; PBT `TestProp_HintConditional` → `TestProp_WhenLabelMatchesContext`
- [x] **T-5** View renders 3 new field blocks — 1 new test
- [x] **T-6** ApplyAndSave sequential resolve — Part B core; 3 resolve blocks (Area, Project, Heading) с auto-clear heading on project change; 13 new tests
- [x] **T-7** PBT batch — 12 new PBTs (CP-2..CP-13); 17 total PBTs in editor_test.go; stable при `-count=2`
- [x] **T-8** GATE — all checks PASS

## Final Verification

- **Tests:** all packages PASS; race detector clean; 17 PBTs stable.
- **Build:** `bin/todushka` compiles.
- **Lint:** 0 issues.
- **Format:** gofmt clean.
- **Manual smoke:** `bin/todushka --help` shows all flags.

## Files Changed

### Modified

- `internal/tui/editor.go` — 3 new textinput fields; new enum constants; NewEditor signature change with pre-fill; `whenLabel` helper; View context-aware label + 3 new field blocks; ApplyAndSave sequential resolve (Area → Project → Heading); auto-clear heading on project change
- `internal/tui/editor_test.go` — обновлены existing tests (rename FieldCountIsSix→Nine, removed obsolete hint tests, updated `TestProp_HintConditional` → `TestProp_WhenLabelMatchesContext`); добавлены 24 new tests (4 pre-fill + 6 label/view + 13 resolve + 1 visual) + 12 new PBTs
- `internal/tui/app.go` — `openEditor` обновлён на новую NewEditor signature
- `internal/tui/shell_test.go` — 2 NewEditor call sites обновлены

### Created

(нет новых файлов — фича изолирована в editor.go)

## Notes

- **Headings now properly resolved** в pre-fill через `HeadingList(projectID)` find-by-ID. Полностью closes F-2 finding из dual-pane review (HeadingGet missing) — фактически через workaround списком.
- **Backward compat preserved:** task.Someday domain field unchanged; existing test suite passes (минус 2 obsolete hint tests заменены на equivalent коverage).
- **Sequential resolve** ensures predictable error ordering: invalid area + invalid project → user видит area error first; fix → видит project error.
- **Project change auto-clears heading** prevents orphan HeadingID когда user меняет project но забывает обновить heading.
- **Out of scope (deferred):** modal picker screen (visual scrollable list), inline tag autocomplete, inline area/project creation.
- **Untracked LICENSE file** — пользователь создал manually вне нашего pipeline; не часть этой фичи; будет закоммичено отдельным chore-коммитом после merge.

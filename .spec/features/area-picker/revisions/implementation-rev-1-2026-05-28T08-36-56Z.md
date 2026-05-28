# Implementation Report: area-picker (BL-8)

## Summary

Замена free-text поля Area на интерактивный picker в обоих редакторах TUI.
Новый компонент `areaPicker` + интеграция в task/project editor. 8 задач.

## Commands Used
- **Test:** `task test` (`go test ./...`)
- **Build:** `task build`
- **Lint:** `task lint`

## Task Execution

- [x] **T-1** Preservation tests — GREEN confirmed (3 tests pass on unmodified code)
- [ ] **T-2** Реализовать компонент `areaPicker` — (in progress)
- [ ] **T-3** Тесты компонента `areaPicker`
- [ ] **T-4** Интегрировать picker в task editor
- [ ] **T-5** Интегрировать picker в project editor
- [ ] **T-6** Интеграционные тесты task editor
- [ ] **T-7** Интеграционные тесты project editor
- [ ] **T-8** Checkpoint

## Final Verification

_(заполняется на T-8)_

## Files Changed

_(заполняется по ходу)_

## Notes

- T-1: preservation-тесты добавлены в `editor_test.go`
  (`TestEditor_ProjectHeadingResolveUnchanged`, `TestEditor_NonAreaFieldsSavePreserved`)
  и `project_editor_test.go` (`TestProjectEditor_NonAreaFieldsSavePreserved`).

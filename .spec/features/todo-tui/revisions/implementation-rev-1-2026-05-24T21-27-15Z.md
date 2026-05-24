# Implementation Report: todushka (todo-tui)

## Summary

Реализация терминального todo-менеджера `todushka` в стиле Things 3 на Go + Bubble Tea + bbolt. Implementation выполняется sequential mode по 10 top-level задачам из task plan.

## Commands Used

- **Test:** `go test ./...`
- **Test (race):** `go test -race ./...`
- **Build:** `go build -o bin/todushka ./cmd/todushka`
- **Lint:** `golangci-lint run`
- **Cross-compile:** `task cross-compile`

## Task Execution

- [x] **T-1** Project bootstrap — done
  - `Taskfile.yml` создан с таргетами test/test-race/build/lint/fmt/tidy/run/cross-compile (9 таргетов, проверено `task --list-all`).
  - `.golangci.yml` v2-format, 8 линтеров (govet/staticcheck/errcheck/gosec/gocritic/revive/unused/ineffassign) + gofmt/goimports formatters.
  - `cmd/todushka/main.go` — пустой `main()`, собирается.
  - Зависимости добавлены: bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1.1.0, bbolt v1.4.3, cobra v1.10.2, ulid/v2 v2.1.1, testify v1.11.1, rapid v1.3.0.
  - `README.md` — install / Quick Start / data location / keybindings.
  - `internal/config/paths.go` + `paths_test.go` — DataDir/StateDir/LogPath с XDG + HOME fallback, 3 теста проходят.
  - Cross-compile успешен: 4 бинаря linux/darwin × amd64/arm64.
- [ ] **T-2** Domain core types — pending
- [ ] **T-3** Repeat rule + Today engine — pending
- [ ] **T-4** Quick Entry parser — pending
- [ ] **T-5** Storage layer — pending
- [ ] **T-6** Application Service — pending
- [ ] **T-7** CLI front-end — pending
- [ ] **T-8** TUI infrastructure — pending
- [ ] **T-9** TUI screens — pending
- [ ] **T-10** GATE — pending

## Final Verification

(Will be filled at GATE.)

## Files Changed

(Updated cumulatively as tasks complete.)

### After T-1
- `Taskfile.yml` [NEW]
- `.golangci.yml` [NEW]
- `go.mod` [MODIFIED] (deps)
- `go.sum` [NEW]
- `cmd/todushka/main.go` [NEW]
- `README.md` [NEW]
- `internal/config/paths.go` [NEW]
- `internal/config/paths_test.go` [NEW]

## Notes

- golangci-lint v2 использует другой формат конфига (map для output.formats, settings nested под linters). Подправлено в первой итерации T-1.
- Все 4 cross-compiled бинаря собираются без CGO (REQ-14.3 верификация на T-10).

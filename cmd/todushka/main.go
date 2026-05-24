package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/cli"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/storage/bbolt"
	"github.com/jtprogru/todushka/internal/tui"
)

func main() {
	exitCode := 0
	defer func() {
		if r := recover(); r != nil {
			logPanic(r)
			os.Exit(2)
		}
		os.Exit(exitCode)
	}()

	dataDir, err := config.DataDir(config.OSEnv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "todushka: %v\n", err)
		exitCode = 1
		return
	}
	repo, err := bbolt.Open(filepath.Join(dataDir, "db"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "todushka: %v\n", err)
		exitCode = 1
		return
	}
	defer func() { _ = repo.Close() }()

	svc := app.New(repo, app.SystemClock{})
	deps := cli.DefaultDeps(svc)
	deps.LaunchTUI = tui.Run

	if err := cli.Execute(deps); err != nil {
		exitCode = 1
	}
}

func logPanic(r any) {
	stateDir, _ := config.StateDir(config.OSEnv)
	if stateDir == "" {
		fmt.Fprintf(os.Stderr, "todushka: panic: %v\n%s\n", r, debug.Stack())
		return
	}
	_ = os.MkdirAll(stateDir, 0o750)
	logPath := filepath.Join(stateDir, "log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // logPath is derived from config dir
	if err != nil {
		fmt.Fprintf(os.Stderr, "todushka: panic: %v\n%s\n", r, debug.Stack())
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "panic: %v\n%s\n", r, debug.Stack())
	fmt.Fprintf(os.Stderr, "todushka: panic: %v (details in %s)\n", r, logPath)
}

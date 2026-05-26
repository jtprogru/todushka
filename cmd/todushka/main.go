package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/cli"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/storage"
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
	dbPath := filepath.Join(dataDir, "db")

	readOnlyFlag := hasReadOnlyFlag(os.Args[1:])

	var repo *bbolt.Repo
	if readOnlyFlag {
		repo, err = bbolt.OpenReadOnly(dbPath)
	} else {
		repo, err = bbolt.Open(dbPath)
		if errors.Is(err, storage.ErrDatabaseLocked) {
			fmt.Fprintln(os.Stderr, "todushka: database is locked by another todushka process")
			fmt.Fprintln(os.Stderr, "  hint: run with --readonly to open in read-only mode for browsing")
			exitCode = 1
			return
		}
	}
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

// hasReadOnlyFlag scans the raw os.Args slice for the persistent
// --readonly (or --ro alias) flag and returns whether it is enabled.
//
// We do this pre-scan because the bbolt repository is opened before
// cobra parses flags (cobra parses inside Execute, but Open happens
// in main before cli.Execute). The scan must accept every standard
// pflag boolean form to stay in sync with cobra's eventual parse:
//
//	--readonly            -> true
//	--readonly=true       -> true
//	--readonly=false      -> false
//	--readonly=           -> true   (pflag treats empty value as true)
//	--readonly=notabool   -> false  (invalid → safe default)
//
// The same forms apply to --ro. Persistent flags may appear anywhere
// on the command line in cobra (before or after a subcommand), so the
// scan does NOT stop at a positional argument.
func hasReadOnlyFlag(args []string) bool {
	result := false
	for _, a := range args {
		switch {
		case a == "--readonly", a == "--ro":
			result = true
		case strings.HasPrefix(a, "--readonly="):
			result = parseFlagValue(strings.TrimPrefix(a, "--readonly="))
		case strings.HasPrefix(a, "--ro="):
			result = parseFlagValue(strings.TrimPrefix(a, "--ro="))
		}
	}
	return result
}

// parseFlagValue interprets the right-hand side of "--flag=value" for a
// boolean flag, following pflag's conventions: empty string means true,
// strconv.ParseBool-compatible values are parsed, anything else is
// treated as false (safe default rather than panicking).
func parseFlagValue(v string) bool {
	if v == "" {
		return true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return b
}

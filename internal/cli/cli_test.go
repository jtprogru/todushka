package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func newTestDeps(t *testing.T, env func(string) string) (Deps, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	repo := fakes.New()
	svc := app.New(repo, fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if env == nil {
		env = func(string) string { return "" }
	}
	deps := Deps{
		Service:   svc,
		Stdout:    stdout,
		Stderr:    stderr,
		Stdin:     strings.NewReader(""),
		Env:       env,
		LaunchTUI: func(_ *app.Service, _ config.AppConfig) error { return nil },
	}
	return deps, stdout, stderr
}

func execute(t *testing.T, deps Deps, args ...string) error {
	t.Helper()
	root := NewRootCmd(deps)
	root.SetArgs(args)
	return root.ExecuteContext(context.Background())
}

func TestCLI_AddCreatesTask(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	err := execute(t, deps, "add", "buy", "milk", "--deadline", "2026-06-01")
	require.NoError(t, err)
	short := strings.TrimSpace(stdout.String())
	require.Len(t, short, id.ShortLen)
}

func TestCLI_AddRejectsEmptyTitle(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	err := execute(t, deps, "add")
	require.Error(t, err)
}

func TestCLI_TodayJSON(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	now := deps.Service.Clock().Now()
	require.NoError(t, execute(t, deps, "add", "x", "--start", now.Format("2006-01-02")))
	stdout.Reset()
	require.NoError(t, execute(t, deps, "today", "--json"))
	require.True(t, strings.HasPrefix(stdout.String(), "["), "expected JSON array, got %q", stdout.String())
}

func TestCLI_CompleteNotFound(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	err := execute(t, deps, "complete", "ZZZZZZ")
	require.Error(t, err)
}

func TestCLI_CompleteHappyPath(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "add", "x"))
	short := strings.TrimSpace(stdout.String())
	stdout.Reset()
	require.NoError(t, execute(t, deps, "complete", short))
	require.Contains(t, stdout.String(), "completed")
}

func TestCLI_NoArgsCallsTUI(t *testing.T) {
	called := false
	deps, _, _ := newTestDeps(t, nil)
	deps.LaunchTUI = func(_ *app.Service, _ config.AppConfig) error {
		called = true
		return nil
	}
	require.NoError(t, execute(t, deps))
	require.True(t, called)
}

func TestCLI_ExportToStdout(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "export"))
	require.Contains(t, stdout.String(), `"schema_version": 1`)
}

func TestCLI_ImportRequiresYes(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	err := execute(t, deps, "import", "--json", "/nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--yes")
}

func TestCLI_ReadOnlyFlagDefaultFalse(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	root := NewRootCmd(deps)
	root.SetArgs([]string{})
	require.NoError(t, root.ExecuteContext(context.Background()))
	got, err := root.PersistentFlags().GetBool("readonly")
	require.NoError(t, err)
	require.False(t, got)
}

func TestCLI_ReadOnlyFlagParsed(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	root := NewRootCmd(deps)
	root.SetArgs([]string{"--readonly"})
	require.NoError(t, root.ExecuteContext(context.Background()))
	got, err := root.PersistentFlags().GetBool("readonly")
	require.NoError(t, err)
	require.True(t, got)
}

func TestCLI_ReadOnlyAliasRO(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	root := NewRootCmd(deps)
	root.SetArgs([]string{"--ro"})
	require.NoError(t, root.ExecuteContext(context.Background()))
	got, err := root.PersistentFlags().GetBool("ro")
	require.NoError(t, err)
	require.True(t, got)
	gotMain, err := root.PersistentFlags().GetBool("readonly")
	require.NoError(t, err)
	require.True(t, gotMain, "--ro alias should also flip --readonly's bound bool")
}

// TestProp_CLI_ReadOnlyFlagAlias verifies CP-14: --readonly and --ro
// are observably equivalent — either spelling flips the bound bool
// behind --readonly.
func TestProp_CLI_ReadOnlyFlagAlias(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		flag := rapid.SampledFrom([]string{"--readonly", "--ro"}).Draw(rt, "flag")
		deps, _, _ := newTestDeps(t, nil)
		root := NewRootCmd(deps)
		root.SetArgs([]string{flag})
		require.NoError(rt, root.ExecuteContext(context.Background()))
		got, err := root.PersistentFlags().GetBool("readonly")
		require.NoError(rt, err)
		require.True(rt, got, "flag %q must set readonly=true", flag)
	})
}

func TestOutput_NoColorFormat(t *testing.T) {
	env := func(k string) string {
		if k == "NO_COLOR" {
			return "1"
		}
		return ""
	}
	require.False(t, ColorEnabled(env))
	require.True(t, ColorEnabled(func(string) string { return "" }))
}

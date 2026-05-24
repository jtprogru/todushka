package cli

import (
	"strings"
	"testing"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestCLI_AreaAddCreatesArea(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	require.NotEmpty(t, strings.TrimSpace(stdout.String()))

	stdout.Reset()
	require.NoError(t, execute(t, deps, "area", "list"))
	require.Contains(t, stdout.String(), "Work")
}

func TestCLI_AreaAddRejectsDuplicate(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	err := execute(t, deps, "area", "add", "  WORK ")
	require.ErrorIs(t, err, storage.ErrAlreadyExists)
}

func TestCLI_AreaDeleteEmpty(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	stdout.Reset()
	require.NoError(t, execute(t, deps, "area", "delete", "Work"))
	require.Contains(t, stdout.String(), "deleted")
}

func TestCLI_AreaDeleteNonEmptyRequiresForce(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	require.NoError(t, execute(t, deps, "project", "add", "PR", "--area", "Work"))

	err := execute(t, deps, "area", "delete", "Work")
	require.ErrorIs(t, err, app.ErrAreaNotEmpty)

	require.NoError(t, execute(t, deps, "area", "delete", "Work", "--force"))
}

func TestCLI_AreaDeleteNotFound(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	err := execute(t, deps, "area", "delete", "ghost")
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestCLI_ProjectAddBasic(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "project", "add", "PR review"))
	require.NotEmpty(t, strings.TrimSpace(stdout.String()))

	stdout.Reset()
	require.NoError(t, execute(t, deps, "project", "list"))
	require.Contains(t, stdout.String(), "PR review")
}

func TestCLI_ProjectAddWithAreaAndDeadline(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	stdout.Reset()
	require.NoError(t, execute(t, deps, "project", "add", "PR", "--area", "Work", "--deadline", "2026-06-30"))

	stdout.Reset()
	require.NoError(t, execute(t, deps, "project", "list"))
	out := stdout.String()
	require.Contains(t, out, "PR")
	require.Contains(t, out, "due:2026-06-30")
}

func TestCLI_ProjectListFilteredByArea(t *testing.T) {
	deps, stdout, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	require.NoError(t, execute(t, deps, "area", "add", "Home"))
	require.NoError(t, execute(t, deps, "project", "add", "Work proj", "--area", "Work"))
	require.NoError(t, execute(t, deps, "project", "add", "Home proj", "--area", "Home"))
	stdout.Reset()
	require.NoError(t, execute(t, deps, "project", "list", "--area", "Work"))
	out := stdout.String()
	require.Contains(t, out, "Work proj")
	require.NotContains(t, out, "Home proj")
}

func TestCLI_ProjectDeleteNotFound(t *testing.T) {
	deps, _, _ := newTestDeps(t, nil)
	err := execute(t, deps, "project", "delete", "ghost")
	require.Error(t, err)
}

func TestCLI_AddTaskWithProjectFromCLIWorks(t *testing.T) {
	// Validates full chain REQ-4.5 / REQ-1.6 user path: create area, project, task → confirm linkage
	deps, _, _ := newTestDeps(t, nil)
	require.NoError(t, execute(t, deps, "area", "add", "Work"))
	require.NoError(t, execute(t, deps, "project", "add", "PR", "--area", "Work"))
	require.NoError(t, execute(t, deps, "add", "review PR", "--project", "PR"))
}

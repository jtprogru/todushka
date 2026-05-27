package tui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ─── T-2: scaffolding tests ────────────────────────────────────────────

func TestShellMode_Projects(t *testing.T) {
	m := newTestModelForProjects()
	m.screen = screenProjects
	require.Equal(t, modeProjects, currentMode(m))
	require.Equal(t, "PROJECTS", modeProjects.modeLabel())
}

func TestShellMode_ProjectTasks(t *testing.T) {
	m := newTestModelForProjects()
	m.screen = screenProjectTasks
	require.Equal(t, modeProjects, currentMode(m))
}

func TestKeyMap_Projects_IsCapitalP(t *testing.T) {
	km := DefaultKeyMap()
	keys := km.Projects.Keys()
	require.Contains(t, keys, "P", "Projects keybinding must use capital P")
	require.NotContains(t, keys, "p", "Projects must not collide with PinToday")
}

func TestKeyMap_ToggleAllStatuses_IsA(t *testing.T) {
	km := DefaultKeyMap()
	require.Contains(t, km.ToggleAllStatuses.Keys(), "a")
}

func TestShellMode_HelpStillOverridesProjects(t *testing.T) {
	m := newTestModelForProjects()
	m.screen = screenHelp
	require.Equal(t, modeHelp, currentMode(m))
}

// newTestModelForProjects returns a minimal Model suitable for shell-mode
// and navigation unit tests. service is nil (tests using it must inject
// their own).
func newTestModelForProjects() Model {
	return Model{
		theme: NewTheme(),
		keys:  DefaultKeyMap(),
	}
}

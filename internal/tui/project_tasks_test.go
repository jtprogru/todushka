package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// ─── T-6: zoom-in screenProjectTasks ───────────────────────────────────

func TestModel_ZoomIntoProject(t *testing.T) {
	ctx := context.Background()
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
	m.screen = screenProjects
	m.projects = []project.Project{p}
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	got := mm.(Model)
	require.Equal(t, screenProjectTasks, got.screen)
	require.NotNil(t, got.activeProjectID)
	require.Equal(t, p.ID, *got.activeProjectID)
	require.NotNil(t, cmd, "Enter must produce fetchProjectTasks Cmd")
}

func TestModel_ZoomOut_Esc(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	pid := p.ID
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	m.projects = []project.Project{p}
	m.projectCursor = 0
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	got := mm.(Model)
	require.Equal(t, screenProjects, got.screen)
	require.Nil(t, got.activeProjectID)
}

func TestModel_PKey_IgnoredInTasksScreen(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	pid := p.ID
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	require.Equal(t, screenProjectTasks, mm.(Model).screen)
}

func TestModel_TabKey_IgnoredInTasksScreen(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	pid := p.ID
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	originalList := m.activeList
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, screenProjectTasks, mm.(Model).screen)
	require.Equal(t, originalList, mm.(Model).activeList)
}

func TestModel_GTDKeys_IgnoredInTasksScreen(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	pid := p.ID
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	originalList := m.activeList
	for _, r := range []rune{'1', '2', '3', '4', '5', '6'} {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mm.(Model)
		require.Equal(t, screenProjectTasks, m.screen, "screen must not change on %c", r)
		require.Equal(t, originalList, m.activeList, "activeList must not change on %c", r)
	}
}

func TestViewProjectTasks_NonEmpty(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	pid := id.New()
	m.activeProjectID = &pid
	m.projects = []project.Project{{ID: pid, Name: "X", Status: project.StatusOpen}}
	m.tasks = []task.Task{{ID: id.New(), Title: "Some task", Status: task.StatusOpen}}
	out := viewProjectTasks(m, 80)
	require.Contains(t, out, "Some task")
	require.Contains(t, out, "X") // project name in header
}

func TestViewProjectTasks_Empty(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	pid := id.New()
	m.activeProjectID = &pid
	m.projects = []project.Project{{ID: pid, Name: "X", Status: project.StatusOpen}}
	m.tasks = nil
	out := viewProjectTasks(m, 80)
	require.Contains(t, out, "(no tasks in this project)")
}

func TestViewProjectTasks_HeadingBadge(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	pid := id.New()
	hid := id.New()
	m.activeProjectID = &pid
	m.projects = []project.Project{{ID: pid, Name: "X", Status: project.StatusOpen}}
	m.headingNamesByID = map[id.ID]string{hid: "Section A"}
	m.tasks = []task.Task{{ID: id.New(), Title: "T", HeadingID: &hid, Status: task.StatusOpen}}
	out := viewProjectTasks(m, 80)
	require.Contains(t, out, "[Section A]")
}

func TestModel_ZoomOut_RestoresGTDList(t *testing.T) {
	ctx := context.Background()
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
	pid := p.ID
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	m.projects = []project.Project{p}
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Equal(t, screenProjects, mm.(Model).screen)
	// After zoom-out, expect fetchProjects Cmd (refresh project list+counts).
	require.NotNil(t, cmd)
}

func TestModel_ProjectTasksCursor_J_K(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	pid := p.ID
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	m.tasks = []task.Task{
		{ID: id.New(), Title: "a", Status: task.StatusOpen},
		{ID: id.New(), Title: "b", Status: task.StatusOpen},
	}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	require.Equal(t, 1, mm.(Model).cursor)
}

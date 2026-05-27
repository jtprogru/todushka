package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// ─── T-3.1: viewProjectList rendering ─────────────────────────────────

func TestViewProjectList_Empty(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	out := viewProjectList(m, 80)
	require.Contains(t, out, "(no projects)")
}

func TestViewProjectList_NonEmpty_HasShortID(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	pid := id.New()
	m.projects = []project.Project{{ID: pid, Name: "x", Status: project.StatusOpen}}
	out := viewProjectList(m, 80)
	require.Contains(t, out, id.Short(pid))
}

func TestViewProjectList_NonEmpty_HasName(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	m.projects = []project.Project{{ID: id.New(), Name: "Architecture Overhaul", Status: project.StatusOpen}}
	out := viewProjectList(m, 80)
	require.Contains(t, out, "Architecture Overhaul")
}

func TestViewProjectList_NonEmpty_HasCounts(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	pid := id.New()
	m.projects = []project.Project{{ID: pid, Name: "x", Status: project.StatusOpen}}
	m.projectCounts = map[id.ID][2]int{pid: {2, 5}}
	out := viewProjectList(m, 80)
	require.Contains(t, out, "[2/5]")
}

func TestViewProjectList_NonEmpty_HasAreaName(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	pid := id.New()
	aid := id.New()
	m.projects = []project.Project{{ID: pid, Name: "x", Status: project.StatusOpen, AreaID: &aid}}
	m.areaNamesByID = map[id.ID]string{aid: "Work"}
	out := viewProjectList(m, 80)
	require.Contains(t, out, "Work")
}

func TestViewProjectList_NonEmpty_StatusIcon_Completed(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	m.projects = []project.Project{{ID: id.New(), Name: "x", Status: project.StatusCompleted}}
	m.projectStatusFilter = psfAll
	out := viewProjectList(m, 80)
	require.Contains(t, out, "✓")
}

func TestViewProjectList_NonEmpty_StatusIcon_Cancelled(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	m.projects = []project.Project{{ID: id.New(), Name: "x", Status: project.StatusCancelled}}
	m.projectStatusFilter = psfAll
	out := viewProjectList(m, 80)
	require.Contains(t, out, "✗")
}

func TestViewProjectList_NonEmpty_Deadline(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	m := newTestModel(t)
	d := task.NewDate(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
	m.projects = []project.Project{{ID: id.New(), Name: "x", Status: project.StatusOpen, Deadline: &d}}
	out := viewProjectList(m, 80)
	require.Contains(t, out, "2026-12-31")
}

// ─── T-3: screen navigation tests ─────────────────────────────────────

func TestModel_EnterScreenProjects_P(t *testing.T) {
	m := newTestModel(t)
	require.Equal(t, screenList, m.screen)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	require.Equal(t, screenProjects, m2.(Model).screen)
}

func TestModel_ExitScreenProjects_P(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	require.Equal(t, screenList, m2.(Model).screen)
}

func TestModel_ExitScreenProjects_Esc(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Equal(t, screenList, m2.(Model).screen)
}

func TestModel_GTDKeysBlocked_InProjects(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	originalList := m.activeList
	// Press '1', '6' — these would switch lists in screenList; here they must no-op.
	for _, r := range []rune{'1', '2', '3', '4', '5', '6'} {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mm.(Model)
		require.Equal(t, originalList, m.activeList, "activeList must not change in screenProjects on key %c", r)
		require.Equal(t, screenProjects, m.screen, "screen must not change in screenProjects on key %c", r)
	}
	// Tab → no-op.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mm.(Model)
	require.Equal(t, originalList, m.activeList)
	require.Equal(t, screenProjects, m.screen)
}

func TestModel_ProjectsCursor_J_K(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	m.projects = []project.Project{
		{ID: id.New(), Name: "a", Status: project.StatusOpen},
		{ID: id.New(), Name: "b", Status: project.StatusOpen},
		{ID: id.New(), Name: "c", Status: project.StatusOpen},
	}
	// Down twice.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = mm.(Model)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = mm.(Model)
	require.Equal(t, 2, m.projectCursor)
	// Down past end — clamped at len-1.
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = mm.(Model)
	require.Equal(t, 2, m.projectCursor)
	// Up.
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = mm.(Model)
	require.Equal(t, 1, m.projectCursor)
}

func TestModel_ProjectsToggleStatusFilter_A(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	require.Equal(t, psfOpen, m.projectStatusFilter)
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mm.(Model)
	require.Equal(t, psfAll, m.projectStatusFilter)
	require.NotNil(t, cmd, "toggle should emit fetchProjects Cmd")
	// Toggle back.
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mm.(Model)
	require.Equal(t, psfOpen, m.projectStatusFilter)
}

func TestModel_ProjectsFilter_Slash(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	m.projects = []project.Project{
		{ID: id.New(), Name: "Foo", Status: project.StatusOpen},
		{ID: id.New(), Name: "Bar", Status: project.StatusOpen},
	}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = mm.(Model)
	require.True(t, m.filtering)
	// Type "fo".
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = mm.(Model)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = mm.(Model)
	require.Equal(t, "fo", m.filterQuery)
	disp := displayedProjects(m)
	require.Len(t, disp, 1)
	require.Equal(t, "Foo", disp[0].Name)
}

func TestDisplayedProjects_EmptyQueryReturnsAll(t *testing.T) {
	m := newTestModel(t)
	m.projects = []project.Project{
		{Name: "a"}, {Name: "b"},
	}
	require.Len(t, displayedProjects(m), 2)
}

func TestDisplayedProjects_CaseInsensitive(t *testing.T) {
	m := newTestModel(t)
	m.projects = []project.Project{
		{Name: "Foo"}, {Name: "BarFOO"},
	}
	m.filterQuery = "foo"
	require.Len(t, displayedProjects(m), 2)
}

// fetchProjects via service integration

func newServiceWithProjects(t *testing.T, names ...string) *app.Service {
	t.Helper()
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	for _, n := range names {
		_, err := svc.AddProject(ctxBg(), app.AddProjectInput{Name: n})
		require.NoError(t, err)
	}
	return svc
}

func ctxBg() context.Context {
	return context.Background()
}

func TestFetchProjects_LoadsAllSorted(t *testing.T) {
	svc := newServiceWithProjects(t, "b", "A", "c")
	msg := fetchProjects(svc, true)()
	pm, ok := msg.(projectsLoadedMsg)
	require.True(t, ok, "expected projectsLoadedMsg, got %T", msg)
	require.Len(t, pm.projects, 3)
	// Position default 0 for all; sort by Name case-fold: "A" < "b" < "c".
	names := []string{pm.projects[0].Name, pm.projects[1].Name, pm.projects[2].Name}
	require.Equal(t, []string{"A", "b", "c"}, names)
}

// ─── T-5: Delete confirm flow ─────────────────────────────────────────

func TestModel_DeleteProject_DKey_OpensConfirm(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	m.screen = screenProjects
	m.projects = []project.Project{p}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	require.NotNil(t, mm.(Model).confirm)
	require.NotNil(t, mm.(Model).confirm.projectID)
	require.Equal(t, p.ID, *mm.(Model).confirm.projectID)
}

func TestModel_DeleteProject_Confirm_Y_Deletes(t *testing.T) {
	ctx := context.Background()
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
	m.screen = screenProjects
	m.projects = []project.Project{p}
	m.confirm = &confirmState{projectID: &p.ID}
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	require.Nil(t, mm.(Model).confirm)
	require.NotNil(t, cmd)
	// Execute the Cmd to trigger the DeleteProject call.
	msg := cmd()
	_, ok := msg.(projectDeletedMsg)
	require.True(t, ok, "expected projectDeletedMsg, got %T", msg)
	// Project soft-deleted.
	got, err := svc.Repo().ProjectGet(ctx, p.ID)
	require.NoError(t, err)
	require.NotNil(t, got.DeletedAt)
}

func TestModel_DeleteProject_Confirm_N_Cancels(t *testing.T) {
	ctx := context.Background()
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
	m.screen = screenProjects
	m.projects = []project.Project{p}
	m.confirm = &confirmState{projectID: &p.ID}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	require.Nil(t, mm.(Model).confirm)
	// Project unchanged.
	got, _ := svc.Repo().ProjectGet(ctx, p.ID)
	require.Nil(t, got.DeletedAt)
}

func TestModel_ReadOnly_D_Blocked(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	m.readOnly = true
	m.screen = screenProjects
	m.projects = []project.Project{p}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	require.Nil(t, mm.(Model).confirm)
	require.Contains(t, mm.(Model).statusMsg, "read-only")
}

func TestModel_DeleteProject_EmptyList_NoOp(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.screen = screenProjects
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	require.Nil(t, mm.(Model).confirm)
}

func TestFetchProjects_OnlyOpen(t *testing.T) {
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	p1, _ := svc.AddProject(ctxBg(), app.AddProjectInput{Name: "open1"})
	_ = p1
	// Manually create a completed project via repo (service has no SetStatus).
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	_ = svc.Repo().ProjectCreate(ctxBg(), project.Project{
		ID: id.New(), Name: "done", Status: project.StatusCompleted,
		CreatedAt: now, UpdatedAt: now,
	})
	msg := fetchProjects(svc, false)()
	pm := msg.(projectsLoadedMsg)
	require.Len(t, pm.projects, 1)
	require.Equal(t, "open1", pm.projects[0].Name)
}

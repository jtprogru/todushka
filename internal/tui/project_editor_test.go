package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
)

// ─── T-4: ProjectEditorModel validation ───────────────────────────────

func newEditorTestService(t *testing.T) *app.Service {
	t.Helper()
	return app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
}

func TestProjectEditor_Save_EmptyName(t *testing.T) {
	svc := newEditorTestService(t)
	m := newProjectEditor(true, nil, svc)
	_, _, err := m.ApplyAndSave(context.Background(), svc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "name")
}

func TestProjectEditor_Save_MalformedDeadline(t *testing.T) {
	svc := newEditorTestService(t)
	m := newProjectEditor(true, nil, svc)
	m.name.SetValue("x")
	m.deadline.SetValue("abc")
	_, _, err := m.ApplyAndSave(context.Background(), svc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deadline")
}

func TestProjectEditor_Save_Create_Valid(t *testing.T) {
	ctx := context.Background()
	svc := newEditorTestService(t)
	m := newProjectEditor(true, nil, svc)
	m.name.SetValue("New PR")
	m.deadline.SetValue("2026-12-31")
	p, created, err := m.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "New PR", p.Name)
	require.NotNil(t, p.Deadline)
	require.Equal(t, "2026-12-31", p.Deadline.Format("2006-01-02"))
	// Project actually persisted.
	got, err := svc.Repo().ProjectGet(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, p.ID, got.ID)
}

func TestProjectEditor_Save_Edit_Valid(t *testing.T) {
	ctx := context.Background()
	svc := newEditorTestService(t)
	orig, err := svc.AddProject(ctx, app.AddProjectInput{Name: "Old"})
	require.NoError(t, err)
	m := newProjectEditor(false, &orig, svc)
	m.name.SetValue("Renamed")
	p, created, err := m.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, "Renamed", p.Name)
	got, _ := svc.Repo().ProjectGet(ctx, orig.ID)
	require.Equal(t, "Renamed", got.Name)
}

func TestProjectEditor_Edit_OpensPrefilled(t *testing.T) {
	ctx := context.Background()
	svc := newEditorTestService(t)
	a, _ := svc.AddArea(ctx, "Work")
	aid := a.ID
	d := projectDateFor(2026, 12, 1)
	orig := project.Project{
		ID: a.ID, Name: "PR Review", Notes: "details", Status: project.StatusOpen,
		AreaID: &aid, Deadline: &d, AutoClose: true,
	}
	// Persist via repo so newProjectEditor can resolve area name.
	require.NoError(t, svc.Repo().ProjectCreate(ctx, orig))
	m := newProjectEditor(false, &orig, svc)
	require.Equal(t, "PR Review", m.name.Value())
	require.Equal(t, "details", m.notes.Value())
	require.Equal(t, "Work", m.areaName)
	require.NotNil(t, m.areaID)
	require.Equal(t, aid, *m.areaID)
	require.Equal(t, "2026-12-01", m.deadline.Value())
	require.True(t, m.autoClose)
	require.NotNil(t, m.original)
}

func TestProjectEditor_New_OpensEmpty(t *testing.T) {
	svc := newEditorTestService(t)
	m := newProjectEditor(true, nil, svc)
	require.Nil(t, m.original)
	require.Equal(t, "", m.name.Value())
	require.False(t, m.autoClose)
}

// Integration: editor flow inside the Model

func TestModel_NewProject_OpensEditor(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.screen = screenProjects
	require.False(t, m.editingProject)
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	require.True(t, mm.(Model).editingProject)
	require.Nil(t, mm.(Model).projectEditor.original)
}

func TestModel_EditProject_OpensEditor(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	m.screen = screenProjects
	m.projects = []project.Project{p}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	require.True(t, mm.(Model).editingProject)
	require.NotNil(t, mm.(Model).projectEditor.original)
	require.Equal(t, "X", mm.(Model).projectEditor.name.Value())
}

func TestModel_EditProject_EmptyList_NoOp(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.screen = screenProjects
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	require.False(t, mm.(Model).editingProject)
}

func TestModel_ProjectEditor_Esc_DismissesWithoutSave(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	m.screen = screenProjects
	m.projects = []project.Project{p}
	m.editingProject = true
	m.projectEditor = newProjectEditor(false, &p, svc)
	m.projectEditor.name.SetValue("Y")
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.False(t, mm.(Model).editingProject)
	// Repo unchanged.
	got, _ := svc.Repo().ProjectGet(context.Background(), p.ID)
	require.Equal(t, "X", got.Name)
}

func TestModel_ReadOnly_N_Blocked(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.readOnly = true
	m.screen = screenProjects
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	require.False(t, mm.(Model).editingProject)
	require.Contains(t, mm.(Model).statusMsg, "read-only")
}

func TestModel_ReadOnly_E_Blocked(t *testing.T) {
	m, svc := newTestModelWithService(t)
	p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
	m.readOnly = true
	m.screen = screenProjects
	m.projects = []project.Project{p}
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	require.False(t, mm.(Model).editingProject)
}

// Defensive: ensure Config defaults are normally available (smoke).
func TestProjectEditor_Smoke_Defaults(t *testing.T) {
	cfg := config.Defaults()
	require.Greater(t, cfg.DualPaneMinWidth, 0)
}

// projectDateFor builds a task.Date for tests.
func projectDateFor(year int, month time.Month, day int) task.Date {
	return task.NewDate(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
}

// ─── BL-8 area-picker: T-7 integration tests ─────────────────────────────

// TestProjectEditor_EnterOnAreaOpensPicker verifies CP-14 (REQ-1.1): Enter
// on the Area field opens the picker sub-state.
func TestProjectEditor_EnterOnAreaOpensPicker(t *testing.T) {
	m, svc := newTestModelWithService(t)
	m.screen = screenProjects
	m.editingProject = true
	m.projectEditor = newProjectEditor(true, nil, svc)
	m.projectEditor.focus = pefArea
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, mm.(Model).projectEditor.picker)
}

// TestProjectEditor_ApplyAndSaveUsesAreaID verifies CP-12 (REQ-6.1): project
// save persists the picker-selected AreaID.
func TestProjectEditor_ApplyAndSaveUsesAreaID(t *testing.T) {
	ctx := context.Background()
	svc := newEditorTestService(t)
	a, err := svc.AddArea(ctx, "Work")
	require.NoError(t, err)
	m := newProjectEditor(true, nil, svc)
	m.name.SetValue("P")
	m.areaID = &a.ID
	m.areaName = "Work"
	p, _, err := m.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.NotNil(t, p.AreaID)
	require.Equal(t, a.ID, *p.AreaID)
}

// TestProjectEditor_NoAreaLabelIsNoArea verifies REQ-2.2: the project editor
// renders "No area" when no area is selected.
func TestProjectEditor_NoAreaLabelIsNoArea(t *testing.T) {
	svc := newEditorTestService(t)
	m := newProjectEditor(true, nil, svc)
	out := m.View(NewTheme(), 80)
	require.Contains(t, out, "No area")
}

// ─── BL-8 area-picker: T-1 preservation test ─────────────────────────────

// TestProjectEditor_NonAreaFieldsSavePreserved locks that name/notes/
// deadline/autoClose round-trip through ApplyAndSave unchanged — must hold
// after the area field becomes a picker.
func TestProjectEditor_NonAreaFieldsSavePreserved(t *testing.T) {
	ctx := context.Background()
	svc := newEditorTestService(t)
	m := newProjectEditor(true, nil, svc)
	m.name.SetValue("My Project")
	m.notes.SetValue("project notes")
	m.deadline.SetValue("2026-09-15")
	m.autoClose = true
	p, created, err := m.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "My Project", p.Name)
	require.Equal(t, "project notes", p.Notes)
	require.NotNil(t, p.Deadline)
	require.Equal(t, "2026-09-15", p.Deadline.Format("2006-01-02"))
	require.True(t, p.AutoClose)
}

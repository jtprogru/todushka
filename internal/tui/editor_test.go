package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestEditor_WhenDefaultsAnytimeForOpenTask(t *testing.T) {
	_, svc := newTestModelWithService(t)
	tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen, Someday: false}
	ed := NewEditor(context.Background(), tk, svc)
	require.False(t, ed.someday)
	require.Nil(t, ed.startDate)
}

func TestEditor_WhenDefaultsSomedayForSomedayTask(t *testing.T) {
	_, svc := newTestModelWithService(t)
	tk := task.Task{ID: id.New(), Title: "x", Someday: true}
	ed := NewEditor(context.Background(), tk, svc)
	require.True(t, ed.someday)
}

func TestEditor_ApplyAndSaveMapsAnytime(t *testing.T) {
	_, svc, tasks := setupModelWithInboxTasks(t, "x")
	ed := NewEditor(context.Background(), tasks[0], svc)
	ed.someday = false
	ed.startDate = nil
	saved, err := ed.ApplyAndSave(context.Background(), svc)
	require.NoError(t, err)
	require.False(t, saved.Someday)
}

func TestEditor_ApplyAndSaveMapsSomeday(t *testing.T) {
	_, svc, tasks := setupModelWithInboxTasks(t, "x")
	ed := NewEditor(context.Background(), tasks[0], svc)
	ed.someday = true
	saved, err := ed.ApplyAndSave(context.Background(), svc)
	require.NoError(t, err)
	require.True(t, saved.Someday)
}

func TestEditor_ViewHidesOldHint(t *testing.T) {
	_, svc := newTestModelWithService(t)
	tk := task.Task{ID: id.New(), Title: "x"}
	ed := NewEditor(context.Background(), tk, svc)
	out := ed.View(NewTheme(), 80)
	require.NotContains(t, out, "will appear in Inbox")
}

func TestEditor_FieldCountIsEight(t *testing.T) {
	require.Equal(t, 8, int(fieldCount))
}

// TestProp_WhenMapping: ApplyAndSave maps the editor's someday state to
// task.Someday at save time.
func TestProp_WhenMapping(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, tasks := setupRapidModel(rt, "x")
		wantSomeday := rapid.Bool().Draw(rt, "wantSomeday")
		ed := NewEditor(context.Background(), tasks[0], svc)
		ed.someday = wantSomeday
		saved, err := ed.ApplyAndSave(context.Background(), svc)
		require.NoError(rt, err)
		require.Equal(rt, wantSomeday, saved.Someday)
	})
}

// TestProp_WhenLabelMatchesContext verifies CP-1 (REQ-1.1, 1.2): the When
// section label is "Inbox" iff task has no Area and no Project; otherwise
// "Anytime". The obsolete v0.5.0 "will appear in Inbox" hint must NOT
// appear in any case (REQ-1.3).
func TestProp_WhenLabelMatchesContext(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		tk := task.Task{ID: id.New(), Title: "x"}
		if hasArea {
			aid := id.New()
			tk.AreaID = &aid
		}
		if hasProject {
			pid := id.New()
			tk.ProjectID = &pid
		}
		ed := NewEditor(context.Background(), tk, svc)
		out := ed.View(NewTheme(), 60)
		expectInbox := !hasArea && !hasProject
		if expectInbox {
			require.Contains(rt, out, "Inbox")
		} else {
			require.Contains(rt, out, "Anytime")
		}
		// Obsolete hint must never appear (REQ-1.3).
		require.NotContains(rt, out, "will appear in Inbox")
	})
}

func TestEditor_NewEditorPrefillArea(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tk := task.Task{ID: id.New(), Title: "x", AreaID: &a.ID}
	ed := NewEditor(ctx, tk, svc)
	require.Equal(t, "work", ed.areaName)
	require.NotNil(t, ed.areaID)
	require.Equal(t, a.ID, *ed.areaID)
}

func TestEditor_NewEditorPrefillProject(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "todushka"})
	require.NoError(t, err)
	tk := task.Task{ID: id.New(), Title: "x", ProjectID: &p.ID}
	ed := NewEditor(ctx, tk, svc)
	require.Equal(t, "todushka", ed.project.Value())
}

func TestEditor_NewEditorPrefillHeading(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "p"})
	require.NoError(t, err)
	h, err := svc.AddHeading(ctx, p.ID, "Q1 Planning")
	require.NoError(t, err)
	tk := task.Task{ID: id.New(), Title: "x", ProjectID: &p.ID, HeadingID: &h.ID}
	ed := NewEditor(ctx, tk, svc)
	require.Equal(t, "Q1 Planning", ed.heading.Value())
}

func TestEditor_NewEditorEmptyArea(t *testing.T) {
	_, svc := newTestModelWithService(t)
	tk := task.Task{ID: id.New(), Title: "x"}
	ed := NewEditor(context.Background(), tk, svc)
	require.Nil(t, ed.areaID)
	require.Empty(t, ed.areaName)
	require.Empty(t, ed.project.Value())
	require.Empty(t, ed.heading.Value())
}

func TestEditor_ViewRendersAllNewFields(t *testing.T) {
	_, svc := newTestModelWithService(t)
	tk := task.Task{ID: id.New(), Title: "x"}
	ed := NewEditor(context.Background(), tk, svc)
	out := ed.View(NewTheme(), 80)
	require.Contains(t, out, "Area")
	require.Contains(t, out, "Project")
	require.Contains(t, out, "Heading")
}

// TestEditor_ApplyAndSaveNilAreaID verifies CP-12 (REQ-6.2): a nil
// picker selection saves AreaID == nil.
func TestEditor_ApplyAndSaveNilAreaID(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: &a.ID})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.areaID = nil
	ed.areaName = ""
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.Nil(t, saved.AreaID)
}

// TestEditor_ApplyAndSaveUsesAreaID verifies CP-12 (REQ-6.1): save
// persists the picker-selected AreaID without name resolution.
func TestEditor_ApplyAndSaveUsesAreaID(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.areaID = &a.ID
	ed.areaName = "work"
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.NotNil(t, saved.AreaID)
	require.Equal(t, a.ID, *saved.AreaID)
}

// TestEditor_EnterOnAreaOpensPicker verifies CP-14 (REQ-1.1): Enter on
// the Area field opens the area picker sub-state.
func TestEditor_EnterOnAreaOpensPicker(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	m.editor = NewEditor(ctx, tk, svc)
	m.editor.focus = fieldArea
	m.screen = screenEditor
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, mm.(Model).editor.picker)
}

func TestEditor_SaveEmptyProjectClearsBothIDs(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "proj"})
	require.NoError(t, err)
	h, err := svc.AddHeading(ctx, p.ID, "h")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", ProjectID: &p.ID, HeadingID: &h.ID})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("")
	ed.heading.SetValue("")
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.Nil(t, saved.ProjectID)
	require.Nil(t, saved.HeadingID)
}

func TestEditor_SaveValidProjectSetsID(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "todushka"})
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("todushka")
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.NotNil(t, saved.ProjectID)
	require.Equal(t, p.ID, *saved.ProjectID)
}

func TestEditor_SaveAmbiguousProjectErrors(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	_, err := svc.AddProject(ctx, app.AddProjectInput{Name: "alpha"})
	require.NoError(t, err)
	_, err = svc.AddProject(ctx, app.AddProjectInput{Name: "alpha"})
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("alpha")
	_, err = ed.ApplyAndSave(ctx, svc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous")
}

func TestEditor_SaveInvalidProjectErrors(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("nonexistent")
	_, err = ed.ApplyAndSave(ctx, svc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "project")
	require.Contains(t, err.Error(), "nonexistent")
	require.Contains(t, err.Error(), "not found")
}

func TestEditor_SaveHeadingWithoutProjectErrors(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.heading.SetValue("Q1")
	_, err = ed.ApplyAndSave(ctx, svc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "heading")
	require.Contains(t, err.Error(), "project")
}

func TestEditor_SaveValidHeadingSetsID(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "proj"})
	require.NoError(t, err)
	h, err := svc.AddHeading(ctx, p.ID, "Q1 Planning")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("proj")
	ed.heading.SetValue("Q1 Planning")
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.NotNil(t, saved.HeadingID)
	require.Equal(t, h.ID, *saved.HeadingID)
}

func TestEditor_SaveInvalidHeadingErrors(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	_, err := svc.AddProject(ctx, app.AddProjectInput{Name: "proj"})
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("proj")
	ed.heading.SetValue("nonexistent")
	_, err = ed.ApplyAndSave(ctx, svc)
	require.Error(t, err)
	require.Contains(t, err.Error(), "heading")
	require.Contains(t, err.Error(), "not found")
}

func TestEditor_SaveCaseInsensitiveHeading(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "proj"})
	require.NoError(t, err)
	h, err := svc.AddHeading(ctx, p.ID, "Bugs")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("proj")
	ed.heading.SetValue("BUGS")
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.Equal(t, h.ID, *saved.HeadingID)
}

func TestEditor_SaveProjectChangeAutoClearsHeading(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	pA, err := svc.AddProject(ctx, app.AddProjectInput{Name: "alpha"})
	require.NoError(t, err)
	hA, err := svc.AddHeading(ctx, pA.ID, "intro")
	require.NoError(t, err)
	_, err = svc.AddProject(ctx, app.AddProjectInput{Name: "beta"})
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", ProjectID: &pA.ID, HeadingID: &hA.ID})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	// Switch project to beta without typing a new heading
	ed.project.SetValue("beta")
	ed.heading.SetValue("")
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.Nil(t, saved.HeadingID, "heading auto-cleared on project change")
}

// TestProp_PreFillRoundTrip verifies CP-2 (REQ-2.2, 3.2, 4.2):
// pre-filling editor with a task that has IDs, then immediately saving
// without edits, preserves AreaID/ProjectID/HeadingID.
func TestProp_PreFillRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		hasHeading := hasProject && rapid.Bool().Draw(rt, "hasHeading")

		var aid, pid, hid *id.ID
		if hasArea {
			a, err := svc.AddArea(ctx, "area-rt")
			require.NoError(rt, err)
			aid = &a.ID
		}
		if hasProject {
			p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "proj-rt"})
			require.NoError(rt, err)
			pid = &p.ID
			if hasHeading {
				h, err := svc.AddHeading(ctx, p.ID, "head-rt")
				require.NoError(rt, err)
				hid = &h.ID
			}
		}
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: aid, ProjectID: pid, HeadingID: hid})
		require.NoError(rt, err)

		ed := NewEditor(ctx, tk, svc)
		saved, err := ed.ApplyAndSave(ctx, svc)
		require.NoError(rt, err)
		if hasArea {
			require.Equal(rt, aid, saved.AreaID)
		} else {
			require.Nil(rt, saved.AreaID)
		}
		if hasProject {
			require.Equal(rt, pid, saved.ProjectID)
		} else {
			require.Nil(rt, saved.ProjectID)
		}
		if hasHeading {
			require.Equal(rt, hid, saved.HeadingID)
		} else {
			require.Nil(rt, saved.HeadingID)
		}
	})
}

// TestProp_EmptyAreaClears verifies CP-12 (REQ-6.2): clearing the area
// selection (areaID = nil) erases the task's AreaID on save.
func TestProp_EmptyAreaClears(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		var aid *id.ID
		if hasArea {
			a, err := svc.AddArea(ctx, "area-c")
			require.NoError(rt, err)
			aid = &a.ID
		}
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: aid})
		require.NoError(rt, err)
		ed := NewEditor(ctx, tk, svc)
		ed.areaID = nil
		ed.areaName = ""
		saved, err := ed.ApplyAndSave(ctx, svc)
		require.NoError(rt, err)
		require.Nil(rt, saved.AreaID)
	})
}

// TestProp_EmptyProjectClearsBoth verifies CP-4 (REQ-3.3, 4.3):
// clearing the project input also clears the heading ID.
func TestProp_EmptyProjectClearsBoth(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "p-c"})
		require.NoError(rt, err)
		h, err := svc.AddHeading(ctx, p.ID, "h-c")
		require.NoError(rt, err)
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", ProjectID: &p.ID, HeadingID: &h.ID})
		require.NoError(rt, err)
		ed := NewEditor(ctx, tk, svc)
		ed.project.SetValue("")
		ed.heading.SetValue("")
		saved, err := ed.ApplyAndSave(ctx, svc)
		require.NoError(rt, err)
		require.Nil(rt, saved.ProjectID)
		require.Nil(rt, saved.HeadingID)
	})
}

// TestProp_AmbiguousProjectErrors verifies CP-6 (REQ-3.5): when more
// than one project shares the typed name, save fails with "ambiguous".
func TestProp_AmbiguousProjectErrors(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		n := rapid.IntRange(2, 4).Draw(rt, "n")
		name := "alpha-" + rapid.StringMatching(`[a-z]{3,8}`).Draw(rt, "suffix")
		for i := 0; i < n; i++ {
			_, err := svc.AddProject(ctx, app.AddProjectInput{Name: name})
			require.NoError(rt, err)
		}
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
		require.NoError(rt, err)
		ed := NewEditor(ctx, tk, svc)
		ed.project.SetValue(name)
		_, err = ed.ApplyAndSave(ctx, svc)
		require.Error(rt, err)
		require.Contains(rt, err.Error(), "ambiguous")
	})
}

// TestProp_HeadingWithoutProject verifies CP-7 (REQ-4.4): setting a
// non-empty heading without a project yields an error mentioning both
// terms. Generator uses [a-zA-Z]+ to avoid whitespace-only strings
// (which the editor trims to empty and treats as no-heading per
// editor.go line 294-295).
func TestProp_HeadingWithoutProject(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
		require.NoError(rt, err)
		headingName := rapid.StringMatching(`[a-zA-Z]{2,10}`).Draw(rt, "name")
		ed := NewEditor(ctx, tk, svc)
		ed.heading.SetValue(headingName)
		_, err = ed.ApplyAndSave(ctx, svc)
		require.Error(rt, err)
		require.Contains(rt, err.Error(), "heading")
		require.Contains(rt, err.Error(), "project")
	})
}

// TestProp_ValidHeadingResolves verifies CP-8 (REQ-4.1, 4.2): a
// (project, heading) pair resolves to the heading's ID.
func TestProp_ValidHeadingResolves(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		headingName := "head-" + rapid.StringMatching(`[a-z]{3,8}`).Draw(rt, "name")
		p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "p-vh"})
		require.NoError(rt, err)
		h, err := svc.AddHeading(ctx, p.ID, headingName)
		require.NoError(rt, err)
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
		require.NoError(rt, err)
		ed := NewEditor(ctx, tk, svc)
		ed.project.SetValue("p-vh")
		ed.heading.SetValue(headingName)
		saved, err := ed.ApplyAndSave(ctx, svc)
		require.NoError(rt, err)
		require.NotNil(rt, saved.HeadingID)
		require.Equal(rt, h.ID, *saved.HeadingID)
	})
}

// TestProp_FieldCountInvariant verifies CP-9: the editor always has
// exactly 8 focusable fields (sanity / deterministic invariant).
func TestProp_FieldCountInvariant(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		require.Equal(rt, 8, int(fieldCount))
	})
}

// TestProp_TabCycleOrder verifies CP-10 (REQ-5.1): nextField is a
// cyclic permutation of length fieldCount; calling it fieldCount
// times from any starting field returns to that field.
func TestProp_TabCycleOrder(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		tk := task.Task{ID: id.New(), Title: "x"}
		ed := NewEditor(context.Background(), tk, svc)
		ed.focus = fieldTitle
		// fieldCount nextField calls should return to fieldTitle
		for i := 0; i < int(fieldCount); i++ {
			ed = ed.nextField()
		}
		require.Equal(rt, fieldTitle, ed.focus)
	})
}

// TestProp_HeadingCaseInsensitive verifies CP-12 (REQ-4.5): heading
// lookup is case-insensitive — both upper and lower case spellings
// of a registered heading resolve to its ID.
func TestProp_HeadingCaseInsensitive(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		// Use ASCII-only for case folding predictability
		headingName := "head-" + rapid.StringMatching(`[a-z]{3,6}`).Draw(rt, "name")
		p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "p-ci-" + rapid.StringMatching(`[a-z]{3,6}`).Draw(rt, "pname")})
		require.NoError(rt, err)
		h, err := svc.AddHeading(ctx, p.ID, headingName)
		require.NoError(rt, err)
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
		require.NoError(rt, err)
		ed := NewEditor(ctx, tk, svc)
		ed.project.SetValue(p.Name)
		// Randomly uppercase or lowercase
		if rapid.Bool().Draw(rt, "upper") {
			ed.heading.SetValue(strings.ToUpper(headingName))
		} else {
			ed.heading.SetValue(strings.ToLower(headingName))
		}
		saved, err := ed.ApplyAndSave(ctx, svc)
		require.NoError(rt, err)
		require.Equal(rt, h.ID, *saved.HeadingID)
	})
}

// TestProp_ProjectChangeClearsOrphanHeading verifies CP-13 (REQ-4.6):
// changing the project while leaving the heading input empty clears
// the orphan HeadingID instead of carrying it over to a new project.
func TestProp_ProjectChangeClearsOrphanHeading(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, _ := setupRapidModel(rt, "seed")
		ctx := context.Background()
		pA, err := svc.AddProject(ctx, app.AddProjectInput{Name: "pA-" + rapid.StringMatching(`[a-z]{3,6}`).Draw(rt, "a")})
		require.NoError(rt, err)
		hA, err := svc.AddHeading(ctx, pA.ID, "intro")
		require.NoError(rt, err)
		pB, err := svc.AddProject(ctx, app.AddProjectInput{Name: "pB-" + rapid.StringMatching(`[a-z]{3,6}`).Draw(rt, "b")})
		require.NoError(rt, err)
		_ = pB
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", ProjectID: &pA.ID, HeadingID: &hA.ID})
		require.NoError(rt, err)
		ed := NewEditor(ctx, tk, svc)
		ed.project.SetValue(pB.Name)
		ed.heading.SetValue("")
		saved, err := ed.ApplyAndSave(ctx, svc)
		require.NoError(rt, err)
		require.Nil(rt, saved.HeadingID)
	})
}

// ─── BL-8 area-picker: T-1 preservation tests ────────────────────────────

// TestEditor_ProjectHeadingResolveUnchanged locks the free-text Project/
// Heading resolution behavior (REQ-7.1) — it must remain identical after
// the area field becomes a picker.
func TestEditor_ProjectHeadingResolveUnchanged(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	p, err := svc.AddProject(ctx, app.AddProjectInput{Name: "proj-pres"})
	require.NoError(t, err)
	h, err := svc.AddHeading(ctx, p.ID, "Heading One")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.project.SetValue("proj-pres")
	ed.heading.SetValue("Heading One")
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.NotNil(t, saved.ProjectID)
	require.Equal(t, p.ID, *saved.ProjectID)
	require.NotNil(t, saved.HeadingID)
	require.Equal(t, h.ID, *saved.HeadingID)
}

// TestEditor_NonAreaFieldsSavePreserved locks that title/notes/startDate/
// deadline/tags/someday round-trip through ApplyAndSave unchanged.
func TestEditor_NonAreaFieldsSavePreserved(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	ed := NewEditor(ctx, tk, svc)
	ed.title.SetValue("renamed title")
	ed.notes.SetValue("some notes")
	sd := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.Local))
	ed.startDate = &sd
	ed.deadline.SetValue("2026-07-01")
	ed.tags.SetValue("alpha, beta")
	ed.someday = true
	saved, err := ed.ApplyAndSave(ctx, svc)
	require.NoError(t, err)
	require.Equal(t, "renamed title", saved.Title)
	require.Equal(t, "some notes", saved.Notes)
	require.NotNil(t, saved.StartDate)
	require.Equal(t, "2026-06-01", saved.StartDate.Format("2006-01-02"))
	require.NotNil(t, saved.Deadline)
	require.Equal(t, "2026-07-01", saved.Deadline.Format("2006-01-02"))
	require.Len(t, saved.Tags, 2)
	require.True(t, saved.Someday)
}

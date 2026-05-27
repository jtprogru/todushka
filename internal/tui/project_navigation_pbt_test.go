package tui

// T-7 — Property-based tests for the project-navigation feature.
//
// Coverage map (design.md §2.6):
//   - CP-1  TestProp_ScreenEntryRoundTrip
//   - CP-2  TestProp_GTDKeysAbsent
//   - CP-3  TestProp_StatusFilterEquiv
//   - CP-4  TestProp_SortStable
//   - CP-5  TestProp_CursorBounded
//   - CP-6  TestProp_DeleteReassignsTasks   (in service)
//   - CP-7  TestProp_SoftDeleteInvisible    (in service)
//   - CP-8  TestProp_EmptyProjectGuard      (in service)
//   - CP-9  TestProp_ReadOnlyBlocks
//   - CP-10 TestProp_EditorSaveRoundTrip
//   - CP-11 TestProp_EditorInvalidStaysOpen
//   - CP-12 TestProp_ZoomRoundTrip
//   - CP-13 TestProp_ProjectTasksFilter
//   - CP-14 TestProp_ModeLabelProjects
//   - CP-15 TestProp_PKeyIgnoredInTasks

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// CP-1
func TestProp_ScreenEntryRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := newTestModel(t)
		require.Equal(rt, screenList, m.screen)
		// Enter screenProjects.
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
		m = mm.(Model)
		require.Equal(rt, screenProjects, m.screen)
		// Either P or Esc returns.
		exitKey := rapid.SampledFrom([]tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'P'}},
			{Type: tea.KeyEsc},
		}).Draw(rt, "exitKey")
		mm, _ = m.Update(exitKey)
		require.Equal(rt, screenList, mm.(Model).screen)
	})
}

// CP-2
func TestProp_GTDKeysAbsent(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		blockedKeys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'1'}},
			{Type: tea.KeyRunes, Runes: []rune{'2'}},
			{Type: tea.KeyRunes, Runes: []rune{'3'}},
			{Type: tea.KeyRunes, Runes: []rune{'4'}},
			{Type: tea.KeyRunes, Runes: []rune{'5'}},
			{Type: tea.KeyRunes, Runes: []rune{'6'}},
			{Type: tea.KeyTab},
			{Type: tea.KeyShiftTab},
		}
		k := rapid.SampledFrom(blockedKeys).Draw(rt, "key")
		m := newTestModel(t)
		m.screen = screenProjects
		origList := m.activeList
		mm, _ := m.Update(k)
		require.Equal(rt, screenProjects, mm.(Model).screen)
		require.Equal(rt, origList, mm.(Model).activeList)
	})
}

// CP-3 — Status filter equivalence: displayedProjects equals expected set.
func TestProp_StatusFilterEquiv(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		all := make([]project.Project, 0, n)
		for i := 0; i < n; i++ {
			st := rapid.SampledFrom([]project.Status{
				project.StatusOpen, project.StatusCompleted, project.StatusCancelled,
			}).Draw(rt, "st")
			all = append(all, project.Project{ID: id.New(), Name: "x", Status: st})
		}
		// Simulate service-side filter (the service drops soft-deleted always).
		filterOpen := make([]project.Project, 0)
		filterAll := make([]project.Project, 0)
		for _, p := range all {
			filterAll = append(filterAll, p)
			if p.Status == project.StatusOpen {
				filterOpen = append(filterOpen, p)
			}
		}
		// The Model receives whatever the service emitted; verify
		// displayedProjects returns it as-is when no filterQuery.
		m := newTestModel(t)
		m.projects = filterOpen
		require.Len(rt, displayedProjects(m), len(filterOpen))
		m.projects = filterAll
		require.Len(rt, displayedProjects(m), len(filterAll))
	})
}

// CP-4
func TestProp_SortStable(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(0, 6).Draw(rt, "n")
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		for i := 0; i < n; i++ {
			name := rapid.StringMatching(`[A-Za-z]{1,5}`).Draw(rt, "name")
			pos := rapid.Int64Range(0, 20).Draw(rt, "pos")
			require.NoError(rt, svc.Repo().ProjectCreate(context.Background(), project.Project{
				ID: id.New(), Name: name, Status: project.StatusOpen, Position: pos,
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			}))
		}
		got1, err := svc.ListProjectsSorted(context.Background(), nil, true)
		require.NoError(rt, err)
		got2, err := svc.ListProjectsSorted(context.Background(), nil, true)
		require.NoError(rt, err)
		require.Equal(rt, got1, got2)
		// Verify sort invariants.
		for i := 1; i < len(got1); i++ {
			a, b := got1[i-1], got1[i]
			if a.Position != b.Position {
				require.LessOrEqual(rt, a.Position, b.Position)
			} else {
				require.LessOrEqual(rt, strings.ToLower(a.Name), strings.ToLower(b.Name))
			}
		}
	})
}

// CP-5
func TestProp_CursorBounded(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(0, 6).Draw(rt, "n")
		m := newTestModel(t)
		m.screen = screenProjects
		m.projects = make([]project.Project, n)
		for i := range m.projects {
			m.projects[i] = project.Project{ID: id.New(), Name: "x", Status: project.StatusOpen}
		}
		keys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'j'}},
			{Type: tea.KeyRunes, Runes: []rune{'k'}},
		}
		steps := rapid.IntRange(0, 15).Draw(rt, "steps")
		for i := 0; i < steps; i++ {
			k := rapid.SampledFrom(keys).Draw(rt, "k")
			mm, _ := m.Update(k)
			m = mm.(Model)
		}
		if n == 0 {
			require.Equal(rt, 0, m.projectCursor)
		} else {
			require.GreaterOrEqual(rt, m.projectCursor, 0)
			require.Less(rt, m.projectCursor, n)
		}
	})
}

// CP-6 — service-level.
func TestProp_DeleteReassignsTasks(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		ctx := context.Background()
		p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
		pid := p.ID
		tIDs := make([]id.ID, 0, n)
		for i := 0; i < n; i++ {
			tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "t", ProjectID: &pid})
			require.NoError(rt, err)
			tIDs = append(tIDs, tk.ID)
		}
		require.NoError(rt, svc.DeleteProject(ctx, pid, true))
		for _, tid := range tIDs {
			got, err := svc.Repo().TaskGet(ctx, tid)
			require.NoError(rt, err)
			require.Nil(rt, got.ProjectID)
			require.Nil(rt, got.HeadingID)
		}
	})
}

// CP-7 — service-level.
func TestProp_SoftDeleteInvisible(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		ctx := context.Background()
		p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
		require.NoError(rt, svc.DeleteProject(ctx, p.ID, true))
		got, err := svc.ListProjectsSorted(ctx, nil, true)
		require.NoError(rt, err)
		for _, x := range got {
			require.NotEqual(rt, p.ID, x.ID)
		}
	})
}

// CP-8 — empty-project guard.
func TestProp_EmptyProjectGuard(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		hasTasks := rapid.Bool().Draw(rt, "hasTasks")
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		ctx := context.Background()
		p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "X"})
		pid := p.ID
		var taskID id.ID
		if hasTasks {
			tk, _ := svc.AddTask(ctx, app.AddTaskInput{Title: "t", ProjectID: &pid})
			taskID = tk.ID
		}
		err := svc.DeleteProject(ctx, pid, false)
		if hasTasks {
			require.ErrorIs(rt, err, app.ErrProjectNotEmpty)
			got, getErr := svc.Repo().TaskGet(ctx, taskID)
			require.NoError(rt, getErr)
			require.NotNil(rt, got.ProjectID)
		} else {
			require.NoError(rt, err)
		}
	})
}

// CP-9
func TestProp_ReadOnlyBlocks(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		keys := []rune{'n', 'e', 'd'}
		k := rapid.SampledFrom(keys).Draw(rt, "key")
		m := newTestModel(t)
		// Pre-seed at least one project so 'e' and 'd' have a cursor target.
		m.projects = []project.Project{{ID: id.New(), Name: "X", Status: project.StatusOpen}}
		m.readOnly = true
		m.screen = screenProjects
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{k}})
		got := mm.(Model)
		require.False(rt, got.editingProject, "editor must not open in RO")
		require.Nil(rt, got.confirm, "confirm must not open in RO")
		require.Contains(rt, got.statusMsg, "read-only")
	})
}

// CP-10
func TestProp_EditorSaveRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		name := rapid.StringMatching(`[A-Za-z][A-Za-z0-9 ]{0,20}`).Draw(rt, "name")
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		m := newProjectEditor(true, nil, svc)
		m.name.SetValue(name)
		p, created, err := m.ApplyAndSave(context.Background(), svc)
		require.NoError(rt, err)
		require.True(rt, created)
		// ApplyAndSave trims whitespace; compare against the trimmed input.
		require.Equal(rt, strings.TrimSpace(name), p.Name)
		list, err := svc.ListProjectsSorted(context.Background(), nil, true)
		require.NoError(rt, err)
		found := false
		for _, x := range list {
			if x.ID == p.ID {
				found = true
				break
			}
		}
		require.True(rt, found)
	})
}

// CP-11
func TestProp_EditorInvalidStaysOpen(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		invalidKind := rapid.IntRange(0, 2).Draw(rt, "kind")
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		m := newProjectEditor(true, nil, svc)
		switch invalidKind {
		case 0:
			// Empty name (default).
		case 1:
			m.name.SetValue("ok")
			m.area.SetValue("nosuch")
		case 2:
			m.name.SetValue("ok")
			m.deadline.SetValue("bad-date")
		}
		_, _, err := m.ApplyAndSave(context.Background(), svc)
		require.Error(rt, err)
	})
}

// CP-12
func TestProp_ZoomRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 4).Draw(rt, "n")
		m, svc := newTestModelWithService(t)
		ctx := context.Background()
		ps := make([]project.Project, 0, n)
		for i := 0; i < n; i++ {
			p, _ := svc.AddProject(ctx, app.AddProjectInput{Name: rapid.StringMatching(`[A-Za-z]+`).Draw(rt, "name")})
			ps = append(ps, p)
		}
		m.projects = ps
		m.screen = screenProjects
		m.projectCursor = rapid.IntRange(0, n-1).Draw(rt, "cursor")
		want := m.projectCursor
		// Enter zoom.
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = mm.(Model)
		require.Equal(rt, screenProjectTasks, m.screen)
		// Esc back.
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = mm.(Model)
		require.Equal(rt, screenProjects, m.screen)
		require.Equal(rt, want, m.projectCursor, "projectCursor must be preserved across zoom round-trip")
	})
}

// CP-13
func TestProp_ProjectTasksFilter(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
		ctx := context.Background()
		pA, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "A"})
		pB, _ := svc.AddProject(ctx, app.AddProjectInput{Name: "B"})
		nA := rapid.IntRange(0, 4).Draw(rt, "nA")
		nB := rapid.IntRange(0, 4).Draw(rt, "nB")
		aid := pA.ID
		bid := pB.ID
		for i := 0; i < nA; i++ {
			_, _ = svc.AddTask(ctx, app.AddTaskInput{Title: "a", ProjectID: &aid})
		}
		for i := 0; i < nB; i++ {
			_, _ = svc.AddTask(ctx, app.AddTaskInput{Title: "b", ProjectID: &bid})
		}
		msg := fetchProjectTasks(svc, aid)()
		ptm := msg.(projectTasksLoadedMsg)
		require.Equal(rt, nA, len(ptm.tasks))
		for _, t := range ptm.tasks {
			require.NotNil(rt, t.ProjectID)
			require.Equal(rt, aid, *t.ProjectID)
		}
	})
}

// CP-14
func TestProp_ModeLabelProjects(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Confirm/filter/help/editor and confirm: explicitly disabled.
		m := newTestModel(t)
		zoom := rapid.Bool().Draw(rt, "zoom")
		if zoom {
			m.screen = screenProjectTasks
			pid := id.New()
			m.activeProjectID = &pid
		} else {
			m.screen = screenProjects
		}
		require.Equal(rt, modeProjects, currentMode(m))
		require.Equal(rt, "PROJECTS", modeProjects.modeLabel())
	})
}

// CP-15
func TestProp_PKeyIgnoredInTasks(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m, svc := newTestModelWithService(t)
		p, _ := svc.AddProject(context.Background(), app.AddProjectInput{Name: "X"})
		pid := p.ID
		m.screen = screenProjectTasks
		m.activeProjectID = &pid
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
		require.Equal(rt, screenProjectTasks, mm.(Model).screen)
	})
}

// Keep imports used (storage, task) referenced for go vet.
var (
	_ = storage.ErrNotFound
	_ = task.StatusOpen
)

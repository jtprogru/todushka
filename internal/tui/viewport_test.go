package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ─── T-1.1: exploration test that reproduces BL-7 ────────────────────

// TestModel_CursorInvisibleOnOverflow demonstrates the bug: with 30 tasks
// and a small terminal (height=12 → visibleRows around 8), pressing j 20
// times moves m.cursor to 20, but the renderer must keep that cursor
// inside the visible window. Before the fix, m.scrollOffset doesn't exist
// and the cursor "leaves" the screen.
func TestModel_CursorInvisibleOnOverflow(t *testing.T) {
	m := newTestModel(t)
	m.height = 12 // small terminal
	m.width = 80
	// Populate 30 tasks.
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: "t", Status: task.StatusOpen}
	}
	m.tasks = tasks

	// Press 'j' 20 times.
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	require.Equal(t, 20, m.cursor)
	vr := visibleRows(m)
	require.Greater(t, vr, 0, "visibleRows must be positive at height=12")
	// After fix: cursor must be inside the visible window [offset, offset+vr).
	require.GreaterOrEqual(t, m.cursor, m.scrollOffset,
		"cursor must be >= scrollOffset (would mean cursor is above window)")
	require.Less(t, m.cursor, m.scrollOffset+vr,
		"cursor must be < scrollOffset+visibleRows (would mean cursor is below window)")
}

// ─── T-1.2: ensureCursorVisible unit tests ──────────────────────────

func TestEnsureCursorVisible_FitsInVisible(t *testing.T) {
	require.Equal(t, 0, ensureCursorVisible(2, 0, 10, 3, 5))
}

func TestEnsureCursorVisible_CursorAtStart(t *testing.T) {
	require.Equal(t, 0, ensureCursorVisible(0, 0, 10, 3, 50))
}

func TestEnsureCursorVisible_CursorAtEnd(t *testing.T) {
	// total=20, visible=10 → maxOffset=10. cursor=19 → must be at the bottom
	// of the window, so offset should be 10 (cursor at index 9 within window).
	require.Equal(t, 10, ensureCursorVisible(19, 0, 10, 3, 20))
}

func TestEnsureCursorVisible_CursorMovesDownIntoBuffer(t *testing.T) {
	// total=20, visible=10, scrolloff=3. offset=0. cursor=7 → still has
	// scrolloff (10-1-7 = 2 < 3) → must shift offset to keep 3 buffer.
	// new offset = cursor - visible + 1 + scrolloff = 7 - 10 + 1 + 3 = 1.
	got := ensureCursorVisible(7, 0, 10, 3, 20)
	require.Equal(t, 1, got)
}

func TestEnsureCursorVisible_CursorMovesUpIntoBuffer(t *testing.T) {
	// total=20, visible=10, scrolloff=3. offset=10, cursor=12.
	// cursor - offset = 2 < scrolloff(3) → shift up: offset = cursor - scrolloff = 9.
	require.Equal(t, 9, ensureCursorVisible(12, 10, 10, 3, 20))
}

func TestEnsureCursorVisible_ClampOnReload(t *testing.T) {
	// offset=10, total shrinks to 5, visible=10 → maxOffset=0 → 0.
	require.Equal(t, 0, ensureCursorVisible(2, 10, 10, 3, 5))
}

func TestEnsureCursorVisible_VisibleZero(t *testing.T) {
	require.Equal(t, 0, ensureCursorVisible(5, 3, 0, 3, 20))
}

func TestEnsureCursorVisible_NegativeCursor(t *testing.T) {
	require.Equal(t, 0, ensureCursorVisible(-1, 5, 10, 3, 20))
}

func TestEnsureCursorVisible_NoMovementWhenInBuffer(t *testing.T) {
	// cursor=10, visible=10, offset=5 → cursor in [5,15), buffer
	// 10-5=5 >= 3 and 14-10=4 >= 3 → no movement, return 5.
	require.Equal(t, 5, ensureCursorVisible(10, 5, 10, 3, 20))
}

// ─── T-1.4: PBT for invariants ───────────────────────────────────────

func TestProp_OffsetBounded(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		total := rapid.IntRange(0, 100).Draw(rt, "total")
		visible := rapid.IntRange(0, 50).Draw(rt, "visible")
		cursor := rapid.IntRange(-5, 100).Draw(rt, "cursor")
		off := rapid.IntRange(-5, 100).Draw(rt, "off")
		got := ensureCursorVisible(cursor, off, visible, 3, total)
		require.GreaterOrEqual(rt, got, 0)
		maxOff := total - visible
		if maxOff < 0 {
			maxOff = 0
		}
		require.LessOrEqual(rt, got, maxOff,
			"offset must be <= max(0, total-visible)")
	})
}

func TestProp_CursorAlwaysInside(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		total := rapid.IntRange(1, 100).Draw(rt, "total")
		visible := rapid.IntRange(1, 50).Draw(rt, "visible")
		cursor := rapid.IntRange(0, total-1).Draw(rt, "cursor")
		off := rapid.IntRange(0, total).Draw(rt, "off")
		got := ensureCursorVisible(cursor, off, visible, 3, total)
		require.GreaterOrEqual(rt, cursor, got,
			"cursor must be >= offset")
		require.Less(rt, cursor, got+visible,
			"cursor must be < offset + visible")
	})
}

// ─── T-4: integration tests ──────────────────────────────────────────

func TestModel_ViewList_CursorVisibleAfterJ(t *testing.T) {
	m := newTestModel(t)
	m.height = 20
	m.width = 80
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: "t", Status: task.StatusOpen}
	}
	m.tasks = tasks
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
		vr := visibleRows(m)
		if vr <= 0 {
			t.Fatalf("visibleRows=0 at iter=%d", i)
		}
		require.GreaterOrEqual(t, m.cursor, m.scrollOffset,
			"cursor must be >= scrollOffset at iter=%d", i)
		require.Less(t, m.cursor, m.scrollOffset+vr,
			"cursor must be < scrollOffset+vr at iter=%d (cursor=%d, off=%d, vr=%d)",
			i, m.cursor, m.scrollOffset, vr)
	}
}

func TestModel_ViewProjectList_CursorVisibleAfterJ(t *testing.T) {
	m := newTestModel(t)
	m.screen = screenProjects
	m.height = 20
	m.width = 80
	projects := make([]projectStub, 30)
	for i := range projects {
		projects[i] = newOpenProject()
	}
	m.projects = stubsToProjects(projects)
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
		// Account for "Projects" header + blank line consuming 2 rows.
		vr := visibleRows(m) - 2
		if vr <= 0 {
			t.Skipf("visibleRows-header=%d <=0 at iter=%d", vr, i)
		}
		require.GreaterOrEqual(t, m.projectCursor, m.projectScrollOffset,
			"projectCursor >= projectScrollOffset at iter=%d", i)
		require.Less(t, m.projectCursor, m.projectScrollOffset+vr,
			"projectCursor < projectScrollOffset+vr at iter=%d (cursor=%d, off=%d, vr=%d)",
			i, m.projectCursor, m.projectScrollOffset, vr)
	}
}

func TestModel_ViewProjectTasks_CursorVisibleAfterJ(t *testing.T) {
	m := newTestModel(t)
	m.height = 20
	m.width = 80
	pid := id.New()
	m.activeProjectID = &pid
	m.screen = screenProjectTasks
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: "t", Status: task.StatusOpen}
	}
	m.tasks = tasks
	m.projectTasks = tasks
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
		vr := visibleRows(m) - 2 // header+blank
		if vr <= 0 {
			t.Skipf("visibleRows-header=%d <=0 at iter=%d", vr, i)
		}
		require.GreaterOrEqual(t, m.cursor, m.scrollOffset)
		require.Less(t, m.cursor, m.scrollOffset+vr)
	}
}

func TestModel_ScrollOffsetResetOnSwitchList(t *testing.T) {
	m := newTestModel(t)
	m.height = 20
	m.width = 80
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: "t", Status: task.StatusOpen}
	}
	m.tasks = tasks
	// Scroll down so offset > 0.
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	require.Greater(t, m.scrollOffset, 0, "precondition: scrollOffset must be > 0")
	// Switch list via Tab.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mm.(Model)
	require.Equal(t, 0, m.scrollOffset, "switchList must reset scrollOffset")
	require.Equal(t, 0, m.cursor, "switchList must reset cursor")
}

func TestModel_ScrollOffsetClampOnTasksReload(t *testing.T) {
	m := newTestModel(t)
	m.height = 20
	m.width = 80
	m.scrollOffset = 25 // pretend we were scrolled deep
	m.cursor = 25
	// Reload with a shorter list.
	short := make([]task.Task, 5)
	for i := range short {
		short[i] = task.Task{ID: id.New(), Title: "t", Status: task.StatusOpen}
	}
	mm, _ := m.Update(tasksLoadedMsg{tasks: short})
	m = mm.(Model)
	// totalRows=5, visibleRows>5 → offset must collapse to 0.
	require.Equal(t, 0, m.scrollOffset)
	require.LessOrEqual(t, m.cursor, len(m.tasks)-1)
}

// Helpers for project stubs.

type projectStub struct{ id id.ID }

func newOpenProject() projectStub { return projectStub{id: id.New()} }

func stubsToProjects(stubs []projectStub) []project.Project {
	out := make([]project.Project, len(stubs))
	for i, s := range stubs {
		out[i] = project.Project{ID: s.id, Name: "p", Status: project.StatusOpen}
	}
	return out
}

func TestProp_ScrollOffBufferRespected(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Pick params where buffer is meaningful: total > visible + 2*scrolloff.
		visible := rapid.IntRange(10, 30).Draw(rt, "visible")
		extra := rapid.IntRange(7, 50).Draw(rt, "extra") // > 2*scrolloff
		total := visible + extra
		cursor := rapid.IntRange(0, total-1).Draw(rt, "cursor")
		off := rapid.IntRange(0, total).Draw(rt, "off")
		got := ensureCursorVisible(cursor, off, visible, 3, total)
		// Top buffer: cursor - got >= 3, unless cursor near list start (cursor < 3).
		if cursor >= 3 {
			require.GreaterOrEqual(rt, cursor-got, 3,
				"top scrolloff buffer must be >= 3 when not near start")
		}
		// Bottom buffer: (got + visible - 1) - cursor >= 3, unless cursor near end.
		if cursor < total-3 {
			require.GreaterOrEqual(rt, (got+visible-1)-cursor, 3,
				"bottom scrolloff buffer must be >= 3 when not near end")
		}
	})
}

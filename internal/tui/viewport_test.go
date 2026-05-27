package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ─── windowLines unit tests ──────────────────────────────────────────

func TestWindowLines_FitsInVisible(t *testing.T) {
	got := windowLines([]string{"a", "b", "c"}, 1, 10, 3)
	require.Equal(t, "a\nb\nc", got)
}

func TestWindowLines_VisibleZero(t *testing.T) {
	got := windowLines([]string{"a", "b", "c"}, 1, 0, 3)
	require.Equal(t, "a\nb\nc", got)
}

func TestWindowLines_CursorMissing(t *testing.T) {
	got := windowLines([]string{"a", "b", "c", "d", "e"}, -1, 2, 3)
	require.Equal(t, "a\nb", got)
}

func TestWindowLines_CursorAtStart(t *testing.T) {
	// 20 lines, visible=10, cursor at line 0, scrolloff=3.
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = fmt.Sprintf("L%02d", i)
	}
	got := windowLines(lines, 0, 10, 3)
	// Window should be lines[0:10].
	require.Equal(t, strings.Join(lines[0:10], "\n"), got)
}

func TestWindowLines_CursorAtEnd(t *testing.T) {
	// 20 lines, visible=10, cursor at line 19.
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = fmt.Sprintf("L%02d", i)
	}
	got := windowLines(lines, 19, 10, 3)
	// Cursor at end → window anchored to end: lines[10:20].
	require.Equal(t, strings.Join(lines[10:20], "\n"), got)
}

func TestWindowLines_CursorInMiddle(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = fmt.Sprintf("L%02d", i)
	}
	// Cursor at line 10, scrolloff=3 → window starts at 10-3=7, ends at 17.
	got := windowLines(lines, 10, 10, 3)
	require.Equal(t, strings.Join(lines[7:17], "\n"), got)
}

func TestWindowLines_NarrowVisible(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e"}
	got := windowLines(lines, 3, 1, 3) // visible=1, scrolloff=3 — narrow window
	require.Equal(t, "d", got, "narrow window must still contain cursor line")
}

func TestProp_WindowLines_CursorAlwaysInside(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		total := rapid.IntRange(1, 100).Draw(rt, "total")
		visible := rapid.IntRange(1, 50).Draw(rt, "visible")
		cursor := rapid.IntRange(0, total-1).Draw(rt, "cursor")
		lines := make([]string, total)
		for i := range lines {
			lines[i] = fmt.Sprintf("L%03d", i)
		}
		out := windowLines(lines, cursor, visible, 3)
		require.Contains(rt, out, fmt.Sprintf("L%03d", cursor),
			"cursor's line must be present in windowed output")
	})
}

// ─── Real-render integration tests: assert cursor marker visible ────

// findCursorLine returns the line in `out` that starts with the cursor
// marker `> `, or empty string if not found. Lines are split on '\n'.
// Output is assumed to have ANSI styling stripped (termenv.Ascii).
func findCursorLine(out string) string {
	for _, ln := range strings.Split(out, "\n") {
		trimmed := strings.TrimLeft(ln, " ")
		if strings.HasPrefix(trimmed, "> ") {
			return ln
		}
	}
	return ""
}

func TestModel_ViewList_CursorInRenderedOutput_NoWrap(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.height = 20
	m.width = 80
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("task-%02d", i), Status: task.StatusOpen}
	}
	m.tasks = tasks
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	require.Equal(t, 20, m.cursor)
	line := findCursorLine(m.View())
	require.NotEmpty(t, line)
	require.Contains(t, line, "task-20")
}

func TestModel_ViewList_CursorInRenderedOutput_LongTitlesWrap(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.height = 20
	m.width = 80
	long := strings.Repeat("very long title text that definitely wraps ", 3)
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("%02d %s", i, long), Status: task.StatusOpen}
	}
	m.tasks = tasks
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	require.NotEmpty(t, findCursorLine(m.View()),
		"cursor marker '> ' must be present in rendered output even with wrapping titles")
}

func TestModel_ViewList_CursorInRenderedOutput_DualPane(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.height = 20
	m.width = 150
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("task-%02d", i), Status: task.StatusOpen}
	}
	m.tasks = tasks
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	line := findCursorLine(m.View())
	require.NotEmpty(t, line)
	require.Contains(t, line, "task-20")
}

func TestModel_ViewList_CursorInRenderedOutput_SmallTerminal(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.height = 12
	m.width = 60
	tasks := make([]task.Task, 40)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("task-%03d", i), Status: task.StatusOpen}
	}
	m.tasks = tasks
	for i := 0; i < 30; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	line := findCursorLine(m.View())
	require.NotEmpty(t, line)
	require.Contains(t, line, "task-030")
}

func TestModel_ViewProjectList_CursorInRenderedOutput(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.screen = screenProjects
	m.height = 20
	m.width = 80
	projects := make([]project.Project, 30)
	for i := range projects {
		projects[i] = project.Project{ID: id.New(), Name: fmt.Sprintf("proj-%02d", i), Status: project.StatusOpen}
	}
	m.projects = projects
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	line := findCursorLine(m.View())
	require.NotEmpty(t, line, "cursor marker must be present in screenProjects after 20×j")
	require.Contains(t, line, "proj-20")
}

func TestModel_ViewProjectTasks_CursorInRenderedOutput(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	pid := id.New()
	m.screen = screenProjectTasks
	m.activeProjectID = &pid
	m.projects = []project.Project{{ID: pid, Name: "P", Status: project.StatusOpen}}
	m.height = 20
	m.width = 80
	tasks := make([]task.Task, 30)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("zoom-%02d", i), Status: task.StatusOpen}
	}
	m.tasks = tasks
	m.projectTasks = tasks
	for i := 0; i < 20; i++ {
		mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = mm.(Model)
	}
	line := findCursorLine(m.View())
	require.NotEmpty(t, line, "cursor marker must be present in screenProjectTasks after 20×j")
	require.Contains(t, line, "zoom-20")
}

func TestModel_ViewList_BodyLineCountBounded(t *testing.T) {
	// After windowing, the body should not exceed the available rows
	// (visibleRows). This prevents lipgloss MaxHeight from cutting any
	// rendered content unexpectedly.
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.height = 15
	m.width = 80
	tasks := make([]task.Task, 50)
	for i := range tasks {
		tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("task-%02d", i), Status: task.StatusOpen}
	}
	m.tasks = tasks
	m.cursor = 25

	body := m.viewList()
	bodyLines := strings.Count(body, "\n") + 1
	require.LessOrEqual(t, bodyLines, visibleRows(m),
		"viewList body must not exceed visibleRows (%d), got %d lines",
		visibleRows(m), bodyLines)
}

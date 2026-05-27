package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// taskWithDates returns a Task with both StartDate and Deadline set.
func taskWithDates(title string, status task.Status) task.Task {
	now := time.Now()
	start := task.NewDate(time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC))
	due := task.NewDate(time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC))
	return task.Task{
		ID:        id.New(),
		Title:     title,
		Status:    status,
		StartDate: &start,
		Deadline:  &due,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// countLeadingSpaces returns the number of leading ASCII spaces in s.
func countLeadingSpaces(s string) int {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return i
}

// expectedTitleStartCol mirrors viewList's prefix composition for the task
// at cursor in the given model (no selection prefix, marker is the styled
// "> ", icon is two spaces for open status, short is theme-Dim 6-char id,
// followed by two literal spaces). Returns lipgloss-aware visual width.
func expectedTitleStartCol(m Model, tk task.Task) int {
	marker := m.theme.Selected.Render("> ")
	icon := "  "
	switch tk.Status {
	case task.StatusCompleted:
		icon = m.theme.StatusInfo.Render("✓ ")
	case task.StatusCancelled:
		icon = m.theme.StatusError.Render("✗ ")
	}
	short := m.theme.Dim.Render(id.Short(tk.ID))
	return lipgloss.Width(marker + icon + short + "  ")
}

// ─── BL-1: dates removed from list ──────────────────────────────────────

func TestTUI_ViewListOmitsDates(t *testing.T) {
	m := newTestModel(t)
	m.tasks = []task.Task{taskWithDates("task with dates", task.StatusOpen)}
	m.activeList = listInbox

	out := m.viewList()

	require.NotContains(t, out, "start:", "viewList must not include start: prefix")
	require.NotContains(t, out, "due:", "viewList must not include due: prefix")
}

func TestTUI_ViewDetailsKeepsDates(t *testing.T) {
	m := newTestModel(t)
	m.tasks = []task.Task{taskWithDates("detail task", task.StatusOpen)}
	m.activeList = listInbox
	m.cursor = 0

	out := viewDetails(m, 60)

	require.Contains(t, out, "Start:", "viewDetails must render Start label")
	require.Contains(t, out, "Due:", "viewDetails must render Due label")
	require.Contains(t, out, "2026-05-26", "viewDetails must show start date value")
	require.Contains(t, out, "2026-05-27", "viewDetails must show due date value")
}

// ─── BL-3: heavy separators ─────────────────────────────────────────────

func TestTUI_RenderSeparatorHeavy(t *testing.T) {
	s := renderSeparator(NewTheme(), 10)
	require.Equal(t, 10, strings.Count(s, "━"), "expected 10 heavy horizontal runes")
	require.Equal(t, 0, strings.Count(s, "─"), "must not contain light horizontal rune")
}

func TestTUI_RenderSeparatorBoundary(t *testing.T) {
	require.Equal(t, "", renderSeparator(NewTheme(), 0))
	require.Equal(t, "", renderSeparator(NewTheme(), -5))
}

// ─── BL-4: title-column wrapping with hanging indent ────────────────────

// longTitle returns a title of given rune length, all letters so it
// soft-wraps cleanly via lipgloss whitespace breaking when spaces are
// present, or hard-wraps when they aren't. We use letters with spaces
// so lipgloss can break at word boundaries.
func longTitle(words int) string {
	parts := make([]string, words)
	for i := range parts {
		parts[i] = "word"
	}
	return strings.Join(parts, " ")
}

func TestTUI_ViewListWrapsTitleWithHangingIndent(t *testing.T) {
	m := newTestModel(t)
	m.width = 60
	m.config.DualPaneMinWidth = 100 // force single-pane
	m.tasks = []task.Task{
		{
			ID:        id.New(),
			Title:     longTitle(50),
			Status:    task.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	m.activeList = listInbox

	out := m.viewList()
	lines := strings.Split(out, "\n")
	require.Greater(t, len(lines), 1, "long title must wrap to more than 1 line at width=60")

	titleStartCol := expectedTitleStartCol(m, m.tasks[0])
	require.Greater(t, titleStartCol, 0, "expected title start column to be positive")

	for i := 1; i < len(lines); i++ {
		require.Equal(t, titleStartCol, countLeadingSpaces(lines[i]),
			"continuation line %d must start with exactly %d leading spaces (hanging indent)", i, titleStartCol)
	}
}

func TestTUI_ViewListNoWrapWhenWidthZero(t *testing.T) {
	m := newTestModel(t)
	m.width = 0
	m.tasks = []task.Task{
		{
			ID:        id.New(),
			Title:     longTitle(50),
			Status:    task.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	m.activeList = listInbox

	out := m.viewList()
	lines := strings.Split(out, "\n")
	require.Equal(t, 1, len(lines), "with m.width=0 each task must render on exactly 1 line")
}

func TestTUI_ViewListStrikethroughOnWrappedLines(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.width = 60
	m.config.DualPaneMinWidth = 100
	now := time.Now()
	m.tasks = []task.Task{
		{
			ID:          id.New(),
			Title:       longTitle(40),
			Status:      task.StatusCompleted,
			CompletedAt: &now,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	m.activeList = listInbox

	out := m.viewList()
	lines := strings.Split(out, "\n")
	require.Greater(t, len(lines), 1, "long completed title must wrap")

	for i, line := range lines {
		require.Contains(t, line, "\x1b[9",
			"wrapped completed line %d must carry strikethrough ANSI escape", i)
	}
}

func TestTUI_ViewListCursorMarkerOnFirstLineOnly(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.width = 60
	m.config.DualPaneMinWidth = 100
	m.tasks = []task.Task{
		{
			ID:        id.New(),
			Title:     longTitle(40),
			Status:    task.StatusOpen,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	m.activeList = listInbox
	m.cursor = 0

	cursorMarker := m.theme.Selected.Render("> ")
	out := m.viewList()
	require.Equal(t, 1, strings.Count(out, cursorMarker),
		"cursor marker must appear exactly once in the wrapped task block")
}

// ─── Property-based tests ───────────────────────────────────────────────

// TestProp_ViewListOmitsDates verifies CP-1: for any combination of
// start/due assignments, the rendered list never contains start:/due:
// prefixes.
func TestProp_ViewListOmitsDates(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 4).Draw(rt, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = fmt.Sprintf("task-%d", i)
		}
		m, _, tasks := setupRapidModel(rt, titles...)
		for i := range tasks {
			if rapid.Bool().Draw(rt, fmt.Sprintf("hasStart-%d", i)) {
				d := task.NewDate(time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC))
				m.tasks[i].StartDate = &d
			}
			if rapid.Bool().Draw(rt, fmt.Sprintf("hasDue-%d", i)) {
				d := task.NewDate(time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC))
				m.tasks[i].Deadline = &d
			}
		}
		out := m.viewList()
		require.NotContains(rt, out, "start:")
		require.NotContains(rt, out, "due:")
	})
}

// TestProp_SeparatorHeavy verifies CP-3: renderSeparator emits exactly
// width heavy horizontal runes for any positive width.
func TestProp_SeparatorHeavy(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(1, 200).Draw(rt, "width")
		s := renderSeparator(NewTheme(), w)
		require.Equal(rt, w, strings.Count(s, "━"), "heavy rune count mismatch")
		require.Equal(rt, 0, strings.Count(s, "─"), "light rune must not appear")
	})
}

// TestProp_SeparatorBoundary verifies CP-4: width <= 0 returns "".
func TestProp_SeparatorBoundary(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(-50, 0).Draw(rt, "width")
		s := renderSeparator(NewTheme(), w)
		require.Equal(rt, "", s)
	})
}

// TestProp_TitleWrapHangingIndent verifies CP-5: continuation lines of
// a wrapping task start at the same visual column as the title's first
// character.
func TestProp_TitleWrapHangingIndent(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		words := rapid.IntRange(20, 60).Draw(rt, "words")
		width := rapid.IntRange(40, 80).Draw(rt, "width")

		m := newTestModel(t)
		m.width = width
		m.config.DualPaneMinWidth = 200 // force single-pane regardless of width
		m.tasks = []task.Task{
			{
				ID:        id.New(),
				Title:     longTitle(words),
				Status:    task.StatusOpen,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		m.activeList = listInbox

		out := m.viewList()
		lines := strings.Split(out, "\n")
		if len(lines) <= 1 {
			return // title fit on one line; nothing to assert
		}
		col := expectedTitleStartCol(m, m.tasks[0])
		for i := 1; i < len(lines); i++ {
			require.Equal(rt, col, countLeadingSpaces(lines[i]),
				"continuation line %d must align with title column", i)
		}
	})
}

// TestProp_NoWrapWidthZero verifies CP-6: when m.width == 0, every
// rendered task occupies exactly 1 line.
func TestProp_NoWrapWidthZero(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 5).Draw(rt, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = longTitle(rapid.IntRange(1, 80).Draw(rt, fmt.Sprintf("w-%d", i)))
		}
		m, _, _ := setupRapidModel(rt, titles...)
		m.width = 0

		out := m.viewList()
		lines := strings.Split(out, "\n")
		require.Equal(rt, n, len(lines), "with m.width=0 number of rendered lines must equal number of tasks")
	})
}

// TestProp_StrikethroughPropagatesAcrossWrap verifies CP-7: for
// Completed/Cancelled tasks whose title wraps, every wrapped line
// carries the strikethrough ANSI escape.
func TestProp_StrikethroughPropagatesAcrossWrap(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	rapid.Check(t, func(rt *rapid.T) {
		words := rapid.IntRange(25, 60).Draw(rt, "words")
		width := rapid.IntRange(40, 70).Draw(rt, "width")
		statusChoice := rapid.IntRange(0, 1).Draw(rt, "status")
		status := task.StatusCompleted
		if statusChoice == 1 {
			status = task.StatusCancelled
		}

		m := newTestModel(t)
		m.width = width
		m.config.DualPaneMinWidth = 200
		now := time.Now()
		tk := task.Task{
			ID:        id.New(),
			Title:     longTitle(words),
			Status:    status,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if status == task.StatusCompleted {
			tk.CompletedAt = &now
		} else {
			tk.CancelledAt = &now
		}
		m.tasks = []task.Task{tk}
		m.activeList = listInbox

		out := m.viewList()
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			require.Contains(rt, line, "\x1b[9",
				"wrapped line %d must carry strikethrough", i)
		}
	})
}

// TestProp_SingleCursorMarker verifies CP-8: the cursor marker appears
// exactly once across all rendered lines, regardless of which task is
// under the cursor.
func TestProp_SingleCursorMarker(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 4).Draw(rt, "n")
		width := rapid.IntRange(40, 80).Draw(rt, "width")
		cursorIdx := rapid.IntRange(0, n-1).Draw(rt, "cursor")

		m := newTestModel(t)
		m.width = width
		m.config.DualPaneMinWidth = 200
		tasks := make([]task.Task, n)
		now := time.Now()
		for i := range tasks {
			words := rapid.IntRange(1, 50).Draw(rt, fmt.Sprintf("w-%d", i))
			tasks[i] = task.Task{
				ID:        id.New(),
				Title:     longTitle(words),
				Status:    task.StatusOpen,
				CreatedAt: now,
				UpdatedAt: now,
			}
		}
		m.tasks = tasks
		m.activeList = listInbox
		m.cursor = cursorIdx

		cursorMarker := m.theme.Selected.Render("> ")
		out := m.viewList()
		require.Equal(rt, 1, strings.Count(out, cursorMarker),
			"cursor marker must appear exactly once across all lines")
	})
}

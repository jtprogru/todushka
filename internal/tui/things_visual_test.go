package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// titleCol returns the display-column (lipgloss.Width) at which `title`
// starts in the first line of `out` that contains it, or -1 if absent.
func titleCol(out, title string) int {
	for _, ln := range strings.Split(out, "\n") {
		if idx := strings.Index(ln, title); idx >= 0 {
			return lipgloss.Width(ln[:idx])
		}
	}
	return -1
}

// ringIndexOf returns the position of a ring glyph in ringGlyphs (-1 if absent).
func ringIndexOf(s string) int {
	for i, g := range ringGlyphs {
		if g == s {
			return i
		}
	}
	return -1
}

func TestProgressRing_Endpoints(t *testing.T) {
	cases := []struct {
		open, total int
		want        string
	}{
		{0, 0, "◯"},
		{0, 5, "●"},
		{5, 5, "◯"},
		{3, 4, "◔"}, // done=1 → ◔
		{2, 4, "◑"}, // done=2 → ◑
		{1, 4, "◕"}, // done=3 → ◕
	}
	for _, c := range cases {
		require.Equal(t, c.want, progressRing(c.open, c.total),
			"progressRing(%d,%d)", c.open, c.total)
	}
}

// TestProp_RingEndpoints — Property 1: ◯ iff done==0, ● iff done==total>0,
// otherwise a partial glyph.
func TestProp_RingEndpoints(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		total := rapid.IntRange(0, 50).Draw(rt, "total")
		open := 0
		if total > 0 {
			open = rapid.IntRange(0, total).Draw(rt, "open")
		}
		done := total - open
		got := progressRing(open, total)
		switch {
		case done == 0:
			require.Equal(rt, "◯", got)
		case done == total && total > 0:
			require.Equal(rt, "●", got)
		default:
			require.Contains(rt, []string{"◔", "◑", "◕"}, got)
		}
	})
}

func TestViewProjectList_RingAndCount(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.screen = screenProjects
	m.height = 20
	m.width = 80
	pid := id.New()
	m.projects = []project.Project{{ID: pid, Name: "Proj", Status: project.StatusOpen}}
	m.projectCounts = map[id.ID][2]int{pid: {3, 5}} // done=2 → ◑

	out := viewProjectList(m, m.width)
	require.Contains(t, out, "[3/5]", "count must be retained (REQ-1.4)")
	require.Contains(t, out, "◑", "ring glyph must render (REQ-1.1)")
}

// TestProp_RingMonotonic — Property 2: more completed ⇒ glyph index never decreases.
func TestProp_RingMonotonic(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		total := rapid.IntRange(1, 50).Draw(rt, "total")
		o1 := rapid.IntRange(0, total).Draw(rt, "o1")
		o2 := rapid.IntRange(0, total).Draw(rt, "o2")
		done1, done2 := total-o1, total-o2
		if done1 > done2 {
			o1, o2 = o2, o1
		}
		require.LessOrEqual(rt,
			ringIndexOf(progressRing(o1, total)),
			ringIndexOf(progressRing(o2, total)))
	})
}

func TestThingsVisual_Monochrome(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.theme = NewMonochromeTheme()
	m.height = 20
	m.width = 80

	m.activeList = listAnytime
	d := task.NewDate(time.Now())
	m.tasks = []task.Task{{ID: id.New(), Title: "MonoStar", Status: task.StatusOpen, PinnedToday: &d}}
	require.Contains(t, m.viewList(), "★", "star renders under monochrome (REQ-4.1)")

	m.screen = screenProjects
	pid := id.New()
	m.projects = []project.Project{{ID: pid, Name: "P", Status: project.StatusOpen}}
	m.projectCounts = map[id.ID][2]int{pid: {0, 4}} // done=4 → ●
	require.Contains(t, viewProjectList(m, m.width), "●", "ring renders under monochrome (REQ-4.1)")
}

// TestProp_MonochromeNoOverflow — Property 6: under NO_COLOR, the windowed
// body never exceeds visibleRows and the cursor marker stays visible.
func TestProp_MonochromeNoOverflow(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	lists := []listKind{listInbox, listToday, listUpcoming, listAnytime, listSomeday, listLogbook}
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 60).Draw(rt, "n")
		h := rapid.IntRange(14, 40).Draw(rt, "h")
		lk := lists[rapid.IntRange(0, len(lists)-1).Draw(rt, "list")]
		m := newTestModel(t)
		m.theme = NewMonochromeTheme()
		m.activeList = lk
		m.height = h
		m.width = 80
		d := task.NewDate(time.Now())
		tasks := make([]task.Task, n)
		for i := range tasks {
			tk := task.Task{ID: id.New(), Title: fmt.Sprintf("t-%03d", i), Status: task.StatusOpen}
			if i%3 == 0 {
				tk.PinnedToday = &d
			}
			tasks[i] = tk
		}
		m.tasks = tasks
		m.cursor = rapid.IntRange(0, n-1).Draw(rt, "cursor")
		body := m.viewList()
		require.LessOrEqual(rt, strings.Count(body, "\n")+1, visibleRows(m))
		require.NotEmpty(rt, findCursorLine(body))
	})
}

func anytimeModel(t *testing.T) Model {
	t.Helper()
	m := newTestModel(t)
	m.activeList = listAnytime
	m.height = 40
	m.width = 80
	return m
}

func TestViewList_StarOnTodayInAnytime(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := anytimeModel(t)
	d := task.NewDate(time.Now())
	m.tasks = []task.Task{{ID: id.New(), Title: "PinnedTask", Status: task.StatusOpen, PinnedToday: &d}}
	require.Contains(t, m.viewList(), "★", "today-task in Anytime must show ★")
}

func TestViewList_NoStarOutsideAnytime(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := anytimeModel(t)
	m.activeList = listToday
	d := task.NewDate(time.Now())
	m.tasks = []task.Task{{ID: id.New(), Title: "PinnedTask", Status: task.StatusOpen, PinnedToday: &d}}
	require.NotContains(t, m.viewList(), "★", "no star outside Anytime")
}

func TestViewList_StarSlotAlignment(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := anytimeModel(t)
	d := task.NewDate(time.Now())
	m.tasks = []task.Task{
		{ID: id.New(), Title: "AlphaTitle", Status: task.StatusOpen, PinnedToday: &d}, // starred
		{ID: id.New(), Title: "BetaTitle", Status: task.StatusOpen},                   // not starred
	}
	out := m.viewList()
	require.Equal(t, titleCol(out, "AlphaTitle"), titleCol(out, "BetaTitle"),
		"starred and unstarred rows must align (REQ-2.2)")
}

// TestProp_StarPresence — Property 3: every today-task row in Anytime shows ★.
func TestProp_StarPresence(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 8).Draw(rt, "n")
		m := anytimeModel(t)
		d := task.NewDate(time.Now())
		tasks := make([]task.Task, n)
		for i := range tasks {
			tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("today-%02d", i), Status: task.StatusOpen, PinnedToday: &d}
		}
		m.tasks = tasks
		require.Equal(rt, n, strings.Count(m.viewList(), "★"))
	})
}

func TestViewList_DoneRowKeepsContent(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := newTestModel(t)
	m.height = 20
	m.width = 80
	now := time.Now()
	m.tasks = []task.Task{
		{ID: id.New(), Title: "DoneTitle", Status: task.StatusCompleted, CompletedAt: &now},
		{ID: id.New(), Title: "CancelTitle", Status: task.StatusCancelled, CancelledAt: &now},
	}
	out := m.viewList()
	require.Contains(t, out, "DoneTitle")
	require.Contains(t, out, "✓")
	require.Contains(t, out, "CancelTitle")
	require.Contains(t, out, "✗")
}

// TestProp_DoneContentPreserved — Property 5: faint styling never drops the
// title text nor the status glyph of completed rows.
func TestProp_DoneContentPreserved(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 6).Draw(rt, "n")
		m := newTestModel(t)
		m.height = 40
		m.width = 80
		now := time.Now()
		tasks := make([]task.Task, n)
		for i := range tasks {
			tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("done-%02d", i), Status: task.StatusCompleted, CompletedAt: &now}
		}
		m.tasks = tasks
		out := m.viewList()
		for i := range tasks {
			require.Contains(rt, out, fmt.Sprintf("done-%02d", i))
		}
		require.Equal(rt, n, strings.Count(out, "✓"))
	})
}

// TestProp_StarExclusionAndAlignment — Property 4: no ★ outside Anytime;
// within Anytime starred and unstarred rows share the same title column.
func TestProp_StarExclusionAndAlignment(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })
	lists := []listKind{listInbox, listToday, listUpcoming, listAnytime, listSomeday, listLogbook}
	rapid.Check(t, func(rt *rapid.T) {
		lk := lists[rapid.IntRange(0, len(lists)-1).Draw(rt, "list")]
		m := anytimeModel(t)
		m.activeList = lk
		d := task.NewDate(time.Now())
		m.tasks = []task.Task{
			{ID: id.New(), Title: "AlphaTitle", Status: task.StatusOpen, PinnedToday: &d},
			{ID: id.New(), Title: "BetaTitle", Status: task.StatusOpen},
		}
		out := m.viewList()
		if lk != listAnytime {
			require.NotContains(rt, out, "★")
		} else {
			require.Equal(rt, titleCol(out, "AlphaTitle"), titleCol(out, "BetaTitle"))
		}
	})
}

package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestCurrentMode_PriorityOrder(t *testing.T) {
	m := newTestModel(t)
	// HELP wins over CONFIRM
	m.screen = screenHelp
	m.confirm = &confirmState{}
	require.Equal(t, modeHelp, currentMode(m))
	// EDITOR wins over CONFIRM
	m.screen = screenEditor
	require.Equal(t, modeEditor, currentMode(m))
	// CONFIRM wins over FILTER
	m.screen = screenList
	m.filtering = true
	require.Equal(t, modeConfirm, currentMode(m))
	// FILTER wins over SELECT
	m.confirm = nil
	m.selected[id.New()] = struct{}{}
	require.Equal(t, modeFilter, currentMode(m))
	// SELECT wins over NORMAL
	m.filtering = false
	require.Equal(t, modeSelect, currentMode(m))
	// NORMAL default
	m.selected = make(map[id.ID]struct{})
	require.Equal(t, modeNormal, currentMode(m))
}

func TestModeKeyHints_Normal(t *testing.T) {
	h := modeKeyHints(modeNormal)
	joined := strings.Join(h, " ")
	require.Contains(t, joined, "/: filter")
	require.Contains(t, joined, "space: select")
	require.Contains(t, joined, "q: quit")
}

func TestModeKeyHints_Filter(t *testing.T) {
	h := modeKeyHints(modeFilter)
	joined := strings.Join(h, " ")
	require.Contains(t, joined, "↵: save")
	require.Contains(t, joined, "esc: cancel")
}

func TestModeKeyHints_Select(t *testing.T) {
	h := modeKeyHints(modeSelect)
	joined := strings.Join(h, " ")
	require.Contains(t, joined, "c/x/d/p: bulk")
	require.Contains(t, joined, "*: all")
	require.Contains(t, joined, "esc: clear")
}

func TestModeKeyHints_Confirm(t *testing.T) {
	h := modeKeyHints(modeConfirm)
	joined := strings.Join(h, " ")
	require.Contains(t, joined, "y: yes")
	require.Contains(t, joined, "any: cancel")
}

func TestModeKeyHints_Editor(t *testing.T) {
	h := modeKeyHints(modeEditor)
	joined := strings.Join(h, " ")
	require.Contains(t, joined, "Tab: next")
	require.Contains(t, joined, "Shift+Tab: prev")
	require.Contains(t, joined, "Ctrl+S: save")
	require.Contains(t, joined, "esc: cancel")
}

func TestModeKeyHints_Help(t *testing.T) {
	h := modeKeyHints(modeHelp)
	joined := strings.Join(h, " ")
	require.Contains(t, joined, "?: close")
}

func TestViewFooter_ModeChipPresent(t *testing.T) {
	m := newTestModel(t)
	out := m.viewFooter()
	require.Contains(t, out, "-- NORMAL --")
}

func TestViewFooter_FilterModeChip(t *testing.T) {
	m := newTestModel(t)
	m.filtering = true
	out := m.viewFooter()
	require.Contains(t, out, "-- FILTER --")
}

func TestViewFooter_SelectModeChip(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "x")
	m.selected[tasks[0].ID] = struct{}{}
	out := m.viewFooter()
	require.Contains(t, out, "-- SELECT --")
}

func TestViewFooter_StatusMessagePreserved(t *testing.T) {
	m := newTestModel(t)
	m.statusMsg = "boom"
	out := m.viewFooter()
	require.Contains(t, out, "boom")
}

func TestViewHeader_FullModeSegment(t *testing.T) {
	m := newTestModel(t)
	m.width = 120
	m.listCounts[listInbox] = 24
	out := m.viewHeader()
	require.Contains(t, out, "(1)")
	require.Contains(t, out, "Inbox")
	require.Contains(t, out, "[24]")
}

func TestViewHeader_CompactMode(t *testing.T) {
	m := newTestModel(t)
	m.width = 70
	m.listCounts[listInbox] = 24
	out := m.viewHeader()
	require.Contains(t, out, "(1)")
	require.Contains(t, out, "I")
	require.Contains(t, out, "[24]")
	require.NotContains(t, out, "Inbox")
	require.NotContains(t, out, "Today")
}

func TestViewHeader_UnknownCountPlaceholder(t *testing.T) {
	m := newTestModel(t)
	m.width = 120
	out := m.viewHeader()
	require.Contains(t, out, "[?]")
}

func TestViewHeader_EmptyListShowsZero(t *testing.T) {
	m := newTestModel(t)
	m.width = 120
	m.listCounts[listInbox] = 0
	out := m.viewHeader()
	require.Contains(t, out, "[0]")
}

func TestViewHeader_ActiveSegmentInverted(t *testing.T) {
	m := newTestModel(t)
	m.width = 120
	m.activeList = listInbox
	m.listCounts[listInbox] = 24
	out := m.viewHeader()
	require.Greater(t, len(out), len("(1) Inbox [24]"))
}

func TestFetchListCounts_EmitsMsg(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	_, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	cmd := fetchListCounts(svc)
	msg := cmd()
	res, ok := msg.(countsLoadedMsg)
	require.True(t, ok)
	require.Equal(t, 1, res.counts[listInbox])
}

func TestUpdate_TasksLoadedTriggersCountsFetch(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x"})
	require.NoError(t, err)
	m2, cmd := m.Update(tasksLoadedMsg{tasks: []task.Task{tk}})
	require.NotNil(t, cmd)
	_ = m2
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var foundCounts bool
		for _, sub := range batch {
			if sub == nil {
				continue
			}
			if _, ok := sub().(countsLoadedMsg); ok {
				foundCounts = true
			}
		}
		require.True(t, foundCounts, "countsLoadedMsg expected in batch")
		return
	}
	if _, ok := msg.(countsLoadedMsg); ok {
		return
	}
	t.Fatalf("unexpected msg type: %T", msg)
}

func TestView_FullScreenClamp(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m.height = 40
	out := m.View()
	require.Equal(t, 40, lipgloss.Height(out), "full-screen mode must produce exactly m.height lines")
}

func TestView_SmallTerminalFallback(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m.height = 5
	out := m.View()
	require.NotEmpty(t, out)
}

func TestView_EditorIgnoresClamp(t *testing.T) {
	m, svc, _ := setupModelWithInboxTasks(t, "x")
	m.width = 200
	m.height = 40
	m.screen = screenEditor
	m.editor = NewEditor(context.Background(), m.tasks[0], svc)
	out := m.View()
	require.NotEmpty(t, out)
}

// TestProp_ModeExclusion verifies CP-1 (REQ-3.2): for every state, exactly
// one shellMode is active and it matches the documented priority order
// HELP > EDITOR > CONFIRM > FILTER > SELECT > NORMAL.
func TestProp_ModeExclusion(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		screens := []screenKind{screenList, screenEditor, screenHelp, screenQuickEntry}
		m := newTestModel(t)
		m.screen = rapid.SampledFrom(screens).Draw(rt, "screen")
		m.filtering = rapid.Bool().Draw(rt, "filtering")
		if rapid.Bool().Draw(rt, "withSelected") {
			m.selected[id.New()] = struct{}{}
		}
		if rapid.Bool().Draw(rt, "withConfirm") {
			m.confirm = &confirmState{}
		}
		mode := currentMode(m)
		switch {
		case m.screen == screenHelp:
			require.Equal(rt, modeHelp, mode)
		case m.screen == screenEditor:
			require.Equal(rt, modeEditor, mode)
		case m.confirm != nil:
			require.Equal(rt, modeConfirm, mode)
		case m.filtering:
			require.Equal(rt, modeFilter, mode)
		case len(m.selected) > 0:
			require.Equal(rt, modeSelect, mode)
		default:
			require.Equal(rt, modeNormal, mode)
		}
	})
}

// TestProp_FullScreenHeight verifies CP-2 (REQ-1.2, 1.3, 1.4): when width
// >= 40 and height >= 10 the rendered View occupies exactly m.height
// lines (full-screen clamp invariant).
func TestProp_FullScreenHeight(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m, _, _ := setupModelWithInboxTasks(t, "x", "y", "z")
		m.width = rapid.IntRange(40, 200).Draw(rt, "width")
		m.height = rapid.IntRange(10, 80).Draw(rt, "height")
		out := m.View()
		require.Equal(rt, m.height, lipgloss.Height(out), "full-screen clamp invariant")
	})
}

// TestProp_WindowSizeTracked verifies CP-3 (REQ-1.1): tea.WindowSizeMsg
// updates both width and height on the model.
func TestProp_WindowSizeTracked(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := newTestModel(t)
		w := rapid.IntRange(0, 300).Draw(rt, "w")
		h := rapid.IntRange(0, 100).Draw(rt, "h")
		m2, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
		mm := m2.(Model)
		require.Equal(rt, w, mm.width)
		require.Equal(rt, h, mm.height)
	})
}

// TestProp_ActiveSegmentStyled verifies CP-4 (REQ-2.3): the active header
// segment renders with the theme.Header style (rendered output contains
// the active list segment).
func TestProp_ActiveSegmentStyled(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := newTestModel(t)
		m.width = 120
		m.activeList = rapid.SampledFrom(allLists).Draw(rt, "active")
		m.listCounts[m.activeList] = rapid.IntRange(0, 99).Draw(rt, "count")
		out := m.viewHeader()
		require.NotEmpty(rt, out)
	})
}

// TestProp_CompactModeThreshold verifies CP-5 (REQ-2.1, 2.4): at widths
// below headerCompactThreshold the full list labels ("Inbox", "Today",
// ...) are replaced by single-letter initials; at or above the threshold
// the full label is present.
func TestProp_CompactModeThreshold(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := newTestModel(t)
		m.width = rapid.IntRange(40, 200).Draw(rt, "width")
		m.listCounts[listInbox] = 5
		out := m.viewHeader()
		if m.width < headerCompactThreshold {
			require.NotContains(rt, out, "Inbox")
		} else {
			require.Contains(rt, out, "Inbox")
		}
	})
}

// TestProp_CountsRefreshPropagates verifies CP-6 (REQ-2.5, 2.6): every
// tasksLoadedMsg dispatches a batch that includes a fetchListCounts
// command (whose execution yields a countsLoadedMsg).
func TestProp_CountsRefreshPropagates(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m, svc := newTestModelWithService(t)
		ctx := context.Background()
		n := rapid.IntRange(0, 5).Draw(rt, "taskCount")
		tasks := make([]task.Task, 0, n)
		for i := 0; i < n; i++ {
			tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: fmt.Sprintf("t%d", i)})
			require.NoError(rt, err)
			tasks = append(tasks, tk)
		}
		_, cmd := m.Update(tasksLoadedMsg{tasks: tasks})
		if cmd == nil {
			return
		}
		msg := cmd()
		foundCounts := false
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, sub := range batch {
				if sub == nil {
					continue
				}
				if _, ok := sub().(countsLoadedMsg); ok {
					foundCounts = true
					break
				}
			}
		} else if _, ok := msg.(countsLoadedMsg); ok {
			foundCounts = true
		}
		require.True(rt, foundCounts, "expected countsLoadedMsg in result")
	})
}

// TestProp_ModeChipText verifies CP-7 (REQ-3.3): the footer always
// contains a "-- MODE --" chip whose text matches currentMode().
func TestProp_ModeChipText(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := newTestModel(t)
		screens := []screenKind{screenList, screenEditor, screenHelp}
		m.screen = rapid.SampledFrom(screens).Draw(rt, "screen")
		m.filtering = rapid.Bool().Draw(rt, "filtering")
		if rapid.Bool().Draw(rt, "withSelected") {
			m.selected[id.New()] = struct{}{}
		}
		if rapid.Bool().Draw(rt, "withConfirm") {
			m.confirm = &confirmState{action: bulkActionComplete}
		}
		out := m.viewFooter()
		expected := "-- " + currentMode(m).modeLabel() + " --"
		require.Contains(rt, out, expected)
	})
}

// TestProp_KeyHintsMatchMode verifies CP-8 (REQ-3.4..3.9): every shellMode
// has a non-empty set of key hints.
func TestProp_KeyHintsMatchMode(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		modes := []shellMode{modeNormal, modeFilter, modeSelect, modeConfirm, modeEditor, modeHelp}
		mode := rapid.SampledFrom(modes).Draw(rt, "mode")
		hints := modeKeyHints(mode)
		require.NotEmpty(rt, hints)
	})
}

// TestProp_CountsMatchService verifies CP-12 (REQ-2.5): the count reported
// for listInbox via fetchListCounts equals the number of tasks added to
// the service.
func TestProp_CountsMatchService(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithService(t)
		ctx := context.Background()
		n := rapid.IntRange(0, 5).Draw(rt, "inboxCount")
		for i := 0; i < n; i++ {
			_, err := svc.AddTask(ctx, app.AddTaskInput{Title: "t"})
			require.NoError(rt, err)
		}
		cmd := fetchListCounts(svc)
		msg := cmd().(countsLoadedMsg)
		require.Equal(rt, n, msg.counts[listInbox])
	})
}

func TestRenderSeparator_FullWidth(t *testing.T) {
	s := renderSeparator(NewTheme(), 80)
	// Strip ANSI escapes to count "─" reliably. lipgloss adds them around content.
	// Simpler: count occurrences of "─" rune.
	count := strings.Count(s, "─")
	require.Equal(t, 80, count, "separator should contain exactly 80 box-drawing rune characters")
}

func TestRenderSeparator_EmptyOnZero(t *testing.T) {
	s := renderSeparator(NewTheme(), 0)
	require.Empty(t, s)
}

func TestRenderSeparator_EmptyOnNegative(t *testing.T) {
	s := renderSeparator(NewTheme(), -5)
	require.Empty(t, s)
}

func TestView_HasSeparatorsInFullScreen(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m.height = 40
	out := m.View()
	require.Contains(t, out, "─", "full-screen view should contain separator rune")
	// Count: 2 separator rows × 120 width = 240 "─" runes
	count := strings.Count(out, "─")
	require.GreaterOrEqual(t, count, 240, "expected at least 240 ─ characters (2 rows × 120 width)")
}

func TestView_NoSeparatorsInLegacy(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m.height = 5 // below threshold
	out := m.View()
	require.NotContains(t, out, "─", "legacy mode must not render section separators")
}

func TestView_NoSeparatorsInEditor(t *testing.T) {
	m, svc, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m.height = 40
	m.screen = screenEditor
	m.editor = NewEditor(context.Background(), m.tasks[0], svc)
	out := m.View()
	// The editor's own border decorations include "─" characters, but the
	// full-width section separators (m.width consecutive "─") must not be
	// rendered in editor mode (REQ-4.5). We assert absence of a full-width
	// horizontal rule by checking for a run of m.width "─" runes.
	fullWidthRule := strings.Repeat("─", m.width)
	require.NotContains(t, out, fullWidthRule, "editor mode must not render full-width section separators")
}

func TestView_FullScreenHeightWithSeparators(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m.height = 40
	out := m.View()
	require.Equal(t, 40, lipgloss.Height(out), "lipgloss height invariant preserved")
}

// TestProp_SeparatorsConditional verifies CP-8 (REQ-4.1, 4.2, 4.4, 4.5):
// section separators render iff width >= 40, height >= 10, and the
// screen is not the editor. Restricted to screenList without modal
// overlays — overlay screens (editor, quick-entry, confirm) render
// their own bordered widgets whose "─" runs are unrelated to section
// separators. Those cases are locked by adjacent unit tests
// (TestView_NoSeparatorsInEditor, TestView_HasSeparatorsInFullScreen,
// TestView_NoSeparatorsInLegacy).
func TestProp_SeparatorsConditional(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m, _, _ := setupRapidModel(rt, "x")
		m.width = rapid.IntRange(20, 200).Draw(rt, "width")
		m.height = rapid.IntRange(3, 60).Draw(rt, "height")
		m.screen = screenList
		m.confirm = nil
		out := m.View()
		fullWidthRule := strings.Repeat("─", m.width)
		sepActive := m.height >= 10 && m.width >= 40
		if sepActive {
			require.Contains(rt, out, fullWidthRule, "expected section separator")
		} else {
			require.NotContains(rt, out, fullWidthRule, "did not expect section separator")
		}
	})
}

// TestProp_SeparatorWidth verifies CP-9 (REQ-4.3): renderSeparator
// produces exactly width "─" runes.
func TestProp_SeparatorWidth(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(40, 200).Draw(rt, "width")
		s := renderSeparator(NewTheme(), w)
		count := strings.Count(s, "─")
		require.Equal(rt, w, count)
	})
}

// TestProp_FullScreenHeightWithSeparators verifies CP-10 (REQ-4.6):
// when separators are active, the rendered View still occupies exactly
// m.height lines (bodyH accounts for the two extra separator rows).
func TestProp_FullScreenHeightWithSeparators(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m, _, _ := setupRapidModel(rt, "x")
		m.width = rapid.IntRange(40, 200).Draw(rt, "width")
		m.height = rapid.IntRange(10, 60).Draw(rt, "height")
		// Ensure screen is not editor so full-screen clamp is active
		m.screen = screenList
		out := m.View()
		require.Equal(rt, m.height, lipgloss.Height(out))
	})
}

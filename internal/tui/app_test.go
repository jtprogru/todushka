package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func newTestModel(t *testing.T) Model {
	t.Helper()
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	return NewModel(svc, NewTheme(), config.Defaults())
}

func newTestModelWithService(t *testing.T) (Model, *app.Service) {
	t.Helper()
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	return NewModel(svc, NewTheme(), config.Defaults()), svc
}

func TestTUI_QuitOnQ(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	require.NotNil(t, cmd)
	// Verify the command is tea.Quit by checking it produces tea.QuitMsg
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok, "expected tea.QuitMsg, got %T", msg)
}

func TestTUI_QuitOnCtrlC(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok)
}

func TestTUI_SwitchListByNumber(t *testing.T) {
	m := newTestModel(t)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	mm := m2.(Model)
	require.Equal(t, listInbox, mm.activeList)

	m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6'}})
	mm = m3.(Model)
	require.Equal(t, listLogbook, mm.activeList)
}

func TestTUI_CursorBoundary(t *testing.T) {
	m := newTestModel(t)
	m.tasks = []task.Task{{Title: "a"}, {Title: "b"}}
	m.cursor = 0

	// Up at top stays at 0
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	require.Equal(t, 0, m2.(Model).cursor)

	// Down moves
	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	require.Equal(t, 1, m3.(Model).cursor)

	// Down at bottom stays at 1
	m4 := m3.(Model)
	m5, _ := m4.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	require.Equal(t, 1, m5.(Model).cursor)
}

func TestTUI_HelpToggle(t *testing.T) {
	m := newTestModel(t)
	require.Equal(t, screenList, m.screen)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	require.Equal(t, screenHelp, m2.(Model).screen)
	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	require.Equal(t, screenList, m3.(Model).screen)
}

func TestTUI_QuickEntryHotkeyOpensModal(t *testing.T) {
	m := newTestModel(t)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	require.Equal(t, screenQuickEntry, m2.(Model).screen)
}

func TestTUI_QuickEntryEscapeCloses(t *testing.T) {
	m := newTestModel(t)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mm := m2.(Model)
	require.Equal(t, screenQuickEntry, mm.screen)

	m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Equal(t, screenList, m3.(Model).screen)
}

func TestTUI_ErrorMsgUpdatesStatusBar(t *testing.T) {
	m := newTestModel(t)
	m2, _ := m.Update(errorMsg{err: errors.New("boom")})
	mm := m2.(Model)
	require.Equal(t, "boom", mm.statusMsg)
	require.True(t, mm.statusUntil.After(time.Now()))
}

func TestTUI_ClearStatusFadesText(t *testing.T) {
	m := newTestModel(t)
	m.statusMsg = "x"
	m.statusUntil = time.Now().Add(-time.Second) // already expired
	m2, _ := m.Update(clearStatusMsg{})
	require.Empty(t, m2.(Model).statusMsg)
}

func TestTUI_TasksLoadedPopulatesModel(t *testing.T) {
	m := newTestModel(t)
	tasks := []task.Task{{Title: "a"}, {Title: "b"}}
	m2, _ := m.Update(tasksLoadedMsg{tasks: tasks})
	require.Equal(t, tasks, m2.(Model).tasks)
}

func TestTUI_ViewContainsListLabels(t *testing.T) {
	m := newTestModel(t)
	out := m.View()
	for _, label := range []string{"Inbox", "Today", "Upcoming", "Anytime", "Someday", "Logbook"} {
		require.Contains(t, out, label)
	}
}

func TestTUI_TabCyclesViews(t *testing.T) {
	m := newTestModel(t)
	require.Equal(t, listToday, m.activeList)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, listUpcoming, m2.(Model).activeList)

	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, listAnytime, m3.(Model).activeList)
}

func TestTUI_ShiftTabCyclesBack(t *testing.T) {
	m := newTestModel(t)
	require.Equal(t, listToday, m.activeList)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	require.Equal(t, listInbox, m2.(Model).activeList)

	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	require.Equal(t, listLogbook, m3.(Model).activeList)
}

func TestTUI_EnterOpensEditor(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.tasks = []task.Task{{Title: "edit me"}}
	m.cursor = 0

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := m2.(Model)
	require.Equal(t, screenEditor, mm.screen)
	require.Equal(t, "edit me", mm.editor.title.Value())
}

func TestTUI_EnterOnEmptyListDoesNotOpenEditor(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.tasks = nil
	m.cursor = 0

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, screenList, m2.(Model).screen)
}

func TestTUI_EditorEscClosesWithoutSave(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.tasks = []task.Task{{Title: "x"}}
	m.cursor = 0
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, screenEditor, m2.(Model).screen)

	m3, _ := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Equal(t, screenList, m3.(Model).screen)
}

func TestTUI_EditorTabCyclesFields(t *testing.T) {
	m, _ := newTestModelWithService(t)
	m.tasks = []task.Task{{Title: "x"}}
	m.cursor = 0
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := m2.(Model)
	require.Equal(t, fieldTitle, mm.editor.focus)

	m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeyTab})
	mm = m3.(Model)
	require.Equal(t, fieldNotes, mm.editor.focus)

	m4, _ := mm.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	mm = m4.(Model)
	require.Equal(t, fieldTitle, mm.editor.focus)
}

func setupModelWithInboxTasks(t *testing.T, titles ...string) (Model, *app.Service, []task.Task) {
	t.Helper()
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	tasks := make([]task.Task, 0, len(titles))
	for _, title := range titles {
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: title})
		require.NoError(t, err)
		tasks = append(tasks, tk)
	}
	m.activeList = listInbox
	m.tasks = tasks
	m.cursor = 0
	return m, svc, tasks
}

func TestTUI_CompleteCursorTaskWhenNoSelection(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "task one", "task two")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.NotNil(t, cmd)
	_ = cmd()

	got, err := svc.Repo().TaskGet(context.Background(), tasks[0].ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, got.Status)
	require.NotNil(t, got.CompletedAt)

	other, err := svc.Repo().TaskGet(context.Background(), tasks[1].ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusOpen, other.Status, "other tasks must remain open")
}

func TestTUI_CancelCursorTaskWhenNoSelection(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "task one", "task two")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	require.NotNil(t, cmd)
	_ = cmd()

	got, err := svc.Repo().TaskGet(context.Background(), tasks[0].ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusCancelled, got.Status)
	require.NotNil(t, got.CancelledAt)

	other, err := svc.Repo().TaskGet(context.Background(), tasks[1].ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusOpen, other.Status)
}

func TestTUI_DeleteCursorTaskWhenNoSelection(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "task one", "task two")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	require.NotNil(t, cmd)
	_ = cmd()

	inbox, err := svc.ListInbox(context.Background())
	require.NoError(t, err)
	require.Len(t, inbox, 1, "deleted task must be filtered out of inbox")
	require.Equal(t, tasks[1].ID, inbox[0].ID, "remaining task is the one we did not delete")
}

func TestTUI_PinCursorTaskWhenNoSelection(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "task one", "task two")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	require.NotNil(t, cmd)
	_ = cmd()

	got, err := svc.Repo().TaskGet(context.Background(), tasks[0].ID)
	require.NoError(t, err)
	require.NotNil(t, got.PinnedToday, "expected PinnedToday to be set")

	other, err := svc.Repo().TaskGet(context.Background(), tasks[1].ID)
	require.NoError(t, err)
	require.Nil(t, other.PinnedToday)
}

func TestSelection_SpaceToggle(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "a", "b")

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	mm := m2.(Model)
	require.Equal(t, 1, len(mm.selected))

	m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeySpace})
	require.Equal(t, 0, len(m3.(Model).selected))
}

func TestSelection_StarSelectsAllVisible(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c")

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}})
	mm := m2.(Model)
	require.Equal(t, 3, len(mm.selected))
	for _, tk := range tasks {
		_, ok := mm.selected[tk.ID]
		require.True(t, ok, "task ID %v expected in selected", tk.ID)
	}
}

func TestSelection_EscClearsSelection(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b")
	m.selected[tasks[0].ID] = struct{}{}
	require.Equal(t, 1, len(m.selected))

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.Equal(t, 0, len(m2.(Model).selected))
}

func TestSelection_PrefixHiddenWhenEmpty(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "task")
	require.Equal(t, 0, len(m.selected))

	out := m.viewList()
	require.NotContains(t, out, "[x]")
	require.NotContains(t, out, "[ ]")
}

func TestSelection_PrefixVisibleWhenNonEmpty(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "task one", "task two")
	m.selected[tasks[0].ID] = struct{}{}

	out := m.viewList()
	require.Contains(t, out, "[x] ")
	require.Contains(t, out, "[ ] ")
}

func TestSelection_StatusBarCounter(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}

	out := m.viewFooter()
	require.Contains(t, out, "Selected: 3")
}

func TestSwitchList_ClearsFilterAndSelection(t *testing.T) {
	// Number key (1) → Inbox
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b")
	m.filterQuery = "foo"
	m.filtering = false
	m.selected[tasks[0].ID] = struct{}{}
	require.NotEmpty(t, m.filterQuery)
	require.Equal(t, 1, len(m.selected))

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	mm := m2.(Model)
	require.Equal(t, "", mm.filterQuery)
	require.False(t, mm.filtering)
	require.Equal(t, 0, len(mm.selected))

	// Tab
	m, _, tasks = setupModelWithInboxTasks(t, "a", "b")
	m.filterQuery = "foo"
	m.filtering = false
	m.selected[tasks[0].ID] = struct{}{}
	require.NotEmpty(t, m.filterQuery)
	require.Equal(t, 1, len(m.selected))

	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	mm = m3.(Model)
	require.Equal(t, "", mm.filterQuery)
	require.False(t, mm.filtering)
	require.Equal(t, 0, len(mm.selected))

	// Shift+Tab
	m, _, tasks = setupModelWithInboxTasks(t, "a", "b")
	m.filterQuery = "foo"
	m.filtering = false
	m.selected[tasks[0].ID] = struct{}{}
	require.NotEmpty(t, m.filterQuery)
	require.Equal(t, 1, len(m.selected))

	m4, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	mm = m4.(Model)
	require.Equal(t, "", mm.filterQuery)
	require.False(t, mm.filtering)
	require.Equal(t, 0, len(mm.selected))
}

func TestHelp_IncludesNewBindings(t *testing.T) {
	m := newTestModel(t)
	out := m.viewHelp()
	require.Contains(t, out, "/", "help must list filter binding")
	require.Contains(t, out, "*", "help must list select-all binding")
	require.Contains(t, out, "space", "help must list toggle-select binding")
}

func TestFooter_IncludesNewHints(t *testing.T) {
	m := newTestModel(t)
	out := m.viewFooter()
	require.Contains(t, out, "/", "footer must mention filter")
	require.Contains(t, out, "space", "footer must mention select")
}

func TestSelectTheme_DefaultDark(t *testing.T) {
	tm := SelectTheme(func(string) string { return "" })
	require.Equal(t, "catppuccin-macchiato", tm.Name)
}

func TestSelectTheme_LightOnRequest(t *testing.T) {
	env := func(k string) string {
		if k == "TODUSHKA_THEME" {
			return "light"
		}
		return ""
	}
	tm := SelectTheme(env)
	require.Equal(t, "catppuccin-latte", tm.Name)
}

func TestSelectTheme_NoColorOverrides(t *testing.T) {
	env := func(k string) string {
		if k == "NO_COLOR" {
			return "1"
		}
		return ""
	}
	tm := SelectTheme(env)
	require.Equal(t, "monochrome", tm.Name)
}

// switchKeyMsgs lists every key that should trigger switchList semantics
// (Tab/Shift+Tab plus the digit keys 1..6).
func switchKeyMsgs() []tea.KeyMsg {
	return []tea.KeyMsg{
		{Type: tea.KeyTab},
		{Type: tea.KeyShiftTab},
		{Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{'2'}},
		{Type: tea.KeyRunes, Runes: []rune{'3'}},
		{Type: tea.KeyRunes, Runes: []rune{'4'}},
		{Type: tea.KeyRunes, Runes: []rune{'5'}},
		{Type: tea.KeyRunes, Runes: []rune{'6'}},
	}
}

// TestProp_SwitchListResetsState verifies CP-4 (REQ-1.6, 2.7):
// every switch key clears the filter query, leaves filtering=false,
// and empties the selection.
func TestProp_SwitchListResetsState(t *testing.T) {
	keys := switchKeyMsgs()
	for i, msg := range keys {
		msg := msg
		t.Run(fmt.Sprintf("key_%d", i), func(t *testing.T) {
			rapid.Check(t, func(rt *rapid.T) {
				query := rapid.StringMatching(`[a-z]{1,8}`).Draw(rt, "query")
				n := rapid.IntRange(1, 5).Draw(rt, "selN")

				m, _, tasks := setupRapidModel(rt, "a", "b", "c", "d", "e")
				m.filterQuery = query
				m.filtering = false
				for j := 0; j < n; j++ {
					m.selected[tasks[j].ID] = struct{}{}
				}
				require.NotEmpty(rt, m.filterQuery)
				require.Equal(rt, n, len(m.selected))

				m2, _ := m.Update(msg)
				mm := m2.(Model)
				require.Equal(rt, "", mm.filterQuery, "switchList must clear filterQuery")
				require.False(rt, mm.filtering, "switchList must leave filtering=false")
				require.Equal(rt, 0, len(mm.selected), "switchList must empty the selection")
			})
		})
	}
}

// pbtModelHelper adapts setupModelWithInboxTasks for rapid.Check
// callbacks. setupModelWithInboxTasks takes a *testing.T (used for
// t.Helper), so we wrap rapid's testing.TB equivalent.
func setupRapidModel(rt *rapid.T, titles ...string) (Model, *app.Service, []task.Task) {
	rt.Helper()
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	m := NewModel(svc, NewTheme(), config.Defaults())
	ctx := context.Background()
	tasks := make([]task.Task, 0, len(titles))
	for _, title := range titles {
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: title})
		require.NoError(rt, err)
		tasks = append(tasks, tk)
	}
	m.activeList = listInbox
	m.tasks = tasks
	m.cursor = 0
	return m, svc, tasks
}

// TestProp_SpaceIsInvolution verifies CP-5 (REQ-2.1): pressing Space
// twice on the same cursor task leaves the selection size unchanged.
func TestProp_SpaceIsInvolution(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 5).Draw(t, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = fmt.Sprintf("task-%d", i)
		}
		m, _, _ := setupRapidModel(t, titles...)
		cursor := rapid.IntRange(0, n-1).Draw(t, "cursor")
		m.cursor = cursor
		initialN := len(m.selected)

		// Press space twice; cursor must not move between presses.
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
		mm := m2.(Model)
		require.Equal(t, cursor, mm.cursor, "Space must not move cursor")
		m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeySpace})
		mm = m3.(Model)
		require.Equal(t, initialN, len(mm.selected), "two Space presses on same cursor must be identity on |selected|")
	})
}

// TestProp_PrefixIffNonEmpty verifies CP-6 (REQ-2.2, 2.3): the
// "[x] / [ ]" prefix renders if and only if |selected| > 0.
func TestProp_PrefixIffNonEmpty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 5).Draw(t, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = fmt.Sprintf("task-%d", i)
		}
		m, _, tasks := setupRapidModel(t, titles...)
		// Choose a subset of indices to mark as selected.
		mask := rapid.SliceOfN(rapid.Bool(), n, n).Draw(t, "mask")
		selN := 0
		for i, b := range mask {
			if b {
				m.selected[tasks[i].ID] = struct{}{}
				selN++
			}
		}
		out := m.viewList()
		if selN > 0 {
			containsAny := strings.Contains(out, "[x]") || strings.Contains(out, "[ ]")
			require.True(t, containsAny, "with non-empty selection the prefix must render")
		} else {
			require.False(t, strings.Contains(out, "[x]"), "with empty selection the [x] prefix must NOT render")
			require.False(t, strings.Contains(out, "[ ]"), "with empty selection the [ ] prefix must NOT render")
		}
	})
}

// TestProp_StarSelectsAllVisible verifies CP-7 (REQ-2.4): '*' adds
// every visible task ID to m.selected.
func TestProp_StarSelectsAllVisible(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Titles drawn from a small alphabet so the query reliably matches
		// at least one (or zero, which is also a valid spec point).
		n := rapid.IntRange(1, 6).Draw(t, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = rapid.StringMatching(`[a-c]{1,5}`).Draw(t, fmt.Sprintf("title-%d", i))
		}
		m, _, _ := setupRapidModel(t, titles...)
		// Optionally apply a filter that may narrow the visible set.
		useFilter := rapid.Bool().Draw(t, "useFilter")
		if useFilter {
			m.filterQuery = rapid.StringMatching(`[a-c]{0,3}`).Draw(t, "query")
		}
		visible := displayedTasks(m)

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}})
		mm := m2.(Model)
		for _, vt := range visible {
			_, ok := mm.selected[vt.ID]
			require.True(t, ok, "task %v expected in selection after '*'", vt.ID)
		}
	})
}

// TestProp_EscClearsSelection verifies CP-8 (REQ-2.5): in list mode,
// Esc empties the selection set.
func TestProp_EscClearsSelection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 5).Draw(t, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = fmt.Sprintf("t-%d", i)
		}
		m, _, tasks := setupRapidModel(t, titles...)
		// Populate selection with a non-empty random subset.
		selCount := rapid.IntRange(1, n).Draw(t, "selCount")
		for i := 0; i < selCount; i++ {
			m.selected[tasks[i].ID] = struct{}{}
		}
		require.Equal(t, selCount, len(m.selected))
		require.False(t, m.filtering)
		require.Nil(t, m.confirm)

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		mm := m2.(Model)
		require.Equal(t, 0, len(mm.selected), "Esc must clear the selection in list mode")
	})
}

// TestProp_SelectionSubsetOfVisible verifies CP-9 (REQ-2.6): for any
// sequence of rune mutations to the filter query, every ID in
// m.selected corresponds to a task in displayedTasks(m).
func TestProp_SelectionSubsetOfVisible(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		titles := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
		m, _, tasks := setupRapidModel(t, titles...)
		// Seed the selection with every task (the most stringent case).
		for _, tk := range tasks {
			m.selected[tk.ID] = struct{}{}
		}
		// Enter filter mode.
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		mm := m2.(Model)
		require.True(t, mm.filtering)

		body := rapid.SliceOfN(rapid.StringMatching(`[a-z]`), 0, 8).Draw(t, "body")
		for _, s := range body {
			r := []rune(s)[0]
			mNext, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			mm = mNext.(Model)
			visible := make(map[id.ID]struct{}, len(mm.tasks))
			for _, vt := range displayedTasks(mm) {
				visible[vt.ID] = struct{}{}
			}
			for sid := range mm.selected {
				_, ok := visible[sid]
				require.True(t, ok, "selected ID %v must be visible after rune %q (query=%q)", sid, r, mm.filterQuery)
			}
		}
	})
}

// bareTestModel constructs a Model without any service dependency
// (sufficient for view-only properties that don't load tasks).
func bareTestModel() Model {
	return NewModel(nil, NewTheme(), config.Defaults())
}

// TestProp_StatusBarShowsCount verifies CP-10 (REQ-2.8): when N tasks
// are selected, viewFooter contains "Selected: N".
func TestProp_StatusBarShowsCount(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 20).Draw(rt, "n")
		m := bareTestModel()
		for i := 0; i < n; i++ {
			m.selected[id.New()] = struct{}{}
		}
		out := m.viewFooter()
		require.Contains(rt, out, fmt.Sprintf("Selected: %d", n))
	})
}

// TestProp_HelpContainsNewKeys verifies CP-17 (REQ-4.1): the help
// view always advertises '/', '*', and 'space', regardless of
// selection size or filter state.
func TestProp_HelpContainsNewKeys(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		selN := rapid.IntRange(0, 10).Draw(rt, "selN")
		filtering := rapid.Bool().Draw(rt, "filtering")
		m := bareTestModel()
		for i := 0; i < selN; i++ {
			m.selected[id.New()] = struct{}{}
		}
		m.filtering = filtering
		out := m.viewHelp()
		require.Contains(rt, out, "/", "help must advertise the filter binding")
		require.Contains(rt, out, "*", "help must advertise the select-all binding")
		require.Contains(rt, out, "space", "help must advertise the toggle-select binding")
	})
}

// TestProp_FooterContainsNewKeys verifies CP-18 (REQ-4.2): the
// list-mode footer surfaces mode-appropriate key hints. In NORMAL mode
// (no selection, not filtering) it advertises '/' and 'space'; in SELECT
// mode (selection present) it advertises the bulk actions and clear key.
func TestProp_FooterContainsNewKeys(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		selN := rapid.IntRange(0, 10).Draw(rt, "selN")
		m := bareTestModel()
		for i := 0; i < selN; i++ {
			m.selected[id.New()] = struct{}{}
		}
		m.filtering = false
		m.confirm = nil
		out := m.viewFooter()
		if selN == 0 {
			require.Contains(rt, out, "-- NORMAL --", "NORMAL mode chip expected")
			require.Contains(rt, out, "/: filter", "footer must mention the filter key")
			require.Contains(rt, out, "space: select", "footer must mention the select key")
		} else {
			require.Contains(rt, out, "-- SELECT --", "SELECT mode chip expected")
			require.Contains(rt, out, "c/x/d/p: bulk", "footer must mention bulk actions")
			require.Contains(rt, out, "esc: clear", "footer must mention clear key")
		}
	})
}

// Preservation tests for single-pane behavior (T-1 for dual-pane-layout).
// These lock the current behavior at width=0 (test fixtures, initial state)
// and width<100 (narrow terminals): no double-line border, full-width list.

func TestTUI_ZeroWidthRendersSinglePane(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "task one")
	require.Equal(t, 0, m.width, "fixture defaults width to 0")
	out := m.View()
	require.NotContains(t, out, "║", "width=0 must render single-pane (no double-line border)")
}

func TestTUI_NarrowWidthRendersSinglePane(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "task one")
	m.width = 80
	out := m.View()
	require.NotContains(t, out, "║", "width<100 must render single-pane")
}

func TestTUI_NarrowWidthShowsListInBody(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "alpha", "beta")
	m.width = 80
	m.cursor = 0
	out := m.View()
	require.Contains(t, out, "alpha", "single-pane must show first task")
	require.Contains(t, out, "beta", "single-pane must show second task")
}

func TestNameCache_LoadedMsgPopulatesModel(t *testing.T) {
	m := newTestModel(t)
	tid := id.New()
	aid := id.New()
	pid := id.New()
	msg := nameCacheLoadedMsg{
		tags:     map[id.ID]string{tid: "work"},
		areas:    map[id.ID]string{aid: "home"},
		projects: map[id.ID]string{pid: "todushka"},
		headings: map[id.ID]string{},
	}
	m2, _ := m.Update(msg)
	mm := m2.(Model)
	require.Equal(t, "work", mm.tagNamesByID[tid])
	require.Equal(t, "home", mm.areaNamesByID[aid])
	require.Equal(t, "todushka", mm.projectNamesByID[pid])
}

func TestNameCache_FetchCmdEmitsMsg(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tg, err := svc.UpsertTag(ctx, "urgent")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{
		Title:  "t1",
		AreaID: &a.ID,
		Tags:   []id.ID{tg.ID},
	})
	require.NoError(t, err)

	cmd := fetchNameCache(svc, []task.Task{tk})
	require.NotNil(t, cmd)
	msg := cmd()
	res, ok := msg.(nameCacheLoadedMsg)
	require.True(t, ok)
	require.Equal(t, "work", res.areas[a.ID])
	require.Equal(t, "urgent", res.tags[tg.ID])
}

func TestWindowSize_BothAxesTracked(t *testing.T) {
	m := newTestModel(t)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	mm := m2.(Model)
	require.Equal(t, 120, mm.width)
	require.Equal(t, 40, mm.height)
}

func TestEditorSavedMsg_InlineSpliceByID(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "old title", "other")
	updated := tasks[0]
	updated.Title = "new title"
	m2, _ := m.Update(editorSavedMsg{updated: updated})
	mm := m2.(Model)
	require.Equal(t, "new title", mm.tasks[0].Title, "inline splice replaces by ID")
	require.Len(t, mm.tasks, 2, "length preserved")
}

func TestEditorSavedMsg_NotFoundSkipsSplice(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "a", "b")
	ghost := task.Task{ID: id.New(), Title: "ghost"}
	m2, _ := m.Update(editorSavedMsg{updated: ghost})
	mm := m2.(Model)
	require.Len(t, mm.tasks, 2)
	require.NotEqual(t, "ghost", mm.tasks[0].Title)
	require.NotEqual(t, "ghost", mm.tasks[1].Title)
}

func TestEditorSavedMsg_PreservesSliceLength(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c")
	require.Len(t, m.tasks, 3)
	upd := tasks[1]
	upd.Title = "modified"
	m2, _ := m.Update(editorSavedMsg{updated: upd})
	mm := m2.(Model)
	require.Len(t, mm.tasks, 3)
}

func TestEditorSavedMsg_FiresBatchedCmd(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "x")
	_, cmd := m.Update(editorSavedMsg{updated: tasks[0]})
	require.NotNil(t, cmd)
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	require.True(t, ok, "expected tea.BatchMsg, got %T", msg)
	foundCounts := false
	foundTasks := false
	for _, sub := range batch {
		if sub == nil {
			continue
		}
		switch sub().(type) {
		case countsLoadedMsg:
			foundCounts = true
		case tasksLoadedMsg:
			foundTasks = true
		}
	}
	require.True(t, foundCounts, "countsLoadedMsg expected in batch")
	require.True(t, foundTasks, "tasksLoadedMsg expected in batch")
}

// TestProp_EditSplicePreservesLength verifies CP-3 (REQ-2.4): an
// editorSavedMsg never changes the length of m.tasks, whether the
// updated task matches an existing ID or is a ghost.
func TestProp_EditSplicePreservesLength(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 6).Draw(rt, "n")
		titles := make([]string, n)
		for i := 0; i < n; i++ {
			titles[i] = fmt.Sprintf("t%d", i)
		}
		m, _, tasks := setupRapidModel(rt, titles...)
		// Sometimes target an existing task, sometimes a ghost
		useGhost := rapid.Bool().Draw(rt, "useGhost")
		var upd task.Task
		if useGhost {
			upd = task.Task{ID: id.New(), Title: "ghost"}
		} else {
			idx := rapid.IntRange(0, n-1).Draw(rt, "idx")
			upd = tasks[idx]
			upd.Title = "modified"
		}
		m2, _ := m.Update(editorSavedMsg{updated: upd})
		mm := m2.(Model)
		require.Equal(rt, n, len(mm.tasks), "length preserved")
	})
}

// TestProp_EditSpliceByID verifies CP-4 (REQ-2.1, 2.3): the inline splice
// targets the task whose ID matches; non-target tasks remain untouched.
func TestProp_EditSpliceByID(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 6).Draw(rt, "n")
		titles := make([]string, n)
		for i := 0; i < n; i++ {
			titles[i] = fmt.Sprintf("orig%d", i)
		}
		m, _, tasks := setupRapidModel(rt, titles...)
		idx := rapid.IntRange(0, n-1).Draw(rt, "idx")
		upd := tasks[idx]
		upd.Title = "spliced"
		m2, _ := m.Update(editorSavedMsg{updated: upd})
		mm := m2.(Model)
		require.Equal(rt, "spliced", mm.tasks[idx].Title)
		for i, tk := range mm.tasks {
			if i != idx {
				require.Equal(rt, tasks[i].Title, tk.Title, "non-target tasks unchanged")
			}
		}
	})
}

// TestProp_RefreshBatchHasBothCmds verifies CP-12 (REQ-2.2): the cmd
// produced after editorSavedMsg is a tea.BatchMsg that includes both a
// countsLoadedMsg producer and a tasksLoadedMsg producer.
func TestProp_RefreshBatchHasBothCmds(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 4).Draw(rt, "n")
		titles := make([]string, n)
		for i := 0; i < n; i++ {
			titles[i] = fmt.Sprintf("t%d", i)
		}
		m, _, tasks := setupRapidModel(rt, titles...)
		_, cmd := m.Update(editorSavedMsg{updated: tasks[0]})
		require.NotNil(rt, cmd)
		msg := cmd()
		batch, ok := msg.(tea.BatchMsg)
		require.True(rt, ok)
		foundCounts := false
		foundTasks := false
		for _, sub := range batch {
			if sub == nil {
				continue
			}
			switch sub().(type) {
			case countsLoadedMsg:
				foundCounts = true
			case tasksLoadedMsg:
				foundTasks = true
			}
		}
		require.True(rt, foundCounts, "countsLoadedMsg expected")
		require.True(rt, foundTasks, "tasksLoadedMsg expected")
	})
}

package tui

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func newTestModel(t *testing.T) Model {
	t.Helper()
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	return NewModel(svc, NewTheme())
}

func newTestModelWithService(t *testing.T) (Model, *app.Service) {
	t.Helper()
	svc := app.New(fakes.New(), fixedClock{now: time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)})
	return NewModel(svc, NewTheme()), svc
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

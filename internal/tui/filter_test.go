package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestDisplayedTasks_EmptyQuery(t *testing.T) {
	m := Model{tasks: []task.Task{{Title: "a"}, {Title: "b"}}}
	require.Equal(t, 2, len(displayedTasks(m)))
}

func TestDisplayedTasks_WhitespaceQueryEquivEmpty(t *testing.T) {
	m := Model{
		tasks:       []task.Task{{Title: "a"}, {Title: "b"}, {Title: "c"}},
		filterQuery: "   ",
	}
	require.Equal(t, 3, len(displayedTasks(m)))
}

func TestDisplayedTasks_SubstringMatch(t *testing.T) {
	m := Model{
		tasks: []task.Task{
			{Title: "foo bar"},
			{Title: "baz qux"},
			{Title: "foo baz"},
		},
		filterQuery: "foo",
	}
	got := displayedTasks(m)
	require.Equal(t, 2, len(got))
	require.Equal(t, "foo bar", got[0].Title)
	require.Equal(t, "foo baz", got[1].Title)
}

func TestDisplayedTasks_CaseInsensitive(t *testing.T) {
	m := Model{
		tasks:       []task.Task{{Title: "Hello World"}, {Title: "GOODBYE"}},
		filterQuery: "hello",
	}
	got := displayedTasks(m)
	require.Equal(t, 1, len(got))
	require.Equal(t, "Hello World", got[0].Title)
}

func TestFilter_SlashEntersMode(t *testing.T) {
	m := newTestModel(t)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := m2.(Model)
	require.True(t, mm.filtering)
	require.Equal(t, "", mm.filterQuery)

	m3, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mm = m3.(Model)
	m4, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	mm = m4.(Model)
	m5, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	mm = m5.(Model)
	require.Equal(t, "abc", mm.filterQuery)
}

func TestFilter_EnterPreservesQuery(t *testing.T) {
	m := newTestModel(t)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := m2.(Model)
	for _, r := range []rune{'k', 'u', 'p'} {
		mNext, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mm = mNext.(Model)
	}
	mFinal, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = mFinal.(Model)
	require.False(t, mm.filtering)
	require.Equal(t, "kup", mm.filterQuery)
}

func TestFilter_EscClearsQuery(t *testing.T) {
	m := newTestModel(t)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := m2.(Model)
	for _, r := range []rune{'k', 'u', 'p'} {
		mNext, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mm = mNext.(Model)
	}
	mFinal, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm = mFinal.(Model)
	require.False(t, mm.filtering)
	require.Equal(t, "", mm.filterQuery)
}

func TestFilter_BackspaceTrimsLastRune(t *testing.T) {
	m := newTestModel(t)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := m2.(Model)
	for _, r := range []rune{'a', 'b', 'c'} {
		mNext, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mm = mNext.(Model)
	}
	mFinal, _ := mm.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	mm = mFinal.(Model)
	require.Equal(t, "ab", mm.filterQuery)
}

func TestFilter_NoMatchesPlaceholder(t *testing.T) {
	m := newTestModel(t)
	m.tasks = []task.Task{{Title: "alpha"}, {Title: "beta"}}
	m.filterQuery = "zzznomatch"

	out := m.viewList()
	require.Contains(t, out, "(no matches)")
}

func TestFilter_NoTasksPlaceholderUnchanged(t *testing.T) {
	m := newTestModel(t)
	m.tasks = nil

	out := m.viewList()
	require.Contains(t, out, "(no tasks)")
}

func TestFilter_ChangeDropsHiddenFromSelection(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "alpha milk", "beta bread", "gamma milk")
	m.selected[tasks[0].ID] = struct{}{}
	m.selected[tasks[2].ID] = struct{}{}
	require.Equal(t, 2, len(m.selected))

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	mm := m2.(Model)
	require.True(t, mm.filtering)

	for _, r := range []rune{'b', 'r', 'e', 'a', 'd'} {
		mNext, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		mm = mNext.(Model)
	}
	require.Equal(t, "bread", mm.filterQuery)
	require.Equal(t, 0, len(mm.selected))
}

// TestProp_FilterIsSubstringSubset verifies CP-1 (REQ-1.2, 1.7):
// displayedTasks returns only tasks whose Title contains TrimSpace(query)
// fold-case, in original order.
func TestProp_FilterIsSubstringSubset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		titles := rapid.SliceOfN(rapid.StringMatching(`[a-zA-Z0-9 ]{1,15}`), 0, 8).Draw(t, "titles")
		query := rapid.StringMatching(`[a-zA-Z0-9 ]{0,8}`).Draw(t, "query")
		tasks := make([]task.Task, len(titles))
		for i, ti := range titles {
			tasks[i] = task.Task{ID: id.New(), Title: ti}
		}
		m := Model{tasks: tasks, filterQuery: query}
		result := displayedTasks(m)
		trimmed := strings.TrimSpace(query)
		if trimmed == "" {
			require.Equal(t, len(tasks), len(result))
			return
		}
		// Every result element must satisfy fold-case substring.
		for _, t2 := range result {
			require.True(t, foldCaseContains(t2.Title, trimmed),
				"result title %q must contain query %q (fold-case)", t2.Title, trimmed)
		}
		// Order preservation: result is in same relative order as tasks.
		idx := 0
		for _, src := range tasks {
			if idx < len(result) && result[idx].ID == src.ID {
				idx++
			}
		}
		require.Equal(t, len(result), idx, "result must preserve relative order of tasks")
	})
}

// TestProp_FilterStateTransitions verifies CP-2 (REQ-1.1, 1.3, 1.4):
// '/' enters filter mode; after a body of runes, Esc clears+exits and
// Enter saves+exits.
func TestProp_FilterStateTransitions(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		body := rapid.SliceOfN(rapid.StringMatching(`[a-z]`), 0, 5).Draw(t, "runes")
		closer := rapid.SampledFrom([]string{"esc", "enter"}).Draw(t, "closer")

		m := NewModel(nil, NewTheme())
		// Enter filter mode via '/'.
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		mm := m2.(Model)
		require.True(t, mm.filtering)
		require.Equal(t, "", mm.filterQuery)

		// Type body.
		var expectQuery strings.Builder
		for _, s := range body {
			r := []rune(s)[0]
			expectQuery.WriteRune(r)
			mNext, _ := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			mm = mNext.(Model)
		}
		require.Equal(t, expectQuery.String(), mm.filterQuery)
		require.True(t, mm.filtering)

		// Close the filter.
		switch closer {
		case "esc":
			mFinal, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
			mm = mFinal.(Model)
			require.False(t, mm.filtering, "Esc must exit filter mode")
			require.Equal(t, "", mm.filterQuery, "Esc must clear the filter query")
		case "enter":
			mFinal, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
			mm = mFinal.(Model)
			require.False(t, mm.filtering, "Enter must exit filter mode")
			require.Equal(t, expectQuery.String(), mm.filterQuery, "Enter must preserve the filter query")
		}
	})
}

// TestProp_NoMatchesShowsPlaceholder verifies CP-3 (REQ-1.5): when
// filtering narrows the visible set to zero, viewList renders the
// "(no matches)" placeholder.
func TestProp_NoMatchesShowsPlaceholder(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		titles := rapid.SliceOfN(rapid.StringMatching(`[a-z]{1,10}`), 1, 8).Draw(t, "titles")
		tasks := make([]task.Task, len(titles))
		for i, ti := range titles {
			tasks[i] = task.Task{ID: id.New(), Title: ti}
		}
		// "~~~" cannot occur in any [a-z] title, so the filter excludes everything.
		m := NewModel(nil, NewTheme())
		m.tasks = tasks
		m.filterQuery = "~~~"

		require.Empty(t, displayedTasks(m), "filter must hide every task")
		out := m.viewList()
		require.Contains(t, out, "(no matches)", "viewList must show the no-matches placeholder")
	})
}

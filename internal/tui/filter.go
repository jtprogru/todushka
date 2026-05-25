package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
)

// foldCaseContains reports whether haystack contains needle using a case-fold
// (lowercase) comparison. Used by the TUI filter to match task titles
// regardless of letter case.
func foldCaseContains(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// displayedTasks returns the slice of tasks visible after applying the model's
// current filter query. When the query is empty (or whitespace-only) the raw
// task slice is returned. Otherwise a new slice is built preserving original
// order with only the tasks whose Title matches the trimmed query (case-fold).
func displayedTasks(m Model) []task.Task {
	q := strings.TrimSpace(m.filterQuery)
	if q == "" {
		return m.tasks
	}
	out := make([]task.Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		if foldCaseContains(t.Title, q) {
			out = append(out, t)
		}
	}
	return out
}

// pruneSelection drops any IDs from m.selected that are not in displayedTasks(m).
// Called after any change to filterQuery so the selection set always
// satisfies REQ-2.6 (Selection ⊆ Visible Tasks).
func pruneSelection(m *Model) {
	visible := make(map[id.ID]struct{}, len(m.tasks))
	for _, t := range displayedTasks(*m) {
		visible[t.ID] = struct{}{}
	}
	for sid := range m.selected {
		if _, ok := visible[sid]; !ok {
			delete(m.selected, sid)
		}
	}
}

// handleFilterKey processes a key message while the TUI is in filter-edit
// mode. It mutates filterQuery / filtering on the returned Model.
func handleFilterKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.filterQuery = ""
		m.filtering = false
		return m, nil
	case tea.KeyEnter:
		m.filtering = false
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		runes := []rune(m.filterQuery)
		if len(runes) > 0 {
			runes = runes[:len(runes)-1]
		}
		m.filterQuery = string(runes)
		pruneSelection(&m)
		return m, nil
	case tea.KeyRunes:
		m.filterQuery += string(msg.Runes)
		pruneSelection(&m)
		return m, nil
	case tea.KeySpace:
		m.filterQuery += " "
		pruneSelection(&m)
		return m, nil
	}
	return m, nil
}

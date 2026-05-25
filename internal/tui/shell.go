// Package tui — shell.go hosts the zellij-style "shell" chrome (mode
// detection + footer) shared across screens. View logic specific to a
// single screen continues to live in its dedicated file (details.go,
// editor.go, etc.). The shell layer reads Model state and renders the
// mode chip + context-aware key hints + status message.
package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
)

// shellMode is the high-level interaction mode displayed in the footer.
// Exactly one mode is active at any given time; precedence is encoded in
// currentMode.
type shellMode int

const (
	modeNormal shellMode = iota
	modeFilter
	modeSelect
	modeConfirm
	modeEditor
	modeHelp
)

// modeLabel returns the human-readable label used inside the footer chip
// (e.g. "NORMAL", "FILTER").
func (m shellMode) modeLabel() string {
	switch m {
	case modeNormal:
		return "NORMAL"
	case modeFilter:
		return "FILTER"
	case modeSelect:
		return "SELECT"
	case modeConfirm:
		return "CONFIRM"
	case modeEditor:
		return "EDITOR"
	case modeHelp:
		return "HELP"
	}
	return "?"
}

// currentMode determines the active mode following the priority order:
// HELP > EDITOR > CONFIRM > FILTER > SELECT > NORMAL. This priority is
// observable via the footer chip and via key hints.
func currentMode(m Model) shellMode {
	switch {
	case m.screen == screenHelp:
		return modeHelp
	case m.screen == screenEditor:
		return modeEditor
	case m.confirm != nil:
		return modeConfirm
	case m.filtering:
		return modeFilter
	case len(m.selected) > 0:
		return modeSelect
	default:
		return modeNormal
	}
}

// modeKeyHints returns the context-aware key hints displayed next to the
// mode chip. Each entry is rendered as a single "key: action" token; the
// caller joins them with a separator.
func modeKeyHints(mode shellMode) []string {
	switch mode {
	case modeNormal:
		return []string{"/: filter", "space: select", "n: quick", "↵: edit", "c: complete", "?: help", "q: quit"}
	case modeFilter:
		return []string{"↵: save", "esc: cancel"}
	case modeSelect:
		return []string{"c/x/d/p: bulk", "*: all", "esc: clear"}
	case modeConfirm:
		return []string{"y: yes", "any: cancel"}
	case modeEditor:
		return []string{"Tab: next", "Shift+Tab: prev", "Ctrl+S: save", "esc: cancel"}
	case modeHelp:
		return []string{"?: close"}
	}
	return nil
}

// viewFooter renders the zellij-style footer:
//
//	[ -- MODE -- ] hint1 │ hint2 │ ...           status
//
// In FILTER mode the live filter query is prepended to the hint list. In
// SELECT mode the selection count is appended. The status message (if
// any) is rendered on the right; in CONFIRM mode it uses the error style
// because confirmation prompts surface destructive operations.
func (m Model) viewFooter() string {
	mode := currentMode(m)
	chip := m.theme.Header.Render(fmt.Sprintf(" -- %s -- ", mode.modeLabel()))

	hints := modeKeyHints(mode)
	if mode == modeFilter {
		hints = append([]string{"Filter: " + m.filterQuery + "_"}, hints...)
	}
	if mode == modeSelect {
		hints = append(hints, fmt.Sprintf("Selected: %d", len(m.selected)))
	}

	hintsRendered := m.theme.Help.Render(strings.Join(hints, " │ "))

	var status string
	if m.statusMsg != "" {
		if mode == modeConfirm {
			status = m.theme.StatusError.Render(m.statusMsg)
		} else {
			status = m.theme.Help.Render(m.statusMsg)
		}
	}

	left := lipgloss.JoinHorizontal(lipgloss.Top, chip, " ", hintsRendered)
	if status != "" {
		return left + "  " + status
	}
	return left
}

const headerCompactThreshold = 80

var listInitials = map[listKind]string{
	listInbox:    "I",
	listToday:    "T",
	listUpcoming: "U",
	listAnytime:  "A",
	listSomeday:  "S",
	listLogbook:  "L",
}

// renderHeaderSegment formats one header segment.
// active=true → entire segment styled with theme.Header (inverted bg).
// compact=true → "(N)I[Count]"; compact=false → "(N) Name [Count]".
// knownCount=false → display "[?]" instead of "[Count]".
func renderHeaderSegment(theme Theme, idx int, label string, count int, knownCount bool, active, compact bool) string {
	n := fmt.Sprintf("(%d)", idx+1)
	countStr := "[?]"
	if knownCount {
		countStr = fmt.Sprintf("[%d]", count)
	}
	var raw string
	if compact {
		initial := label[:1]
		if i, ok := listInitials[allLists[idx]]; ok {
			initial = i
		}
		raw = n + initial + countStr
	} else {
		raw = n + " " + label + " " + countStr
	}
	if active {
		return theme.Header.Render(raw)
	}
	var b strings.Builder
	if compact {
		initial := label[:1]
		if i, ok := listInitials[allLists[idx]]; ok {
			initial = i
		}
		b.WriteString(theme.Selected.Render(n))
		b.WriteString(theme.HeaderDim.Render(initial))
		b.WriteString(theme.Dim.Render(countStr))
	} else {
		b.WriteString(theme.Selected.Render(n))
		b.WriteString(" ")
		b.WriteString(theme.HeaderDim.Render(label))
		b.WriteString(" ")
		b.WriteString(theme.Dim.Render(countStr))
	}
	return b.String()
}

func (m Model) viewHeader() string {
	compact := m.width > 0 && m.width < headerCompactThreshold
	labels := []string{"Inbox", "Today", "Upcoming", "Anytime", "Someday", "Logbook"}
	parts := make([]string, 0, len(allLists))
	for i, l := range allLists {
		count, known := m.listCounts[l]
		active := l == m.activeList
		parts = append(parts, renderHeaderSegment(m.theme, i, labels[i], count, known, active, compact))
	}
	return strings.Join(parts, " ")
}

func fetchListCounts(svc *app.Service) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		counts := make(map[listKind]int, 6)
		if list, err := svc.ListInbox(ctx); err == nil {
			counts[listInbox] = len(list)
		}
		if list, err := svc.ListToday(ctx); err == nil {
			counts[listToday] = len(list)
		}
		if list, err := svc.ListUpcoming(ctx); err == nil {
			counts[listUpcoming] = len(list)
		}
		if list, err := svc.ListAnytime(ctx); err == nil {
			counts[listAnytime] = len(list)
		}
		if list, err := svc.ListSomeday(ctx); err == nil {
			counts[listSomeday] = len(list)
		}
		if list, err := svc.ListLogbook(ctx); err == nil {
			counts[listLogbook] = len(list)
		}
		return countsLoadedMsg{counts: counts}
	}
}

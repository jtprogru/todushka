package tui

import (
	"github.com/jtprogru/todushka/internal/domain/task"
)

// Internal Bubble Tea messages.

type tasksLoadedMsg struct {
	tasks []task.Task
}

type errorMsg struct{ err error }

type clearStatusMsg struct{}

type quickEntrySubmittedMsg struct{ raw string }

type screenKind int

const (
	screenList screenKind = iota
	screenQuickEntry
	screenHelp
	screenEditor
)

type editorSavedMsg struct{ updated task.Task }

type listKind int

const (
	listInbox listKind = iota
	listToday
	listUpcoming
	listAnytime
	listSomeday
	listLogbook
)

func (l listKind) String() string {
	switch l {
	case listInbox:
		return "Inbox"
	case listToday:
		return "Today"
	case listUpcoming:
		return "Upcoming"
	case listAnytime:
		return "Anytime"
	case listSomeday:
		return "Someday"
	case listLogbook:
		return "Logbook"
	}
	return "?"
}

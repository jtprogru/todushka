package tui

import (
	"github.com/jtprogru/todushka/internal/domain/id"
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

// bulkResultMsg is emitted by runBulk after iterating all selected IDs.
// Fatal=true when the run was halted by storage.ErrDatabaseLocked or
// context.Canceled; in that case the selection set is preserved.
type bulkResultMsg struct {
	action    bulkAction
	succeeded int
	failed    int
	lastErr   error
	fatal     bool
}

// singleActionDoneMsg is emitted by per-cursor write Cmds after the service
// call returns. Success → reload chained by handler; err != nil → status
// fade only (no reload — keeps the optimistic splice visible until manual r).
type singleActionDoneMsg struct {
	action bulkAction
	tid    id.ID
	err    error
}

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

// nameCacheLoadedMsg carries names resolved by fetchNameCache.
// Per-map keys are not exclusive — Update merges them additively into Model.
type nameCacheLoadedMsg struct {
	tags     map[id.ID]string
	areas    map[id.ID]string
	projects map[id.ID]string
	headings map[id.ID]string
}

// countsLoadedMsg carries per-list task counts after fetchListCounts Cmd.
type countsLoadedMsg struct {
	counts map[listKind]int
}

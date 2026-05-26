package tui

import (
	"context"
	"errors"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/storage"
)

type bulkAction int

const (
	bulkActionComplete bulkAction = iota
	bulkActionCancel
	bulkActionDelete
	bulkActionPin
)

// confirmState tracks a pending bulk operation awaiting user confirmation.
type confirmState struct {
	action bulkAction
	ids    []id.ID
}

func (a bulkAction) label() string {
	switch a {
	case bulkActionComplete:
		return "Complete"
	case bulkActionCancel:
		return "Cancel"
	case bulkActionDelete:
		return "Delete"
	case bulkActionPin:
		return "Pin"
	}
	return "?"
}

// dispatch routes an action key based on |m.selected|. Empty selection falls
// back to per-cursor; 1..m.config.BulkConfirmThreshold-1 runs immediately;
// >=threshold installs a confirm modal.
func dispatch(m Model, action bulkAction) (Model, tea.Cmd) {
	if m.readOnly {
		m.statusMsg = "read-only mode: writes disabled"
		m.statusUntil = time.Now().Add(statusFadeDuration)
		return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
	}
	if len(m.selected) == 0 {
		if action == bulkActionDelete && m.config.ConfirmDelete {
			sel := m.selectedTask()
			if sel == nil {
				return m, nil
			}
			m.confirm = &confirmState{action: bulkActionDelete, ids: []id.ID{sel.ID}}
			return m, nil
		}
		return perCursorAction(m, action)
	}
	ids := selectionIDs(m)
	if len(ids) < m.config.BulkConfirmThreshold {
		return m, runBulk(m.service, action, ids)
	}
	m.confirm = &confirmState{action: action, ids: ids}
	return m, nil
}

// selectionIDs snapshots m.selected into a slice. Order is unspecified but
// stable within a single dispatch call.
func selectionIDs(m Model) []id.ID {
	out := make([]id.ID, 0, len(m.selected))
	for tid := range m.selected {
		out = append(out, tid)
	}
	return out
}

// perCursorAction applies an action to the cursor task with an optimistic
// inline splice of m.tasks (REQ-2.1..2.4) and a Cmd that fires the service
// write asynchronously. On completion the Cmd emits a singleActionDoneMsg
// so the Update loop can chain loadCurrentList for an accurate refresh.
func perCursorAction(m Model, action bulkAction) (Model, tea.Cmd) {
	switch action {
	case bulkActionComplete:
		return m.completeSelected()
	case bulkActionCancel:
		return m.cancelSelected()
	case bulkActionDelete:
		return m.deleteSelected()
	case bulkActionPin:
		return m.pinSelected()
	}
	return m, nil
}

// runBulk applies action to each ID sequentially. Recoverable errors are
// counted into bulkResultMsg.failed. context.Canceled and
// storage.ErrDatabaseLocked are fatal and abort the run.
func runBulk(svc *app.Service, action bulkAction, ids []id.ID) tea.Cmd {
	return func() tea.Msg {
		res := bulkResultMsg{action: action}
		ctx := context.Background()
		for _, tid := range ids {
			err := applyAction(ctx, svc, action, tid)
			if err == nil {
				res.succeeded++
				continue
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, storage.ErrDatabaseLocked) {
				res.fatal = true
				res.lastErr = err
				return res
			}
			res.failed++
			res.lastErr = err
		}
		return res
	}
}

// applyAction is the per-task service call for a bulk action.
func applyAction(ctx context.Context, svc *app.Service, action bulkAction, tid id.ID) error {
	switch action {
	case bulkActionComplete:
		_, err := svc.CompleteTask(ctx, tid)
		return err
	case bulkActionCancel:
		return svc.CancelTask(ctx, tid)
	case bulkActionDelete:
		return svc.DeleteTask(ctx, tid, false)
	case bulkActionPin:
		return svc.PinToToday(ctx, tid)
	}
	return nil
}

// handleConfirmKey processes keys while a confirm modal is active.
// 'y' dispatches the pending action; any other key dismisses without action.
// In both cases the modal closes (m.confirm = nil). A single-task confirm
// (len(c.ids)==1) routes through the per-id splice helper so the action
// targets the originally-confirmed task, even if the cursor moved due to
// an async reload between 'd' and 'y'.
func handleConfirmKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	c := m.confirm
	m.confirm = nil
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'y' {
		if len(c.ids) == 1 {
			return singleActionByID(m, c.action, c.ids[0])
		}
		return m, runBulk(m.service, c.action, c.ids)
	}
	return m, nil
}

// singleActionByID applies a per-cursor-style action to the task with the
// given ID rather than the cursor task. Used by the single-task confirm
// modal so the originally-confirmed task is the one acted on, regardless
// of any cursor drift caused by an async reload in flight.
func singleActionByID(m Model, action bulkAction, tid id.ID) (Model, tea.Cmd) {
	idx := -1
	for i := range m.tasks {
		if m.tasks[i].ID == tid {
			idx = i
			break
		}
	}
	if idx < 0 {
		// Task no longer in displayed list; let the service write
		// proceed; an upcoming loadCurrentList will reconcile.
		return m, fireSingleAction(m.service, action, tid)
	}
	prevCursor := m.cursor
	m.cursor = idx
	m2, cmd := perCursorAction(m, action)
	if action != bulkActionDelete {
		m2.cursor = prevCursor
	}
	return m2, cmd
}

// fireSingleAction issues a service write for a task no longer in m.tasks
// (edge case for singleActionByID). Mirrors the Cmds in completeSelected
// etc., but without the optimistic splice (the task is already absent).
func fireSingleAction(svc *app.Service, action bulkAction, tid id.ID) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch action {
		case bulkActionComplete:
			_, err = svc.CompleteTask(context.Background(), tid)
		case bulkActionCancel:
			err = svc.CancelTask(context.Background(), tid)
		case bulkActionDelete:
			err = svc.DeleteTask(context.Background(), tid, false)
		case bulkActionPin:
			err = svc.PinToToday(context.Background(), tid)
		}
		if err != nil {
			return singleActionDoneMsg{action: action, tid: tid, err: err}
		}
		return singleActionDoneMsg{action: action, tid: tid}
	}
}

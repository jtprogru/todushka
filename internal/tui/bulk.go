package tui

import (
	"context"
	"errors"

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

const bulkConfirmThreshold = 5

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
// back to per-cursor; 1..bulkConfirmThreshold-1 runs immediately; >=threshold
// installs a confirm modal.
func dispatch(m Model, action bulkAction) (Model, tea.Cmd) {
	if len(m.selected) == 0 {
		return m, perCursorCmd(m, action)
	}
	ids := selectionIDs(m)
	if len(ids) < bulkConfirmThreshold {
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

// perCursorCmd returns the existing per-cursor Cmd for an action. This
// delegates to the methods on Model that were the entry points before T-5.
func perCursorCmd(m Model, action bulkAction) tea.Cmd {
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
	return nil
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
// 'y' dispatches the pending bulk run; any other key dismisses without action.
// In both cases the modal closes (m.confirm = nil).
func handleConfirmKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	c := m.confirm
	m.confirm = nil
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'y' {
		return m, runBulk(m.service, c.action, c.ids)
	}
	return m, nil
}

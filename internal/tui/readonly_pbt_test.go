package tui

// T-8 — Property-based tests for the read-only mode feature (TUI side).
//
// Coverage map (design.md §2.6):
//   - CP-8 (Model.readOnly reflects Repo): covered by
//     TestTUI_ModelReadOnlyReflectsRepo in shell_test.go.
//   - CP-9 (currentMode priority): covered below by
//     TestProp_TUI_CurrentModePriority — random subsets of transient
//     modes; assert highest-priority enabled mode wins.
//   - CP-10 (mode chip "READ-ONLY"): covered by TestTUI_ModeChipReadOnly.
//   - CP-11 + CP-12 (write key in RO blocks service AND sets status):
//     covered below by TestProp_TUI_WriteKeyBlockedInRO over the
//     {c,x,d,p,n} key set.
//   - CP-13 (editor save in RO): covered by TestTUI_EditorSaveBlockedInRO.

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// TestProp_TUI_CurrentModePriority verifies CP-9: regardless of which
// subset of {filter, select, confirm, editor, help, readOnly} is
// enabled, currentMode returns the highest-priority enabled mode per
// the documented order HELP > EDITOR > CONFIRM > FILTER > SELECT >
// READ-ONLY > NORMAL.
func TestProp_TUI_CurrentModePriority(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		help := rapid.Bool().Draw(rt, "help")
		editor := rapid.Bool().Draw(rt, "editor")
		confirm := rapid.Bool().Draw(rt, "confirm")
		filter := rapid.Bool().Draw(rt, "filter")
		sel := rapid.Bool().Draw(rt, "select")
		ro := rapid.Bool().Draw(rt, "readOnly")

		m := newTestModel(t)
		switch {
		case help:
			m.screen = screenHelp
		case editor:
			m.screen = screenEditor
		default:
			m.screen = screenList
		}
		if confirm {
			m.confirm = &confirmState{}
		}
		m.filtering = filter
		if sel {
			m.selected[id.New()] = struct{}{}
		}
		m.readOnly = ro

		want := modeNormal
		switch {
		case help:
			want = modeHelp
		case editor:
			want = modeEditor
		case confirm:
			want = modeConfirm
		case filter:
			want = modeFilter
		case sel:
			want = modeSelect
		case ro:
			want = modeReadOnly
		}
		require.Equal(rt, want, currentMode(m))
	})
}

// TestProp_TUI_WriteKeyBlockedInRO verifies CP-11 + CP-12 jointly: for
// any write key in {c, x, d, p, n} pressed on a RO model, the resulting
// state must (a) carry a "read-only" status message, (b) leave the
// backing service repository unchanged, and (c) for 'n' specifically,
// keep the screen on screenList (no quick-entry modal opens).
func TestProp_TUI_WriteKeyBlockedInRO(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		key := rapid.SampledFrom([]rune{'c', 'x', 'd', 'p', 'n'}).Draw(rt, "key")

		m, svc, _ := setupModelWithInboxTasks(t, "alpha", "beta", "gamma")
		m.readOnly = true

		ctx := context.Background()
		beforeList, err := svc.Repo().TaskList(ctx, storage.TaskFilter{})
		require.NoError(rt, err)
		beforeStatuses := snapshotStatuses(beforeList)

		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		mm := m2.(Model)
		require.Contains(rt, mm.statusMsg, "read-only", "key %q must set RO status", key)

		afterList, err := svc.Repo().TaskList(ctx, storage.TaskFilter{})
		require.NoError(rt, err)
		require.Len(rt, afterList, len(beforeList), "task count unchanged for key %q", key)
		require.Equal(rt, beforeStatuses, snapshotStatuses(afterList), "statuses unchanged for key %q", key)

		if key == 'n' {
			require.Equal(rt, screenList, mm.screen, "quick-entry modal must not open in RO")
		}
	})
}

func snapshotStatuses(ts []task.Task) map[id.ID]task.Status {
	out := make(map[id.ID]task.Status, len(ts))
	for _, t := range ts {
		out[t.ID] = t.Status
	}
	return out
}

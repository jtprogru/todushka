package tui

import (
	"context"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestBulk_EmptyDispatchesPerCursor(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "x", "y")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.NotNil(t, cmd)
	_ = cmd()
	got, err := svc.Repo().TaskGet(context.Background(), tasks[0].ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, got.Status)
}

func TestBulk_BelowThresholdNoConfirm(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c", "d")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	require.Equal(t, 4, len(m.selected))
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	mm := m2.(Model)
	require.Nil(t, mm.confirm)
	require.NotNil(t, cmd)
}

func TestBulk_AtThresholdShowsConfirm(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c", "d", "e")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	require.Equal(t, 5, len(m.selected))
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	mm := m2.(Model)
	require.NotNil(t, mm.confirm)
	require.Equal(t, bulkActionComplete, mm.confirm.action)
	require.Equal(t, 5, len(mm.confirm.ids))
	require.Nil(t, cmd)
}

func TestBulk_YConfirmsAndDispatches(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "a", "b", "c", "d", "e")
	ids := make([]id.ID, len(tasks))
	for i, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
		ids[i] = tk.ID
	}
	// Trigger modal
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.Nil(t, cmd)
	mm := m2.(Model)
	require.NotNil(t, mm.confirm)
	// Confirm with 'y'
	m3, cmd2 := mm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	mm = m3.(Model)
	require.Nil(t, mm.confirm)
	require.NotNil(t, cmd2)
	// Execute the bulk Cmd and check that some tasks transitioned
	msg := cmd2()
	res, ok := msg.(bulkResultMsg)
	require.True(t, ok)
	require.Equal(t, bulkActionComplete, res.action)
	require.Equal(t, 5, res.succeeded)
	require.Equal(t, 0, res.failed)
	// Spot-check repo
	got, err := svc.Repo().TaskGet(context.Background(), ids[0])
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, got.Status)
}

func TestBulk_NonYDismissesNoDispatch(t *testing.T) {
	m, svc, tasks := setupModelWithInboxTasks(t, "a", "b", "c", "d", "e")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	// Press 'n'
	m3, cmd := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mm := m3.(Model)
	require.Nil(t, mm.confirm)
	require.Nil(t, cmd)
	// Tasks must remain open
	got, err := svc.Repo().TaskGet(context.Background(), tasks[0].ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusOpen, got.Status)
}

func TestBulk_SuccessClearsSelection(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	require.Equal(t, 3, len(m.selected))
	// Inject a bulkResultMsg directly (below threshold so the modal isn't involved)
	m2, _ := m.Update(bulkResultMsg{action: bulkActionComplete, succeeded: 3, failed: 0, fatal: false})
	mm := m2.(Model)
	require.Equal(t, 0, len(mm.selected))
}

func TestBulk_FatalPreservesSelection(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	require.Equal(t, 2, len(m.selected))
	m2, _ := m.Update(bulkResultMsg{action: bulkActionDelete, fatal: true, lastErr: storage.ErrDatabaseLocked})
	mm := m2.(Model)
	require.Equal(t, 2, len(mm.selected), "fatal must NOT clear selection")
}

func TestBulk_AggregateMath(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "real1", "real2", "real3")
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	// Add 2 ghost IDs (never persisted) — service will return ErrTaskNotFound for them
	ghost1 := id.New()
	ghost2 := id.New()
	m.selected[ghost1] = struct{}{}
	m.selected[ghost2] = struct{}{}
	require.Equal(t, 5, len(m.selected)) // crosses confirm threshold

	// Trigger modal
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	require.NotNil(t, m2.(Model).confirm)

	// Confirm with 'y'
	m3, cmd := m2.(Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	require.NotNil(t, cmd)
	_ = m3

	// Execute bulk Cmd
	res := cmd().(bulkResultMsg)
	require.Equal(t, 3, res.succeeded)
	require.Equal(t, 2, res.failed)
	require.NotNil(t, res.lastErr)
}

// bulkActionKeys enumerates the four action keys that route through dispatch.
func bulkActionKeys() []struct {
	key    rune
	action bulkAction
} {
	return []struct {
		key    rune
		action bulkAction
	}{
		{'c', bulkActionComplete},
		{'x', bulkActionCancel},
		{'d', bulkActionDelete},
		{'p', bulkActionPin},
	}
}

// TestProp_EmptySelectionEquivCursor verifies CP-11 (REQ-3.1): with
// no selection, each action key transitions the cursor task as the
// per-cursor handler would.
func TestProp_EmptySelectionEquivCursor(t *testing.T) {
	for _, ak := range bulkActionKeys() {
		ak := ak
		t.Run(string(ak.key), func(t *testing.T) {
			rapid.Check(t, func(rt *rapid.T) {
				n := rapid.IntRange(1, 5).Draw(rt, "n")
				titles := make([]string, n)
				for i := range titles {
					titles[i] = fmt.Sprintf("task-%d", i)
				}
				m, svc, tasks := setupRapidModel(rt, titles...)
				cursorIdx := rapid.IntRange(0, n-1).Draw(rt, "cursor")
				m.cursor = cursorIdx
				require.Empty(rt, m.selected)

				_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ak.key}})
				require.NotNil(rt, cmd, "per-cursor action must return a non-nil Cmd")
				_ = cmd()

				ctx := context.Background()
				switch ak.action {
				case bulkActionComplete:
					got, err := svc.Repo().TaskGet(ctx, tasks[cursorIdx].ID)
					require.NoError(rt, err)
					require.Equal(rt, task.StatusCompleted, got.Status)
				case bulkActionCancel:
					got, err := svc.Repo().TaskGet(ctx, tasks[cursorIdx].ID)
					require.NoError(rt, err)
					require.Equal(rt, task.StatusCancelled, got.Status)
				case bulkActionDelete:
					inbox, err := svc.ListInbox(ctx)
					require.NoError(rt, err)
					require.Equal(rt, n-1, len(inbox), "delete must remove exactly one task from inbox")
				case bulkActionPin:
					got, err := svc.Repo().TaskGet(ctx, tasks[cursorIdx].ID)
					require.NoError(rt, err)
					require.NotNil(rt, got.PinnedToday)
				}
			})
		})
	}
}

// TestProp_BulkThresholdGate verifies CP-12 (REQ-3.2, 3.3): for
// |selected|=N, pressing 'c' opens a confirm modal iff N ≥ 5.
func TestProp_BulkThresholdGate(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(rt, "n")
		titles := make([]string, n)
		for i := range titles {
			titles[i] = fmt.Sprintf("task-%d", i)
		}
		m, _, tasks := setupRapidModel(rt, titles...)
		for _, tk := range tasks {
			m.selected[tk.ID] = struct{}{}
		}
		require.Equal(rt, n, len(m.selected))

		m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		mm := m2.(Model)
		if n < config.Defaults().BulkConfirmThreshold {
			require.Nil(rt, mm.confirm, "below threshold: must NOT open modal")
			require.NotNil(rt, cmd, "below threshold: must dispatch immediately")
		} else {
			require.NotNil(rt, mm.confirm, "at/above threshold: must open modal")
			require.Nil(rt, cmd, "at/above threshold: must wait for confirmation")
		}
	})
}

// TestProp_OnlyYConfirms verifies CP-13 (REQ-3.4): once the confirm
// modal is open, only 'y' triggers the bulk run; every other key
// dismisses without dispatching.
func TestProp_OnlyYConfirms(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Setup at threshold so the modal opens.
		titles := []string{"a", "b", "c", "d", "e"}
		m, _, tasks := setupRapidModel(rt, titles...)
		for _, tk := range tasks {
			m.selected[tk.ID] = struct{}{}
		}
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
		mm := m2.(Model)
		require.NotNil(rt, mm.confirm, "precondition: modal must be open")

		// Draw a key: 'y' or any non-'y' rune in [a-z] excluding 'y' (and
		// 'q' to avoid tea.Quit pre-empting via the global quit handler).
		useY := rapid.Bool().Draw(rt, "useY")
		var keyMsg tea.KeyMsg
		if useY {
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
		} else {
			r := rapid.SampledFrom([]rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'r', 's', 't', 'u', 'v', 'w', 'x', 'z'}).Draw(rt, "rune")
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		}

		m3, cmd := mm.Update(keyMsg)
		mm = m3.(Model)
		require.Nil(rt, mm.confirm, "any key must dismiss the modal")
		if useY {
			require.NotNil(rt, cmd, "'y' must dispatch the bulk Cmd")
		} else {
			require.Nil(rt, cmd, "non-'y' keys must NOT dispatch")
		}
	})
}

// TestProp_BulkAggregateMath verifies CP-14 (REQ-3.5):
// succeeded+failed == len(ids), and the partition matches real vs
// ghost IDs.
func TestProp_BulkAggregateMath(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		numReal := rapid.IntRange(1, 5).Draw(rt, "numReal")
		numGhost := rapid.IntRange(0, 5).Draw(rt, "numGhost")

		titles := make([]string, numReal)
		for i := range titles {
			titles[i] = fmt.Sprintf("real-%d", i)
		}
		_, svc, tasks := setupRapidModel(rt, titles...)

		ids := make([]id.ID, 0, numReal+numGhost)
		for _, tk := range tasks {
			ids = append(ids, tk.ID)
		}
		for i := 0; i < numGhost; i++ {
			ids = append(ids, id.New())
		}

		cmd := runBulk(svc, bulkActionComplete, ids)
		require.NotNil(rt, cmd)
		res, ok := cmd().(bulkResultMsg)
		require.True(rt, ok)

		require.Equal(rt, len(ids), res.succeeded+res.failed, "succeeded+failed must equal |ids|")
		require.Equal(rt, numReal, res.succeeded, "every real ID should succeed")
		require.Equal(rt, numGhost, res.failed, "every ghost ID should fail")
		require.False(rt, res.fatal, "non-fatal errors must not flip fatal=true")
	})
}

// TestProp_SuccessClearsSelection verifies CP-15 (REQ-3.6): a
// non-fatal bulkResultMsg empties m.selected regardless of size.
func TestProp_SuccessClearsSelection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(rt, "n")
		m := bareTestModel()
		for i := 0; i < n; i++ {
			m.selected[id.New()] = struct{}{}
		}
		require.Equal(rt, n, len(m.selected))

		m2, _ := m.Update(bulkResultMsg{action: bulkActionComplete, succeeded: n, failed: 0, fatal: false})
		mm := m2.(Model)
		require.Equal(rt, 0, len(mm.selected), "successful bulk must clear selection")
	})
}

// TestProp_FatalPreservesSelection verifies CP-16 (REQ-3.7): a
// fatal bulkResultMsg leaves m.selected unchanged.
func TestProp_FatalPreservesSelection(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(rt, "n")
		m := bareTestModel()
		ids := make([]id.ID, 0, n)
		for i := 0; i < n; i++ {
			tid := id.New()
			m.selected[tid] = struct{}{}
			ids = append(ids, tid)
		}
		require.Equal(rt, n, len(m.selected))

		m2, _ := m.Update(bulkResultMsg{action: bulkActionDelete, fatal: true, lastErr: storage.ErrDatabaseLocked})
		mm := m2.(Model)
		require.Equal(rt, n, len(mm.selected), "fatal bulk must preserve selection")
		for _, sid := range ids {
			_, ok := mm.selected[sid]
			require.True(rt, ok, "every previously selected ID must remain")
		}
	})
}

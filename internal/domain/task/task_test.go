package task

import (
	"testing"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/stretchr/testify/require"
)

func TestTask_ValidateEmptyTitle(t *testing.T) {
	tk := Task{Title: "  ", Status: StatusOpen}
	err := tk.Validate()
	require.ErrorIs(t, err, ErrEmptyTitle)
}

func TestTask_ValidateStatusInvariants(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		task    Task
		wantErr error
	}{
		{
			name:    "open_with_no_timestamps_ok",
			task:    Task{Title: "x", Status: StatusOpen},
			wantErr: nil,
		},
		{
			name:    "open_with_completed_at_fails",
			task:    Task{Title: "x", Status: StatusOpen, CompletedAt: &now},
			wantErr: ErrStatusInvariant,
		},
		{
			name:    "open_with_cancelled_at_fails",
			task:    Task{Title: "x", Status: StatusOpen, CancelledAt: &now},
			wantErr: ErrStatusInvariant,
		},
		{
			name:    "completed_requires_completed_at",
			task:    Task{Title: "x", Status: StatusCompleted},
			wantErr: ErrStatusInvariant,
		},
		{
			name:    "completed_with_completed_at_ok",
			task:    Task{Title: "x", Status: StatusCompleted, CompletedAt: &now},
			wantErr: nil,
		},
		{
			name:    "completed_and_cancelled_fails",
			task:    Task{Title: "x", Status: StatusCompleted, CompletedAt: &now, CancelledAt: &now},
			wantErr: ErrStatusInvariant,
		},
		{
			name:    "cancelled_with_cancelled_at_ok",
			task:    Task{Title: "x", Status: StatusCancelled, CancelledAt: &now},
			wantErr: nil,
		},
		{
			name:    "cancelled_requires_cancelled_at",
			task:    Task{Title: "x", Status: StatusCancelled},
			wantErr: ErrStatusInvariant,
		},
		{
			name:    "unknown_status_rejected",
			task:    Task{Title: "x", Status: "weird"},
			wantErr: ErrInvalidStatus,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.task.Validate()
			if c.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, c.wantErr)
			}
		})
	}
}

func TestTask_ChecklistValidation(t *testing.T) {
	t.Run("empty_text_rejected", func(t *testing.T) {
		ci := ChecklistItem{ID: id.New(), Text: "   "}
		tk := Task{Title: "x", Status: StatusOpen, Checklist: []ChecklistItem{ci}}
		require.ErrorIs(t, tk.Validate(), ErrEmptyChecklistText)
	})
	t.Run("missing_id_rejected", func(t *testing.T) {
		ci := ChecklistItem{Text: "x"}
		tk := Task{Title: "x", Status: StatusOpen, Checklist: []ChecklistItem{ci}}
		require.ErrorIs(t, tk.Validate(), ErrInvalidChecklistID)
	})
	t.Run("duplicate_id_rejected", func(t *testing.T) {
		sameID := id.New()
		tk := Task{Title: "x", Status: StatusOpen, Checklist: []ChecklistItem{
			{ID: sameID, Text: "a"},
			{ID: sameID, Text: "b"},
		}}
		require.ErrorIs(t, tk.Validate(), ErrChecklistDuplicateID)
	})
}

func TestDate_JSONRoundTrip(t *testing.T) {
	d := NewDate(time.Date(2026, 5, 25, 14, 30, 0, 0, time.Local))
	data, err := d.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `"2026-05-25"`, string(data))

	var d2 Date
	require.NoError(t, d2.UnmarshalJSON(data))
	require.True(t, d.Equal(d2))
}

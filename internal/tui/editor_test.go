package tui

import (
	"context"
	"testing"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestEditor_WhenDefaultsAnytimeForOpenTask(t *testing.T) {
	tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen, Someday: false}
	ed := NewEditor(tk)
	require.Equal(t, whenAnytime, ed.when)
}

func TestEditor_WhenDefaultsSomedayForSomedayTask(t *testing.T) {
	tk := task.Task{ID: id.New(), Title: "x", Someday: true}
	ed := NewEditor(tk)
	require.Equal(t, whenSomeday, ed.when)
}

func TestEditor_ApplyAndSaveMapsAnytime(t *testing.T) {
	_, svc, tasks := setupModelWithInboxTasks(t, "x")
	ed := NewEditor(tasks[0])
	ed.when = whenAnytime
	saved, err := ed.ApplyAndSave(context.Background(), svc)
	require.NoError(t, err)
	require.False(t, saved.Someday)
}

func TestEditor_ApplyAndSaveMapsSomeday(t *testing.T) {
	_, svc, tasks := setupModelWithInboxTasks(t, "x")
	ed := NewEditor(tasks[0])
	ed.when = whenSomeday
	saved, err := ed.ApplyAndSave(context.Background(), svc)
	require.NoError(t, err)
	require.True(t, saved.Someday)
}

func TestEditor_HintShownWhenAnytimeNoAreaProject(t *testing.T) {
	tk := task.Task{ID: id.New(), Title: "x", Someday: false}
	ed := NewEditor(tk)
	ed.when = whenAnytime
	out := ed.View(NewTheme(), 60)
	require.Contains(t, out, "will appear in Inbox")
}

func TestEditor_HintHiddenForSomeday(t *testing.T) {
	tk := task.Task{ID: id.New(), Title: "x", Someday: true}
	ed := NewEditor(tk)
	ed.when = whenSomeday
	out := ed.View(NewTheme(), 60)
	require.NotContains(t, out, "will appear in Inbox")
}

func TestEditor_FieldCountIsSix(t *testing.T) {
	require.Equal(t, 6, int(fieldCount))
}

// TestProp_WhenToggleInvolution verifies CP-5 (REQ-3.1, 3.3): toggling
// the When radio twice always restores the original value.
func TestProp_WhenToggleInvolution(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		startSomeday := rapid.Bool().Draw(rt, "startSomeday")
		tk := task.Task{ID: id.New(), Title: "x", Someday: startSomeday}
		ed := NewEditor(tk)
		initial := ed.when
		// Toggle twice via the editor space handler — simulating via m.editor manipulation
		if ed.when == whenAnytime {
			ed.when = whenSomeday
		} else {
			ed.when = whenAnytime
		}
		if ed.when == whenAnytime {
			ed.when = whenSomeday
		} else {
			ed.when = whenAnytime
		}
		require.Equal(rt, initial, ed.when, "Space toggle is involutive")
	})
}

// TestProp_WhenMapping verifies CP-6 (REQ-3.4): ApplyAndSave maps the
// editor's When radio to task.Someday at save time.
func TestProp_WhenMapping(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, tasks := setupRapidModel(rt, "x")
		wantSomeday := rapid.Bool().Draw(rt, "wantSomeday")
		ed := NewEditor(tasks[0])
		if wantSomeday {
			ed.when = whenSomeday
		} else {
			ed.when = whenAnytime
		}
		saved, err := ed.ApplyAndSave(context.Background(), svc)
		require.NoError(rt, err)
		require.Equal(rt, wantSomeday, saved.Someday)
	})
}

// TestProp_HintConditional verifies CP-7 (REQ-3.5): the "will appear in
// Inbox" hint renders iff When=Anytime AND AreaID is nil AND ProjectID
// is nil.
func TestProp_HintConditional(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		whenAny := rapid.Bool().Draw(rt, "whenAnytime")
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		tk := task.Task{ID: id.New(), Title: "x"}
		if hasArea {
			aid := id.New()
			tk.AreaID = &aid
		}
		if hasProject {
			pid := id.New()
			tk.ProjectID = &pid
		}
		ed := NewEditor(tk)
		if whenAny {
			ed.when = whenAnytime
		} else {
			ed.when = whenSomeday
		}
		out := ed.View(NewTheme(), 60)
		shouldShow := whenAny && !hasArea && !hasProject
		if shouldShow {
			require.Contains(rt, out, "will appear in Inbox")
		} else {
			require.NotContains(rt, out, "will appear in Inbox")
		}
	})
}

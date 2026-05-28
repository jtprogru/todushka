package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

var (
	keyEnter = tea.KeyMsg{Type: tea.KeyEnter}
	keyEsc   = tea.KeyMsg{Type: tea.KeyEsc}
)

func TestWhenPicker_SelectToday(t *testing.T) {
	now := time.Now()
	p := newWhenPicker(nil, false, now)
	p.cursor = whenRowToday
	_, res := p.Update(keyEnter)
	require.Equal(t, whenSelected, res.outcome)
	require.NotNil(t, res.startDate)
	require.True(t, res.startDate.Equal(task.NewDate(now)))
	require.False(t, res.someday)
}

func TestWhenPicker_SelectSomeday(t *testing.T) {
	p := newWhenPicker(nil, false, time.Now())
	p.cursor = whenRowSomeday
	_, res := p.Update(keyEnter)
	require.Equal(t, whenSelected, res.outcome)
	require.Nil(t, res.startDate)
	require.True(t, res.someday)
}

func TestWhenPicker_SelectAnytime(t *testing.T) {
	d := task.NewDate(time.Now())
	p := newWhenPicker(&d, false, time.Now())
	p.cursor = whenRowAnytime
	_, res := p.Update(keyEnter)
	require.Equal(t, whenSelected, res.outcome)
	require.Nil(t, res.startDate)
	require.False(t, res.someday)
}

func TestWhenPicker_PickDateValid(t *testing.T) {
	p := newWhenPicker(nil, false, time.Now())
	p.cursor = whenRowPickDate
	p, res := p.Update(keyEnter)
	require.Equal(t, whenNone, res.outcome)
	require.True(t, p.entering)
	p.dateInput.SetValue("2030-01-02")
	_, res = p.Update(keyEnter)
	require.Equal(t, whenSelected, res.outcome)
	require.NotNil(t, res.startDate)
	require.Equal(t, "2030-01-02", res.startDate.Format("2006-01-02"))
	require.False(t, res.someday)
}

func TestWhenPicker_PickDateInvalid(t *testing.T) {
	p := newWhenPicker(nil, false, time.Now())
	p.cursor = whenRowPickDate
	p, _ = p.Update(keyEnter)
	p.dateInput.SetValue("nope")
	p, res := p.Update(keyEnter)
	require.Equal(t, whenNone, res.outcome)
	require.NotEmpty(t, p.err)
}

func TestWhenPicker_EscInDateReturnsToList(t *testing.T) {
	p := newWhenPicker(nil, false, time.Now())
	p.cursor = whenRowPickDate
	p, _ = p.Update(keyEnter)
	require.True(t, p.entering)
	p, res := p.Update(keyEsc)
	require.False(t, p.entering)
	require.Equal(t, whenNone, res.outcome)
}

func TestWhenPicker_EscInListCancels(t *testing.T) {
	p := newWhenPicker(nil, false, time.Now())
	_, res := p.Update(keyEsc)
	require.Equal(t, whenCancel, res.outcome)
}

func TestWhenPicker_InitialCursor(t *testing.T) {
	now := time.Now()
	today := task.NewDate(now)
	future := task.NewDate(now.AddDate(0, 0, 5))
	cases := []struct {
		start   *task.Date
		someday bool
		want    int
	}{
		{nil, false, whenRowAnytime},
		{nil, true, whenRowSomeday},
		{&today, false, whenRowToday},
		{&future, false, whenRowPickDate},
	}
	for _, c := range cases {
		require.Equal(t, c.want, newWhenPicker(c.start, c.someday, now).cursor)
	}
}

func TestWhenDisplay_Mapping(t *testing.T) {
	now := time.Now()
	today := task.NewDate(now)
	future := task.NewDate(now.AddDate(0, 0, 5))
	require.Equal(t, "Someday", whenDisplay(nil, true, now))
	require.Equal(t, "Anytime", whenDisplay(nil, false, now))
	require.Equal(t, "Today", whenDisplay(&today, false, now))
	require.Equal(t, future.Format("2006-01-02"), whenDisplay(&future, false, now))
}

// TestProp_WhenChoiceMapping — Property 1: each non-date row maps to the
// documented (startDate, someday).
func TestProp_WhenChoiceMapping(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		now := time.Now()
		row := rapid.SampledFrom([]int{whenRowToday, whenRowSomeday, whenRowAnytime}).Draw(rt, "row")
		p := newWhenPicker(nil, false, now)
		p.cursor = row
		_, res := p.Update(keyEnter)
		require.Equal(rt, whenSelected, res.outcome)
		switch row {
		case whenRowToday:
			require.True(rt, res.startDate != nil && res.startDate.Equal(task.NewDate(now)) && !res.someday)
		case whenRowSomeday:
			require.True(rt, res.startDate == nil && res.someday)
		case whenRowAnytime:
			require.True(rt, res.startDate == nil && !res.someday)
		}
	})
}

// TestProp_WhenExclusivity — Property 2: never someday AND a start date.
func TestProp_WhenExclusivity(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		now := time.Now()
		row := rapid.IntRange(0, 3).Draw(rt, "row")
		p := newWhenPicker(nil, false, now)
		p.cursor = row
		p, res := p.Update(keyEnter)
		if row == whenRowPickDate {
			off := rapid.IntRange(-30, 30).Draw(rt, "off")
			p.dateInput.SetValue(now.AddDate(0, 0, off).Format("2006-01-02"))
			_, res = p.Update(keyEnter)
		}
		if res.outcome == whenSelected {
			require.False(rt, res.someday && res.startDate != nil)
		}
	})
}

// TestProp_WhenInvalidDateInert — Property 3: a string that fails to parse
// leaves the picker open with an error and emits no selection.
func TestProp_WhenInvalidDateInert(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		s := rapid.String().Draw(rt, "s")
		p := newWhenPicker(nil, false, time.Now())
		p.cursor = whenRowPickDate
		p, _ = p.Update(keyEnter)
		p.dateInput.SetValue(s)
		stored := strings.TrimSpace(p.dateInput.Value())
		_, perr := time.ParseInLocation("2006-01-02", stored, time.Local)
		p2, res := p.Update(keyEnter)
		if perr != nil {
			require.Equal(rt, whenNone, res.outcome)
			require.NotEmpty(rt, p2.err)
		} else {
			require.Equal(rt, whenSelected, res.outcome)
		}
	})
}

// TestProp_WhenDisplayMapping — Property 5.
func TestProp_WhenDisplayMapping(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		now := time.Now()
		someday := rapid.Bool().Draw(rt, "someday")
		var sd *task.Date
		if rapid.Bool().Draw(rt, "hasDate") {
			d := task.NewDate(now.AddDate(0, 0, rapid.IntRange(-10, 10).Draw(rt, "off")))
			sd = &d
		}
		got := whenDisplay(sd, someday, now)
		switch {
		case someday:
			require.Equal(rt, "Someday", got)
		case sd == nil:
			require.Equal(rt, "Anytime", got)
		case sd.Equal(task.NewDate(now)):
			require.Equal(rt, "Today", got)
		default:
			require.Equal(rt, sd.Format("2006-01-02"), got)
		}
	})
}

// ─── editor integration ──────────────────────────────────────────────

func TestEditor_NoStartField(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ed := NewEditor(context.Background(), task.Task{ID: id.New(), Title: "x"}, svc)
	out := ed.View(NewTheme(), 80)
	require.NotContains(t, out, "Start")
	require.Contains(t, out, "When")
}

func TestEditor_EnterOnWhenOpensPicker(t *testing.T) {
	m, svc := newTestModelWithService(t)
	m.screen = screenEditor
	m.editor = NewEditor(context.Background(), task.Task{ID: id.New(), Title: "x"}, svc)
	m.editor.focus = fieldWhen
	mm, _ := m.Update(keyEnter)
	m = mm.(Model)
	require.NotNil(t, m.editor.whenPicker)
}

func TestEditor_ApplyAndSave_WhenState(t *testing.T) {
	_, svc, tasks := setupModelWithInboxTasks(t, "x")
	ed := NewEditor(context.Background(), tasks[0], svc)
	d := task.NewDate(time.Now())
	ed.startDate = &d
	ed.someday = false
	saved, err := ed.ApplyAndSave(context.Background(), svc)
	require.NoError(t, err)
	require.NotNil(t, saved.StartDate)
	require.True(t, saved.StartDate.Equal(task.NewDate(time.Now())))
	require.False(t, saved.Someday)
}

// TestProp_WhenOpenRoundTrip — Property 4: picker cursor and editor When
// label agree for any (startDate, someday) state.
func TestProp_WhenOpenRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		now := time.Now()
		someday := rapid.Bool().Draw(rt, "someday")
		var sd *task.Date
		if !someday && rapid.Bool().Draw(rt, "hasDate") {
			d := task.NewDate(now.AddDate(0, 0, rapid.IntRange(-5, 5).Draw(rt, "off")))
			sd = &d
		}
		wp := newWhenPicker(sd, someday, now)
		label := whenDisplay(sd, someday, now)
		switch wp.cursor {
		case whenRowToday:
			require.Equal(rt, "Today", label)
		case whenRowPickDate:
			require.Equal(rt, sd.Format("2006-01-02"), label)
		case whenRowSomeday:
			require.Equal(rt, "Someday", label)
		case whenRowAnytime:
			require.Equal(rt, "Anytime", label)
		}
	})
}

// TestProp_EditorNoStartInertCancel — Property 6: exactly 4 rows and Esc in
// list mode cancels without emitting a selection, for any state.
func TestProp_EditorNoStartInertCancel(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		now := time.Now()
		someday := rapid.Bool().Draw(rt, "someday")
		var sd *task.Date
		if !someday && rapid.Bool().Draw(rt, "hasDate") {
			d := task.NewDate(now.AddDate(0, 0, rapid.IntRange(-5, 5).Draw(rt, "off")))
			sd = &d
		}
		require.Len(rt, whenRows, 4)
		wp := newWhenPicker(sd, someday, now)
		_, res := wp.Update(keyEsc)
		require.Equal(rt, whenCancel, res.outcome)
	})
}

// TestProp_WhenSavePreserves — Property 7: ApplyAndSave writes startDate/
// someday and preserves other fields.
func TestProp_WhenSavePreserves(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc, tasks := setupRapidModel(rt, "x")
		now := time.Now()
		someday := rapid.Bool().Draw(rt, "someday")
		var sd *task.Date
		if !someday && rapid.Bool().Draw(rt, "hasDate") {
			d := task.NewDate(now.AddDate(0, 0, rapid.IntRange(-5, 5).Draw(rt, "off")))
			sd = &d
		}
		ed := NewEditor(context.Background(), tasks[0], svc)
		ed.title.SetValue("kept")
		ed.startDate = sd
		ed.someday = someday
		saved, err := ed.ApplyAndSave(context.Background(), svc)
		require.NoError(rt, err)
		require.Equal(rt, "kept", saved.Title)
		require.Equal(rt, someday, saved.Someday)
		if sd == nil {
			require.Nil(rt, saved.StartDate)
		} else {
			require.NotNil(rt, saved.StartDate)
			require.True(rt, saved.StartDate.Equal(*sd))
		}
	})
}

// TestProp_ReadOnlyBlocksWhenSave — Property 8: a read-only editor never
// writes, regardless of the chosen When state.
func TestProp_ReadOnlyBlocksWhenSave(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m, svc, tasks := setupRapidModel(rt, "x")
		m.readOnly = true
		m.editor = NewEditor(context.Background(), tasks[0], svc)
		m.editor.someday = rapid.Bool().Draw(rt, "someday")
		mm, cmd := m.saveEditor()
		m = mm.(Model)
		require.Nil(rt, cmd)
		require.NotEmpty(rt, m.editor.err)
		got, err := svc.Repo().TaskGet(context.Background(), tasks[0].ID)
		require.NoError(rt, err)
		require.Equal(rt, tasks[0].Someday, got.Someday)
	})
}

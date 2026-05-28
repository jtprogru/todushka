package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/domain/task"
)

// whenOutcome communicates the result of routing one key into the picker.
type whenOutcome int

const (
	whenNone     whenOutcome = iota // still open, keep routing
	whenCancel                      // Esc — close, no change
	whenSelected                    // close, apply (startDate, someday)
)

// whenResult is returned by whenPicker.Update.
type whenResult struct {
	outcome   whenOutcome
	startDate *task.Date // valid when whenSelected (nil for Someday/Anytime)
	someday   bool
}

// whenRows are the picker rows in display order.
var whenRows = []string{"Today", "Pick date…", "Someday", "Anytime"}

const (
	whenRowToday    = 0
	whenRowPickDate = 1
	whenRowSomeday  = 2
	whenRowAnytime  = 3
)

// whenPicker is a pure overlay for choosing a task's When state. Unlike
// areaPicker it needs no service (no DB). It maps a choice to
// (startDate, someday) per the design: Today→(today,false),
// date→(date,false), Someday→(nil,true), Anytime→(nil,false).
type whenPicker struct {
	cursor    int             // index into whenRows
	entering  bool            // true → date sub-mode
	dateInput textinput.Model // active while entering
	now       time.Time       // drives "today" and initial cursor
	err       string          // invalid-date message (keeps picker open)
}

// newWhenPicker positions the cursor on the row matching the current state.
func newWhenPicker(startDate *task.Date, someday bool, now time.Time) whenPicker {
	in := textinput.New()
	in.Placeholder = "YYYY-MM-DD"
	in.CharLimit = 10

	cursor := whenRowAnytime
	switch {
	case someday:
		cursor = whenRowSomeday
	case startDate != nil && startDate.Equal(task.NewDate(now)):
		cursor = whenRowToday
	case startDate != nil:
		cursor = whenRowPickDate
	}
	return whenPicker{cursor: cursor, dateInput: in, now: now}
}

// Update routes one key. Pure: navigation, cancel, and selection only.
func (p whenPicker) Update(msg tea.KeyMsg) (whenPicker, whenResult) {
	if p.entering {
		return p.updateEntering(msg)
	}
	switch {
	case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && string(msg.Runes) == "k"):
		if p.cursor > 0 {
			p.cursor--
		}
		return p, whenResult{outcome: whenNone}
	case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && string(msg.Runes) == "j"):
		if p.cursor < len(whenRows)-1 {
			p.cursor++
		}
		return p, whenResult{outcome: whenNone}
	case msg.Type == tea.KeyEsc:
		return p, whenResult{outcome: whenCancel}
	case msg.Type == tea.KeyEnter:
		return p.selectCurrent()
	}
	return p, whenResult{outcome: whenNone}
}

func (p whenPicker) selectCurrent() (whenPicker, whenResult) {
	switch p.cursor {
	case whenRowToday:
		d := task.NewDate(p.now)
		return p, whenResult{outcome: whenSelected, startDate: &d, someday: false}
	case whenRowPickDate:
		p.entering = true
		p.err = ""
		p.dateInput.SetValue("")
		p.dateInput.Focus()
		return p, whenResult{outcome: whenNone}
	case whenRowSomeday:
		return p, whenResult{outcome: whenSelected, startDate: nil, someday: true}
	default: // whenRowAnytime
		return p, whenResult{outcome: whenSelected, startDate: nil, someday: false}
	}
}

func (p whenPicker) updateEntering(msg tea.KeyMsg) (whenPicker, whenResult) {
	switch msg.Type {
	case tea.KeyEsc:
		p.entering = false
		p.err = ""
		return p, whenResult{outcome: whenNone}
	case tea.KeyEnter:
		s := strings.TrimSpace(p.dateInput.Value())
		d, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			p.err = "invalid date (YYYY-MM-DD)"
			return p, whenResult{outcome: whenNone}
		}
		nd := task.NewDate(d)
		return p, whenResult{outcome: whenSelected, startDate: &nd, someday: false}
	}
	var cmd tea.Cmd
	p.dateInput, cmd = p.dateInput.Update(msg)
	_ = cmd
	return p, whenResult{outcome: whenNone}
}

// whenDisplay renders the When label shown in the editor's When field.
func whenDisplay(startDate *task.Date, someday bool, now time.Time) string {
	switch {
	case someday:
		return "Someday"
	case startDate == nil:
		return "Anytime"
	case startDate.Equal(task.NewDate(now)):
		return "Today"
	default:
		return startDate.Format("2006-01-02")
	}
}

// View renders the picker as a modal.
func (p whenPicker) View(theme Theme, width int) string {
	if p.entering {
		body := lipgloss.JoinVertical(lipgloss.Left,
			theme.Title.Render("Pick date"),
			"",
			theme.Label.Render("Date"),
			theme.Field.Render(p.dateInput.View()),
			"",
			theme.Help.Render("Enter: set  Esc: back"),
		)
		if p.err != "" {
			body = lipgloss.JoinVertical(lipgloss.Left, body, theme.StatusError.Render(p.err))
		}
		return theme.Modal.Render(body)
	}

	lines := []string{theme.Title.Render("When"), ""}
	for i, label := range whenRows {
		if i == p.cursor {
			lines = append(lines, theme.Selected.Render("> "+label))
		} else {
			lines = append(lines, "  "+label)
		}
	}
	lines = append(lines, "", theme.Help.Render("↑/↓: move  Enter: select  Esc: cancel"))
	return theme.Modal.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

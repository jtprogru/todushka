package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/storage"
)

// pickerOutcome communicates the result of routing one key into the picker.
type pickerOutcome int

const (
	pickerNone     pickerOutcome = iota // still open, keep routing
	pickerCancel                        // Esc — close, no change
	pickerSelected                      // close, apply areaID/areaName
)

// pickerResult is returned by areaPicker.Update.
type pickerResult struct {
	outcome  pickerOutcome
	areaID   *id.ID // valid when outcome == pickerSelected (nil = no-area)
	areaName string // display name for the selection ("" for no-area)
}

const createRowLabel = "+ New area…"

// areaPicker is an overlay list for choosing an area, clearing it, or
// creating a new one inline. It is editor-agnostic and communicates the
// chosen value via pickerResult.
type areaPicker struct {
	areas       []area.Area     // from svc.ListAreas (display order preserved)
	cursor      int             // 0 = no-area; 1..N = areas; N+1 = create row
	noAreaLabel string          // "Inbox" (task) | "No area" (project)
	readOnly    bool            // true → no create row, AddArea never called
	creating    bool            // true → name-input mode
	nameInput   textinput.Model // active only while creating
	err         string          // inline-create error (keeps picker open)
}

// newAreaPicker builds a picker positioned on `current` (the no-area row
// when current == nil or not found). readOnly hides the create row.
func newAreaPicker(areas []area.Area, current *id.ID, noAreaLabel string, readOnly bool) areaPicker {
	in := textinput.New()
	in.Placeholder = "new area name"
	in.CharLimit = 100

	cursor := 0
	if current != nil {
		for i, a := range areas {
			if a.ID == *current {
				cursor = i + 1
				break
			}
		}
	}
	return areaPicker{
		areas:       areas,
		cursor:      cursor,
		noAreaLabel: noAreaLabel,
		readOnly:    readOnly,
		nameInput:   in,
	}
}

// rowLabels returns the rendered row labels in display order:
// no-area, each area, then the create row (only when !readOnly).
func (p areaPicker) rowLabels() []string {
	labels := make([]string, 0, len(p.areas)+2)
	labels = append(labels, p.noAreaLabel)
	for _, a := range p.areas {
		labels = append(labels, a.Name)
	}
	if !p.readOnly {
		labels = append(labels, createRowLabel)
	}
	return labels
}

func (p areaPicker) lastRowIndex() int {
	return len(p.rowLabels()) - 1
}

// isCreateRow reports whether index i is the create row.
func (p areaPicker) isCreateRow(i int) bool {
	return !p.readOnly && i == len(p.areas)+1
}

// Update routes one key. svc is used only for inline-create (AddArea /
// AreaFindByNormalized). Navigation and cancel keys are pure.
func (p areaPicker) Update(msg tea.KeyMsg, svc *app.Service) (areaPicker, pickerResult) {
	if p.creating {
		return p.updateCreating(msg, svc)
	}
	switch {
	case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && string(msg.Runes) == "k"):
		if p.cursor > 0 {
			p.cursor--
		}
		return p, pickerResult{outcome: pickerNone}
	case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && string(msg.Runes) == "j"):
		if p.cursor < p.lastRowIndex() {
			p.cursor++
		}
		return p, pickerResult{outcome: pickerNone}
	case msg.Type == tea.KeyEsc:
		return p, pickerResult{outcome: pickerCancel}
	case msg.Type == tea.KeyEnter:
		return p.selectCurrent()
	}
	return p, pickerResult{outcome: pickerNone}
}

// selectCurrent resolves the row under the cursor.
func (p areaPicker) selectCurrent() (areaPicker, pickerResult) {
	if p.cursor == 0 {
		return p, pickerResult{outcome: pickerSelected, areaID: nil, areaName: ""}
	}
	if p.isCreateRow(p.cursor) {
		p.creating = true
		p.err = ""
		p.nameInput.SetValue("")
		p.nameInput.Focus()
		return p, pickerResult{outcome: pickerNone}
	}
	a := p.areas[p.cursor-1]
	aid := a.ID
	return p, pickerResult{outcome: pickerSelected, areaID: &aid, areaName: a.Name}
}

// updateCreating handles keys while in name-input mode.
func (p areaPicker) updateCreating(msg tea.KeyMsg, svc *app.Service) (areaPicker, pickerResult) {
	switch msg.Type {
	case tea.KeyEsc:
		p.creating = false
		p.err = ""
		return p, pickerResult{outcome: pickerNone}
	case tea.KeyEnter:
		return p.confirmCreate(svc)
	}
	var cmd tea.Cmd
	p.nameInput, cmd = p.nameInput.Update(msg)
	_ = cmd
	return p, pickerResult{outcome: pickerNone}
}

// confirmCreate validates the typed name and creates/selects the area.
func (p areaPicker) confirmCreate(svc *app.Service) (areaPicker, pickerResult) {
	ctx := context.Background()
	name := strings.TrimSpace(p.nameInput.Value())
	if name == "" {
		p.err = "name required"
		return p, pickerResult{outcome: pickerNone}
	}
	a, err := svc.AddArea(ctx, name)
	if err == nil {
		aid := a.ID
		return p, pickerResult{outcome: pickerSelected, areaID: &aid, areaName: a.Name}
	}
	if errors.Is(err, storage.ErrAlreadyExists) {
		existing, lookupErr := svc.Repo().AreaFindByNormalized(ctx, area.Normalize(name))
		if lookupErr == nil {
			eid := existing.ID
			return p, pickerResult{outcome: pickerSelected, areaID: &eid, areaName: existing.Name}
		}
		p.err = lookupErr.Error()
		return p, pickerResult{outcome: pickerNone}
	}
	p.err = err.Error()
	return p, pickerResult{outcome: pickerNone}
}

// View renders the picker panel as a modal.
func (p areaPicker) View(theme Theme, width int) string {
	if p.creating {
		body := lipgloss.JoinVertical(lipgloss.Left,
			theme.Title.Render("New area"),
			"",
			theme.Label.Render("Name"),
			theme.Field.Render(p.nameInput.View()),
			"",
			theme.Help.Render("Enter: create  Esc: back"),
		)
		if p.err != "" {
			body = lipgloss.JoinVertical(lipgloss.Left, body, theme.StatusError.Render(p.err))
		}
		return theme.Modal.Render(body)
	}

	lines := []string{theme.Title.Render("Select area"), ""}
	for i, label := range p.rowLabels() {
		if i == p.cursor {
			lines = append(lines, theme.Selected.Render("> "+label))
		} else {
			lines = append(lines, "  "+label)
		}
	}
	lines = append(lines, "", theme.Help.Render("↑/↓: move  Enter: select  Esc: cancel"))
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if p.err != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, theme.StatusError.Render(p.err))
	}
	return theme.Modal.Render(body)
}

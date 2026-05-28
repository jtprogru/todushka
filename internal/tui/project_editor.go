package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
)

type projectEditorField int

const (
	pefName projectEditorField = iota
	pefNotes
	pefArea
	pefDeadline
	pefAutoClose
	pefCount
)

// ProjectEditorModel is the create/edit form for a Project. When original
// is nil → create flow; otherwise → edit flow on a copy of the original.
type ProjectEditorModel struct {
	original  *project.Project
	name      textinput.Model
	notes     textarea.Model
	deadline  textinput.Model
	autoClose bool
	focus     projectEditorField
	areaID    *id.ID      // selected area (nil = No area)
	areaName  string      // display name of selected area ("" = No area)
	picker    *areaPicker // non-nil while the area picker is open
	err       string
}

func newProjectEditor(create bool, p *project.Project, svc *app.Service) ProjectEditorModel {
	nameIn := textinput.New()
	nameIn.CharLimit = 200
	nameIn.Placeholder = "project name"
	nameIn.Focus()

	notesIn := textarea.New()
	notesIn.CharLimit = 4000
	notesIn.SetHeight(4)

	deadlineIn := textinput.New()
	deadlineIn.Placeholder = "YYYY-MM-DD"
	deadlineIn.CharLimit = 10

	m := ProjectEditorModel{
		name:     nameIn,
		notes:    notesIn,
		deadline: deadlineIn,
		focus:    pefName,
	}
	if !create && p != nil {
		cp := *p
		m.original = &cp
		m.name.SetValue(p.Name)
		m.notes.SetValue(p.Notes)
		m.autoClose = p.AutoClose
		if p.Deadline != nil {
			m.deadline.SetValue(p.Deadline.Format("2006-01-02"))
		}
		m.areaID = p.AreaID
		if p.AreaID != nil && svc != nil {
			if a, err := svc.Repo().AreaGet(context.Background(), *p.AreaID); err == nil {
				m.areaName = a.Name
			}
		}
	}
	return m
}

// View renders the editor as a modal panel.
func (m ProjectEditorModel) View(theme Theme, width int) string {
	if m.picker != nil {
		return m.picker.View(theme, width)
	}
	label := func(name string) string { return theme.Label.Render(name) }
	field := func(name, content string, focused bool) string {
		style := theme.Field
		if focused {
			style = theme.FieldFocus
		}
		if width > 6 {
			style = style.Width(width - 4)
		}
		return label(name) + "\n" + style.Render(content)
	}
	areaDisplay := m.areaName
	if areaDisplay == "" {
		areaDisplay = "No area"
	}
	title := "Edit project"
	if m.original == nil {
		title = "New project"
	}
	acBullet := "[ ]"
	if m.autoClose {
		acBullet = "[•]"
	}
	var acSection string
	if m.focus == pefAutoClose {
		acSection = theme.Selected.Render("▶ Auto-close") + "\n" + theme.Selected.Render(acBullet+" when all tasks done")
	} else {
		acSection = theme.Dim.Render("  Auto-close") + "\n" + theme.Dim.Render(acBullet+" when all tasks done")
	}
	body := lipgloss.JoinVertical(lipgloss.Left,
		theme.Title.Render(title),
		"",
		field("Name", m.name.View(), m.focus == pefName),
		field("Notes", m.notes.View(), m.focus == pefNotes),
		field("Area", areaDisplay, m.focus == pefArea),
		field("Deadline", m.deadline.View(), m.focus == pefDeadline),
		acSection,
		"",
		theme.Help.Render("Tab/Shift+Tab: field  Ctrl+S: save  Esc: cancel  Space: toggle Auto-close"),
	)
	if m.err != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, theme.StatusError.Render(m.err))
	}
	return theme.Modal.Render(body)
}

func (m *ProjectEditorModel) focusCurrent() tea.Cmd {
	m.name.Blur()
	m.notes.Blur()
	m.deadline.Blur()
	switch m.focus {
	case pefName:
		return m.name.Focus()
	case pefNotes:
		return m.notes.Focus()
	case pefDeadline:
		return m.deadline.Focus()
	}
	return nil
}

func (m ProjectEditorModel) nextField() ProjectEditorModel {
	m.focus = projectEditorField((int(m.focus) + 1) % int(pefCount))
	return m
}

func (m ProjectEditorModel) prevField() ProjectEditorModel {
	m.focus = projectEditorField((int(m.focus) - 1 + int(pefCount)) % int(pefCount))
	return m
}

// UpdateForm dispatches msg to the focused sub-widget.
func (m ProjectEditorModel) UpdateForm(msg tea.Msg) (ProjectEditorModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case pefName:
		m.name, cmd = m.name.Update(msg)
	case pefNotes:
		m.notes, cmd = m.notes.Update(msg)
	case pefDeadline:
		m.deadline, cmd = m.deadline.Update(msg)
	}
	return m, cmd
}

// ApplyAndSave validates the form, applies changes via the service, and
// returns the resulting project plus a "created" flag. Validation errors
// keep the modal open (caller checks the returned error).
func (m ProjectEditorModel) ApplyAndSave(ctx context.Context, svc *app.Service) (project.Project, bool, error) {
	name := strings.TrimSpace(m.name.Value())
	if name == "" {
		return project.Project{}, false, errors.New("name required")
	}
	notes := strings.TrimSpace(m.notes.Value())

	// Area is selected via the picker (REQ-6.1/6.2) — no name resolution.
	areaIDPtr := m.areaID

	var deadlinePtr *task.Date
	if s := strings.TrimSpace(m.deadline.Value()); s != "" {
		d, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			return project.Project{}, false, fmt.Errorf("deadline: %w", err)
		}
		dd := task.NewDate(d)
		deadlinePtr = &dd
	}

	if m.original == nil {
		in := app.AddProjectInput{
			Name:      name,
			Notes:     notes,
			AreaID:    areaIDPtr,
			Deadline:  deadlinePtr,
			AutoClose: m.autoClose,
		}
		p, err := svc.AddProject(ctx, in)
		if err != nil {
			return project.Project{}, false, err
		}
		return p, true, nil
	}

	// Edit
	p := *m.original
	p.Name = name
	p.Notes = notes
	p.Deadline = deadlinePtr
	p.AutoClose = m.autoClose
	p.AreaID = areaIDPtr
	if err := svc.EditProject(ctx, p); err != nil {
		return project.Project{}, false, err
	}
	return p, false, nil
}

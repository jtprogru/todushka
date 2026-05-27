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
	"github.com/jtprogru/todushka/internal/storage"
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
	area      textinput.Model
	deadline  textinput.Model
	autoClose bool
	focus     projectEditorField
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

	areaIn := textinput.New()
	areaIn.Placeholder = "area name (empty = none)"
	areaIn.CharLimit = 100

	deadlineIn := textinput.New()
	deadlineIn.Placeholder = "YYYY-MM-DD"
	deadlineIn.CharLimit = 10

	m := ProjectEditorModel{
		name:     nameIn,
		notes:    notesIn,
		area:     areaIn,
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
		if p.AreaID != nil && svc != nil {
			if a, err := svc.Repo().AreaGet(context.Background(), *p.AreaID); err == nil {
				m.area.SetValue(a.Name)
			}
		}
	}
	return m
}

// View renders the editor as a modal panel.
func (m ProjectEditorModel) View(theme Theme, width int) string {
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
		field("Area", m.area.View(), m.focus == pefArea),
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
	m.area.Blur()
	m.deadline.Blur()
	switch m.focus {
	case pefName:
		return m.name.Focus()
	case pefNotes:
		return m.notes.Focus()
	case pefArea:
		return m.area.Focus()
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
	case pefArea:
		m.area, cmd = m.area.Update(msg)
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

	var areaIDPtr *id.ID
	areaName := strings.TrimSpace(m.area.Value())
	if areaName != "" {
		a, err := svc.Repo().AreaFindByNormalized(ctx, areaName)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return project.Project{}, false, fmt.Errorf("area %q not found", areaName)
			}
			return project.Project{}, false, fmt.Errorf("area %q: %w", areaName, err)
		}
		aid := a.ID
		areaIDPtr = &aid
	}

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

package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
)

type editorField int

const (
	fieldTitle editorField = iota
	fieldNotes
	fieldStart
	fieldDeadline
	fieldTags
	fieldSomeday
	fieldCount
)

// EditorModel is a form that edits a single Task. Tab/Shift+Tab cycle fields;
// Ctrl+S saves (or Enter while on a single-line field); Esc cancels.
type EditorModel struct {
	original task.Task

	title    textinput.Model
	notes    textarea.Model
	start    textinput.Model
	deadline textinput.Model
	tags     textinput.Model
	someday  bool
	focus    editorField

	err string
}

func NewEditor(t task.Task) EditorModel {
	titleIn := textinput.New()
	titleIn.SetValue(t.Title)
	titleIn.CharLimit = 200
	titleIn.Focus()

	notesIn := textarea.New()
	notesIn.SetValue(t.Notes)
	notesIn.CharLimit = 4000
	notesIn.SetHeight(4)

	startIn := textinput.New()
	startIn.Placeholder = "YYYY-MM-DD"
	startIn.CharLimit = 10
	if t.StartDate != nil {
		startIn.SetValue(t.StartDate.Format("2006-01-02"))
	}

	dlIn := textinput.New()
	dlIn.Placeholder = "YYYY-MM-DD"
	dlIn.CharLimit = 10
	if t.Deadline != nil {
		dlIn.SetValue(t.Deadline.Format("2006-01-02"))
	}

	tagsIn := textinput.New()
	tagsIn.Placeholder = "comma-separated names"
	tagsIn.CharLimit = 500
	// We don't have tag names available here without a DB roundtrip; the
	// model receives an external `tagNames` value via SetTagNames if needed.

	return EditorModel{
		original: t,
		title:    titleIn,
		notes:    notesIn,
		start:    startIn,
		deadline: dlIn,
		tags:     tagsIn,
		someday:  t.Someday,
		focus:    fieldTitle,
	}
}

// SetTagNames pre-fills the tags input with comma-separated names of existing tags.
func (m *EditorModel) SetTagNames(names []string) {
	m.tags.SetValue(strings.Join(names, ", "))
}

func (m *EditorModel) focusCurrent() tea.Cmd {
	m.title.Blur()
	m.notes.Blur()
	m.start.Blur()
	m.deadline.Blur()
	m.tags.Blur()
	switch m.focus {
	case fieldTitle:
		return m.title.Focus()
	case fieldNotes:
		return m.notes.Focus()
	case fieldStart:
		return m.start.Focus()
	case fieldDeadline:
		return m.deadline.Focus()
	case fieldTags:
		return m.tags.Focus()
	}
	return nil
}

func (m EditorModel) nextField() EditorModel {
	m.focus = editorField((int(m.focus) + 1) % int(fieldCount))
	return m
}

func (m EditorModel) prevField() EditorModel {
	m.focus = editorField((int(m.focus) - 1 + int(fieldCount)) % int(fieldCount))
	return m
}

// UpdateForm dispatches key/msg to whichever sub-component is focused.
// Returns the new model and any tea.Cmd produced by the widget.
func (m EditorModel) UpdateForm(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case fieldTitle:
		m.title, cmd = m.title.Update(msg)
	case fieldNotes:
		m.notes, cmd = m.notes.Update(msg)
	case fieldStart:
		m.start, cmd = m.start.Update(msg)
	case fieldDeadline:
		m.deadline, cmd = m.deadline.Update(msg)
	case fieldTags:
		m.tags, cmd = m.tags.Update(msg)
	}
	return m, cmd
}

// ApplyAndSave validates the form, applies it to the underlying task, and
// persists via the service. Returns the updated task on success.
func (m EditorModel) ApplyAndSave(ctx context.Context, svc *app.Service) (task.Task, error) {
	t := m.original
	t.Title = strings.TrimSpace(m.title.Value())
	t.Notes = strings.TrimSpace(m.notes.Value())
	t.Someday = m.someday

	if s := strings.TrimSpace(m.start.Value()); s == "" {
		t.StartDate = nil
	} else {
		d, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			return task.Task{}, fmt.Errorf("start: %w", err)
		}
		nd := task.NewDate(d)
		t.StartDate = &nd
	}
	if s := strings.TrimSpace(m.deadline.Value()); s == "" {
		t.Deadline = nil
	} else {
		d, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			return task.Task{}, fmt.Errorf("deadline: %w", err)
		}
		nd := task.NewDate(d)
		t.Deadline = &nd
	}

	tagNames := splitTags(m.tags.Value())
	tagIDs := make([]id.ID, 0, len(tagNames))
	for _, name := range tagNames {
		tg, err := svc.UpsertTag(ctx, name)
		if err != nil {
			return task.Task{}, fmt.Errorf("tag %q: %w", name, err)
		}
		tagIDs = append(tagIDs, tg.ID)
	}
	t.Tags = tagIDs

	if err := svc.EditTask(ctx, t); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

func splitTags(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// View renders the editor form for the given theme and width.
func (m EditorModel) View(theme Theme, width int) string {
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

	someday := "[ ] Someday"
	if m.someday {
		someday = "[x] Someday"
	}
	if m.focus == fieldSomeday {
		someday = theme.Selected.Render("▶ " + someday)
	} else {
		someday = theme.Dim.Render("  " + someday)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		theme.Title.Render("Edit task"),
		"",
		field("Title", m.title.View(), m.focus == fieldTitle),
		field("Notes", m.notes.View(), m.focus == fieldNotes),
		field("Start", m.start.View(), m.focus == fieldStart),
		field("Deadline", m.deadline.View(), m.focus == fieldDeadline),
		field("Tags", m.tags.View(), m.focus == fieldTags),
		someday,
		"",
		theme.Help.Render("Tab/Shift+Tab: field  Ctrl+S: save  Esc: cancel  Space: toggle Someday"),
	)
	if m.err != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, theme.StatusError.Render(m.err))
	}
	return theme.Modal.Render(body)
}

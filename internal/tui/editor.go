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
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
)

type editorField int

const (
	fieldTitle editorField = iota
	fieldNotes
	fieldStart
	fieldDeadline
	fieldArea    // [NEW]
	fieldProject // [NEW]
	fieldHeading // [NEW]
	fieldTags
	fieldWhen
	fieldCount // = 9
)

type shellEditorWhen int

const (
	whenAnytime shellEditorWhen = iota
	whenSomeday
)

// EditorModel is a form that edits a single Task. Tab/Shift+Tab cycle fields;
// Ctrl+S saves (or Enter while on a single-line field); Esc cancels.
type EditorModel struct {
	original task.Task

	title    textinput.Model
	notes    textarea.Model
	start    textinput.Model
	deadline textinput.Model
	area     textinput.Model // [NEW]
	project  textinput.Model // [NEW]
	heading  textinput.Model // [NEW]
	tags     textinput.Model
	when     shellEditorWhen
	focus    editorField

	err string
}

func NewEditor(ctx context.Context, t task.Task, svc *app.Service) EditorModel {
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

	when := whenAnytime
	if t.Someday {
		when = whenSomeday
	}

	areaIn := textinput.New()
	areaIn.Placeholder = "area name (empty = Inbox)"
	areaIn.CharLimit = 100
	if t.AreaID != nil {
		if a, err := svc.Repo().AreaGet(ctx, *t.AreaID); err == nil {
			areaIn.SetValue(a.Name)
		}
	}

	projectIn := textinput.New()
	projectIn.Placeholder = "project name"
	projectIn.CharLimit = 100
	if t.ProjectID != nil {
		if p, err := svc.Repo().ProjectGet(ctx, *t.ProjectID); err == nil {
			projectIn.SetValue(p.Name)
		}
	}

	headingIn := textinput.New()
	headingIn.Placeholder = "heading name (requires project)"
	headingIn.CharLimit = 100
	if t.HeadingID != nil && t.ProjectID != nil {
		if hs, err := svc.Repo().HeadingList(ctx, *t.ProjectID); err == nil {
			for _, h := range hs {
				if h.ID == *t.HeadingID {
					headingIn.SetValue(h.Name)
					break
				}
			}
		}
	}

	return EditorModel{
		original: t,
		title:    titleIn,
		notes:    notesIn,
		start:    startIn,
		deadline: dlIn,
		area:     areaIn,
		project:  projectIn,
		heading:  headingIn,
		tags:     tagsIn,
		when:     when,
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
	m.area.Blur()
	m.project.Blur()
	m.heading.Blur()
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
	case fieldArea:
		return m.area.Focus()
	case fieldProject:
		return m.project.Focus()
	case fieldHeading:
		return m.heading.Focus()
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
	case fieldArea:
		m.area, cmd = m.area.Update(msg)
	case fieldProject:
		m.project, cmd = m.project.Update(msg)
	case fieldHeading:
		m.heading, cmd = m.heading.Update(msg)
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
	t.Someday = m.when == whenSomeday

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

	// Resolve Area
	areaName := strings.TrimSpace(m.area.Value())
	if areaName == "" {
		t.AreaID = nil
	} else {
		a, err := svc.Repo().AreaFindByNormalized(ctx, areaName)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return task.Task{}, fmt.Errorf("area %q not found", areaName)
			}
			return task.Task{}, fmt.Errorf("area %q: %w", areaName, err)
		}
		t.AreaID = &a.ID
	}

	// Resolve Project
	projectName := strings.TrimSpace(m.project.Value())
	if projectName == "" {
		t.ProjectID = nil
		t.HeadingID = nil
	} else {
		matches, err := svc.Repo().ProjectFindByName(ctx, projectName)
		if err != nil {
			return task.Task{}, fmt.Errorf("project %q: %w", projectName, err)
		}
		if len(matches) == 0 {
			return task.Task{}, fmt.Errorf("project %q not found", projectName)
		}
		if len(matches) > 1 {
			return task.Task{}, fmt.Errorf("project %q is ambiguous (%d matches), use CLI to disambiguate", projectName, len(matches))
		}
		newPID := matches[0].ID
		// Detect project change → auto-clear heading (ADR-5)
		if m.original.ProjectID == nil || newPID != *m.original.ProjectID {
			t.HeadingID = nil
		}
		t.ProjectID = &newPID
	}

	// Resolve Heading
	headingName := strings.TrimSpace(m.heading.Value())
	if headingName == "" {
		t.HeadingID = nil
	} else {
		if t.ProjectID == nil {
			return task.Task{}, fmt.Errorf("heading %q requires a project", headingName)
		}
		headings, err := svc.Repo().HeadingList(ctx, *t.ProjectID)
		if err != nil {
			return task.Task{}, fmt.Errorf("heading %q: %w", headingName, err)
		}
		var found bool
		for _, h := range headings {
			if strings.EqualFold(h.Name, headingName) {
				t.HeadingID = &h.ID
				found = true
				break
			}
		}
		if !found {
			return task.Task{}, fmt.Errorf("heading %q not found in project", headingName)
		}
	}

	if err := svc.EditTask(ctx, t); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

// whenLabel returns the "When" section's primary label based on whether the
// task is currently routed to Inbox (no Area/Project) or Anytime (has either).
//
//   - "Inbox"   when t.AreaID == nil AND t.ProjectID == nil
//   - "Anytime" otherwise
//
// Pure function; called from View() to make the When label honest about the
// task's actual bucket.
func whenLabel(t task.Task) string {
	if t.AreaID == nil && t.ProjectID == nil {
		return "Inbox"
	}
	return "Anytime"
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

	anytimeBullet := "[ ]"
	somedayBullet := "[ ]"
	if m.when == whenAnytime {
		anytimeBullet = "[•]"
	} else {
		somedayBullet = "[•]"
	}
	primaryLabel := whenLabel(m.original)
	whenBody := fmt.Sprintf("%s %s\n%s Someday", anytimeBullet, primaryLabel, somedayBullet)
	var whenSection string
	if m.focus == fieldWhen {
		whenSection = theme.Selected.Render("▶ When") + "\n" + theme.Selected.Render(whenBody)
	} else {
		whenSection = theme.Dim.Render("  When") + "\n" + theme.Dim.Render(whenBody)
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		theme.Title.Render("Edit task"),
		"",
		field("Title", m.title.View(), m.focus == fieldTitle),
		field("Notes", m.notes.View(), m.focus == fieldNotes),
		field("Start", m.start.View(), m.focus == fieldStart),
		field("Deadline", m.deadline.View(), m.focus == fieldDeadline),
		field("Area", m.area.View(), m.focus == fieldArea),
		field("Project", m.project.View(), m.focus == fieldProject),
		field("Heading", m.heading.View(), m.focus == fieldHeading),
		field("Tags", m.tags.View(), m.focus == fieldTags),
		whenSection,
		"",
		theme.Help.Render("Tab/Shift+Tab: field  Ctrl+S: save  Esc: cancel  Space: toggle When"),
	)
	if m.err != "" {
		body = lipgloss.JoinVertical(lipgloss.Left, body, theme.StatusError.Render(m.err))
	}
	return theme.Modal.Render(body)
}

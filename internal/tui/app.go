// Package tui implements the Bubble Tea TUI. Update is pure: it returns the
// next Model and an optional Cmd; I/O happens inside Cmd closures that hit
// the application service.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
)

const statusFadeDuration = 5 * time.Second

type Model struct {
	service *app.Service
	keys    KeyMap
	theme   Theme

	screen      screenKind
	activeList  listKind
	tasks       []task.Task
	cursor      int
	statusMsg   string
	statusUntil time.Time
	quickInput  textinput.Model
	editor      EditorModel
	width       int
	selected    map[id.ID]struct{}
	confirm     *confirmState
	filterQuery string
	filtering   bool
	readOnly    bool

	tagNamesByID     map[id.ID]string
	areaNamesByID    map[id.ID]string
	projectsByID     map[id.ID]project.Project
	headingNamesByID map[id.ID]string

	height     int
	config     config.AppConfig
	listCounts map[listKind]int

	// Projects screen state (BL-5)
	projects            []project.Project
	projectCounts       map[id.ID][2]int
	projectCursor       int
	activeProjectID     *id.ID
	projectStatusFilter projectStatusFilterMode
	projectTasks        []task.Task
	projectEditor       ProjectEditorModel
	editingProject      bool
}

// allLists is the canonical order used by Tab/Shift+Tab cycling and the header.
var allLists = []listKind{listInbox, listToday, listUpcoming, listAnytime, listSomeday, listLogbook}

// NewModel constructs the root TUI model.
func NewModel(svc *app.Service, theme Theme, cfg config.AppConfig) Model {
	ti := textinput.New()
	ti.Placeholder = "what to do? — tokens: #tag @today @project !YYYY-MM-DD"
	ti.CharLimit = 256
	var ro bool
	if svc != nil {
		ro = svc.Repo().ReadOnly()
	}
	return Model{
		service:          svc,
		keys:             DefaultKeyMap(),
		theme:            theme,
		screen:           screenList,
		activeList:       listToday,
		quickInput:       ti,
		selected:         make(map[id.ID]struct{}),
		tagNamesByID:     make(map[id.ID]string),
		areaNamesByID:    make(map[id.ID]string),
		projectsByID:     make(map[id.ID]project.Project),
		headingNamesByID: make(map[id.ID]string),
		config:           cfg,
		listCounts:       make(map[listKind]int),
		readOnly:         ro,
		projectCounts:    make(map[id.ID][2]int),
	}
}

func (m Model) Init() tea.Cmd { return m.loadCurrentList() }

// blockWriteIfReadOnly mutates m to surface the RO status message and
// returns true when the model is in read-only mode. Callers must use a
// pointer receiver method or inline the equivalent for value-receiver
// flows (see dispatch in bulk.go).
func (m *Model) blockWriteIfReadOnly() bool {
	if !m.readOnly {
		return false
	}
	m.statusMsg = "read-only mode: writes disabled"
	m.statusUntil = time.Now().Add(statusFadeDuration)
	return true
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = max(0, len(m.tasks)-1)
		}
		return m, tea.Batch(
			fetchNameCache(m.service, m.tasks),
			fetchListCounts(m.service),
		)

	case countsLoadedMsg:
		m.listCounts = msg.counts
		return m, nil

	case projectsLoadedMsg:
		m.projects = msg.projects
		m.projectCounts = msg.counts
		// Refresh area name cache for any new areas referenced by projects.
		areaIDs := make([]id.ID, 0)
		for _, p := range msg.projects {
			if p.AreaID != nil {
				areaIDs = append(areaIDs, *p.AreaID)
			}
		}
		if m.projectCursor >= len(displayedProjects(m)) {
			m.projectCursor = max(0, len(displayedProjects(m))-1)
		}
		return m, fetchAreaNames(m.service, areaIDs)

	case projectTasksLoadedMsg:
		if m.activeProjectID == nil || msg.projectID != *m.activeProjectID {
			return m, nil
		}
		m.projectTasks = msg.tasks
		// Mirror into m.tasks so existing helpers (selectedTask,
		// completeSelected, displayedTasks, dispatch, etc.) operate on
		// the zoom view without per-helper screen checks.
		m.tasks = msg.tasks
		if m.cursor >= len(msg.tasks) {
			m.cursor = max(0, len(msg.tasks)-1)
		}
		return m, fetchNameCache(m.service, msg.tasks)

	case projectSavedMsg:
		m.editingProject = false
		m.projectEditor = ProjectEditorModel{}
		return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)

	case projectEditorErrMsg:
		m.projectEditor.err = msg.err
		return m, nil

	case projectDeletedMsg:
		return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)

	case nameCacheLoadedMsg:
		for k, v := range msg.tags {
			m.tagNamesByID[k] = v
		}
		for k, v := range msg.areas {
			m.areaNamesByID[k] = v
		}
		for k, v := range msg.projects {
			m.projectsByID[k] = v
		}
		for k, v := range msg.headings {
			m.headingNamesByID[k] = v
		}
		return m, nil

	case bulkResultMsg:
		if msg.fatal {
			m.statusMsg = msg.lastErr.Error()
			m.statusUntil = time.Now().Add(statusFadeDuration)
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		m.selected = make(map[id.ID]struct{})
		total := msg.succeeded + msg.failed
		if msg.failed > 0 {
			m.statusMsg = fmt.Sprintf("%s: %d/%d succeeded, %d failed", msg.action.label(), msg.succeeded, total, msg.failed)
		} else {
			m.statusMsg = fmt.Sprintf("%s: %d done", msg.action.label(), msg.succeeded)
		}
		m.statusUntil = time.Now().Add(statusFadeDuration)
		return m, tea.Batch(
			m.reloadDisplayedTasks(),
			tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} }),
		)

	case singleActionDoneMsg:
		if msg.err != nil {
			m.statusMsg = msg.err.Error()
			m.statusUntil = time.Now().Add(statusFadeDuration)
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		return m, tea.Batch(m.reloadDisplayedTasks(), fetchListCounts(m.service))

	case errorMsg:
		m.statusMsg = msg.err.Error()
		m.statusUntil = time.Now().Add(statusFadeDuration)
		return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })

	case clearStatusMsg:
		if time.Now().After(m.statusUntil) {
			m.statusMsg = ""
		}
		return m, nil

	case quickEntrySubmittedMsg:
		raw := msg.raw
		svc := m.service
		return m, func() tea.Msg {
			if _, err := svc.QuickEntry(context.Background(), raw); err != nil {
				return errorMsg{err}
			}
			return quickEntryDoneMsg{}
		}

	case quickEntryDoneMsg:
		return m, tea.Batch(m.loadCurrentList(), fetchListCounts(m.service))

	case editorSavedMsg:
		// Restore origin screen: zoom mode keeps us in screenProjectTasks
		// so the user does not get bounced back to GTD lists after editing.
		if m.activeProjectID != nil {
			m.screen = screenProjectTasks
		} else {
			m.screen = screenList
		}
		// Inline splice for immediate visual update (REQ-2.1, 2.3, 2.4).
		for i := range m.tasks {
			if m.tasks[i].ID == msg.updated.ID {
				m.tasks[i] = msg.updated
				break
			}
		}
		// Async refresh handles sort order, header counts, and name cache.
		return m, tea.Batch(
			m.reloadDisplayedTasks(),
			fetchListCounts(m.service),
		)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirm != nil {
		return handleConfirmKey(m, msg)
	}
	if m.filtering {
		return handleFilterKey(m, msg)
	}
	switch m.screen {
	case screenQuickEntry:
		return m.handleQuickEntryKey(msg)
	case screenEditor:
		return m.handleEditorKey(msg)
	case screenProjects:
		return m.handleProjectsKey(msg)
	case screenProjectTasks:
		return m.handleProjectTasksKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.ClearSelection) && m.screen == screenList && len(m.selected) > 0:
		m.selected = make(map[id.ID]struct{})
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		if m.screen == screenHelp {
			m.screen = screenList
		} else {
			m.screen = screenHelp
		}
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		m.filtering = true
		m.filterQuery = ""
		return m, nil
	case key.Matches(msg, m.keys.Projects):
		m.screen = screenProjects
		m.projectCursor = 0
		m.filterQuery = ""
		m.filtering = false
		return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)
	case key.Matches(msg, m.keys.NextView):
		return m.switchList(nextList(m.activeList, +1))
	case key.Matches(msg, m.keys.PrevView):
		return m.switchList(nextList(m.activeList, -1))
	case key.Matches(msg, m.keys.Inbox):
		return m.switchList(listInbox)
	case key.Matches(msg, m.keys.Today):
		return m.switchList(listToday)
	case key.Matches(msg, m.keys.Upcoming):
		return m.switchList(listUpcoming)
	case key.Matches(msg, m.keys.Anytime):
		return m.switchList(listAnytime)
	case key.Matches(msg, m.keys.Someday):
		return m.switchList(listSomeday)
	case key.Matches(msg, m.keys.Logbook):
		return m.switchList(listLogbook)
	case key.Matches(msg, m.keys.ToggleSelect):
		sel := m.selectedTask()
		if sel != nil {
			if _, ok := m.selected[sel.ID]; ok {
				delete(m.selected, sel.ID)
			} else {
				m.selected[sel.ID] = struct{}{}
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.SelectAll):
		for _, t := range displayedTasks(m) {
			m.selected[t.ID] = struct{}{}
		}
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.PageUp):
		m.cursor = clampCursor(m.cursor-pageStep(m), len(m.tasks))
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m.cursor = clampCursor(m.cursor+pageStep(m), len(m.tasks))
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		return m.openEditor()
	case key.Matches(msg, m.keys.QuickEntry):
		if m.blockWriteIfReadOnly() {
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		m.screen = screenQuickEntry
		m.quickInput.SetValue("")
		m.quickInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Complete):
		return dispatch(m, bulkActionComplete)
	case key.Matches(msg, m.keys.Cancel):
		return dispatch(m, bulkActionCancel)
	case key.Matches(msg, m.keys.Delete):
		return dispatch(m, bulkActionDelete)
	case key.Matches(msg, m.keys.PinToday):
		return dispatch(m, bulkActionPin)
	case key.Matches(msg, m.keys.Refresh):
		return m, m.loadCurrentList()
	}
	return m, nil
}

func nextList(cur listKind, step int) listKind {
	for i, l := range allLists {
		if l == cur {
			idx := (i + step + len(allLists)) % len(allLists)
			return allLists[idx]
		}
	}
	return cur
}

func (m Model) openEditor() (tea.Model, tea.Cmd) {
	sel := m.selectedTask()
	if sel == nil {
		return m, nil
	}
	m.editor = NewEditor(context.Background(), *sel, m.service)
	tagNames, err := m.lookupTagNames(sel.Tags)
	if err == nil {
		m.editor.SetTagNames(tagNames)
	}
	m.screen = screenEditor
	return m, m.editor.focusCurrent()
}

func (m Model) lookupTagNames(ids []id.ID) ([]string, error) {
	names := make([]string, 0, len(ids))
	for _, tid := range ids {
		t, err := m.service.Repo().TagGet(context.Background(), tid)
		if err != nil {
			return nil, err
		}
		names = append(names, t.Name)
	}
	return names, nil
}

func (m Model) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editor.picker != nil {
		np, res := m.editor.picker.Update(msg, m.service)
		m.editor.picker = &np
		switch res.outcome {
		case pickerCancel:
			m.editor.picker = nil
		case pickerSelected:
			m.editor.areaID = res.areaID
			m.editor.areaName = res.areaName
			m.editor.picker = nil
		}
		return m, nil
	}
	switch {
	case key.Matches(msg, m.keys.CloseModal):
		if m.activeProjectID != nil {
			m.screen = screenProjectTasks
		} else {
			m.screen = screenList
		}
		return m, nil
	case key.Matches(msg, m.keys.Save):
		return m.saveEditor()
	case key.Matches(msg, m.keys.NextField):
		m.editor = m.editor.nextField()
		return m, m.editor.focusCurrent()
	case key.Matches(msg, m.keys.PrevField):
		m.editor = m.editor.prevField()
		return m, m.editor.focusCurrent()
	case m.editor.focus == fieldArea && msg.Type == tea.KeyEnter:
		areas, err := m.service.ListAreas(context.Background())
		if err != nil {
			m.editor.err = err.Error()
			return m, nil
		}
		p := newAreaPicker(areas, m.editor.areaID, "Inbox", m.readOnly)
		m.editor.picker = &p
		return m, nil
	case m.editor.focus == fieldWhen && msg.Type == tea.KeySpace:
		if m.editor.when == whenAnytime {
			m.editor.when = whenSomeday
		} else {
			m.editor.when = whenAnytime
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.UpdateForm(msg)
	return m, cmd
}

func (m Model) saveEditor() (tea.Model, tea.Cmd) {
	if m.readOnly {
		m.editor.err = "read-only mode: writes disabled"
		return m, nil
	}
	svc := m.service
	ed := m.editor
	return m, func() tea.Msg {
		t, err := ed.ApplyAndSave(context.Background(), svc)
		if err != nil {
			return errorMsg{err}
		}
		return editorSavedMsg{updated: t}
	}
}

// handleProjectsKey dispatches key events in screenProjects. GTD keys
// (1..6, Tab/Shift+Tab) and task bulk-actions are intentionally ignored
// here — they are not part of the projects screen contract (REQ-1.4).
func (m Model) handleProjectsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingProject {
		return m.handleProjectEditorKey(msg)
	}
	if m.confirm != nil {
		return handleConfirmKey(m, msg)
	}
	if m.filtering {
		nm, cmd := handleFilterKey(m, msg)
		// Clamp cursor when filter shrinks the visible list.
		if nm.projectCursor >= len(displayedProjects(nm)) {
			nm.projectCursor = max(0, len(displayedProjects(nm))-1)
		}
		return nm, cmd
	}
	switch {
	case key.Matches(msg, m.keys.CloseModal):
		m.screen = screenList
		m.projectCursor = 0
		m.filterQuery = ""
		return m, nil
	case key.Matches(msg, m.keys.Projects):
		m.screen = screenList
		m.projectCursor = 0
		m.filterQuery = ""
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.projectCursor > 0 {
			m.projectCursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.projectCursor < len(displayedProjects(m))-1 {
			m.projectCursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.PageUp):
		m.projectCursor = clampCursor(m.projectCursor-pageStep(m), len(displayedProjects(m)))
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m.projectCursor = clampCursor(m.projectCursor+pageStep(m), len(displayedProjects(m)))
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		m.filtering = true
		m.filterQuery = ""
		return m, nil
	case key.Matches(msg, m.keys.ToggleAllStatuses):
		if m.projectStatusFilter == psfOpen {
			m.projectStatusFilter = psfAll
		} else {
			m.projectStatusFilter = psfOpen
		}
		return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)
	case key.Matches(msg, m.keys.QuickEntry):
		// 'n' in screenProjects = new project (REQ-3.1).
		if m.blockWriteIfReadOnly() {
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		m.projectEditor = newProjectEditor(true, nil, m.service)
		m.editingProject = true
		return m, m.projectEditor.focusCurrent()
	case msg.String() == "e":
		// REQ-3.2: edit selected project.
		if m.blockWriteIfReadOnly() {
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		p := projectAtCursor(m)
		if p == nil {
			return m, nil
		}
		m.projectEditor = newProjectEditor(false, p, m.service)
		m.editingProject = true
		return m, m.projectEditor.focusCurrent()
	case key.Matches(msg, m.keys.Delete):
		// REQ-3.8: confirm delete.
		if m.blockWriteIfReadOnly() {
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		p := projectAtCursor(m)
		if p == nil {
			return m, nil
		}
		pid := p.ID
		m.confirm = &confirmState{projectID: &pid}
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		// REQ-4.1: zoom into project.
		p := projectAtCursor(m)
		if p == nil {
			return m, nil
		}
		pid := p.ID
		m.screen = screenProjectTasks
		m.activeProjectID = &pid
		m.cursor = 0
		m.filterQuery = ""
		m.filtering = false
		m.selected = make(map[id.ID]struct{})
		return m, fetchProjectTasks(m.service, pid)
	}
	// All other keys (1..6, Tab, Shift+Tab, c/x/p, etc.) are no-ops in
	// screenProjects (REQ-1.4).
	return m, nil
}

// handleProjectTasksKey processes keys in the project-tasks zoom screen.
// All GTD-navigation keys (P, Tab, Shift+Tab, 1..6) are blocked. Esc
// returns to screenProjects. The remaining keys reuse the existing task
// helpers because m.tasks is mirrored from m.projectTasks while zoomed.
func (m Model) handleProjectTasksKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirm != nil {
		return handleConfirmKey(m, msg)
	}
	if m.filtering {
		return handleFilterKey(m, msg)
	}
	// Blocked keys (REQ-4.6): no-op so user cannot escape into GTD nav.
	switch {
	case key.Matches(msg, m.keys.Projects),
		key.Matches(msg, m.keys.NextView),
		key.Matches(msg, m.keys.PrevView),
		key.Matches(msg, m.keys.Inbox),
		key.Matches(msg, m.keys.Today),
		key.Matches(msg, m.keys.Upcoming),
		key.Matches(msg, m.keys.Anytime),
		key.Matches(msg, m.keys.Someday),
		key.Matches(msg, m.keys.Logbook):
		return m, nil
	}
	// Esc returns to screenProjects (or clears selection if any).
	if key.Matches(msg, m.keys.ClearSelection) && len(m.selected) > 0 {
		m.selected = make(map[id.ID]struct{})
		return m, nil
	}
	if key.Matches(msg, m.keys.CloseModal) {
		m.screen = screenProjects
		m.activeProjectID = nil
		m.projectTasks = nil
		m.tasks = nil
		m.cursor = 0
		m.selected = make(map[id.ID]struct{})
		return m, fetchProjects(m.service, m.projectStatusFilter == psfAll)
	}
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		if m.screen == screenHelp {
			m.screen = screenProjectTasks
		} else {
			m.screen = screenHelp
		}
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		m.filtering = true
		m.filterQuery = ""
		return m, nil
	case key.Matches(msg, m.keys.ToggleSelect):
		sel := m.selectedTask()
		if sel != nil {
			if _, ok := m.selected[sel.ID]; ok {
				delete(m.selected, sel.ID)
			} else {
				m.selected[sel.ID] = struct{}{}
			}
		}
		return m, nil
	case key.Matches(msg, m.keys.SelectAll):
		for _, t := range displayedTasks(m) {
			m.selected[t.ID] = struct{}{}
		}
		return m, nil
	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}
		return m, nil
	case key.Matches(msg, m.keys.PageUp):
		m.cursor = clampCursor(m.cursor-pageStep(m), len(m.tasks))
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m.cursor = clampCursor(m.cursor+pageStep(m), len(m.tasks))
		return m, nil
	case key.Matches(msg, m.keys.Enter):
		return m.openEditor()
	case key.Matches(msg, m.keys.Complete):
		return dispatch(m, bulkActionComplete)
	case key.Matches(msg, m.keys.Cancel):
		return dispatch(m, bulkActionCancel)
	case key.Matches(msg, m.keys.Delete):
		return dispatch(m, bulkActionDelete)
	case key.Matches(msg, m.keys.PinToday):
		return dispatch(m, bulkActionPin)
	case key.Matches(msg, m.keys.Refresh):
		return m, m.reloadDisplayedTasks()
	}
	return m, nil
}

// handleProjectEditorKey processes key events while the ProjectEditorModel
// modal is active in screenProjects. Esc cancels, Ctrl+S validates and
// saves, Tab/Shift+Tab cycle fields, Space on AutoClose toggles bool,
// everything else dispatches into the focused sub-widget.
func (m Model) handleProjectEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.projectEditor.picker != nil {
		np, res := m.projectEditor.picker.Update(msg, m.service)
		m.projectEditor.picker = &np
		switch res.outcome {
		case pickerCancel:
			m.projectEditor.picker = nil
		case pickerSelected:
			m.projectEditor.areaID = res.areaID
			m.projectEditor.areaName = res.areaName
			m.projectEditor.picker = nil
		}
		return m, nil
	}
	switch {
	case key.Matches(msg, m.keys.CloseModal):
		m.editingProject = false
		m.projectEditor = ProjectEditorModel{}
		return m, nil
	case key.Matches(msg, m.keys.Save):
		svc := m.service
		editor := m.projectEditor
		return m, func() tea.Msg {
			p, created, err := editor.ApplyAndSave(context.Background(), svc)
			if err != nil {
				return projectEditorErrMsg{err: err.Error()}
			}
			return projectSavedMsg{project: p, created: created}
		}
	case key.Matches(msg, m.keys.NextField):
		m.projectEditor = m.projectEditor.nextField()
		return m, m.projectEditor.focusCurrent()
	case key.Matches(msg, m.keys.PrevField):
		m.projectEditor = m.projectEditor.prevField()
		return m, m.projectEditor.focusCurrent()
	case m.projectEditor.focus == pefArea && msg.Type == tea.KeyEnter:
		areas, err := m.service.ListAreas(context.Background())
		if err != nil {
			m.projectEditor.err = err.Error()
			return m, nil
		}
		p := newAreaPicker(areas, m.projectEditor.areaID, "No area", m.readOnly)
		m.projectEditor.picker = &p
		return m, nil
	case m.projectEditor.focus == pefAutoClose && msg.Type == tea.KeySpace:
		m.projectEditor.autoClose = !m.projectEditor.autoClose
		return m, nil
	}
	var cmd tea.Cmd
	m.projectEditor, cmd = m.projectEditor.UpdateForm(msg)
	return m, cmd
}

func (m Model) handleQuickEntryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenList
		return m, nil
	case tea.KeyEnter:
		raw := m.quickInput.Value()
		m.screen = screenList
		if m.blockWriteIfReadOnly() {
			return m, tea.Tick(statusFadeDuration, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
		return m, func() tea.Msg { return quickEntrySubmittedMsg{raw: raw} }
	default:
		var cmd tea.Cmd
		m.quickInput, cmd = m.quickInput.Update(msg)
		return m, cmd
	}
}

func (m Model) switchList(l listKind) (tea.Model, tea.Cmd) {
	m.filterQuery = ""
	m.filtering = false
	m.selected = make(map[id.ID]struct{})
	m.activeList = l
	m.cursor = 0
	return m, m.loadCurrentList()
}

// reloadDisplayedTasks returns the right Cmd to refresh the currently
// visible task list: tasks-of-active-project in screenProjectTasks zoom
// mode, otherwise the GTD list selected by m.activeList.
func (m Model) reloadDisplayedTasks() tea.Cmd {
	if m.screen == screenProjectTasks && m.activeProjectID != nil {
		return fetchProjectTasks(m.service, *m.activeProjectID)
	}
	return m.loadCurrentList()
}

func (m Model) loadCurrentList() tea.Cmd {
	svc := m.service
	kind := m.activeList
	return func() tea.Msg {
		ctx := context.Background()
		var tasks []task.Task
		var err error
		switch kind {
		case listInbox:
			tasks, err = svc.ListInbox(ctx)
		case listToday:
			tasks, err = svc.ListToday(ctx)
		case listUpcoming:
			tasks, err = svc.ListUpcoming(ctx)
		case listAnytime:
			tasks, err = svc.ListAnytime(ctx)
		case listSomeday:
			tasks, err = svc.ListSomeday(ctx)
		case listLogbook:
			tasks, err = svc.ListLogbook(ctx)
		}
		if err != nil {
			return errorMsg{err}
		}
		return tasksLoadedMsg{tasks: tasks}
	}
}

func (m Model) selectedTask() *task.Task {
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return nil
	}
	return &m.tasks[m.cursor]
}

func (m Model) completeSelected() (Model, tea.Cmd) {
	sel := m.selectedTask()
	if sel == nil {
		return m, nil
	}
	svc := m.service
	tid := sel.ID
	now := time.Now()
	for i := range m.tasks {
		if m.tasks[i].ID == tid {
			m.tasks[i].Status = task.StatusCompleted
			m.tasks[i].CompletedAt = &now
			m.tasks[i].CancelledAt = nil
			break
		}
	}
	return m, func() tea.Msg {
		if _, err := svc.CompleteTask(context.Background(), tid); err != nil {
			return singleActionDoneMsg{action: bulkActionComplete, tid: tid, err: err}
		}
		return singleActionDoneMsg{action: bulkActionComplete, tid: tid}
	}
}

func (m Model) cancelSelected() (Model, tea.Cmd) {
	sel := m.selectedTask()
	if sel == nil {
		return m, nil
	}
	svc := m.service
	tid := sel.ID
	now := time.Now()
	for i := range m.tasks {
		if m.tasks[i].ID == tid {
			m.tasks[i].Status = task.StatusCancelled
			m.tasks[i].CancelledAt = &now
			m.tasks[i].CompletedAt = nil
			break
		}
	}
	return m, func() tea.Msg {
		if err := svc.CancelTask(context.Background(), tid); err != nil {
			return singleActionDoneMsg{action: bulkActionCancel, tid: tid, err: err}
		}
		return singleActionDoneMsg{action: bulkActionCancel, tid: tid}
	}
}

func (m Model) deleteSelected() (Model, tea.Cmd) {
	sel := m.selectedTask()
	if sel == nil {
		return m, nil
	}
	svc := m.service
	tid := sel.ID
	for i := range m.tasks {
		if m.tasks[i].ID == tid {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}
	if m.cursor >= len(m.tasks) {
		if len(m.tasks) == 0 {
			m.cursor = 0
		} else {
			m.cursor = len(m.tasks) - 1
		}
	}
	return m, func() tea.Msg {
		if err := svc.DeleteTask(context.Background(), tid, false); err != nil {
			return singleActionDoneMsg{action: bulkActionDelete, tid: tid, err: err}
		}
		return singleActionDoneMsg{action: bulkActionDelete, tid: tid}
	}
}

func (m Model) pinSelected() (Model, tea.Cmd) {
	sel := m.selectedTask()
	if sel == nil {
		return m, nil
	}
	svc := m.service
	tid := sel.ID
	d := task.NewDate(time.Now())
	for i := range m.tasks {
		if m.tasks[i].ID == tid {
			m.tasks[i].PinnedToday = &d
			break
		}
	}
	return m, func() tea.Msg {
		if err := svc.PinToToday(context.Background(), tid); err != nil {
			return singleActionDoneMsg{action: bulkActionPin, tid: tid, err: err}
		}
		return singleActionDoneMsg{action: bulkActionPin, tid: tid}
	}
}

// View

// viewBody dispatches between single-pane and dual-pane rendering.
// Returns the body content (header and footer are wrapped separately by View).
// When single-pane, viewBody equals viewList — preserves the legacy layout.
func (m Model) viewBody() string {
	if m.screen == screenProjects {
		body := viewProjectList(m, m.width)
		if m.editingProject {
			editor := m.projectEditor.View(m.theme, m.editorWidth())
			body = lipgloss.JoinVertical(lipgloss.Left, body, editor)
		}
		return body
	}
	if m.screen == screenProjectTasks {
		return viewProjectTasks(m, m.width)
	}
	if !isDualPane(m) {
		return m.viewList()
	}
	listW, detailsW := paneWidths(m)
	left := lipgloss.NewStyle().Width(listW).Render(m.viewList())
	right := lipgloss.NewStyle().
		Width(detailsW).
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		BorderForeground(m.theme.Help.GetForeground()).
		PaddingLeft(1).
		Render(viewDetails(m, detailsW-2))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// renderSeparator returns a one-line horizontal "━" rule of width characters,
// styled via theme.Help. Returns "" if width <= 0.
func renderSeparator(theme Theme, width int) string {
	if width <= 0 {
		return ""
	}
	return theme.Help.Render(strings.Repeat("━", width))
}

func (m Model) View() string {
	// Editor takes full body — no clamp.
	if m.screen == screenEditor {
		body := m.editor.View(m.theme, m.editorWidth())
		return lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
	}
	var body string
	if m.confirm != nil {
		var modal string
		if m.confirm.projectID != nil {
			name := id.Short(*m.confirm.projectID)
			for _, p := range m.projects {
				if p.ID == *m.confirm.projectID {
					name = p.Name
					break
				}
			}
			modal = m.theme.Modal.Render(fmt.Sprintf("Delete project %q? Tasks will move to Inbox. (y/n)", name))
		} else {
			modal = m.theme.Modal.Render(fmt.Sprintf("%s %d tasks? (y/n)", m.confirm.action.label(), len(m.confirm.ids)))
		}
		body = lipgloss.JoinVertical(lipgloss.Left, m.viewBody(), modal)
	} else {
		switch m.screen {
		case screenHelp:
			body = m.viewHelp()
		case screenQuickEntry:
			body = lipgloss.JoinVertical(lipgloss.Left, m.viewBody(), m.viewQuickEntry())
		default:
			body = m.viewBody()
		}
	}
	if m.height >= 10 && m.width >= 40 && m.screen != screenEditor {
		header := m.viewHeader()
		footer := m.viewFooter()
		sep := renderSeparator(m.theme, m.width)
		headerH := lipgloss.Height(header)
		footerH := lipgloss.Height(footer)
		bodyH := m.height - headerH - footerH - 2 // -2 for the two separator rows
		if bodyH < 0 {
			bodyH = 0
		}
		clampedBody := lipgloss.NewStyle().Height(bodyH).MaxHeight(bodyH).Render(body)
		return lipgloss.JoinVertical(lipgloss.Left, header, sep, clampedBody, sep, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
}

func (m Model) editorWidth() int {
	if m.width <= 0 {
		return 60
	}
	if m.width > 80 {
		return 80
	}
	return m.width
}

// wrapTitleColumn soft-wraps title to availWidth (lipgloss-width counted)
// and prepends prefixWidth ASCII spaces to lines [1..]. Returns
// []string{title} when availWidth <= 0 or prefixWidth <= 0 (no-op
// safeguard for unknown terminal size or degenerate narrow panes).
func wrapTitleColumn(title string, prefixWidth, availWidth int) []string {
	if availWidth <= 0 || prefixWidth <= 0 {
		return []string{title}
	}
	wrapped := lipgloss.NewStyle().Width(availWidth).Render(title)
	parts := strings.Split(wrapped, "\n")
	if len(parts) == 1 {
		return parts
	}
	indent := strings.Repeat(" ", prefixWidth)
	for i := 1; i < len(parts); i++ {
		parts[i] = indent + parts[i]
	}
	return parts
}

func (m Model) viewList() string {
	disp := displayedTasks(m)
	if len(disp) == 0 {
		if m.filterQuery != "" {
			return m.theme.Dim.Render("\n  (no matches)\n")
		}
		return m.theme.Dim.Render("\n  (no tasks)\n")
	}
	paneWidth := m.width
	if isDualPane(m) {
		paneWidth, _ = paneWidths(m)
	}
	// Render every task row and remember which line in the output
	// hosts the cursor's first line. Windowing then happens in line
	// space (BL-7): logical-row scroll is incorrect when titles wrap
	// because lipgloss MaxHeight clamps the rendered string by lines.
	var lines []string
	cursorLineIdx := -1
	showPrefix := len(m.selected) > 0
	for i, t := range disp {
		prefix := ""
		if showPrefix {
			if _, ok := m.selected[t.ID]; ok {
				prefix = "[x] "
			} else {
				prefix = "[ ] "
			}
		}
		marker := "  "
		if i == m.cursor {
			marker = m.theme.Selected.Render("> ")
		}
		icon := "  "
		switch t.Status {
		case task.StatusCompleted:
			icon = m.theme.StatusInfo.Render("✓ ")
		case task.StatusCancelled:
			icon = m.theme.StatusError.Render("✗ ")
		}
		short := m.theme.Dim.Render(id.Short(t.ID))
		firstLinePrefix := fmt.Sprintf("%s%s%s%s  ", prefix, marker, icon, short)
		prefixWidth := lipgloss.Width(firstLinePrefix)

		titleLines := wrapTitleColumn(t.Title, prefixWidth, paneWidth-prefixWidth)

		if t.Status == task.StatusCompleted || t.Status == task.StatusCancelled {
			done := lipgloss.NewStyle().Strikethrough(true).Faint(true)
			for j := range titleLines {
				if j == 0 {
					titleLines[j] = done.Render(titleLines[j])
					continue
				}
				content := strings.TrimLeft(titleLines[j], " ")
				indentLen := len(titleLines[j]) - len(content)
				titleLines[j] = strings.Repeat(" ", indentLen) + done.Render(content)
			}
		}

		if i == m.cursor {
			cursorLineIdx = len(lines)
		}
		lines = append(lines, firstLinePrefix+titleLines[0])
		lines = append(lines, titleLines[1:]...)
	}
	return windowLines(lines, cursorLineIdx, visibleRows(m), scrolloff)
}

func (m Model) viewQuickEntry() string {
	return m.theme.Modal.Render("Quick Entry\n" + m.quickInput.View() + "\n\nEnter=add  Esc=cancel")
}

func (m Model) viewHelp() string {
	binds := []key.Binding{
		m.keys.Help, m.keys.Quit,
		m.keys.Inbox, m.keys.Today, m.keys.Upcoming, m.keys.Anytime, m.keys.Someday, m.keys.Logbook,
		m.keys.Up, m.keys.Down, m.keys.PageUp, m.keys.PageDown,
		m.keys.QuickEntry, m.keys.Complete, m.keys.Cancel, m.keys.Delete, m.keys.PinToday, m.keys.Refresh,
		m.keys.Filter, m.keys.ToggleSelect, m.keys.SelectAll, m.keys.ClearSelection,
	}
	lines := []string{m.theme.Title.Render("Keybindings")}
	for _, b := range binds {
		h := b.Help()
		lines = append(lines, "  "+m.theme.Selected.Render(h.Key)+"  "+h.Desc)
	}
	return strings.Join(lines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

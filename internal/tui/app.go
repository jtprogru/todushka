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
	"github.com/jtprogru/todushka/internal/domain/id"
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
}

// allLists is the canonical order used by Tab/Shift+Tab cycling and the header.
var allLists = []listKind{listInbox, listToday, listUpcoming, listAnytime, listSomeday, listLogbook}

// NewModel constructs the root TUI model.
func NewModel(svc *app.Service, theme Theme) Model {
	ti := textinput.New()
	ti.Placeholder = "what to do? — tokens: #tag @today @project !YYYY-MM-DD"
	ti.CharLimit = 256
	return Model{
		service:    svc,
		keys:       DefaultKeyMap(),
		theme:      theme,
		screen:     screenList,
		activeList: listToday,
		quickInput: ti,
	}
}

func (m Model) Init() tea.Cmd { return m.loadCurrentList() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		if m.cursor >= len(m.tasks) {
			m.cursor = max(0, len(m.tasks)-1)
		}
		return m, nil

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
			_, err := svc.QuickEntry(context.Background(), raw)
			if err != nil {
				return errorMsg{err}
			}
			return nil
		}

	case editorSavedMsg:
		m.screen = screenList
		return m, m.loadCurrentList()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenQuickEntry:
		return m.handleQuickEntryKey(msg)
	case screenEditor:
		return m.handleEditorKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		if m.screen == screenHelp {
			m.screen = screenList
		} else {
			m.screen = screenHelp
		}
		return m, nil
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
	case key.Matches(msg, m.keys.Enter):
		return m.openEditor()
	case key.Matches(msg, m.keys.QuickEntry):
		m.screen = screenQuickEntry
		m.quickInput.SetValue("")
		m.quickInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Complete):
		return m, m.completeSelected()
	case key.Matches(msg, m.keys.Cancel):
		return m, m.cancelSelected()
	case key.Matches(msg, m.keys.Delete):
		return m, m.deleteSelected()
	case key.Matches(msg, m.keys.PinToday):
		return m, m.pinSelected()
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
	m.editor = NewEditor(*sel)
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
	switch {
	case key.Matches(msg, m.keys.CloseModal):
		m.screen = screenList
		return m, nil
	case key.Matches(msg, m.keys.Save):
		return m.saveEditor()
	case key.Matches(msg, m.keys.NextField):
		m.editor = m.editor.nextField()
		return m, m.editor.focusCurrent()
	case key.Matches(msg, m.keys.PrevField):
		m.editor = m.editor.prevField()
		return m, m.editor.focusCurrent()
	case m.editor.focus == fieldSomeday && msg.Type == tea.KeySpace:
		m.editor.someday = !m.editor.someday
		return m, nil
	}
	var cmd tea.Cmd
	m.editor, cmd = m.editor.UpdateForm(msg)
	return m, cmd
}

func (m Model) saveEditor() (tea.Model, tea.Cmd) {
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

func (m Model) handleQuickEntryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.screen = screenList
		return m, nil
	case tea.KeyEnter:
		raw := m.quickInput.Value()
		m.screen = screenList
		return m, tea.Batch(
			func() tea.Msg { return quickEntrySubmittedMsg{raw: raw} },
			m.loadCurrentList(),
		)
	default:
		var cmd tea.Cmd
		m.quickInput, cmd = m.quickInput.Update(msg)
		return m, cmd
	}
}

func (m Model) switchList(l listKind) (tea.Model, tea.Cmd) {
	m.activeList = l
	m.cursor = 0
	return m, m.loadCurrentList()
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

func (m Model) completeSelected() tea.Cmd {
	sel := m.selectedTask()
	if sel == nil {
		return nil
	}
	svc := m.service
	tid := sel.ID
	return func() tea.Msg {
		if _, err := svc.CompleteTask(context.Background(), tid); err != nil {
			return errorMsg{err}
		}
		return nil
	}
}

func (m Model) cancelSelected() tea.Cmd {
	sel := m.selectedTask()
	if sel == nil {
		return nil
	}
	svc := m.service
	tid := sel.ID
	return func() tea.Msg {
		if err := svc.CancelTask(context.Background(), tid); err != nil {
			return errorMsg{err}
		}
		return nil
	}
}

func (m Model) deleteSelected() tea.Cmd {
	sel := m.selectedTask()
	if sel == nil {
		return nil
	}
	svc := m.service
	tid := sel.ID
	return func() tea.Msg {
		if err := svc.DeleteTask(context.Background(), tid, false); err != nil {
			return errorMsg{err}
		}
		return nil
	}
}

func (m Model) pinSelected() tea.Cmd {
	sel := m.selectedTask()
	if sel == nil {
		return nil
	}
	svc := m.service
	tid := sel.ID
	return func() tea.Msg {
		if err := svc.PinToToday(context.Background(), tid); err != nil {
			return errorMsg{err}
		}
		return nil
	}
}

// View

func (m Model) View() string {
	var body string
	switch m.screen {
	case screenHelp:
		body = m.viewHelp()
	case screenQuickEntry:
		body = lipgloss.JoinVertical(lipgloss.Left, m.viewList(), m.viewQuickEntry())
	case screenEditor:
		body = m.editor.View(m.theme, m.editorWidth())
	default:
		body = m.viewList()
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

func (m Model) viewHeader() string {
	parts := make([]string, 0, len(allLists))
	for _, l := range allLists {
		if l == m.activeList {
			parts = append(parts, m.theme.Header.Render(l.String()))
		} else {
			parts = append(parts, m.theme.HeaderDim.Render(l.String()))
		}
	}
	return strings.Join(parts, "")
}

func (m Model) viewList() string {
	if len(m.tasks) == 0 {
		return m.theme.Dim.Render("\n  (no tasks)\n")
	}
	lines := make([]string, 0, len(m.tasks))
	for i, t := range m.tasks {
		marker := "  "
		if i == m.cursor {
			marker = m.theme.Selected.Render("> ")
		}
		title := t.Title
		dates := ""
		if t.StartDate != nil {
			dates += " start:" + t.StartDate.Format("2006-01-02")
		}
		if t.Deadline != nil {
			dl := " due:" + t.Deadline.Format("2006-01-02")
			dates += m.theme.Deadline.Render(dl)
		}
		short := m.theme.Dim.Render(id.Short(t.ID))
		lines = append(lines, fmt.Sprintf("%s%s  %s%s", marker, short, title, dates))
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewQuickEntry() string {
	return m.theme.Modal.Render("Quick Entry\n" + m.quickInput.View() + "\n\nEnter=add  Esc=cancel")
}

func (m Model) viewHelp() string {
	binds := []key.Binding{
		m.keys.Help, m.keys.Quit,
		m.keys.Inbox, m.keys.Today, m.keys.Upcoming, m.keys.Anytime, m.keys.Someday, m.keys.Logbook,
		m.keys.Up, m.keys.Down,
		m.keys.QuickEntry, m.keys.Complete, m.keys.Cancel, m.keys.Delete, m.keys.PinToday, m.keys.Refresh,
	}
	lines := []string{m.theme.Title.Render("Keybindings")}
	for _, b := range binds {
		h := b.Help()
		lines = append(lines, "  "+m.theme.Selected.Render(h.Key)+"  "+h.Desc)
	}
	return strings.Join(lines, "\n")
}

func (m Model) viewFooter() string {
	hints := "?: help  ⇥: next view  n: quick  ↵: edit  c: complete  q: quit"
	if m.screen == screenEditor {
		hints = "Tab: field  Ctrl+S: save  Esc: cancel"
	}
	left := m.theme.Help.Render(hints)
	right := ""
	if m.statusMsg != "" {
		right = m.theme.StatusError.Render(m.statusMsg)
	}
	return left + "  " + right
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

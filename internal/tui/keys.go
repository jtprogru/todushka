package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit           key.Binding
	Help           key.Binding
	Inbox          key.Binding
	Today          key.Binding
	Upcoming       key.Binding
	Anytime        key.Binding
	Someday        key.Binding
	Logbook        key.Binding
	Up             key.Binding
	Down           key.Binding
	Enter          key.Binding
	QuickEntry     key.Binding
	Complete       key.Binding
	Cancel         key.Binding
	Delete         key.Binding
	PinToday       key.Binding
	Refresh        key.Binding
	CloseModal     key.Binding
	NextView       key.Binding
	PrevView       key.Binding
	NextField      key.Binding
	PrevField      key.Binding
	Save           key.Binding
	ToggleSelect   key.Binding
	SelectAll      key.Binding
	ClearSelection key.Binding
	Filter         key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:           key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q / Ctrl+C", "quit")),
		Help:           key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Inbox:          key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "Inbox")),
		Today:          key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "Today")),
		Upcoming:       key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "Upcoming")),
		Anytime:        key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "Anytime")),
		Someday:        key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "Someday")),
		Logbook:        key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "Logbook")),
		Up:             key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k / ↑", "up")),
		Down:           key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j / ↓", "down")),
		Enter:          key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "details")),
		QuickEntry:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "quick entry")),
		Complete:       key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "complete")),
		Cancel:         key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "cancel")),
		Delete:         key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		PinToday:       key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pin to today")),
		Refresh:        key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		CloseModal:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
		NextView:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next view")),
		PrevView:       key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev view")),
		NextField:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		PrevField:      key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
		Save:           key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		ToggleSelect:   key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle select")),
		SelectAll:      key.NewBinding(key.WithKeys("*"), key.WithHelp("*", "select all visible")),
		ClearSelection: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear selection")),
		Filter:         key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	}
}

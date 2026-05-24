package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
)

// Run launches the TUI on the given service. Theme is selected from env
// (NO_COLOR / TODUSHKA_THEME). Blocks until the user quits.
func Run(svc *app.Service) error {
	theme := SelectTheme(os.Getenv)
	m := NewModel(svc, theme)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

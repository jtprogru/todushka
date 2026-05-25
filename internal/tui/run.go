package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
)

// Run launches the TUI on the given service with the provided config.
// Theme falls back to NO_COLOR / TODUSHKA_THEME env if cfg.Theme is empty.
// Blocks until the user quits.
func Run(svc *app.Service, cfg config.AppConfig) error {
	theme := selectThemeFromConfig(cfg.Theme, os.Getenv)
	m := NewModel(svc, theme, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// selectThemeFromConfig picks the theme based on the config value, with env
// fallbacks. NO_COLOR has absolute precedence (REQ-1.6); "auto"/"system"
// trigger OS dark-mode detection via resolveAutoTheme (REQ-1.1).
func selectThemeFromConfig(name string, env func(string) string) Theme {
	if env == nil {
		env = func(string) string { return "" }
	}
	if env("NO_COLOR") != "" {
		return NewMonochromeTheme()
	}
	if name == "auto" || name == "system" {
		name = resolveAutoTheme()
	}
	switch name {
	case "latte", "light":
		return newColorTheme(latte)
	case "mono", "monochrome":
		return NewMonochromeTheme()
	case "macchiato", "dark":
		return newColorTheme(macchiato)
	}
	// Fallback to legacy env-based selection
	return SelectTheme(env)
}

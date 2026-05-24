package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme bundles the styles used across the TUI. SelectTheme picks the right
// palette for the runtime environment.
type Theme struct {
	Name string

	Background lipgloss.Color
	Surface    lipgloss.Color
	Overlay    lipgloss.Color
	Text       lipgloss.Color
	Subtext    lipgloss.Color
	Accent     lipgloss.Color // active list, focused field
	AccentAlt  lipgloss.Color // tags
	Warning    lipgloss.Color // deadlines
	Error      lipgloss.Color // overdue, status errors
	Success    lipgloss.Color // status info

	Title       lipgloss.Style
	Header      lipgloss.Style
	HeaderDim   lipgloss.Style
	Selected    lipgloss.Style
	Dim         lipgloss.Style
	Deadline    lipgloss.Style
	Overdue     lipgloss.Style
	Tag         lipgloss.Style
	Help        lipgloss.Style
	StatusError lipgloss.Style
	StatusInfo  lipgloss.Style
	Modal       lipgloss.Style
	Field       lipgloss.Style
	FieldFocus  lipgloss.Style
	Label       lipgloss.Style
}

type palette struct {
	name       string
	background string
	surface    string
	overlay    string
	text       string
	subtext    string
	accent     string
	accentAlt  string
	warning    string
	errorC     string
	success    string
}

// macchiato is the Catppuccin Macchiato palette (dark).
var macchiato = palette{
	name:       "catppuccin-macchiato",
	background: "#24273a", // base
	surface:    "#363a4f", // surface0
	overlay:    "#6e738d", // overlay0
	text:       "#cad3f5", // text
	subtext:    "#a5adcb", // subtext0
	accent:     "#8aadf4", // blue
	accentAlt:  "#a6da95", // green
	warning:    "#eed49f", // yellow
	errorC:     "#ed8796", // red
	success:    "#a6da95", // green
}

// latte is the Catppuccin Latte palette (light).
var latte = palette{
	name:       "catppuccin-latte",
	background: "#eff1f5", // base
	surface:    "#ccd0da", // surface0
	overlay:    "#9ca0b0", // overlay0
	text:       "#4c4f69", // text
	subtext:    "#6c6f85", // subtext0
	accent:     "#1e66f5", // blue
	accentAlt:  "#40a02b", // green
	warning:    "#df8e1d", // yellow
	errorC:     "#d20f39", // red
	success:    "#40a02b", // green
}

// SelectTheme picks the appropriate theme based on environment.
// Respects NO_COLOR (monochrome fallback) and TODUSHKA_THEME=light|dark.
// Default: Catppuccin Macchiato (dark).
func SelectTheme(env func(string) string) Theme {
	if env == nil {
		env = func(string) string { return "" }
	}
	if env("NO_COLOR") != "" {
		return NewMonochromeTheme()
	}
	want := strings.ToLower(strings.TrimSpace(env("TODUSHKA_THEME")))
	if want == "light" || want == "latte" {
		return newColorTheme(latte)
	}
	return newColorTheme(macchiato)
}

func newColorTheme(p palette) Theme {
	t := Theme{
		Name:       p.name,
		Background: lipgloss.Color(p.background),
		Surface:    lipgloss.Color(p.surface),
		Overlay:    lipgloss.Color(p.overlay),
		Text:       lipgloss.Color(p.text),
		Subtext:    lipgloss.Color(p.subtext),
		Accent:     lipgloss.Color(p.accent),
		AccentAlt:  lipgloss.Color(p.accentAlt),
		Warning:    lipgloss.Color(p.warning),
		Error:      lipgloss.Color(p.errorC),
		Success:    lipgloss.Color(p.success),
	}
	t.Title = lipgloss.NewStyle().Bold(true).Foreground(t.Accent)
	t.Header = lipgloss.NewStyle().Bold(true).Foreground(t.Background).Background(t.Accent).Padding(0, 1)
	t.HeaderDim = lipgloss.NewStyle().Foreground(t.Subtext).Padding(0, 1)
	t.Selected = lipgloss.NewStyle().Bold(true).Foreground(t.Background).Background(t.Accent)
	t.Dim = lipgloss.NewStyle().Foreground(t.Overlay)
	t.Deadline = lipgloss.NewStyle().Foreground(t.Warning)
	t.Overdue = lipgloss.NewStyle().Bold(true).Foreground(t.Error)
	t.Tag = lipgloss.NewStyle().Foreground(t.AccentAlt)
	t.Help = lipgloss.NewStyle().Foreground(t.Subtext).Italic(true)
	t.StatusError = lipgloss.NewStyle().Bold(true).Foreground(t.Error)
	t.StatusInfo = lipgloss.NewStyle().Foreground(t.Success)
	t.Modal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Foreground(t.Text).
		Padding(1, 2)
	t.Field = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Overlay).
		Foreground(t.Text).
		Padding(0, 1)
	t.FieldFocus = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Accent).
		Foreground(t.Text).
		Padding(0, 1)
	t.Label = lipgloss.NewStyle().Bold(true).Foreground(t.Subtext)
	return t
}

// NewTheme returns the default coloured theme (Macchiato). Kept for tests.
func NewTheme() Theme { return newColorTheme(macchiato) }

func NewMonochromeTheme() Theme {
	bold := lipgloss.NewStyle().Bold(true)
	underline := lipgloss.NewStyle().Underline(true)
	dim := lipgloss.NewStyle().Faint(true)
	return Theme{
		Name:        "monochrome",
		Title:       bold,
		Header:      bold,
		HeaderDim:   dim,
		Selected:    underline,
		Dim:         dim,
		Deadline:    bold,
		Overdue:     bold.Underline(true),
		Tag:         underline,
		Help:        dim,
		StatusError: bold,
		StatusInfo:  bold,
		Modal:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
		Field:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
		FieldFocus:  lipgloss.NewStyle().Border(lipgloss.ThickBorder()).Padding(0, 1),
		Label:       bold,
	}
}

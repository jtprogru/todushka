package config

// AppConfig holds user-tunable runtime settings loaded from YAML / env / flag.
type AppConfig struct {
	Theme                string  `yaml:"theme"`
	DualPaneMinWidth     int     `yaml:"dual_pane_min_width"`
	ListPaneShare        float64 `yaml:"list_pane_share"`
	BulkConfirmThreshold int     `yaml:"bulk_confirm_threshold"`
	NotesMaxLines        int     `yaml:"notes_max_lines"`
	ConfirmDelete        bool    `yaml:"confirm_delete"`
}

// Defaults returns the built-in fallback values.
func Defaults() AppConfig {
	return AppConfig{
		Theme:                "macchiato",
		DualPaneMinWidth:     100,
		ListPaneShare:        0.60,
		BulkConfirmThreshold: 5,
		NotesMaxLines:        8,
		ConfirmDelete:        true,
	}
}

// Validate returns a corrected AppConfig (invalid fields replaced with
// defaults) AND a slice of warning messages describing each correction.
func (c AppConfig) Validate() (AppConfig, []string) {
	def := Defaults()
	var warns []string
	switch c.Theme {
	case "macchiato", "latte", "mono", "auto", "system", "":
		// ok
	default:
		warns = append(warns, "invalid theme '"+c.Theme+"', using '"+def.Theme+"'")
		c.Theme = def.Theme
	}
	if c.Theme == "" {
		c.Theme = def.Theme
	}
	if c.DualPaneMinWidth < 40 {
		if c.DualPaneMinWidth != 0 {
			warns = append(warns, "dual_pane_min_width must be >= 40, using default")
		}
		c.DualPaneMinWidth = def.DualPaneMinWidth
	}
	if c.ListPaneShare <= 0 || c.ListPaneShare >= 1 {
		if c.ListPaneShare != 0 {
			warns = append(warns, "list_pane_share must be in (0, 1), using default")
		}
		c.ListPaneShare = def.ListPaneShare
	}
	if c.BulkConfirmThreshold < 1 {
		if c.BulkConfirmThreshold != 0 {
			warns = append(warns, "bulk_confirm_threshold must be >= 1, using default")
		}
		c.BulkConfirmThreshold = def.BulkConfirmThreshold
	}
	if c.NotesMaxLines < 1 {
		if c.NotesMaxLines != 0 {
			warns = append(warns, "notes_max_lines must be >= 1, using default")
		}
		c.NotesMaxLines = def.NotesMaxLines
	}
	return c, warns
}

package tui

// detectDarkModeFn is package-level so tests can override it. Production
// uses the platform-specific detectDarkMode from detect_dark_*.go.
var detectDarkModeFn = detectDarkMode

// resolveAutoTheme returns the theme name to use when AppConfig.Theme is
// "auto" or "system". Calls detectDarkModeFn(); maps:
//
//	dark  → "macchiato"
//	light → "latte"
//	error/unsupported → "macchiato" (default to dark per ADR-1)
func resolveAutoTheme() string {
	isDark, err := detectDarkModeFn()
	if err != nil || isDark {
		return "macchiato"
	}
	return "latte"
}

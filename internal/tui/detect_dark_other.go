//go:build !darwin && !linux

package tui

// detectDarkMode for unsupported platforms returns (false, nil).
// Callers should fall back to a default theme (macchiato per design ADR).
func detectDarkMode() (bool, error) {
	return false, nil
}

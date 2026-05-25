//go:build linux

package tui

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// detectDarkMode reports whether Linux/GNOME is in dark mode via:
//
//	gsettings get org.gnome.desktop.interface color-scheme
//
// Output formats: 'prefer-dark' | 'prefer-light' | 'default'.
// Returns (true, nil) if 'dark' substring; (false, nil) if 'light';
// (false, err) on tool/timeout error.
func detectDarkMode() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	s := strings.ToLower(strings.TrimSpace(string(out)))
	if strings.Contains(s, "dark") {
		return true, nil
	}
	return false, nil
}

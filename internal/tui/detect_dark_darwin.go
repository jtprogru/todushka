//go:build darwin

package tui

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// detectDarkMode reports whether macOS is in dark mode via:
//
//	defaults read -g AppleInterfaceStyle
//
// Returns (true, nil) when output is "Dark"; (false, nil) when command
// exits non-zero (which is macOS's way of saying "light mode"); (false, err)
// only on unexpected errors.
func detectDarkMode() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "defaults", "read", "-g", "AppleInterfaceStyle")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(string(out)) == "Dark", nil
}

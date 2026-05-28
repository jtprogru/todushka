package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// scrolloff is the minimum number of rows kept visible above and below
// the cursor before the viewport starts following it. Mirrors vim's
// `scrolloff=3` default.
const scrolloff = 3

// windowLines returns a slice of `lines` of length up to `visibleRows`
// that contains `cursorLineIdx` with at least `scrolloff` lines of
// buffer above and below (clamped by the start/end of `lines`).
// When `visibleRows <= 0` or `len(lines) <= visibleRows`, the entire
// list is joined (no truncation). When `cursorLineIdx < 0` (cursor
// not in current view), the window is taken from the start.
//
// Line-space windowing is the correct model for the renderer because
// lipgloss `MaxHeight` clamps the body by lines, not by logical rows;
// when titles wrap into multiple lines a logical-row offset can place
// the cursor row past the clamp.
func windowLines(lines []string, cursorLineIdx, visibleRows, scrolloff int) string {
	if visibleRows <= 0 || len(lines) <= visibleRows {
		return strings.Join(lines, "\n")
	}
	if cursorLineIdx < 0 {
		return strings.Join(lines[:visibleRows], "\n")
	}
	start := cursorLineIdx - scrolloff
	if start < 0 {
		start = 0
	}
	end := start + visibleRows
	if end > len(lines) {
		end = len(lines)
		start = end - visibleRows
		if start < 0 {
			start = 0
		}
	}
	// Hard-bounds re-anchor — guarantees cursor visible even when the
	// window is narrower than the buffer (visibleRows < 2*scrolloff+1).
	if cursorLineIdx >= end {
		end = cursorLineIdx + 1
		if end > len(lines) {
			end = len(lines)
		}
		start = end - visibleRows
		if start < 0 {
			start = 0
		}
	}
	if cursorLineIdx < start {
		start = cursorLineIdx
		end = start + visibleRows
		if end > len(lines) {
			end = len(lines)
		}
	}
	return strings.Join(lines[start:end], "\n")
}

// pageStep reports how many rows PgUp/PgDown moves the cursor: one
// bodyful of rows (visibleRows). Falls back to a sane default before the
// first WindowSizeMsg so paging still works when the terminal size is
// not yet known.
func pageStep(m Model) int {
	if v := visibleRows(m); v > 0 {
		return v
	}
	return 10
}

// clampCursor bounds idx to [0, n-1]; returns 0 when n <= 0.
func clampCursor(idx, n int) int {
	if n <= 0 || idx < 0 {
		return 0
	}
	if idx > n-1 {
		return n - 1
	}
	return idx
}

// visibleRows reports how many list rows can be rendered in the body
// pane: m.height - height(viewHeader) - height(viewFooter) - 2
// (two separator rows). Returns 0 before the first WindowSizeMsg
// (m.height == 0) or when the result would be negative.
func visibleRows(m Model) int {
	if m.height <= 0 {
		return 0
	}
	headerH := lipgloss.Height(m.viewHeader())
	footerH := lipgloss.Height(m.viewFooter())
	body := m.height - headerH - footerH - 2
	if body < 0 {
		return 0
	}
	return body
}

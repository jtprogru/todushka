package tui

import "github.com/charmbracelet/lipgloss"

// scrolloff is the minimum number of rows kept visible above and below
// the cursor before the viewport starts following it. Mirrors vim's
// `scrolloff=3` default.
const scrolloff = 3

// ensureCursorVisible returns a new scrollOffset such that `cursor` is
// rendered within the window [offset, offset+visibleCount), with at
// least `scrolloff` rows of buffer above and below (clamped by the list
// ends and the visibleCount).
//
//   - If visibleCount <= 0 or totalCount <= visibleCount, the result is 0:
//     everything fits or nothing renders.
//   - If cursor < 0, it is treated as 0 (defensive).
//   - The returned offset is always clamped to [0, max(0, totalCount-visibleCount)].
func ensureCursorVisible(cursor, offset, visibleCount, scrolloff, totalCount int) int {
	if visibleCount <= 0 {
		return 0
	}
	if totalCount <= visibleCount {
		return 0
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= totalCount {
		cursor = totalCount - 1
	}
	maxOffset := totalCount - visibleCount

	// Hard bounds — cursor must be inside [offset, offset+visibleCount).
	// These also dominate when visibleCount < 2*scrolloff+1 (narrow window
	// where both buffer rules would otherwise conflict).
	minOffset := cursor - visibleCount + 1
	if minOffset < 0 {
		minOffset = 0
	}
	maxOffsetForCursor := cursor
	if maxOffsetForCursor > maxOffset {
		maxOffsetForCursor = maxOffset
	}

	// Top buffer: cursor approaching the top of the window.
	if cursor-offset < scrolloff {
		offset = cursor - scrolloff
	}
	// Bottom buffer: cursor approaching the bottom of the window.
	if cursor-offset > visibleCount-1-scrolloff {
		offset = cursor - visibleCount + 1 + scrolloff
	}

	// Clamp into [minOffset, maxOffsetForCursor]: guarantees cursor visible.
	if offset < minOffset {
		offset = minOffset
	}
	if offset > maxOffsetForCursor {
		offset = maxOffsetForCursor
	}
	return offset
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

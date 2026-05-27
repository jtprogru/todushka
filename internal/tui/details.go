package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
)

// isDualPane reports whether the renderer should use horizontal split.
// Activated only when:
//   - terminal width >= m.config.DualPaneMinWidth (excludes width==0 initial state)
//   - screen is screenList — editor/help force single-pane (REQ-1.6/1.7)
//
// Filter mode and quick-entry overlay do NOT disable dual-pane.
func isDualPane(m Model) bool {
	if m.width < m.config.DualPaneMinWidth {
		return false
	}
	if m.screen == screenEditor || m.screen == screenHelp {
		return false
	}
	return true
}

// paneWidths returns (listWidth, detailsWidth) for the dual-pane layout.
// The 1-column border between panes is allocated separately. Invariant:
// listWidth + 1 + detailsWidth == m.width.
func paneWidths(m Model) (int, int) {
	list := int(float64(m.width-1) * m.config.ListPaneShare)
	details := m.width - 1 - list
	return list, details
}

// fetchNameCache returns a Cmd that resolves all referenced
// tag/area/project IDs in tasks via the Repository, emitting
// nameCacheLoadedMsg with the result. Per-ID errors are silently skipped —
// missing names fall back to short-IDs in views (REQ-4.3).
//
// NOTE: Heading names are not resolved in v1 because the Repository
// interface lacks a direct HeadingGet(ctx, id) method (only
// HeadingList(ctx, projectID)). Heading IDs will display via short-ID
// fallback until a HeadingGet method is added in v2.
func fetchNameCache(svc *app.Service, tasks []task.Task) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		tags := make(map[id.ID]string)
		areas := make(map[id.ID]string)
		projects := make(map[id.ID]project.Project)
		headings := make(map[id.ID]string)

		tagSet := make(map[id.ID]struct{})
		areaSet := make(map[id.ID]struct{})
		projectSet := make(map[id.ID]struct{})
		for _, t := range tasks {
			for _, tg := range t.Tags {
				tagSet[tg] = struct{}{}
			}
			if t.AreaID != nil {
				areaSet[*t.AreaID] = struct{}{}
			}
			if t.ProjectID != nil {
				projectSet[*t.ProjectID] = struct{}{}
			}
		}

		repo := svc.Repo()
		for tid := range tagSet {
			if tg, err := repo.TagGet(ctx, tid); err == nil {
				tags[tid] = tg.Name
			}
		}
		for aid := range areaSet {
			if a, err := repo.AreaGet(ctx, aid); err == nil {
				areas[aid] = a.Name
			}
		}
		for pid := range projectSet {
			if p, err := repo.ProjectGet(ctx, pid); err == nil {
				projects[pid] = p
			}
		}
		return nameCacheLoadedMsg{tags: tags, areas: areas, projects: projects, headings: headings}
	}
}

// cursorTask returns the task at m.cursor inside displayedTasks(m), or nil
// when cursor is out of range or the visible list is empty.
func cursorTask(m Model) *task.Task {
	disp := displayedTasks(m)
	if m.cursor < 0 || m.cursor >= len(disp) {
		return nil
	}
	return &disp[m.cursor]
}

// wrapAndTruncate soft-wraps text to width via lipgloss and truncates to
// at most maxLines lines, appending "…" if truncation occurred.
func wrapAndTruncate(text string, width, maxLines int) string {
	if text == "" || width <= 0 {
		return ""
	}
	wrapped := lipgloss.NewStyle().Width(width).Render(text)
	lines := strings.Split(wrapped, "\n")
	if len(lines) <= maxLines {
		return wrapped
	}
	return strings.Join(lines[:maxLines], "\n") + "\n…"
}

// resolveName looks up an ID in a name cache, falling back to id.Short(tid)
// when missing (REQ-4.3).
func resolveName(cache map[id.ID]string, tid id.ID) string {
	if n, ok := cache[tid]; ok && n != "" {
		return n
	}
	return id.Short(tid)
}

// resolveProjectName looks up a Project by ID and returns its Name, falling
// back to id.Short(pid) when the cache has no entry or an empty Name.
func resolveProjectName(cache map[id.ID]project.Project, pid id.ID) string {
	if p, ok := cache[pid]; ok && p.Name != "" {
		return p.Name
	}
	return id.Short(pid)
}

// statusLabel maps task.Status to a user-facing label.
func statusLabel(s task.Status) string {
	switch s {
	case task.StatusOpen:
		return "Open"
	case task.StatusCompleted:
		return "Completed"
	case task.StatusCancelled:
		return "Cancelled"
	}
	return string(s)
}

// viewDetails renders the right pane content for dual-pane mode. Pure
// function — reads only m and width. Returns "(no task selected)" when
// cursorTask(m) is nil.
//
// Output is a sequence of logical groups (Title, Status, Notes, Dates,
// Relations, Tags, Someday). Groups are separated by exactly one blank
// line; absent groups produce no orphan blank lines (REQ-1.3, REQ-1.4).
// Labels are styled via theme.DetailLabel (Bold + Accent in color
// themes); Project/Heading appear on separate lines with project
// sub-fields nested between them.
func viewDetails(m Model, width int) string {
	t := cursorTask(m)
	if t == nil {
		return m.theme.Dim.Render("(no task selected)")
	}
	label := func(s string) string { return m.theme.DetailLabel.Render(s) }

	var groups [][]string

	// Group: Title
	groups = append(groups, []string{m.theme.Title.Render(wrapAndTruncate(t.Title, width, 4))})

	// Group: Status
	groups = append(groups, []string{label("Status:") + " " + statusLabel(t.Status)})

	// Group: Notes
	if t.Notes != "" {
		groups = append(groups, []string{wrapAndTruncate(t.Notes, width, m.config.NotesMaxLines)})
	}

	// Group: Dates
	var dates []string
	if t.StartDate != nil {
		dates = append(dates, label("Start:")+"  "+t.StartDate.Format("2006-01-02"))
	}
	if t.Deadline != nil {
		dates = append(dates, label("Due:")+"    "+t.Deadline.Format("2006-01-02"))
	}
	if t.PinnedToday != nil {
		dates = append(dates, label("Pinned:")+" "+t.PinnedToday.Format("2006-01-02"))
	}
	if len(dates) > 0 {
		groups = append(groups, dates)
	}

	// Group: Relations (Area / Project + sub-fields / Heading on separate lines)
	var relations []string
	if t.AreaID != nil {
		relations = append(relations, label("Area:")+"    "+resolveName(m.areaNamesByID, *t.AreaID))
	}
	if t.ProjectID != nil {
		relations = append(relations, label("Project:")+" "+resolveProjectName(m.projectsByID, *t.ProjectID))
		if p, ok := m.projectsByID[*t.ProjectID]; ok && p.Name != "" {
			if p.Status != project.StatusOpen && p.Status != "" {
				relations = append(relations, "  "+label("Project status:")+" "+string(p.Status))
			}
			if p.Deadline != nil {
				relations = append(relations, "  "+label("Project due:")+" "+p.Deadline.Format("2006-01-02"))
			}
			if p.Notes != "" {
				relations = append(relations, "  "+label("Project notes:")+" "+wrapAndTruncate(p.Notes, width-2, 3))
			}
		}
	}
	if t.HeadingID != nil {
		relations = append(relations, label("Heading:")+" "+resolveName(m.headingNamesByID, *t.HeadingID))
	}
	if len(relations) > 0 {
		groups = append(groups, relations)
	}

	// Group: Tags
	if len(t.Tags) > 0 {
		names := make([]string, 0, len(t.Tags))
		for _, tg := range t.Tags {
			names = append(names, resolveName(m.tagNamesByID, tg))
		}
		groups = append(groups, []string{label("Tags:") + " " + strings.Join(names, ", ")})
	}

	// Group: Someday flag
	if t.Someday {
		groups = append(groups, []string{m.theme.Dim.Render("Someday")})
	}

	// Join groups with one blank line between each non-empty group.
	var out []string
	for i, g := range groups {
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, g...)
	}
	return strings.Join(out, "\n")
}

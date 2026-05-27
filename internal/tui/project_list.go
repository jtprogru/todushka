package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
)

// displayedProjects applies the current filter query (case-fold contains)
// to m.projects and returns the matching subset.
func displayedProjects(m Model) []project.Project {
	q := strings.TrimSpace(m.filterQuery)
	if q == "" {
		return m.projects
	}
	out := make([]project.Project, 0, len(m.projects))
	for _, p := range m.projects {
		if foldCaseContains(p.Name, q) {
			out = append(out, p)
		}
	}
	return out
}

// projectAtCursor returns the project at m.projectCursor inside
// displayedProjects(m), or nil when the cursor is out of range.
func projectAtCursor(m Model) *project.Project {
	disp := displayedProjects(m)
	if m.projectCursor < 0 || m.projectCursor >= len(disp) {
		return nil
	}
	return &disp[m.projectCursor]
}

// fetchProjects returns a Cmd that loads projects via
// ListProjectsSorted and emits projectsLoadedMsg with per-project counts.
func fetchProjects(svc *app.Service, includeAll bool) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		projects, err := svc.ListProjectsSorted(ctx, nil, includeAll)
		if err != nil {
			return errorMsg{err}
		}
		counts := make(map[id.ID][2]int, len(projects))
		for _, p := range projects {
			open, total, _ := svc.CountProjectTasks(ctx, p.ID)
			counts[p.ID] = [2]int{open, total}
		}
		return projectsLoadedMsg{projects: projects, counts: counts}
	}
}

// fetchAreaNames loads area names for the given IDs and merges them into
// the name cache via nameCacheLoadedMsg. Errors are silently skipped
// (existing convention from fetchNameCache).
func fetchAreaNames(svc *app.Service, areaIDs []id.ID) tea.Cmd {
	if len(areaIDs) == 0 {
		return nil
	}
	return func() tea.Msg {
		ctx := context.Background()
		areas := make(map[id.ID]string, len(areaIDs))
		repo := svc.Repo()
		for _, aid := range areaIDs {
			if a, err := repo.AreaGet(ctx, aid); err == nil {
				areas[aid] = a.Name
			}
		}
		return nameCacheLoadedMsg{areas: areas}
	}
}

// projectStatusIcon returns a short rendered icon reflecting Status.
// Open → blank, Completed → ✓ (info color), Cancelled → ✗ (error color).
func projectStatusIcon(theme Theme, st project.Status) string {
	switch st {
	case project.StatusCompleted:
		return theme.StatusInfo.Render("✓")
	case project.StatusCancelled:
		return theme.StatusError.Render("✗")
	}
	return " "
}

// viewProjectList renders the project-list pane content. width is the
// available column count. Returns a dim placeholder for empty lists.
func viewProjectList(m Model, width int) string {
	disp := displayedProjects(m)
	header := m.theme.Title.Render("Projects")
	filterStatus := "Open"
	if m.projectStatusFilter == psfAll {
		filterStatus = "All"
	}
	header += "  " + m.theme.Dim.Render(fmt.Sprintf("(%s)", filterStatus))

	if len(disp) == 0 {
		if m.filterQuery != "" {
			return header + m.theme.Dim.Render("\n  (no matches)\n")
		}
		return header + m.theme.Dim.Render("\n  (no projects)\n")
	}

	lines := []string{header, ""}
	for i, p := range disp {
		marker := "  "
		if i == m.projectCursor {
			marker = m.theme.Selected.Render("> ")
		}
		icon := projectStatusIcon(m.theme, p.Status) + " "
		short := m.theme.Dim.Render(id.Short(p.ID))

		var countsStr string
		if c, ok := m.projectCounts[p.ID]; ok {
			countsStr = m.theme.Dim.Render(fmt.Sprintf(" [%d/%d]", c[0], c[1]))
		}

		var areaStr string
		if p.AreaID != nil {
			if name, ok := m.areaNamesByID[*p.AreaID]; ok && name != "" {
				areaStr = m.theme.Dim.Render("  (area: " + name + ")")
			}
		}

		var deadlineStr string
		if p.Deadline != nil {
			deadlineStr = "  " + m.theme.Deadline.Render(p.Deadline.Format("2006-01-02"))
		}

		row := fmt.Sprintf("%s%s%s  %s%s%s%s", marker, icon, short, p.Name, countsStr, areaStr, deadlineStr)
		lines = append(lines, row)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
)

// fetchProjectTasks loads tasks filtered by ProjectID for the
// screenProjectTasks zoom-in view. Errors surface via errorMsg.
func fetchProjectTasks(svc *app.Service, pid id.ID) tea.Cmd {
	return func() tea.Msg {
		tasks, err := svc.Repo().TaskList(context.Background(), storage.TaskFilter{ProjectID: &pid})
		if err != nil {
			return errorMsg{err}
		}
		return projectTasksLoadedMsg{projectID: pid, tasks: tasks}
	}
}

// projectName looks up the active project's display name from the model's
// projects slice. Falls back to short-id when not found.
func projectName(m Model) string {
	if m.activeProjectID == nil {
		return ""
	}
	for _, p := range m.projects {
		if p.ID == *m.activeProjectID {
			return p.Name
		}
	}
	return id.Short(*m.activeProjectID)
}

// viewProjectTasks renders the task list filtered to the active project.
// Each task is rendered with the same wrap-style as viewList; tasks
// belonging to a heading get an inline [heading] badge after the title.
func viewProjectTasks(m Model, width int) string {
	header := m.theme.Title.Render(projectName(m))
	disp := displayedTasks(m)
	if len(disp) == 0 {
		if m.filterQuery != "" {
			return header + "\n" + m.theme.Dim.Render("  (no matches)")
		}
		return header + "\n" + m.theme.Dim.Render("  (no tasks in this project)")
	}
	// Apply viewport scroll (BL-7). Header + blank line occupy 2 rows.
	vr := visibleRows(m) - 2
	off := m.scrollOffset
	if vr > 0 && len(disp) > vr {
		if off > len(disp)-vr {
			off = len(disp) - vr
		}
		if off < 0 {
			off = 0
		}
		end := off + vr
		if end > len(disp) {
			end = len(disp)
		}
		disp = disp[off:end]
	} else {
		off = 0
	}
	lines := []string{header, ""}
	for i, t := range disp {
		absIdx := i + off
		marker := "  "
		if absIdx == m.cursor {
			marker = m.theme.Selected.Render("> ")
		}
		icon := "  "
		switch t.Status {
		case task.StatusCompleted:
			icon = m.theme.StatusInfo.Render("✓ ")
		case task.StatusCancelled:
			icon = m.theme.StatusError.Render("✗ ")
		}
		short := m.theme.Dim.Render(id.Short(t.ID))
		title := t.Title
		if t.HeadingID != nil {
			if name, ok := m.headingNamesByID[*t.HeadingID]; ok && name != "" {
				title = title + " " + m.theme.Tag.Render(fmt.Sprintf("[%s]", name))
			} else {
				title = title + " " + m.theme.Tag.Render(fmt.Sprintf("[%s]", id.Short(*t.HeadingID)))
			}
		}
		row := fmt.Sprintf("%s%s%s  %s", marker, icon, short, title)
		lines = append(lines, row)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// keep strings import to satisfy go vet for indirect TrimSpace usage
// (handleFilterKey is in a sibling file and reaches into m.filterQuery).
var _ = strings.TrimSpace

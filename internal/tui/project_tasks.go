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
	// Build row-by-row, track cursor line, then window in line space (BL-7).
	rowLines := []string{}
	cursorLineIdx := -1
	for i, t := range disp {
		if i == m.cursor {
			cursorLineIdx = len(rowLines)
		}
		marker := "  "
		if i == m.cursor {
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
		rowLines = append(rowLines, row)
	}
	// Header + blank line consume 2 of visibleRows.
	body := windowLines(rowLines, cursorLineIdx, visibleRows(m)-2, scrolloff)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}

var _ = strings.TrimSpace
var _ = lipgloss.JoinVertical

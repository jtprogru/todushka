package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
)

// FormatTaskList renders a task list, either in human-readable text or JSON.
// The colorEnabled flag is honoured for text mode; JSON output is always raw.
func FormatTaskList(w io.Writer, tasks []task.Task, jsonMode, colorEnabled bool) error {
	if jsonMode {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(tasks)
	}
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(w, "(no tasks)")
		return err
	}
	for _, t := range tasks {
		line := formatTaskLine(t, colorEnabled)
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func formatTaskLine(t task.Task, colorEnabled bool) string {
	parts := []string{id.Short(t.ID), t.Title}
	if t.Deadline != nil {
		parts = append(parts, "due:"+t.Deadline.Format("2006-01-02"))
	}
	if t.StartDate != nil {
		parts = append(parts, "start:"+t.StartDate.Format("2006-01-02"))
	}
	out := strings.Join(parts, "  ")
	if colorEnabled {
		// Minimal: don't pull in lipgloss here — that belongs to the TUI.
		// CLI text mode is intentionally plain.
		return out
	}
	return out
}

// ColorEnabled reports whether ANSI colour is allowed for the current
// environment.
func ColorEnabled(env func(string) string) bool {
	if env == nil {
		return false
	}
	return env("NO_COLOR") == ""
}

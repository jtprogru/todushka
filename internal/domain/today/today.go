// Package today implements the pure Today Engine: given a snapshot of tasks
// and a local "now", returns the subset that belongs in the Today list.
package today

import (
	"time"

	"github.com/jtprogru/todushka/internal/domain/task"
)

// ComputeToday returns tasks that satisfy the Today inclusion rules (REQ-8.1..8.5):
//
//   - status == open
//   - deleted_at == nil
//   - and at least one of:
//   - start_date != nil && start_date <= today()
//   - start_date == nil && deadline != nil && deadline <= today() + deadlineWindowDays
//   - pinned_today == today() (auto-reset at midnight, ADR-7)
//
// The function is referentially transparent — `now` is injected, no globals are read.
func ComputeToday(tasks []task.Task, now time.Time, deadlineWindowDays int) []task.Task {
	today := startOfDay(now)
	windowEnd := today.AddDate(0, 0, deadlineWindowDays)

	out := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Status != task.StatusOpen {
			continue
		}
		if t.DeletedAt != nil {
			continue
		}
		if includesInToday(t, today, windowEnd) {
			out = append(out, t)
		}
	}
	return out
}

func includesInToday(t task.Task, today, windowEnd time.Time) bool {
	if t.PinnedToday != nil {
		pinned := startOfDay(t.PinnedToday.Time)
		if pinned.Equal(today) {
			return true
		}
	}
	if t.StartDate != nil {
		sd := startOfDay(t.StartDate.Time)
		return !sd.After(today)
	}
	if t.Deadline != nil {
		dl := startOfDay(t.Deadline.Time)
		if !dl.After(windowEnd) {
			return true
		}
	}
	return false
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

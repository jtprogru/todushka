package today

import (
	"testing"
	"time"

	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func dateAt(t *testing.T, y int, m time.Month, d int) task.Date {
	t.Helper()
	return task.NewDate(time.Date(y, m, d, 0, 0, 0, 0, time.UTC))
}

func nowAt(t *testing.T, y int, m time.Month, d int) time.Time {
	t.Helper()
	return time.Date(y, m, d, 10, 30, 0, 0, time.UTC)
}

func dp(d task.Date) *task.Date { return &d }

func TestComputeToday_Scenarios(t *testing.T) {
	now := nowAt(t, 2026, 5, 25)
	today := dateAt(t, 2026, 5, 25)
	tomorrow := dateAt(t, 2026, 5, 26)
	yesterday := dateAt(t, 2026, 5, 24)
	completedAt := nowAt(t, 2026, 5, 24)

	cases := []struct {
		name   string
		task   task.Task
		want   bool
		window int
	}{
		{
			name: "start_today_included",
			task: task.Task{Title: "x", Status: task.StatusOpen, StartDate: dp(today)},
			want: true,
		},
		{
			name: "start_yesterday_included",
			task: task.Task{Title: "x", Status: task.StatusOpen, StartDate: dp(yesterday)},
			want: true,
		},
		{
			name: "start_tomorrow_excluded",
			task: task.Task{Title: "x", Status: task.StatusOpen, StartDate: dp(tomorrow)},
			want: false,
		},
		{
			name: "no_dates_excluded",
			task: task.Task{Title: "x", Status: task.StatusOpen},
			want: false,
		},
		{
			name: "deadline_today_window_zero_included",
			task: task.Task{Title: "x", Status: task.StatusOpen, Deadline: dp(today)},
			want: true,
		},
		{
			name: "deadline_tomorrow_window_zero_excluded",
			task: task.Task{Title: "x", Status: task.StatusOpen, Deadline: dp(tomorrow)},
			want: false,
		},
		{
			name:   "deadline_tomorrow_window_1_included",
			task:   task.Task{Title: "x", Status: task.StatusOpen, Deadline: dp(tomorrow)},
			window: 1,
			want:   true,
		},
		{
			name: "pinned_today_included_even_if_start_in_future",
			task: task.Task{Title: "x", Status: task.StatusOpen, StartDate: dp(tomorrow), PinnedToday: dp(today)},
			want: true,
		},
		{
			name: "pinned_yesterday_auto_reset_excluded",
			task: task.Task{Title: "x", Status: task.StatusOpen, StartDate: dp(tomorrow), PinnedToday: dp(yesterday)},
			want: false,
		},
		{
			name: "completed_excluded",
			task: task.Task{Title: "x", Status: task.StatusCompleted, StartDate: dp(today), CompletedAt: &completedAt},
			want: false,
		},
		{
			name: "cancelled_excluded",
			task: task.Task{Title: "x", Status: task.StatusCancelled, StartDate: dp(today), CancelledAt: &completedAt},
			want: false,
		},
		{
			name: "deleted_excluded",
			task: task.Task{Title: "x", Status: task.StatusOpen, StartDate: dp(today), DeletedAt: &completedAt},
			want: false,
		},
		{
			name: "deadline_present_with_start_future_uses_start_rule",
			task: task.Task{
				Title: "x", Status: task.StatusOpen,
				StartDate: dp(tomorrow),
				Deadline:  dp(today),
			},
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ComputeToday([]task.Task{c.task}, now, c.window)
			if c.want {
				require.Len(t, got, 1)
			} else {
				require.Empty(t, got)
			}
		})
	}
}

// PBT: determinism (CP-4).
func TestProp_TodayDeterminism(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tasks := genTasks(rt)
		now := nowAt(t, 2026, 5, 25)
		window := rapid.IntRange(0, 7).Draw(rt, "window")
		a := ComputeToday(tasks, now, window)
		b := ComputeToday(tasks, now, window)
		require.Equal(rt, a, b)
	})
}

// PBT: inclusion (CP-2) — if any inclusion clause holds for an open undeleted task,
// it must be in the result.
func TestProp_TodayInclusion(t *testing.T) {
	now := nowAt(t, 2026, 5, 25)
	today := dateAt(t, 2026, 5, 25)
	rapid.Check(t, func(rt *rapid.T) {
		// Choose one of three inclusion clauses to satisfy.
		clause := rapid.IntRange(0, 2).Draw(rt, "clause")
		var tk task.Task
		tk.Title = "x"
		tk.Status = task.StatusOpen
		switch clause {
		case 0:
			d := dateAt(t, 2026, 5, rapid.IntRange(1, 25).Draw(rt, "start_day"))
			tk.StartDate = &d
		case 1:
			d := dateAt(t, 2026, 5, 25)
			tk.Deadline = &d
		case 2:
			d := dateAt(t, 2026, 5, 25)
			tk.PinnedToday = &d
			future := dateAt(t, 2026, 6, 1)
			tk.StartDate = &future
		}
		_ = today
		got := ComputeToday([]task.Task{tk}, now, 0)
		require.Len(rt, got, 1)
	})
}

// PBT: exclusion (CP-3).
func TestProp_TodayExclusion(t *testing.T) {
	now := nowAt(t, 2026, 5, 25)
	rapid.Check(t, func(rt *rapid.T) {
		scenario := rapid.IntRange(0, 3).Draw(rt, "scenario")
		tk := task.Task{Title: "x", Status: task.StatusOpen}
		ts := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
		switch scenario {
		case 0:
			tk.Status = task.StatusCompleted
			tk.CompletedAt = &ts
			sd := dateAt(t, 2026, 5, 25)
			tk.StartDate = &sd
		case 1:
			tk.Status = task.StatusCancelled
			tk.CancelledAt = &ts
		case 2:
			tk.DeletedAt = &ts
			sd := dateAt(t, 2026, 5, 25)
			tk.StartDate = &sd
		case 3:
			// start in the future, no deadline, not pinned today
			future := dateAt(t, 2026, 6, rapid.IntRange(1, 28).Draw(rt, "day"))
			tk.StartDate = &future
		}
		got := ComputeToday([]task.Task{tk}, now, 0)
		require.Empty(rt, got)
	})
}

func genTasks(rt *rapid.T) []task.Task {
	n := rapid.IntRange(0, 30).Draw(rt, "n")
	tasks := make([]task.Task, 0, n)
	for i := 0; i < n; i++ {
		status := rapid.SampledFrom([]task.Status{task.StatusOpen, task.StatusCompleted, task.StatusCancelled}).Draw(rt, "status")
		tk := task.Task{Title: "t", Status: status}
		ts := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
		switch status {
		case task.StatusCompleted:
			tk.CompletedAt = &ts
		case task.StatusCancelled:
			tk.CancelledAt = &ts
		}
		if rapid.Bool().Draw(rt, "has_start") {
			d := task.NewDate(time.Date(2026, time.Month(rapid.IntRange(1, 12).Draw(rt, "m")),
				rapid.IntRange(1, 28).Draw(rt, "d"), 0, 0, 0, 0, time.UTC))
			tk.StartDate = &d
		}
		if rapid.Bool().Draw(rt, "has_deadline") {
			d := task.NewDate(time.Date(2026, time.Month(rapid.IntRange(1, 12).Draw(rt, "m")),
				rapid.IntRange(1, 28).Draw(rt, "d"), 0, 0, 0, 0, time.UTC))
			tk.Deadline = &d
		}
		if rapid.Bool().Draw(rt, "has_pin") {
			d := task.NewDate(time.Date(2026, time.Month(rapid.IntRange(1, 12).Draw(rt, "m")),
				rapid.IntRange(1, 28).Draw(rt, "d"), 0, 0, 0, 0, time.UTC))
			tk.PinnedToday = &d
		}
		tasks = append(tasks, tk)
	}
	return tasks
}

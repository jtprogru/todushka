package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// ─── BL-1.1: label styling ───────────────────────────────────────────────

func TestTheme_DetailLabelColorTheme(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	th := NewTheme()
	out := th.DetailLabel.Render("Start:")
	require.Contains(t, out, "\x1b[1", "DetailLabel must include bold ANSI escape")
	// lipgloss collapses bold + foreground into one CSI sequence: \x1b[1;38;2;R;G;Bm
	require.Regexp(t, `\x1b\[[\d;]*38;2;\d+;\d+;\d+`, out, "DetailLabel must include truecolor foreground ANSI in color theme")
}

func TestTheme_DetailLabelMonochrome(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	th := NewMonochromeTheme()
	out := th.DetailLabel.Render("Start:")
	require.Contains(t, out, "\x1b[1", "DetailLabel must include bold ANSI escape in monochrome")
	require.NotContains(t, out, "\x1b[38;", "DetailLabel must not include foreground color ANSI in monochrome")
}

func TestProp_DetailLabelHasBold(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	rapid.Check(t, func(rt *rapid.T) {
		mono := rapid.Bool().Draw(rt, "mono")
		var th Theme
		if mono {
			th = NewMonochromeTheme()
		} else {
			th = NewTheme()
		}
		out := th.DetailLabel.Render("X")
		require.Contains(rt, out, "\x1b[1", "DetailLabel must always be bold")
	})
}

// labelStyledAroundFunc returns a model used in viewDetails styling tests.
// All optional fields are populated so every label can be asserted.
func makeFullModel(t *testing.T) Model {
	t.Helper()
	m := newTestModel(t)
	now := time.Now()
	start := task.NewDate(time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC))
	due := task.NewDate(time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC))
	pinned := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
	aid := id.New()
	pid := id.New()
	hid := id.New()
	tagID := id.New()
	tk := task.Task{
		ID:          id.New(),
		Title:       "the title",
		Notes:       "the notes",
		Status:      task.StatusOpen,
		StartDate:   &start,
		Deadline:    &due,
		PinnedToday: &pinned,
		AreaID:      &aid,
		ProjectID:   &pid,
		HeadingID:   &hid,
		Tags:        []id.ID{tagID},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.tasks = []task.Task{tk}
	m.cursor = 0
	m.areaNamesByID[aid] = "work"
	m.projectsByID[pid] = project.Project{Name: "todushka", Status: project.StatusOpen}
	m.headingNamesByID[hid] = "ideas"
	m.tagNamesByID[tagID] = "urgent"
	return m
}

func TestViewDetails_LabelsStyled(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	m := makeFullModel(t)
	out := viewDetails(m, 80)

	for _, label := range []string{"Status:", "Start:", "Due:", "Pinned:", "Area:", "Project:", "Heading:", "Tags:"} {
		styled := m.theme.DetailLabel.Render(label)
		require.Contains(t, out, styled, "label %q must be wrapped via DetailLabel", label)
	}
}

func TestProp_AllLabelsStyled(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

	rapid.Check(t, func(rt *rapid.T) {
		hasStart := rapid.Bool().Draw(rt, "hasStart")
		hasDue := rapid.Bool().Draw(rt, "hasDue")
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		hasTags := rapid.Bool().Draw(rt, "hasTags")

		m := newTestModel(t)
		now := time.Now()
		tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen, CreatedAt: now, UpdatedAt: now}
		if hasStart {
			d := task.NewDate(now)
			tk.StartDate = &d
		}
		if hasDue {
			d := task.NewDate(now)
			tk.Deadline = &d
		}
		if hasArea {
			aid := id.New()
			tk.AreaID = &aid
			m.areaNamesByID[aid] = "a"
		}
		if hasProject {
			pid := id.New()
			tk.ProjectID = &pid
			m.projectsByID[pid] = project.Project{Name: "p", Status: project.StatusOpen}
		}
		if hasTags {
			tg := id.New()
			tk.Tags = []id.ID{tg}
			m.tagNamesByID[tg] = "t"
		}
		m.tasks = []task.Task{tk}
		m.cursor = 0

		out := viewDetails(m, 80)
		// Status is always present
		require.Contains(rt, out, m.theme.DetailLabel.Render("Status:"))
		if hasStart {
			require.Contains(rt, out, m.theme.DetailLabel.Render("Start:"))
		}
		if hasDue {
			require.Contains(rt, out, m.theme.DetailLabel.Render("Due:"))
		}
		if hasArea {
			require.Contains(rt, out, m.theme.DetailLabel.Render("Area:"))
		}
		if hasProject {
			require.Contains(rt, out, m.theme.DetailLabel.Render("Project:"))
		}
		if hasTags {
			require.Contains(rt, out, m.theme.DetailLabel.Render("Tags:"))
		}
	})
}

func TestViewDetails_NoOrphanBlankLines(t *testing.T) {
	m := newTestModel(t)
	now := time.Now()
	m.tasks = []task.Task{{ID: id.New(), Title: "x", Status: task.StatusOpen, CreatedAt: now, UpdatedAt: now}}
	m.cursor = 0

	out := viewDetails(m, 60)
	require.NotContains(t, out, "\n\n\n", "must not emit two consecutive blank lines")
}

func TestViewDetails_EmptyLineBetweenGroups(t *testing.T) {
	m := newTestModel(t)
	now := time.Now()
	start := task.NewDate(now)
	aid := id.New()
	tagID := id.New()
	tk := task.Task{
		ID:        id.New(),
		Title:     "x",
		Status:    task.StatusOpen,
		StartDate: &start,
		AreaID:    &aid,
		Tags:      []id.ID{tagID},
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.tasks = []task.Task{tk}
	m.cursor = 0
	m.areaNamesByID[aid] = "work"
	m.tagNamesByID[tagID] = "urgent"

	out := viewDetails(m, 60)
	lines := strings.Split(out, "\n")
	// Expect at least 4 non-empty groups (Title, Status, Date, Area, Tags) →
	// at least 4 blank-line separators inside the output.
	blank := 0
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			blank++
		}
	}
	require.GreaterOrEqual(t, blank, 4, "expected blank lines between groups")
}

func TestProp_NoConsecutiveBlankLines(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		hasStart := rapid.Bool().Draw(rt, "hasStart")
		hasDue := rapid.Bool().Draw(rt, "hasDue")
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		hasNotes := rapid.Bool().Draw(rt, "hasNotes")
		hasTags := rapid.Bool().Draw(rt, "hasTags")
		isSomeday := rapid.Bool().Draw(rt, "someday")

		m := newTestModel(t)
		now := time.Now()
		tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen, CreatedAt: now, UpdatedAt: now}
		if hasStart {
			d := task.NewDate(now)
			tk.StartDate = &d
		}
		if hasDue {
			d := task.NewDate(now)
			tk.Deadline = &d
		}
		if hasArea {
			aid := id.New()
			tk.AreaID = &aid
			m.areaNamesByID[aid] = "a"
		}
		if hasProject {
			pid := id.New()
			tk.ProjectID = &pid
			m.projectsByID[pid] = project.Project{Name: "p", Status: project.StatusOpen}
		}
		if hasNotes {
			tk.Notes = "notes"
		}
		if hasTags {
			tg := id.New()
			tk.Tags = []id.ID{tg}
			m.tagNamesByID[tg] = "t"
		}
		tk.Someday = isSomeday
		m.tasks = []task.Task{tk}
		m.cursor = 0

		out := viewDetails(m, 60)
		require.NotContains(rt, out, "\n\n\n", "no consecutive blank lines")
	})
}

// ─── BL-2: details pane width ────────────────────────────────────────────

func TestConfig_DefaultsListPaneShareIs06(t *testing.T) {
	require.InDelta(t, 0.60, config.Defaults().ListPaneShare, 1e-9)
}

func TestPaneWidths_DetailsAtMost40Percent(t *testing.T) {
	for _, w := range []int{100, 120, 150, 200, 300} {
		m := newTestModel(t)
		m.width = w
		_, details := paneWidths(m)
		ratio := float64(details) / float64(w)
		require.LessOrEqual(t, ratio, 0.40, "details ratio %.3f exceeds 0.40 at width=%d", ratio, w)
	}
}

func TestConfig_ValidatePreservesValidListPaneShare(t *testing.T) {
	for _, v := range []float64{0.20, 0.30, 0.45, 0.50, 0.80} {
		c, _ := config.AppConfig{ListPaneShare: v, DualPaneMinWidth: 100, NotesMaxLines: 8, BulkConfirmThreshold: 5, Theme: "macchiato"}.Validate()
		require.InDelta(t, v, c.ListPaneShare, 1e-9, "Validate must preserve ListPaneShare=%v", v)
	}
}

func TestProp_DetailsLeq40Percent(t *testing.T) {
	// Integer floor in paneWidths can cause up to a 1-column drift above
	// strict 0.40 at small/odd widths (e.g. w=102 → 41/102 ≈ 0.402).
	// We assert the headline invariant with a 1-column slack: details
	// share never exceeds 0.40 + 1/w. At w=100 this is 0.41; at larger
	// widths the drift shrinks toward zero.
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(100, 400).Draw(rt, "width")
		m := newTestModel(t)
		m.width = w
		_, details := paneWidths(m)
		bound := 0.40 + 1.0/float64(w)
		require.LessOrEqual(rt, float64(details)/float64(w), bound)
	})
}

func TestProp_ListPaneShareRoundtrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		v := rapid.Float64Range(0.01, 0.99).Draw(rt, "v")
		c, _ := config.AppConfig{ListPaneShare: v, DualPaneMinWidth: 100, NotesMaxLines: 8, BulkConfirmThreshold: 5, Theme: "macchiato"}.Validate()
		require.InDelta(rt, v, c.ListPaneShare, 1e-9)
	})
}

// ─── BL-6: project info cache flow ───────────────────────────────────────

func TestNameCache_FetchEmitsFullProject(t *testing.T) {
	_, svc := newTestModelWithService(t)
	ctx := context.Background()
	due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	p, err := svc.AddProject(ctx, app.AddProjectInput{
		Name:     "todushka",
		Notes:    "project notes",
		Deadline: &due,
	})
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "t", ProjectID: &p.ID})
	require.NoError(t, err)

	cmd := fetchNameCache(svc, []task.Task{tk})
	msg := cmd()
	res, ok := msg.(nameCacheLoadedMsg)
	require.True(t, ok)
	got := res.projects[p.ID]
	require.Equal(t, "todushka", got.Name)
	require.Equal(t, "project notes", got.Notes)
	require.NotNil(t, got.Deadline)
	require.Equal(t, "2026-06-01", got.Deadline.Format("2006-01-02"))
	require.Equal(t, project.StatusOpen, got.Status)
}

func TestNameCache_UpdateStoresFullProject(t *testing.T) {
	m := newTestModel(t)
	pid := id.New()
	due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	msg := nameCacheLoadedMsg{
		tags:     map[id.ID]string{},
		areas:    map[id.ID]string{},
		projects: map[id.ID]project.Project{pid: {Name: "foo", Status: project.StatusCompleted, Deadline: &due}},
		headings: map[id.ID]string{},
	}
	m2, _ := m.Update(msg)
	mm := m2.(Model)
	got := mm.projectsByID[pid]
	require.Equal(t, "foo", got.Name)
	require.Equal(t, project.StatusCompleted, got.Status)
	require.NotNil(t, got.Deadline)
}

func TestProp_ProjectFlowsEndToEnd(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithService(t)
		ctx := context.Background()
		name := rapid.StringMatching(`[a-z]{1,15}`).Draw(rt, "name")
		notes := rapid.StringMatching(`[a-z ]{0,20}`).Draw(rt, "notes")
		hasDeadline := rapid.Bool().Draw(rt, "hasDeadline")
		var dl *task.Date
		if hasDeadline {
			d := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
			dl = &d
		}
		p, err := svc.AddProject(ctx, app.AddProjectInput{Name: name, Notes: notes, Deadline: dl})
		require.NoError(rt, err)
		tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "t", ProjectID: &p.ID})
		require.NoError(rt, err)
		msg := fetchNameCache(svc, []task.Task{tk})()
		res := msg.(nameCacheLoadedMsg)
		got := res.projects[p.ID]
		require.Equal(rt, name, got.Name)
		require.Equal(rt, notes, got.Notes)
		if hasDeadline {
			require.NotNil(rt, got.Deadline)
		} else {
			require.Nil(rt, got.Deadline)
		}
	})
}

// ─── BL-6: project sub-fields in viewDetails ─────────────────────────────

// projectModel builds a Model with one task pointing at a Project entry in
// the cache. Use setStatus/setDeadline/setNotes to customise.
func projectModel(t *testing.T, p project.Project) (Model, id.ID) {
	t.Helper()
	m := newTestModel(t)
	pid := id.New()
	now := time.Now()
	tk := task.Task{
		ID:        id.New(),
		Title:     "task",
		Status:    task.StatusOpen,
		ProjectID: &pid,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.tasks = []task.Task{tk}
	m.cursor = 0
	if p.Name == "" {
		p.Name = "todushka"
	}
	m.projectsByID[pid] = p
	return m, pid
}

func TestViewDetails_ProjectStatusSubField(t *testing.T) {
	m, _ := projectModel(t, project.Project{Status: project.StatusCompleted})
	out := viewDetails(m, 60)
	require.Contains(t, out, "Project status:")
	require.Contains(t, out, string(project.StatusCompleted))
}

func TestViewDetails_ProjectDeadlineSubField(t *testing.T) {
	due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	m, _ := projectModel(t, project.Project{Status: project.StatusOpen, Deadline: &due})
	out := viewDetails(m, 60)
	require.Contains(t, out, "Project due:")
	require.Contains(t, out, "2026-06-01")
}

func TestViewDetails_ProjectNotesSubField(t *testing.T) {
	m, _ := projectModel(t, project.Project{Status: project.StatusOpen, Notes: "important context"})
	out := viewDetails(m, 60)
	require.Contains(t, out, "Project notes:")
	require.Contains(t, out, "important context")
}

func TestViewDetails_ProjectSubFieldsHiddenWhenOpenAndEmpty(t *testing.T) {
	m, _ := projectModel(t, project.Project{Status: project.StatusOpen})
	out := viewDetails(m, 60)
	require.Contains(t, out, "Project:")
	require.NotContains(t, out, "Project status:")
	require.NotContains(t, out, "Project due:")
	require.NotContains(t, out, "Project notes:")
}

func TestViewDetails_ProjectFallbackOnCacheMiss(t *testing.T) {
	m := newTestModel(t)
	pid := id.New()
	now := time.Now()
	tk := task.Task{
		ID:        id.New(),
		Title:     "task",
		Status:    task.StatusOpen,
		ProjectID: &pid,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.tasks = []task.Task{tk}
	m.cursor = 0
	// projectsByID intentionally empty

	out := viewDetails(m, 60)
	require.Contains(t, out, "Project:")
	require.Contains(t, out, id.Short(pid), "fallback must show short-id")
	require.NotContains(t, out, "Project status:")
	require.NotContains(t, out, "Project due:")
	require.NotContains(t, out, "Project notes:")
}

func TestViewDetails_ProjectAndHeadingSeparateLines(t *testing.T) {
	m := newTestModel(t)
	pid := id.New()
	hid := id.New()
	now := time.Now()
	tk := task.Task{
		ID:        id.New(),
		Title:     "task",
		Status:    task.StatusOpen,
		ProjectID: &pid,
		HeadingID: &hid,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.tasks = []task.Task{tk}
	m.cursor = 0
	m.projectsByID[pid] = project.Project{Name: "p"}
	m.headingNamesByID[hid] = "h"

	out := viewDetails(m, 60)
	for _, line := range strings.Split(out, "\n") {
		hasP := strings.Contains(line, "Project:")
		hasH := strings.Contains(line, "Heading:")
		require.False(t, hasP && hasH, "Project: and Heading: must not co-occur on one line: %q", line)
	}
}

func TestViewDetails_RegressionContains(t *testing.T) {
	m := makeFullModel(t)
	out := viewDetails(m, 80)
	for _, label := range []string{"Status:", "Start:", "Due:", "Area:", "Project:", "Tags:"} {
		require.Contains(t, out, label, "regression: substring %q must remain present", label)
	}
}

func TestProp_ProjectSubFieldsVisibilityMatchesCache(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		isCompleted := rapid.Bool().Draw(rt, "completed")
		hasDeadline := rapid.Bool().Draw(rt, "deadline")
		hasNotes := rapid.Bool().Draw(rt, "notes")
		status := project.StatusOpen
		if isCompleted {
			status = project.StatusCompleted
		}
		p := project.Project{Name: "p", Status: status}
		if hasDeadline {
			d := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
			p.Deadline = &d
		}
		if hasNotes {
			p.Notes = "notes content"
		}
		m, _ := projectModel(t, p)
		out := viewDetails(m, 60)

		if isCompleted {
			require.Contains(rt, out, "Project status:")
		} else {
			require.NotContains(rt, out, "Project status:")
		}
		if hasDeadline {
			require.Contains(rt, out, "Project due:")
		} else {
			require.NotContains(rt, out, "Project due:")
		}
		if hasNotes {
			require.Contains(rt, out, "Project notes:")
		} else {
			require.NotContains(rt, out, "Project notes:")
		}
	})
}

func TestProp_ProjectFallbackOnMissing(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 3).Draw(rt, "n")
		m := newTestModel(t)
		now := time.Now()
		for i := 0; i < n; i++ {
			pid := id.New()
			tk := task.Task{
				ID:        id.New(),
				Title:     fmt.Sprintf("t-%d", i),
				Status:    task.StatusOpen,
				ProjectID: &pid,
				CreatedAt: now,
				UpdatedAt: now,
			}
			m.tasks = append(m.tasks, tk)
		}
		m.cursor = rapid.IntRange(0, n-1).Draw(rt, "cursor")

		out := viewDetails(m, 60)
		require.Contains(rt, out, "Project:")
		require.NotContains(rt, out, "Project status:")
		require.NotContains(rt, out, "Project due:")
		require.NotContains(rt, out, "Project notes:")
	})
}

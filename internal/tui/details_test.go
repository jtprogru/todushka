package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestWrapAndTruncate_EmptyReturnsEmpty(t *testing.T) {
	require.Equal(t, "", wrapAndTruncate("", 40, 8))
}

func TestWrapAndTruncate_ShortPreserved(t *testing.T) {
	out := wrapAndTruncate("hello", 40, 8)
	require.Equal(t, "hello", strings.TrimRight(out, " "))
}

func TestWrapAndTruncate_LongIsTruncated(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	input := strings.Join(lines, "\n")
	out := wrapAndTruncate(input, 40, 8)
	require.Contains(t, out, "…")
	require.Equal(t, 9, len(strings.Split(out, "\n")))
}

func TestViewDetails_EmptyTasksPlaceholder(t *testing.T) {
	m := newTestModel(t)
	m.tasks = nil
	out := viewDetails(m, 60)
	require.Contains(t, out, "(no task selected)")
}

func TestViewDetails_OutOfRangeCursorPlaceholder(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "a", "b")
	m.cursor = 5
	out := viewDetails(m, 60)
	require.Contains(t, out, "(no task selected)")
}

func TestViewDetails_TitleAndStatus(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "buy milk")
	out := viewDetails(m, 60)
	require.Contains(t, out, "buy milk")
	require.Contains(t, out, "Open")
}

func TestViewDetails_Notes(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", Notes: "some note here"})
	require.NoError(t, err)
	m.tasks = []task.Task{tk}
	m.cursor = 0
	out := viewDetails(m, 60)
	require.Contains(t, out, "some note here")
}

func TestViewDetails_Dates(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	start := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
	due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	tk, err := svc.AddTask(ctx, app.AddTaskInput{
		Title:     "x",
		StartDate: &start,
		Deadline:  &due,
	})
	require.NoError(t, err)
	m.tasks = []task.Task{tk}
	m.cursor = 0
	out := viewDetails(m, 60)
	require.Contains(t, out, "Start:")
	require.Contains(t, out, "2026-05-25")
	require.Contains(t, out, "Due:")
	require.Contains(t, out, "2026-06-01")
}

func TestViewDetails_OmitsNilFields(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "minimal")
	out := viewDetails(m, 60)
	require.NotContains(t, out, "Start:")
	require.NotContains(t, out, "Due:")
	require.NotContains(t, out, "Pinned:")
	require.NotContains(t, out, "Area:")
	require.NotContains(t, out, "Project:")
	require.NotContains(t, out, "Tags:")
	require.NotContains(t, out, "Someday")
}

func TestViewDetails_RelationsAndTags(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tg, err := svc.UpsertTag(ctx, "urgent")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{
		Title:  "x",
		AreaID: &a.ID,
		Tags:   []id.ID{tg.ID},
	})
	require.NoError(t, err)
	m.tasks = []task.Task{tk}
	m.cursor = 0
	m.tagNamesByID[tg.ID] = "urgent"
	m.areaNamesByID[a.ID] = "work"
	out := viewDetails(m, 60)
	require.Contains(t, out, "Area:")
	require.Contains(t, out, "work")
	require.Contains(t, out, "Tags:")
	require.Contains(t, out, "urgent")
}

func TestViewDetails_ShortIDFallback(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: &a.ID})
	require.NoError(t, err)
	m.tasks = []task.Task{tk}
	m.cursor = 0
	// Do NOT populate areaNamesByID — fallback expected.
	out := viewDetails(m, 60)
	require.Contains(t, out, "Area:")
	require.Contains(t, out, id.Short(a.ID))
}

func TestIsDualPane_WideTerminal(t *testing.T) {
	m := newTestModel(t)
	m.width = 100
	require.True(t, isDualPane(m))
}

func TestIsDualPane_NarrowTerminal(t *testing.T) {
	m := newTestModel(t)
	m.width = 99
	require.False(t, isDualPane(m))
}

func TestIsDualPane_ZeroWidth(t *testing.T) {
	m := newTestModel(t)
	// m.width is 0 by default
	require.False(t, isDualPane(m))
}

func TestIsDualPane_EditorScreen(t *testing.T) {
	m := newTestModel(t)
	m.width = 200
	m.screen = screenEditor
	require.False(t, isDualPane(m))
}

func TestIsDualPane_HelpScreen(t *testing.T) {
	m := newTestModel(t)
	m.width = 200
	m.screen = screenHelp
	require.False(t, isDualPane(m))
}

func TestIsDualPane_FilteringAllowed(t *testing.T) {
	m := newTestModel(t)
	m.width = 200
	m.filtering = true
	require.True(t, isDualPane(m))
}

func TestPaneWidths_SumEqualsTotal(t *testing.T) {
	for _, w := range []int{100, 120, 150, 200, 300} {
		m := newTestModel(t)
		m.width = w
		list, details := paneWidths(m)
		require.Equal(t, w, list+1+details, "width %d", w)
	}
}

func TestViewBody_SinglePaneNoBorder(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "task one")
	m.width = 80
	out := m.viewBody()
	require.NotContains(t, out, "║")
	require.Contains(t, out, "task one")
}

func TestViewBody_DualPaneRendersBothAndBorder(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "task one")
	m.width = 120
	out := m.viewBody()
	require.Contains(t, out, "║", "double-line border must be present")
	require.Contains(t, out, "task one", "list content must be present")
	require.Contains(t, out, "Status:", "details content must be present")
}

func TestViewBody_FilterPreservesDualPane(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "task one")
	m.width = 120
	m.filtering = true
	m.filterQuery = "task"
	out := m.viewBody()
	require.Contains(t, out, "║", "filter mode must NOT disable dual-pane")
}

func TestCursorChange_DetailsUpdates(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "alpha-task", "beta-task")
	m.width = 120
	m.cursor = 0
	out0 := m.viewBody()
	require.Contains(t, out0, "alpha-task")
	m.cursor = 1
	out1 := m.viewBody()
	require.Contains(t, out1, "beta-task")
	require.NotEqual(t, out0, out1, "cursor change must alter details pane")
}

func TestViewDetails_FieldOrder(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tg, err := svc.UpsertTag(ctx, "urgent")
	require.NoError(t, err)
	start := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
	due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	pin := task.NewDate(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC))
	tk, err := svc.AddTask(ctx, app.AddTaskInput{
		Title:     "T",
		Notes:     "N",
		AreaID:    &a.ID,
		StartDate: &start,
		Deadline:  &due,
		Tags:      []id.ID{tg.ID},
	})
	require.NoError(t, err)
	tk.PinnedToday = &pin
	tk.Someday = true
	m.tasks = []task.Task{tk}
	m.cursor = 0
	m.areaNamesByID[a.ID] = "work"
	m.tagNamesByID[tg.ID] = "urgent"
	out := viewDetails(m, 60)
	iStatus := strings.Index(out, "Status:")
	iStart := strings.Index(out, "Start:")
	iDue := strings.Index(out, "Due:")
	iPinned := strings.Index(out, "Pinned:")
	iArea := strings.Index(out, "Area:")
	iTags := strings.Index(out, "Tags:")
	iSomeday := strings.Index(out, "Someday")
	require.Less(t, iStatus, iStart)
	require.Less(t, iStart, iDue)
	require.Less(t, iDue, iPinned)
	require.Less(t, iPinned, iArea)
	require.Less(t, iArea, iTags)
	require.Less(t, iTags, iSomeday)
}

func TestLayout_EditorFullPaneIgnoresWidth(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 200
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm := m2.(Model)
	require.Equal(t, screenEditor, mm.screen)
	out := mm.View()
	require.NotContains(t, out, "║", "editor screen must NOT render double-pane border")
}

func TestLayout_HelpFullPaneIgnoresWidth(t *testing.T) {
	m := newTestModel(t)
	m.width = 200
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	mm := m2.(Model)
	require.Equal(t, screenHelp, mm.screen)
	out := mm.View()
	require.NotContains(t, out, "║", "help screen must NOT render double-pane border")
}

func TestLayout_ConfirmStacksBelowDual(t *testing.T) {
	m, _, tasks := setupModelWithInboxTasks(t, "a", "b", "c", "d", "e")
	m.width = 120
	for _, tk := range tasks {
		m.selected[tk.ID] = struct{}{}
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	mm := m2.(Model)
	require.NotNil(t, mm.confirm)
	out := mm.View()
	require.Contains(t, out, "║", "dual-pane border must still render")
	require.Contains(t, out, "Complete 5 tasks?", "confirm modal must render")
	iBorder := strings.Index(out, "║")
	iModal := strings.Index(out, "Complete 5 tasks?")
	require.Less(t, iBorder, iModal, "modal must stack below split body")
}

func TestLayout_QuickEntryStacksBelowDual(t *testing.T) {
	m, _, _ := setupModelWithInboxTasks(t, "x")
	m.width = 120
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	mm := m2.(Model)
	require.Equal(t, screenQuickEntry, mm.screen)
	out := mm.View()
	require.Contains(t, out, "║", "dual-pane must still render under quick-entry")
	require.Contains(t, out, "Quick Entry", "quick-entry modal must render")
}

// TestProp_LayoutModeExclusion verifies CP-1 (REQ-1.1, 1.2, 1.3, 1.6, 1.7):
// the dual-pane border `║` appears in viewBody iff isDualPane(m) returns true.
// Single-pane and dual-pane are mutually exclusive.
func TestProp_LayoutModeExclusion(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		width := rapid.IntRange(0, 300).Draw(rt, "width")
		screen := rapid.SampledFrom([]screenKind{
			screenList, screenEditor, screenHelp, screenQuickEntry,
		}).Draw(rt, "screen")

		m := bareTestModel()
		m.width = width
		m.screen = screen
		// Seed with one task so viewBody has content to render.
		m.tasks = []task.Task{{ID: id.New(), Title: "x", Status: task.StatusOpen}}
		m.cursor = 0

		out := m.viewBody()
		hasBorder := strings.Contains(out, "║")
		if isDualPane(m) {
			require.True(rt, hasBorder, "dual-pane mode must render ║")
		} else {
			require.False(rt, hasBorder, "single-pane mode must NOT render ║")
		}
	})
}

// TestProp_PaneWidthArithmetic verifies CP-2 (REQ-1.4): for any total width,
// paneWidths returns (list, details) such that list + 1 + details == total.
func TestProp_PaneWidthArithmetic(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(100, 400).Draw(rt, "width")
		m := bareTestModel()
		m.width = w
		list, details := paneWidths(m)
		require.Equal(rt, w, list+1+details,
			"listW(%d) + 1 + detailsW(%d) must equal totalW(%d)", list, details, w)
		require.GreaterOrEqual(rt, list, 0, "listW must be non-negative")
		require.GreaterOrEqual(rt, details, 0, "detailsW must be non-negative")
	})
}

// TestProp_BorderRenders verifies CP-3 (REQ-1.5): when isDualPane(m) is true,
// viewBody contains the double-line vertical border `║`.
func TestProp_BorderRenders(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		// Width drawn from values ≥ config.Defaults().DualPaneMinWidth so dual-pane is active.
		w := rapid.IntRange(config.Defaults().DualPaneMinWidth, 300).Draw(rt, "width")
		m := bareTestModel()
		m.width = w
		m.screen = screenList
		m.tasks = []task.Task{{ID: id.New(), Title: "task one", Status: task.StatusOpen}}
		m.cursor = 0
		require.True(rt, isDualPane(m), "precondition: must be dual-pane")
		out := m.viewBody()
		require.Contains(rt, out, "║", "dual-pane must render double-line border")
	})
}

// TestProp_EditorHelpForceFullPane verifies CP-4 (REQ-1.6, 1.7): even at
// arbitrarily wide widths, screenEditor and screenHelp force single-pane
// layout (isDualPane(m) == false).
func TestProp_EditorHelpForceFullPane(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		screen := rapid.SampledFrom([]screenKind{screenEditor, screenHelp}).Draw(rt, "screen")
		w := rapid.IntRange(config.Defaults().DualPaneMinWidth, 400).Draw(rt, "width")
		m := bareTestModel()
		m.width = w
		m.screen = screen
		require.False(rt, isDualPane(m),
			"editor/help screens must force single-pane regardless of width=%d", w)
	})
}

// TestProp_FieldOrderInvariant verifies CP-5 (REQ-2.1..2.10, 2.12): when
// every optional field is populated, the rendered details pane preserves
// the fixed order Status < Start < Due < Pinned < Area < Tags < Someday.
func TestProp_FieldOrderInvariant(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		title := rapid.StringMatching(`[a-zA-Z ]{1,12}`).Draw(rt, "title")
		// Build a task with every optional positional field populated.
		aid := id.New()
		tag := id.New()
		start := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
		due := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
		pin := task.NewDate(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC))
		tk := task.Task{
			ID:          id.New(),
			Title:       title,
			Status:      task.StatusOpen,
			AreaID:      &aid,
			Tags:        []id.ID{tag},
			StartDate:   &start,
			Deadline:    &due,
			PinnedToday: &pin,
			Someday:     true,
		}
		m := bareTestModel()
		m.tasks = []task.Task{tk}
		m.cursor = 0
		m.areaNamesByID[aid] = "work"
		m.tagNamesByID[tag] = "urgent"
		out := viewDetails(m, 60)

		iStatus := strings.Index(out, "Status:")
		iStart := strings.Index(out, "Start:")
		iDue := strings.Index(out, "Due:")
		iPinned := strings.Index(out, "Pinned:")
		iArea := strings.Index(out, "Area:")
		iTags := strings.Index(out, "Tags:")
		iSomeday := strings.Index(out, "Someday")

		require.NotEqual(rt, -1, iStatus, "Status must be present")
		require.NotEqual(rt, -1, iStart, "Start must be present")
		require.NotEqual(rt, -1, iDue, "Due must be present")
		require.NotEqual(rt, -1, iPinned, "Pinned must be present")
		require.NotEqual(rt, -1, iArea, "Area must be present")
		require.NotEqual(rt, -1, iTags, "Tags must be present")
		require.NotEqual(rt, -1, iSomeday, "Someday must be present")

		require.Less(rt, iStatus, iStart, "Status must precede Start")
		require.Less(rt, iStart, iDue, "Start must precede Due")
		require.Less(rt, iDue, iPinned, "Due must precede Pinned")
		require.Less(rt, iPinned, iArea, "Pinned must precede Area")
		require.Less(rt, iArea, iTags, "Area must precede Tags")
		require.Less(rt, iTags, iSomeday, "Tags must precede Someday")
	})
}

// TestProp_NilFieldsOmitted verifies CP-6 (REQ-2.11): each label is rendered
// iff its corresponding field is set. Randomly nil/empty fields must be
// absent from the output.
func TestProp_NilFieldsOmitted(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		hasStart := rapid.Bool().Draw(rt, "hasStart")
		hasDue := rapid.Bool().Draw(rt, "hasDue")
		hasPin := rapid.Bool().Draw(rt, "hasPin")
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		hasTags := rapid.Bool().Draw(rt, "hasTags")
		someday := rapid.Bool().Draw(rt, "someday")

		tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen}
		if hasStart {
			d := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
			tk.StartDate = &d
		}
		if hasDue {
			d := task.NewDate(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
			tk.Deadline = &d
		}
		if hasPin {
			d := task.NewDate(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC))
			tk.PinnedToday = &d
		}
		if hasArea {
			a := id.New()
			tk.AreaID = &a
		}
		if hasProject {
			p := id.New()
			tk.ProjectID = &p
		}
		if hasTags {
			tk.Tags = []id.ID{id.New()}
		}
		tk.Someday = someday

		m := bareTestModel()
		m.tasks = []task.Task{tk}
		m.cursor = 0
		out := viewDetails(m, 60)

		// Labels present iff corresponding field is set.
		require.Equal(rt, hasStart, strings.Contains(out, "Start:"))
		require.Equal(rt, hasDue, strings.Contains(out, "Due:"))
		require.Equal(rt, hasPin, strings.Contains(out, "Pinned:"))
		require.Equal(rt, hasArea, strings.Contains(out, "Area:"))
		require.Equal(rt, hasProject, strings.Contains(out, "Project:"))
		require.Equal(rt, hasTags, strings.Contains(out, "Tags:"))
		require.Equal(rt, someday, strings.Contains(out, "Someday"))
	})
}

// TestProp_NotesTruncationCorrect verifies CP-7 (REQ-2.3): notes are
// wrapped/truncated to at most config.Defaults().NotesMaxLines (8) lines, with an
// ellipsis line appended iff truncation occurred.
func TestProp_NotesTruncationCorrect(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		nLines := rapid.IntRange(1, 30).Draw(rt, "nLines")
		// Lines short enough to avoid soft-wrap inflation at width=60.
		lines := make([]string, nLines)
		for i := range lines {
			lines[i] = fmt.Sprintf("note line %d", i)
		}
		notes := strings.Join(lines, "\n")
		out := wrapAndTruncate(notes, 60, config.Defaults().NotesMaxLines)
		outLines := strings.Split(out, "\n")

		if nLines > config.Defaults().NotesMaxLines {
			require.LessOrEqual(rt, len(outLines), config.Defaults().NotesMaxLines+1,
				"truncated output must be ≤ maxLines+1 lines (got %d for input %d)", len(outLines), nLines)
			require.Contains(rt, out, "…", "truncated output must contain ellipsis")
		} else {
			require.NotContains(rt, out, "…",
				"non-truncated output (input %d ≤ %d) must NOT contain ellipsis", nLines, config.Defaults().NotesMaxLines)
		}
	})
}

// TestProp_PlaceholderOnEmptyCursor verifies CP-8 (REQ-3.1, 3.2): viewDetails
// renders "(no task selected)" iff cursorTask(m) returns nil — i.e., when
// the visible task list is empty or the cursor is out of range.
func TestProp_PlaceholderOnEmptyCursor(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(0, 5).Draw(rt, "n")
		cursor := rapid.IntRange(-5, 20).Draw(rt, "cursor")

		tasks := make([]task.Task, n)
		for i := range tasks {
			tasks[i] = task.Task{ID: id.New(), Title: fmt.Sprintf("t-%d", i), Status: task.StatusOpen}
		}
		m := bareTestModel()
		m.tasks = tasks
		m.cursor = cursor

		out := viewDetails(m, 60)
		inRange := n > 0 && cursor >= 0 && cursor < n
		if inRange {
			require.NotContains(rt, out, "(no task selected)",
				"in-range cursor (n=%d cursor=%d) must NOT show placeholder", n, cursor)
		} else {
			require.Contains(rt, out, "(no task selected)",
				"empty list or out-of-range cursor (n=%d cursor=%d) must show placeholder", n, cursor)
		}
	})
}

// TestProp_NoRepoAccessInView verifies CP-9 (REQ-4.2): viewDetails is a pure
// read of the model and name caches — it never panics for arbitrary
// populated/missing cache entries, and by inspection never calls the
// Repository (all name resolutions go through map lookups via resolveName).
func TestProp_NoRepoAccessInView(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		hasArea := rapid.Bool().Draw(rt, "hasArea")
		hasTag := rapid.Bool().Draw(rt, "hasTag")
		hasProject := rapid.Bool().Draw(rt, "hasProject")
		cacheHasArea := rapid.Bool().Draw(rt, "cacheHasArea")
		cacheHasTag := rapid.Bool().Draw(rt, "cacheHasTag")
		cacheHasProject := rapid.Bool().Draw(rt, "cacheHasProject")

		// Construct a Model without any service — viewDetails must not need it.
		m := Model{
			theme:            NewTheme(),
			selected:         make(map[id.ID]struct{}),
			tagNamesByID:     make(map[id.ID]string),
			areaNamesByID:    make(map[id.ID]string),
			projectsByID:     make(map[id.ID]project.Project),
			headingNamesByID: make(map[id.ID]string),
		}
		tk := task.Task{ID: id.New(), Title: "x", Status: task.StatusOpen}
		if hasArea {
			aid := id.New()
			tk.AreaID = &aid
			if cacheHasArea {
				m.areaNamesByID[aid] = "work"
			}
		}
		if hasTag {
			tg := id.New()
			tk.Tags = []id.ID{tg}
			if cacheHasTag {
				m.tagNamesByID[tg] = "urgent"
			}
		}
		if hasProject {
			pid := id.New()
			tk.ProjectID = &pid
			if cacheHasProject {
				m.projectsByID[pid] = project.Project{Name: "todushka"}
			}
		}
		m.tasks = []task.Task{tk}
		m.cursor = 0
		// m.service is nil — any Repo() access would panic. NotPanics confirms
		// viewDetails is repository-free.
		require.NotPanics(rt, func() { _ = viewDetails(m, 60) })
	})
}

// TestProp_NameCachePopulation verifies CP-10 (REQ-4.1): after running the
// fetchNameCache Cmd, every referenced area/tag ID across the input tasks
// appears in the resulting nameCacheLoadedMsg maps.
func TestProp_NameCachePopulation(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		_, svc := newTestModelWithServiceRapid(rt)
		ctx := context.Background()

		nTags := rapid.IntRange(1, 3).Draw(rt, "nTags")
		nTasks := rapid.IntRange(1, 3).Draw(rt, "nTasks")
		// Allocate a shared area used by every task.
		a, err := svc.AddArea(ctx, "work")
		require.NoError(rt, err)
		// Create a small pool of real tags.
		tagIDs := make([]id.ID, 0, nTags)
		for i := 0; i < nTags; i++ {
			tg, err := svc.UpsertTag(ctx, fmt.Sprintf("tag-%d", i))
			require.NoError(rt, err)
			tagIDs = append(tagIDs, tg.ID)
		}

		tasks := make([]task.Task, 0, nTasks)
		referencedTags := make(map[id.ID]struct{})
		for i := 0; i < nTasks; i++ {
			// Each task references a random tag from the pool.
			ti := rapid.IntRange(0, nTags-1).Draw(rt, fmt.Sprintf("taskTag-%d", i))
			tag := tagIDs[ti]
			tk, err := svc.AddTask(ctx, app.AddTaskInput{
				Title:  fmt.Sprintf("t-%d", i),
				AreaID: &a.ID,
				Tags:   []id.ID{tag},
			})
			require.NoError(rt, err)
			tasks = append(tasks, tk)
			referencedTags[tag] = struct{}{}
		}

		cmd := fetchNameCache(svc, tasks)
		require.NotNil(rt, cmd)
		msg := cmd()
		res, ok := msg.(nameCacheLoadedMsg)
		require.True(rt, ok, "fetchNameCache must emit nameCacheLoadedMsg")
		// Every referenced area appears in res.areas.
		require.Equal(rt, "work", res.areas[a.ID],
			"area referenced by all tasks must be present in cache")
		// Every referenced tag appears in res.tags.
		for tg := range referencedTags {
			_, has := res.tags[tg]
			require.True(rt, has, "tag %v must be in resolved cache", tg)
		}
	})
}

// TestProp_ShortIDFallback verifies CP-11 (REQ-4.3): when a referenced ID
// has no cache entry, viewDetails renders the short-ID via id.Short.
func TestProp_ShortIDFallback(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		width := rapid.IntRange(40, 120).Draw(rt, "width")
		tag := id.New()
		tk := task.Task{
			ID:     id.New(),
			Title:  "x",
			Status: task.StatusOpen,
			Tags:   []id.ID{tag},
		}
		m := bareTestModel()
		m.tasks = []task.Task{tk}
		m.cursor = 0
		// Deliberately do NOT populate tagNamesByID for tag.
		out := viewDetails(m, width)
		require.Contains(rt, out, "Tags:", "tag label must render")
		require.Contains(rt, out, id.Short(tag),
			"missing cache entry must fall back to short-ID")
	})
}

// TestProp_CursorChangeReflects verifies CP-12 (REQ-5.1): moving the cursor
// to a different task produces different details-pane output.
func TestProp_CursorChangeReflects(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(2, 6).Draw(rt, "n")
		// Distinct titles so different cursors yield different output.
		tasks := make([]task.Task, n)
		for i := range tasks {
			tasks[i] = task.Task{
				ID:     id.New(),
				Title:  fmt.Sprintf("task-%d-uniq", i),
				Status: task.StatusOpen,
			}
		}
		m := bareTestModel()
		m.tasks = tasks
		i := rapid.IntRange(0, n-1).Draw(rt, "i")
		j := rapid.IntRange(0, n-1).Filter(func(v int) bool { return v != i }).Draw(rt, "j")

		m.cursor = i
		outI := viewDetails(m, 60)
		m.cursor = j
		outJ := viewDetails(m, 60)
		require.NotEqual(rt, outI, outJ,
			"different cursors (i=%d j=%d) must produce different output", i, j)
	})
}

// TestProp_FilterPreservesDualPane verifies CP-13 (REQ-6.1): when in a wide
// terminal with filter mode active, isDualPane(m) remains true regardless of
// the filter query content.
func TestProp_FilterPreservesDualPane(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(config.Defaults().DualPaneMinWidth, 300).Draw(rt, "width")
		query := rapid.StringMatching(`[a-zA-Z0-9 ]{0,10}`).Draw(rt, "query")
		m := bareTestModel()
		m.width = w
		m.screen = screenList
		m.filtering = true
		m.filterQuery = query
		require.True(rt, isDualPane(m),
			"filter mode must NOT disable dual-pane (width=%d, query=%q)", w, query)
	})
}

// TestProp_ConfirmStacksBelowDual verifies CP-14 (REQ-1.8): when a confirm
// modal is active in dual-pane mode, View() output contains both the
// double-line border and the modal text, with the modal positioned after
// (below) the border.
func TestProp_ConfirmStacksBelowDual(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		w := rapid.IntRange(config.Defaults().DualPaneMinWidth, 300).Draw(rt, "width")
		action := rapid.SampledFrom([]bulkAction{
			bulkActionComplete, bulkActionCancel, bulkActionDelete, bulkActionPin,
		}).Draw(rt, "action")
		nIDs := rapid.IntRange(1, 10).Draw(rt, "nIDs")

		ids := make([]id.ID, nIDs)
		for i := range ids {
			ids[i] = id.New()
		}
		m := bareTestModel()
		m.width = w
		m.screen = screenList
		// Seed at least one task so viewBody has content.
		m.tasks = []task.Task{{ID: id.New(), Title: "task one", Status: task.StatusOpen}}
		m.cursor = 0
		m.confirm = &confirmState{action: action, ids: ids}

		out := m.View()
		require.Contains(rt, out, "║",
			"dual-pane border must still render with confirm modal active")
		label := action.label()
		require.Contains(rt, out, label,
			"modal must render its action label %q", label)

		iBorder := strings.Index(out, "║")
		iModal := strings.Index(out, label)
		require.Less(rt, iBorder, iModal,
			"modal (idx=%d) must stack below the dual-pane body (border idx=%d)", iModal, iBorder)
	})
}

// newTestModelWithServiceRapid is the rapid.T flavor of newTestModelWithService.
// Reuses setupRapidModel with zero titles to obtain a fresh service + model.
func newTestModelWithServiceRapid(rt *rapid.T) (Model, *app.Service) {
	rt.Helper()
	m, svc, _ := setupRapidModel(rt)
	return m, svc
}

func TestUpdate_TasksLoadedDispatchesNameCacheFetch(t *testing.T) {
	m, svc := newTestModelWithService(t)
	ctx := context.Background()
	a, err := svc.AddArea(ctx, "work")
	require.NoError(t, err)
	tk, err := svc.AddTask(ctx, app.AddTaskInput{Title: "x", AreaID: &a.ID})
	require.NoError(t, err)
	m2, cmd := m.Update(tasksLoadedMsg{tasks: []task.Task{tk}})
	mm := m2.(Model)
	require.Equal(t, []task.Task{tk}, mm.tasks)
	require.NotNil(t, cmd, "tasksLoadedMsg must trigger fetchNameCache Cmd")
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			if sub == nil {
				continue
			}
			if nm, ok := sub().(nameCacheLoadedMsg); ok {
				require.Equal(t, "work", nm.areas[a.ID])
				return
			}
		}
		t.Fatal("no nameCacheLoadedMsg in batch")
	}
	if nm, ok := msg.(nameCacheLoadedMsg); ok {
		require.Equal(t, "work", nm.areas[a.ID])
		return
	}
	t.Fatalf("unexpected msg type: %T", msg)
}

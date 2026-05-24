package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/repeat"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/jtprogru/todushka/internal/storage/fakes"
	"github.com/stretchr/testify/require"
)

type fixedClock struct{ now time.Time }

func (f fixedClock) Now() time.Time { return f.now }

func newService(t *testing.T, now time.Time) *Service {
	t.Helper()
	r := fakes.New()
	return New(r, fixedClock{now: now})
}

func TestSystemClock_NowIsLocal(t *testing.T) {
	x := SystemClock{}.Now()
	require.Equal(t, time.Local, x.Location())
}

func TestService_AddTaskCreatesInRepo(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	tk, err := s.AddTask(ctx, AddTaskInput{Title: "milk"})
	require.NoError(t, err)
	require.Equal(t, task.StatusOpen, tk.Status)
	got, err := s.repo.TaskGet(ctx, tk.ID)
	require.NoError(t, err)
	require.Equal(t, tk, got)
}

func TestService_AddTaskRejectsEmptyTitle(t *testing.T) {
	s := newService(t, time.Now())
	_, err := s.AddTask(context.Background(), AddTaskInput{Title: "  "})
	require.ErrorIs(t, err, ErrEmptyTitle)
}

func TestService_AddTaskRejectsDeadlineBeforeStart(t *testing.T) {
	s := newService(t, time.Now())
	start := task.NewDate(time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC))
	deadline := task.NewDate(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC))
	_, err := s.AddTask(context.Background(), AddTaskInput{
		Title: "x", StartDate: &start, Deadline: &deadline,
	})
	require.ErrorIs(t, err, ErrDeadlineBeforeStart)
}

func TestService_CompleteTaskFiresRepeat(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	rule := repeat.Rule{Kind: repeat.KindDaily}
	tk, err := s.AddTask(ctx, AddTaskInput{Title: "exercise", Repeat: &rule})
	require.NoError(t, err)

	next, err := s.CompleteTask(ctx, tk.ID)
	require.NoError(t, err)
	require.NotNil(t, next)
	require.Equal(t, task.StatusOpen, next.Status)
	require.NotNil(t, next.StartDate)
	require.Equal(t, "2026-05-26", next.StartDate.Format("2006-01-02"))

	orig, err := s.repo.TaskGet(ctx, tk.ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, orig.Status)
	require.NotNil(t, orig.CompletedAt)
}

func TestService_CompleteTaskNoRepeatNoSuccessor(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	tk, err := s.AddTask(ctx, AddTaskInput{Title: "x"})
	require.NoError(t, err)
	next, err := s.CompleteTask(ctx, tk.ID)
	require.NoError(t, err)
	require.Nil(t, next)
}

func TestService_CancelTaskNoSuccessor(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	rule := repeat.Rule{Kind: repeat.KindDaily}
	tk, err := s.AddTask(ctx, AddTaskInput{Title: "x", Repeat: &rule})
	require.NoError(t, err)

	require.NoError(t, s.CancelTask(ctx, tk.ID))

	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{IncludeDeleted: true})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
}

func TestService_DeleteAreaMovesChildren(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	a, err := s.AddArea(ctx, "Work")
	require.NoError(t, err)
	p, err := s.AddProject(ctx, AddProjectInput{Name: "PR review", AreaID: &a.ID})
	require.NoError(t, err)
	aid := a.ID
	orphan, err := s.AddTask(ctx, AddTaskInput{Title: "standalone", AreaID: &aid})
	require.NoError(t, err)

	require.NoError(t, s.DeleteArea(ctx, a.ID, true))

	got, err := s.repo.ProjectGet(ctx, p.ID)
	require.NoError(t, err)
	require.Nil(t, got.AreaID)

	gotTask, err := s.repo.TaskGet(ctx, orphan.ID)
	require.NoError(t, err)
	require.Nil(t, gotTask.AreaID)

	_, err = s.repo.AreaGet(ctx, a.ID)
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestService_DeleteAreaWithoutConfirm(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	a, _ := s.AddArea(ctx, "Work")
	_, _ = s.AddProject(ctx, AddProjectInput{Name: "x", AreaID: &a.ID})
	err := s.DeleteArea(ctx, a.ID, false)
	require.ErrorIs(t, err, ErrAreaNotEmpty)
}

func TestService_AutoCloseProject(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	p, _ := s.AddProject(ctx, AddProjectInput{Name: "x", AutoClose: true})
	pid := p.ID
	tk, _ := s.AddTask(ctx, AddTaskInput{Title: "a", ProjectID: &pid})

	// Project still open: tasks still open
	got, _ := s.repo.ProjectGet(ctx, p.ID)
	require.Equal(t, project.StatusOpen, got.Status)

	// Complete the only task: project should auto-close on next edit
	_, _ = s.CompleteTask(ctx, tk.ID)
	require.NoError(t, s.EditProject(ctx, got))

	got, _ = s.repo.ProjectGet(ctx, p.ID)
	require.Equal(t, project.StatusCompleted, got.Status)
}

func TestService_QuickEntryEmptyInput(t *testing.T) {
	s := newService(t, time.Now())
	_, err := s.QuickEntry(context.Background(), "   ")
	require.ErrorIs(t, err, ErrEmptyInput)
}

func TestService_QuickEntryWithTagAndDate(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.Local)
	s := newService(t, now)
	tk, err := s.QuickEntry(ctx, "buy milk #shop @today !2026-06-01")
	require.NoError(t, err)
	require.Equal(t, "buy milk", tk.Title)
	require.Len(t, tk.Tags, 1)
	require.NotNil(t, tk.StartDate)
	require.Equal(t, "2026-05-25", tk.StartDate.Format("2006-01-02"))
	require.NotNil(t, tk.Deadline)
	require.Equal(t, "2026-06-01", tk.Deadline.Format("2006-01-02"))
}

func TestService_QuickEntryAmbiguousProject(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	// Force two projects with identical names via direct repo writes
	now := time.Now()
	require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{ID: id.New(), Name: "work", Status: project.StatusOpen, CreatedAt: now, UpdatedAt: now}))
	require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{ID: id.New(), Name: "WORK", Status: project.StatusOpen, CreatedAt: now, UpdatedAt: now}))

	_, err := s.QuickEntry(ctx, "review @work")
	require.ErrorIs(t, err, ErrAmbiguousProject)
}

func TestService_QuickEntryUnknownProject(t *testing.T) {
	s := newService(t, time.Now())
	_, err := s.QuickEntry(context.Background(), "review @nope")
	require.ErrorIs(t, err, ErrUnknownProject)
}

func TestService_QuickEntryInvalidDate(t *testing.T) {
	s := newService(t, time.Now())
	_, err := s.QuickEntry(context.Background(), "x !2026-13-99")
	require.Error(t, err)
}

func TestService_ListInbox(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	s := newService(t, now)
	// Inbox task: open, no area, no project, no start_date, no someday
	t1, _ := s.AddTask(ctx, AddTaskInput{Title: "inbox"})
	// Anytime: has area
	a, _ := s.AddArea(ctx, "Work")
	aid := a.ID
	_, _ = s.AddTask(ctx, AddTaskInput{Title: "work_anytime", AreaID: &aid})
	// Someday: open with someday flag
	_, _ = s.AddTask(ctx, AddTaskInput{Title: "someday", Someday: true})

	got, err := s.ListInbox(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, t1.ID, got[0].ID)
}

func TestService_ListTodayUsesEngine(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	s := newService(t, now)
	startToday := task.NewDate(now)
	startTomorrow := task.NewDate(now.AddDate(0, 0, 1))
	_, _ = s.AddTask(ctx, AddTaskInput{Title: "today", StartDate: &startToday})
	_, _ = s.AddTask(ctx, AddTaskInput{Title: "tomorrow", StartDate: &startTomorrow})
	got, err := s.ListToday(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "today", got[0].Title)
}

func TestService_ExportImportRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	a, _ := s.AddArea(ctx, "Work")
	_, _ = s.AddProject(ctx, AddProjectInput{Name: "x", AreaID: &a.ID})
	_, _ = s.QuickEntry(ctx, "buy milk #shop @today")

	var buf bytes.Buffer
	require.NoError(t, s.ExportJSON(ctx, &buf))
	exported := buf.String()
	require.Contains(t, exported, `"schema_version": 1`)

	// Import into a fresh service
	s2 := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	_, err := s2.ImportJSON(ctx, strings.NewReader(exported))
	require.NoError(t, err)

	// Both should produce the same snapshot
	snap1, _ := s.ExportSnapshot(ctx)
	snap2, _ := s2.ExportSnapshot(ctx)
	require.Equal(t, len(snap1.Tasks), len(snap2.Tasks))
	require.Equal(t, len(snap1.Tags), len(snap2.Tags))
	require.Equal(t, len(snap1.Areas), len(snap2.Areas))
	require.Equal(t, len(snap1.Projects), len(snap2.Projects))
}

func TestService_ImportRejectsFutureSchema(t *testing.T) {
	s := newService(t, time.Now())
	payload := `{"schema_version": 999, "exported_at": "2026-05-25T10:00:00Z"}`
	_, err := s.ImportJSON(context.Background(), strings.NewReader(payload))
	require.ErrorIs(t, err, ErrSchemaTooNew)
}

func TestService_ImportRejectsMalformedJSON(t *testing.T) {
	s := newService(t, time.Now())
	_, err := s.ImportJSON(context.Background(), strings.NewReader("not json"))
	require.ErrorIs(t, err, ErrInvalidImport)
}

func TestService_FindTaskByShort(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	tk, _ := s.AddTask(ctx, AddTaskInput{Title: "x"})
	prefix := id.Short(tk.ID)
	got, err := s.FindTaskByShort(ctx, prefix)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, tk.ID, got[0].ID)
}

func TestService_PinUnpinToday(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	s := newService(t, now)
	future := task.NewDate(now.AddDate(0, 0, 30))
	tk, _ := s.AddTask(ctx, AddTaskInput{Title: "x", StartDate: &future})

	// Initially not in Today
	got, _ := s.ListToday(ctx)
	require.Empty(t, got)

	require.NoError(t, s.PinToToday(ctx, tk.ID))
	got, _ = s.ListToday(ctx)
	require.Len(t, got, 1)

	require.NoError(t, s.UnpinFromToday(ctx, tk.ID))
	got, _ = s.ListToday(ctx)
	require.Empty(t, got)
}

func TestService_TagRenameCollision(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	a, _ := s.UpsertTag(ctx, "shop")
	_, _ = s.UpsertTag(ctx, "home")
	err := s.RenameTag(ctx, a.ID, "Home")
	require.ErrorIs(t, err, ErrTagAlreadyExists)
}

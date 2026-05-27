package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/repeat"
	"github.com/jtprogru/todushka/internal/domain/tag"
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

// readOnlyRepo wraps an InMemRepo and returns storage.ErrReadOnly on every
// write method. Used to verify service-layer fail-fast behavior under a
// read-only repository (REQ-6.5).
type readOnlyRepo struct{ inner storage.Repository }

func (r *readOnlyRepo) Close() error                                 { return r.inner.Close() }
func (r *readOnlyRepo) ReadOnly() bool                               { return true }
func (r *readOnlyRepo) SchemaVersion(c context.Context) (int, error) { return r.inner.SchemaVersion(c) }
func (r *readOnlyRepo) Migrate(c context.Context, t int) error       { return storage.ErrReadOnly }

func (r *readOnlyRepo) TaskCreate(c context.Context, t task.Task) error { return storage.ErrReadOnly }
func (r *readOnlyRepo) TaskGet(c context.Context, i id.ID) (task.Task, error) {
	return r.inner.TaskGet(c, i)
}
func (r *readOnlyRepo) TaskList(c context.Context, f storage.TaskFilter) ([]task.Task, error) {
	return r.inner.TaskList(c, f)
}
func (r *readOnlyRepo) TaskUpdate(c context.Context, t task.Task) error { return storage.ErrReadOnly }
func (r *readOnlyRepo) TaskDelete(c context.Context, i id.ID, soft bool) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) TaskMatchShort(c context.Context, p string) ([]task.Task, error) {
	return r.inner.TaskMatchShort(c, p)
}

func (r *readOnlyRepo) ProjectCreate(c context.Context, p project.Project) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) ProjectGet(c context.Context, i id.ID) (project.Project, error) {
	return r.inner.ProjectGet(c, i)
}
func (r *readOnlyRepo) ProjectList(c context.Context, f storage.ProjectFilter) ([]project.Project, error) {
	return r.inner.ProjectList(c, f)
}
func (r *readOnlyRepo) ProjectUpdate(c context.Context, p project.Project) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) ProjectDelete(c context.Context, i id.ID, soft bool) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) ProjectFindByName(c context.Context, n string) ([]project.Project, error) {
	return r.inner.ProjectFindByName(c, n)
}
func (r *readOnlyRepo) HeadingCreate(c context.Context, h project.Heading) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) HeadingUpdate(c context.Context, h project.Heading) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) HeadingDelete(c context.Context, i id.ID) error { return storage.ErrReadOnly }
func (r *readOnlyRepo) HeadingList(c context.Context, pid id.ID) ([]project.Heading, error) {
	return r.inner.HeadingList(c, pid)
}

func (r *readOnlyRepo) AreaCreate(c context.Context, a area.Area) error { return storage.ErrReadOnly }
func (r *readOnlyRepo) AreaGet(c context.Context, i id.ID) (area.Area, error) {
	return r.inner.AreaGet(c, i)
}
func (r *readOnlyRepo) AreaList(c context.Context) ([]area.Area, error) { return r.inner.AreaList(c) }
func (r *readOnlyRepo) AreaUpdate(c context.Context, a area.Area) error { return storage.ErrReadOnly }
func (r *readOnlyRepo) AreaDelete(c context.Context, i id.ID, soft bool) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) AreaFindByNormalized(c context.Context, n string) (area.Area, error) {
	return r.inner.AreaFindByNormalized(c, n)
}

func (r *readOnlyRepo) TagCreate(c context.Context, t tag.Tag) error { return storage.ErrReadOnly }
func (r *readOnlyRepo) TagUpsert(c context.Context, n string) (tag.Tag, error) {
	return tag.Tag{}, storage.ErrReadOnly
}
func (r *readOnlyRepo) TagGet(c context.Context, i id.ID) (tag.Tag, error) {
	return r.inner.TagGet(c, i)
}
func (r *readOnlyRepo) TagList(c context.Context) ([]tag.Tag, error) { return r.inner.TagList(c) }
func (r *readOnlyRepo) TagRename(c context.Context, i id.ID, n string) error {
	return storage.ErrReadOnly
}
func (r *readOnlyRepo) TagDelete(c context.Context, i id.ID) error { return storage.ErrReadOnly }

// ─── BL-5: DeleteProject ─────────────────────────────────────────────────

func TestService_DeleteProject_NotFound(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	err := s.DeleteProject(ctx, id.New(), true)
	require.ErrorIs(t, err, ErrProjectNotFound)
}

func TestService_DeleteProject_NonEmpty_NoConfirm(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	p, err := s.AddProject(ctx, AddProjectInput{Name: "PRs"})
	require.NoError(t, err)
	pid := p.ID
	tk, err := s.AddTask(ctx, AddTaskInput{Title: "review", ProjectID: &pid})
	require.NoError(t, err)

	err = s.DeleteProject(ctx, pid, false)
	require.ErrorIs(t, err, ErrProjectNotEmpty)

	// Project unchanged.
	got, err := s.repo.ProjectGet(ctx, pid)
	require.NoError(t, err)
	require.Nil(t, got.DeletedAt)
	// Task unchanged.
	gotTask, err := s.repo.TaskGet(ctx, tk.ID)
	require.NoError(t, err)
	require.NotNil(t, gotTask.ProjectID)
	require.Equal(t, pid, *gotTask.ProjectID)
}

func TestService_DeleteProject_NonEmpty_Confirm(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	p, err := s.AddProject(ctx, AddProjectInput{Name: "PRs"})
	require.NoError(t, err)
	pid := p.ID
	h, err := s.AddHeading(ctx, pid, "Section A")
	require.NoError(t, err)
	hid := h.ID
	tk1, err := s.AddTask(ctx, AddTaskInput{Title: "t1", ProjectID: &pid})
	require.NoError(t, err)
	tk2, err := s.AddTask(ctx, AddTaskInput{Title: "t2", ProjectID: &pid, HeadingID: &hid})
	require.NoError(t, err)

	require.NoError(t, s.DeleteProject(ctx, pid, true))

	// Tasks: ProjectID=nil, HeadingID=nil.
	g1, _ := s.repo.TaskGet(ctx, tk1.ID)
	require.Nil(t, g1.ProjectID)
	require.Nil(t, g1.HeadingID)
	g2, _ := s.repo.TaskGet(ctx, tk2.ID)
	require.Nil(t, g2.ProjectID)
	require.Nil(t, g2.HeadingID)

	// Project soft-deleted.
	gp, err := s.repo.ProjectGet(ctx, pid)
	require.NoError(t, err)
	require.NotNil(t, gp.DeletedAt)
}

func TestService_DeleteProject_Empty_Confirm(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC))
	p, _ := s.AddProject(ctx, AddProjectInput{Name: "PRs"})

	require.NoError(t, s.DeleteProject(ctx, p.ID, true))

	// Excluded from default ProjectList.
	list, err := s.repo.ProjectList(ctx, storage.ProjectFilter{})
	require.NoError(t, err)
	for _, x := range list {
		require.NotEqual(t, p.ID, x.ID)
	}
	// But visible with IncludeDeleted.
	allList, _ := s.repo.ProjectList(ctx, storage.ProjectFilter{IncludeDeleted: true})
	found := false
	for _, x := range allList {
		if x.ID == p.ID {
			found = true
			require.NotNil(t, x.DeletedAt)
			break
		}
	}
	require.True(t, found, "soft-deleted project must be retrievable via IncludeDeleted filter")
}

func TestService_DeleteProject_Empty_NoConfirm(t *testing.T) {
	ctx := context.Background()
	s := newService(t, time.Now())
	p, _ := s.AddProject(ctx, AddProjectInput{Name: "x"})

	// No tasks → confirm=false should succeed (empty-guard does not trigger).
	require.NoError(t, s.DeleteProject(ctx, p.ID, false))
}

func TestService_DeleteProject_ReadOnly(t *testing.T) {
	ctx := context.Background()
	inner := fakes.New()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	// Seed an empty project so ProjectGet succeeds.
	pid := id.New()
	require.NoError(t, inner.ProjectCreate(ctx, project.Project{
		ID: pid, Name: "x", Status: project.StatusOpen,
		CreatedAt: now, UpdatedAt: now,
	}))
	ro := &readOnlyRepo{inner: inner}
	s := New(ro, fixedClock{now: now})
	err := s.DeleteProject(ctx, pid, true)
	require.ErrorIs(t, err, storage.ErrReadOnly)
}

// ─── BL-5: ListProjectsSorted + CountProjectTasks ───────────────────────

func TestService_ListProjectsSorted_Basic(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	s := newService(t, now)
	mk := func(name string, pos int64) id.ID {
		pid := id.New()
		require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{
			ID: pid, Name: name, Status: project.StatusOpen, Position: pos,
			CreatedAt: now, UpdatedAt: now,
		}))
		return pid
	}
	mk("b", 10)
	mk("A", 5)
	mk("c", 5)

	got, err := s.ListProjectsSorted(ctx, nil, true)
	require.NoError(t, err)
	require.Len(t, got, 3)
	// Expected order: Position=5 first ("A", "c" by case-fold), then Position=10 ("b").
	require.Equal(t, "A", got[0].Name)
	require.Equal(t, "c", got[1].Name)
	require.Equal(t, "b", got[2].Name)
}

func TestService_ListProjectsSorted_OnlyOpen(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := newService(t, now)
	mk := func(name string, st project.Status) {
		require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{
			ID: id.New(), Name: name, Status: st, CreatedAt: now, UpdatedAt: now,
		}))
	}
	mk("open1", project.StatusOpen)
	mk("done", project.StatusCompleted)
	mk("cancel", project.StatusCancelled)

	got, err := s.ListProjectsSorted(ctx, nil, false)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "open1", got[0].Name)
}

func TestService_ListProjectsSorted_All(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := newService(t, now)
	mk := func(name string, st project.Status) {
		require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{
			ID: id.New(), Name: name, Status: st, CreatedAt: now, UpdatedAt: now,
		}))
	}
	mk("open1", project.StatusOpen)
	mk("done", project.StatusCompleted)
	mk("cancel", project.StatusCancelled)

	got, err := s.ListProjectsSorted(ctx, nil, true)
	require.NoError(t, err)
	require.Len(t, got, 3)
}

func TestService_ListProjectsSorted_AreaFilter(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := newService(t, now)
	a, _ := s.AddArea(ctx, "Work")
	aid := a.ID
	require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{
		ID: id.New(), Name: "work-proj", Status: project.StatusOpen, AreaID: &aid,
		CreatedAt: now, UpdatedAt: now,
	}))
	require.NoError(t, s.repo.ProjectCreate(ctx, project.Project{
		ID: id.New(), Name: "no-area", Status: project.StatusOpen,
		CreatedAt: now, UpdatedAt: now,
	}))

	got, err := s.ListProjectsSorted(ctx, &aid, true)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "work-proj", got[0].Name)
}

func TestService_ListProjectsSorted_ExcludesSoftDeleted(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	s := newService(t, now)
	p, _ := s.AddProject(ctx, AddProjectInput{Name: "x"})
	require.NoError(t, s.DeleteProject(ctx, p.ID, true))

	got, err := s.ListProjectsSorted(ctx, nil, true)
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestService_CountProjectTasks(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	s := newService(t, now)
	p, _ := s.AddProject(ctx, AddProjectInput{Name: "PRs"})
	pid := p.ID
	t1, _ := s.AddTask(ctx, AddTaskInput{Title: "t1", ProjectID: &pid})
	_, _ = s.AddTask(ctx, AddTaskInput{Title: "t2", ProjectID: &pid})
	t3, _ := s.AddTask(ctx, AddTaskInput{Title: "t3", ProjectID: &pid})
	// Complete one
	_, err := s.CompleteTask(ctx, t1.ID)
	require.NoError(t, err)
	// Cancel one
	require.NoError(t, s.CancelTask(ctx, t3.ID))

	open, total, err := s.CountProjectTasks(ctx, pid)
	require.NoError(t, err)
	require.Equal(t, 1, open)
	require.Equal(t, 3, total)
}

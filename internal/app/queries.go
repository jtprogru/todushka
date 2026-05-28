package app

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/domain/today"
	"github.com/jtprogru/todushka/internal/storage"
)

// ListInbox returns open tasks with no Area, no Project, no start date,
// and not marked as Someday.
func (s *Service) ListInbox(ctx context.Context) ([]task.Task, error) {
	noStart := false
	someday := false
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{
		Statuses:     []task.Status{task.StatusOpen},
		HasStartDate: &noStart,
		Someday:      &someday,
	})
	if err != nil {
		return nil, err
	}
	out := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.AreaID != nil || t.ProjectID != nil {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *Service) ListToday(ctx context.Context) ([]task.Task, error) {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{Statuses: []task.Status{task.StatusOpen}})
	if err != nil {
		return nil, err
	}
	return today.ComputeToday(tasks, s.clock.Now(), 0), nil
}

func (s *Service) ListUpcoming(ctx context.Context) ([]task.Task, error) {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{Statuses: []task.Status{task.StatusOpen}})
	if err != nil {
		return nil, err
	}
	todayStart := startOfLocalDay(s.clock.Now())
	out := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.StartDate == nil {
			continue
		}
		if !t.StartDate.Time.After(todayStart) { //nolint:staticcheck
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartDate.Time.Before(out[j].StartDate.Time) //nolint:staticcheck
	})
	return out, nil
}

func (s *Service) ListAnytime(ctx context.Context) ([]task.Task, error) {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{Statuses: []task.Status{task.StatusOpen}})
	if err != nil {
		return nil, err
	}
	todayStart := startOfLocalDay(s.clock.Now())
	out := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Someday {
			continue
		}
		if t.AreaID == nil && t.ProjectID == nil {
			continue
		}
		if t.StartDate != nil && t.StartDate.Time.After(todayStart) { //nolint:staticcheck
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func (s *Service) ListSomeday(ctx context.Context) ([]task.Task, error) {
	someday := true
	return s.repo.TaskList(ctx, storage.TaskFilter{
		Statuses: []task.Status{task.StatusOpen},
		Someday:  &someday,
	})
}

func (s *Service) ListLogbook(ctx context.Context) ([]task.Task, error) {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{
		Statuses: []task.Status{task.StatusCompleted, task.StatusCancelled},
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		return mostRecentEnd(tasks[i]).After(mostRecentEnd(tasks[j]))
	})
	return tasks, nil
}

func mostRecentEnd(t task.Task) time.Time {
	if t.CompletedAt != nil {
		return *t.CompletedAt
	}
	if t.CancelledAt != nil {
		return *t.CancelledAt
	}
	return t.UpdatedAt
}

func (s *Service) ListTrash(ctx context.Context) ([]task.Task, error) {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{IncludeDeleted: true})
	if err != nil {
		return nil, err
	}
	out := make([]task.Task, 0)
	for _, t := range tasks {
		if t.DeletedAt != nil {
			out = append(out, t)
		}
	}
	return out, nil
}

func (s *Service) ListAreas(ctx context.Context) ([]area.Area, error) {
	areas, err := s.repo.AreaList(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(areas, func(i, j int) bool {
		if areas[i].Position != areas[j].Position {
			return areas[i].Position < areas[j].Position
		}
		return strings.ToLower(areas[i].Name) < strings.ToLower(areas[j].Name)
	})
	return areas, nil
}

func (s *Service) ListProjects(ctx context.Context, areaID *id.ID) ([]project.Project, error) {
	return s.repo.ProjectList(ctx, storage.ProjectFilter{AreaID: areaID})
}

// ListProjectsSorted returns projects filtered by areaID (nil = all areas)
// sorted by Position ASC, then Name case-fold ASC. When includeAllStatuses
// is false, only Status==StatusOpen are returned. Soft-deleted projects are
// always excluded (REQ-7.1/7.2/7.3).
func (s *Service) ListProjectsSorted(ctx context.Context, areaID *id.ID, includeAllStatuses bool) ([]project.Project, error) {
	filter := storage.ProjectFilter{AreaID: areaID}
	if !includeAllStatuses {
		filter.Statuses = []project.Status{project.StatusOpen}
	}
	projects, err := s.repo.ProjectList(ctx, filter)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].Position != projects[j].Position {
			return projects[i].Position < projects[j].Position
		}
		if ni, nj := strings.ToLower(projects[i].Name), strings.ToLower(projects[j].Name); ni != nj {
			return ni < nj
		}
		// Final tiebreaker on ID keeps the order deterministic when Position
		// and Name are equal (repo iteration order is non-deterministic).
		return projects[i].ID < projects[j].ID
	})
	return projects, nil
}

// CountProjectTasks returns the number of open and total non-deleted tasks
// referencing the given project (REQ-2.1 / display counts).
func (s *Service) CountProjectTasks(ctx context.Context, pid id.ID) (open, total int, err error) {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{ProjectID: &pid})
	if err != nil {
		return 0, 0, err
	}
	for _, t := range tasks {
		total++
		if t.Status == task.StatusOpen {
			open++
		}
	}
	return open, total, nil
}

func (s *Service) GetProject(ctx context.Context, pid id.ID) (project.Project, error) {
	return s.repo.ProjectGet(ctx, pid)
}

// FindTaskByShort matches a short ID prefix (first 6 chars by convention) or
// a full ULID against the repository. Returns the unique task, or all
// candidates if the prefix is ambiguous.
func (s *Service) FindTaskByShort(ctx context.Context, shortOrFull string) ([]task.Task, error) {
	prefix := strings.ToUpper(strings.TrimSpace(shortOrFull))
	if prefix == "" {
		return nil, ErrEmptyInput
	}
	return s.repo.TaskMatchShort(ctx, prefix)
}

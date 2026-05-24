// Package app orchestrates use-cases between the domain layer and storage.
// It is the only layer that mutates state; TUI and CLI talk to Service.
package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/repeat"
	"github.com/jtprogru/todushka/internal/domain/tag"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
)

type Service struct {
	repo  storage.Repository
	clock Clock
}

func New(repo storage.Repository, clock Clock) *Service {
	if clock == nil {
		clock = SystemClock{}
	}
	return &Service{repo: repo, clock: clock}
}

// AddTaskInput holds the editable fields used to create a Task.
type AddTaskInput struct {
	Title     string
	Notes     string
	AreaID    *id.ID
	ProjectID *id.ID
	HeadingID *id.ID
	Tags      []id.ID
	StartDate *task.Date
	Deadline  *task.Date
	Someday   bool
	Checklist []task.ChecklistItem
	Repeat    *repeat.Rule
}

func (s *Service) AddTask(ctx context.Context, in AddTaskInput) (task.Task, error) {
	if err := validateDates(in.StartDate, in.Deadline); err != nil {
		return task.Task{}, err
	}
	if in.Repeat != nil {
		if err := repeat.Validate(*in.Repeat); err != nil {
			return task.Task{}, err
		}
	}
	now := s.clock.Now()
	t := task.Task{
		ID:        id.New(),
		Title:     in.Title,
		Notes:     in.Notes,
		Status:    task.StatusOpen,
		AreaID:    in.AreaID,
		ProjectID: in.ProjectID,
		HeadingID: in.HeadingID,
		Tags:      in.Tags,
		StartDate: in.StartDate,
		Deadline:  in.Deadline,
		Someday:   in.Someday,
		Checklist: in.Checklist,
		Repeat:    in.Repeat,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := t.Validate(); err != nil {
		return task.Task{}, mapTaskValidationError(err)
	}
	if err := s.repo.TaskCreate(ctx, t); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

func (s *Service) EditTask(ctx context.Context, t task.Task) error {
	if err := validateDates(t.StartDate, t.Deadline); err != nil {
		return err
	}
	if t.Repeat != nil {
		if err := repeat.Validate(*t.Repeat); err != nil {
			return err
		}
	}
	t.UpdatedAt = s.clock.Now()
	if err := t.Validate(); err != nil {
		return mapTaskValidationError(err)
	}
	return s.repo.TaskUpdate(ctx, t)
}

func (s *Service) CompleteTask(ctx context.Context, taskID id.ID) (*task.Task, error) {
	t, err := s.repo.TaskGet(ctx, taskID)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if t.Status != task.StatusOpen {
		return nil, nil
	}
	now := s.clock.Now()
	t.Status = task.StatusCompleted
	t.CompletedAt = &now
	t.UpdatedAt = now
	if err := s.repo.TaskUpdate(ctx, t); err != nil {
		return nil, err
	}
	if t.Repeat == nil {
		return nil, nil
	}
	nextTime, err := repeat.NextOccurrence(*t.Repeat, now)
	if err != nil {
		return nil, err
	}
	nextDate := task.NewDate(nextTime)
	next := task.Task{
		ID:        id.New(),
		Title:     t.Title,
		Notes:     t.Notes,
		Status:    task.StatusOpen,
		AreaID:    t.AreaID,
		ProjectID: t.ProjectID,
		HeadingID: t.HeadingID,
		Tags:      append([]id.ID(nil), t.Tags...),
		StartDate: &nextDate,
		Deadline:  t.Deadline,
		Checklist: resetChecklist(t.Checklist),
		Repeat:    t.Repeat,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.TaskCreate(ctx, next); err != nil {
		return nil, err
	}
	return &next, nil
}

func (s *Service) CancelTask(ctx context.Context, taskID id.ID) error {
	t, err := s.repo.TaskGet(ctx, taskID)
	if err != nil {
		return mapNotFound(err)
	}
	if t.Status != task.StatusOpen {
		return nil
	}
	now := s.clock.Now()
	t.Status = task.StatusCancelled
	t.CancelledAt = &now
	t.UpdatedAt = now
	return s.repo.TaskUpdate(ctx, t)
}

func (s *Service) DeleteTask(ctx context.Context, taskID id.ID, hard bool) error {
	if err := s.repo.TaskDelete(ctx, taskID, !hard); err != nil {
		return mapNotFound(err)
	}
	return nil
}

func (s *Service) PinToToday(ctx context.Context, taskID id.ID) error {
	t, err := s.repo.TaskGet(ctx, taskID)
	if err != nil {
		return mapNotFound(err)
	}
	d := task.NewDate(s.clock.Now())
	t.PinnedToday = &d
	t.UpdatedAt = s.clock.Now()
	return s.repo.TaskUpdate(ctx, t)
}

func (s *Service) UnpinFromToday(ctx context.Context, taskID id.ID) error {
	t, err := s.repo.TaskGet(ctx, taskID)
	if err != nil {
		return mapNotFound(err)
	}
	t.PinnedToday = nil
	t.UpdatedAt = s.clock.Now()
	return s.repo.TaskUpdate(ctx, t)
}

// MoveTarget describes the destination of MoveTask.
type MoveTarget struct {
	AreaID    *id.ID
	ProjectID *id.ID
	HeadingID *id.ID
	ToInbox   bool
}

func (s *Service) MoveTask(ctx context.Context, taskID id.ID, target MoveTarget) error {
	t, err := s.repo.TaskGet(ctx, taskID)
	if err != nil {
		return mapNotFound(err)
	}
	if target.ToInbox {
		t.AreaID = nil
		t.ProjectID = nil
		t.HeadingID = nil
	} else {
		t.AreaID = target.AreaID
		t.ProjectID = target.ProjectID
		t.HeadingID = target.HeadingID
	}
	t.UpdatedAt = s.clock.Now()
	return s.repo.TaskUpdate(ctx, t)
}

// Projects, areas, headings

type AddProjectInput struct {
	Name      string
	Notes     string
	AreaID    *id.ID
	Deadline  *task.Date
	AutoClose bool
}

func (s *Service) AddProject(ctx context.Context, in AddProjectInput) (project.Project, error) {
	now := s.clock.Now()
	p := project.Project{
		ID:        id.New(),
		Name:      in.Name,
		Notes:     in.Notes,
		AreaID:    in.AreaID,
		Deadline:  in.Deadline,
		Status:    project.StatusOpen,
		AutoClose: in.AutoClose,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := p.Validate(); err != nil {
		return project.Project{}, err
	}
	if err := s.repo.ProjectCreate(ctx, p); err != nil {
		return project.Project{}, err
	}
	return p, nil
}

func (s *Service) EditProject(ctx context.Context, p project.Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	p.UpdatedAt = s.clock.Now()
	if err := s.repo.ProjectUpdate(ctx, p); err != nil {
		return err
	}
	return s.maybeAutoCloseProject(ctx, p)
}

func (s *Service) maybeAutoCloseProject(ctx context.Context, p project.Project) error {
	if !p.AutoClose || p.Status != project.StatusOpen {
		return nil
	}
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{ProjectID: &p.ID})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	for _, t := range tasks {
		if t.Status == task.StatusOpen {
			return nil
		}
	}
	now := s.clock.Now()
	p.Status = project.StatusCompleted
	p.CompletedAt = &now
	p.UpdatedAt = now
	return s.repo.ProjectUpdate(ctx, p)
}

func (s *Service) AddArea(ctx context.Context, name string) (area.Area, error) {
	now := s.clock.Now()
	a := area.Area{ID: id.New(), Name: name, CreatedAt: now, UpdatedAt: now}
	if err := a.Validate(); err != nil {
		return area.Area{}, err
	}
	if err := s.repo.AreaCreate(ctx, a); err != nil {
		return area.Area{}, err
	}
	return a, nil
}

func (s *Service) DeleteArea(ctx context.Context, aid id.ID, confirm bool) error {
	projects, err := s.repo.ProjectList(ctx, storage.ProjectFilter{AreaID: &aid})
	if err != nil {
		return err
	}
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{AreaID: &aid})
	if err != nil {
		return err
	}
	hasOrphanTasks := false
	for _, t := range tasks {
		if t.ProjectID == nil {
			hasOrphanTasks = true
			break
		}
	}
	if !confirm && (len(projects) > 0 || hasOrphanTasks) {
		return ErrAreaNotEmpty
	}
	now := s.clock.Now()
	for _, p := range projects {
		p.AreaID = nil
		p.UpdatedAt = now
		if err := s.repo.ProjectUpdate(ctx, p); err != nil {
			return err
		}
	}
	for _, t := range tasks {
		if t.ProjectID != nil {
			continue
		}
		t.AreaID = nil
		t.UpdatedAt = now
		if err := s.repo.TaskUpdate(ctx, t); err != nil {
			return err
		}
	}
	return s.repo.AreaDelete(ctx, aid, false)
}

func (s *Service) AddHeading(ctx context.Context, projectID id.ID, name string) (project.Heading, error) {
	now := s.clock.Now()
	h := project.Heading{
		ID:        id.New(),
		ProjectID: projectID,
		Name:      name,
		CreatedAt: now,
	}
	if err := h.Validate(); err != nil {
		return project.Heading{}, err
	}
	if err := s.repo.HeadingCreate(ctx, h); err != nil {
		return project.Heading{}, err
	}
	return h, nil
}

func (s *Service) RenameTag(ctx context.Context, tagID id.ID, newName string) error {
	if err := s.repo.TagRename(ctx, tagID, newName); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return ErrTagAlreadyExists
		}
		return err
	}
	return nil
}

func (s *Service) DeleteTag(ctx context.Context, tagID id.ID) error {
	return s.repo.TagDelete(ctx, tagID)
}

func (s *Service) UpsertTag(ctx context.Context, name string) (tag.Tag, error) {
	return s.repo.TagUpsert(ctx, name)
}

func (s *Service) Clock() Clock { return s.clock }

func (s *Service) Repo() storage.Repository { return s.repo }

// helpers

func validateDates(start, deadline *task.Date) error {
	if start == nil || deadline == nil {
		return nil
	}
	if deadline.Time.Before(start.Time) { //nolint:staticcheck
		return ErrDeadlineBeforeStart
	}
	return nil
}

func resetChecklist(items []task.ChecklistItem) []task.ChecklistItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]task.ChecklistItem, len(items))
	for i, ci := range items {
		out[i] = task.ChecklistItem{ID: id.New(), Text: ci.Text}
	}
	return out
}

func mapTaskValidationError(err error) error {
	switch {
	case errors.Is(err, task.ErrEmptyTitle):
		return ErrEmptyTitle
	default:
		return fmt.Errorf("app: validation: %w", err)
	}
}

func mapNotFound(err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return ErrTaskNotFound
	}
	return err
}

// startOfLocalDay aligns the given time to local midnight.
func startOfLocalDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

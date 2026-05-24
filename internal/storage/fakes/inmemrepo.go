// Package fakes provides an in-memory storage.Repository implementation
// used by service tests. NOT for production use.
package fakes

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/tag"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
)

type InMemRepo struct {
	mu       sync.RWMutex
	schema   int
	tasks    map[id.ID]task.Task
	projects map[id.ID]project.Project
	headings map[id.ID]project.Heading
	areas    map[id.ID]area.Area
	tags     map[id.ID]tag.Tag
}

func New() *InMemRepo {
	return &InMemRepo{
		schema:   storage.CurrentSchemaVersion,
		tasks:    make(map[id.ID]task.Task),
		projects: make(map[id.ID]project.Project),
		headings: make(map[id.ID]project.Heading),
		areas:    make(map[id.ID]area.Area),
		tags:     make(map[id.ID]tag.Tag),
	}
}

func (r *InMemRepo) Close() error { return nil }

func (r *InMemRepo) SchemaVersion(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.schema, nil
}

func (r *InMemRepo) Migrate(_ context.Context, target int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if target < r.schema {
		return nil
	}
	if target > storage.CurrentSchemaVersion {
		return storage.ErrSchemaTooNew
	}
	r.schema = target
	return nil
}

// Tasks

func (r *InMemRepo) TaskCreate(_ context.Context, t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[t.ID]; exists {
		return storage.ErrAlreadyExists
	}
	r.tasks[t.ID] = t
	return nil
}

func (r *InMemRepo) TaskGet(_ context.Context, taskID id.ID) (task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[taskID]
	if !ok {
		return task.Task{}, storage.ErrNotFound
	}
	return t, nil
}

func (r *InMemRepo) TaskUpdate(_ context.Context, t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[t.ID]; !exists {
		return storage.ErrNotFound
	}
	r.tasks[t.ID] = t
	return nil
}

func (r *InMemRepo) TaskDelete(_ context.Context, taskID id.ID, soft bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[taskID]
	if !ok {
		return storage.ErrNotFound
	}
	if soft {
		now := time.Now()
		t.DeletedAt = &now
		r.tasks[taskID] = t
		return nil
	}
	delete(r.tasks, taskID)
	return nil
}

func (r *InMemRepo) TaskList(_ context.Context, f storage.TaskFilter) ([]task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]task.Task, 0)
	for _, t := range r.tasks {
		if !f.IncludeDeleted && t.DeletedAt != nil {
			continue
		}
		if !taskMatchesFilter(t, f) {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func taskMatchesFilter(t task.Task, f storage.TaskFilter) bool {
	if len(f.Statuses) > 0 && !containsStatus(f.Statuses, t.Status) {
		return false
	}
	if f.AreaID != nil {
		if t.AreaID == nil || *t.AreaID != *f.AreaID {
			return false
		}
	}
	if f.ProjectID != nil {
		if t.ProjectID == nil || *t.ProjectID != *f.ProjectID {
			return false
		}
	}
	if f.HeadingID != nil {
		if t.HeadingID == nil || *t.HeadingID != *f.HeadingID {
			return false
		}
	}
	if f.Someday != nil && t.Someday != *f.Someday {
		return false
	}
	if f.HasStartDate != nil {
		has := t.StartDate != nil
		if has != *f.HasStartDate {
			return false
		}
	}
	if f.StartDateFrom != nil && (t.StartDate == nil || t.StartDate.Time.Before(*f.StartDateFrom)) { //nolint:staticcheck // Date.Before shadows time.Time.Before
		return false
	}
	if f.StartDateTo != nil && (t.StartDate == nil || t.StartDate.Time.After(*f.StartDateTo)) { //nolint:staticcheck
		return false
	}
	if f.DeadlineTo != nil && (t.Deadline == nil || t.Deadline.Time.After(*f.DeadlineTo)) { //nolint:staticcheck
		return false
	}
	if len(f.TagsAll) > 0 {
		for _, want := range f.TagsAll {
			found := false
			for _, have := range t.Tags {
				if have == want {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func containsStatus(ss []task.Status, s task.Status) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func (r *InMemRepo) TaskMatchShort(_ context.Context, prefix string) ([]task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pfx := strings.ToUpper(prefix)
	out := make([]task.Task, 0)
	for tid, t := range r.tasks {
		upper := strings.ToUpper(string(tid))
		if upper == pfx {
			out = append(out, t)
			continue
		}
		const offset = 10
		if len(upper) >= offset+len(pfx) && upper[offset:offset+len(pfx)] == pfx {
			out = append(out, t)
		}
	}
	return out, nil
}

// Projects

func (r *InMemRepo) ProjectCreate(_ context.Context, p project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.projects[p.ID]; exists {
		return storage.ErrAlreadyExists
	}
	r.projects[p.ID] = p
	return nil
}

func (r *InMemRepo) ProjectGet(_ context.Context, pid id.ID) (project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.projects[pid]
	if !ok {
		return project.Project{}, storage.ErrNotFound
	}
	return p, nil
}

func (r *InMemRepo) ProjectUpdate(_ context.Context, p project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[p.ID]; !ok {
		return storage.ErrNotFound
	}
	r.projects[p.ID] = p
	return nil
}

func (r *InMemRepo) ProjectDelete(_ context.Context, pid id.ID, soft bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.projects[pid]
	if !ok {
		return storage.ErrNotFound
	}
	if soft {
		now := time.Now()
		p.DeletedAt = &now
		r.projects[pid] = p
		return nil
	}
	delete(r.projects, pid)
	return nil
}

func (r *InMemRepo) ProjectList(_ context.Context, f storage.ProjectFilter) ([]project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]project.Project, 0)
	for _, p := range r.projects {
		if !f.IncludeDeleted && p.DeletedAt != nil {
			continue
		}
		if f.AreaID != nil {
			if p.AreaID == nil || *p.AreaID != *f.AreaID {
				continue
			}
		}
		if len(f.Statuses) > 0 {
			found := false
			for _, s := range f.Statuses {
				if p.Status == s {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *InMemRepo) ProjectFindByName(_ context.Context, name string) ([]project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	target := strings.ToLower(strings.TrimSpace(name))
	out := make([]project.Project, 0)
	for _, p := range r.projects {
		if p.DeletedAt != nil {
			continue
		}
		if strings.ToLower(strings.TrimSpace(p.Name)) == target {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *InMemRepo) HeadingCreate(_ context.Context, h project.Heading) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.headings[h.ID]; exists {
		return storage.ErrAlreadyExists
	}
	r.headings[h.ID] = h
	return nil
}

func (r *InMemRepo) HeadingUpdate(_ context.Context, h project.Heading) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.headings[h.ID]; !ok {
		return storage.ErrNotFound
	}
	r.headings[h.ID] = h
	return nil
}

func (r *InMemRepo) HeadingDelete(_ context.Context, hid id.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.headings[hid]; !ok {
		return storage.ErrNotFound
	}
	delete(r.headings, hid)
	return nil
}

func (r *InMemRepo) HeadingList(_ context.Context, projectID id.ID) ([]project.Heading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]project.Heading, 0)
	for _, h := range r.headings {
		if h.ProjectID == projectID {
			out = append(out, h)
		}
	}
	return out, nil
}

// Areas

func (r *InMemRepo) AreaCreate(_ context.Context, a area.Area) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	norm := area.Normalize(a.Name)
	for _, existing := range r.areas {
		if existing.DeletedAt == nil && area.Normalize(existing.Name) == norm {
			return storage.ErrAlreadyExists
		}
	}
	r.areas[a.ID] = a
	return nil
}

func (r *InMemRepo) AreaGet(_ context.Context, aid id.ID) (area.Area, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.areas[aid]
	if !ok {
		return area.Area{}, storage.ErrNotFound
	}
	return a, nil
}

func (r *InMemRepo) AreaList(_ context.Context) ([]area.Area, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]area.Area, 0, len(r.areas))
	for _, a := range r.areas {
		if a.DeletedAt != nil {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *InMemRepo) AreaUpdate(_ context.Context, a area.Area) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.areas[a.ID]; !ok {
		return storage.ErrNotFound
	}
	r.areas[a.ID] = a
	return nil
}

func (r *InMemRepo) AreaDelete(_ context.Context, aid id.ID, soft bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.areas[aid]
	if !ok {
		return storage.ErrNotFound
	}
	if soft {
		now := time.Now()
		a.DeletedAt = &now
		r.areas[aid] = a
		return nil
	}
	delete(r.areas, aid)
	return nil
}

func (r *InMemRepo) AreaFindByNormalized(_ context.Context, normalized string) (area.Area, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.areas {
		if a.DeletedAt == nil && area.Normalize(a.Name) == normalized {
			return a, nil
		}
	}
	return area.Area{}, storage.ErrNotFound
}

// Tags

func (r *InMemRepo) TagCreate(_ context.Context, t tag.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tags[t.ID]; exists {
		return storage.ErrAlreadyExists
	}
	for _, existing := range r.tags {
		if existing.Normalized == t.Normalized {
			return storage.ErrAlreadyExists
		}
	}
	r.tags[t.ID] = t
	return nil
}

func (r *InMemRepo) TagUpsert(_ context.Context, name string) (tag.Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	norm := tag.Normalize(name)
	if norm == "" {
		return tag.Tag{}, tag.ErrEmptyName
	}
	for _, existing := range r.tags {
		if existing.Normalized == norm {
			return existing, nil
		}
	}
	t := tag.Tag{
		ID:         id.New(),
		Name:       strings.TrimSpace(name),
		Normalized: norm,
		CreatedAt:  time.Now(),
	}
	r.tags[t.ID] = t
	return t, nil
}

func (r *InMemRepo) TagGet(_ context.Context, tid id.ID) (tag.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tags[tid]
	if !ok {
		return tag.Tag{}, storage.ErrNotFound
	}
	return t, nil
}

func (r *InMemRepo) TagList(_ context.Context) ([]tag.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]tag.Tag, 0, len(r.tags))
	for _, t := range r.tags {
		out = append(out, t)
	}
	return out, nil
}

func (r *InMemRepo) TagRename(_ context.Context, tid id.ID, newName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tags[tid]
	if !ok {
		return storage.ErrNotFound
	}
	newNorm := tag.Normalize(newName)
	if newNorm == "" {
		return tag.ErrEmptyName
	}
	for tid2, other := range r.tags {
		if tid2 == tid {
			continue
		}
		if other.Normalized == newNorm {
			return storage.ErrAlreadyExists
		}
	}
	t.Name = strings.TrimSpace(newName)
	t.Normalized = newNorm
	r.tags[tid] = t
	return nil
}

func (r *InMemRepo) TagDelete(_ context.Context, tid id.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tags[tid]; !ok {
		return storage.ErrNotFound
	}
	delete(r.tags, tid)
	// Strip tag from all tasks
	for taskID, t := range r.tasks {
		filtered := t.Tags[:0]
		for _, x := range t.Tags {
			if x != tid {
				filtered = append(filtered, x)
			}
		}
		if len(filtered) != len(t.Tags) {
			t.Tags = filtered
			r.tasks[taskID] = t
		}
	}
	return nil
}

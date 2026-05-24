package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/tag"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
)

// Snapshot is the wire format for ExportJSON / ImportJSON.
type Snapshot struct {
	SchemaVersion int               `json:"schema_version"`
	ExportedAt    time.Time         `json:"exported_at"`
	Areas         []area.Area       `json:"areas"`
	Projects      []project.Project `json:"projects"`
	Headings      []project.Heading `json:"headings"`
	Tags          []tag.Tag         `json:"tags"`
	Tasks         []task.Task       `json:"tasks"`
}

func (s *Service) ExportJSON(ctx context.Context, w io.Writer) error {
	snap, err := s.buildSnapshot(ctx)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(snap)
}

func (s *Service) ExportSnapshot(ctx context.Context) (Snapshot, error) {
	return s.buildSnapshot(ctx)
}

func (s *Service) buildSnapshot(ctx context.Context) (Snapshot, error) {
	areas, err := s.repo.AreaList(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	projects, err := s.repo.ProjectList(ctx, storage.ProjectFilter{IncludeDeleted: true})
	if err != nil {
		return Snapshot{}, err
	}
	headings := make([]project.Heading, 0)
	for _, p := range projects {
		hs, err := s.repo.HeadingList(ctx, p.ID)
		if err != nil {
			return Snapshot{}, err
		}
		headings = append(headings, hs...)
	}
	tags, err := s.repo.TagList(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{IncludeDeleted: true})
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{
		SchemaVersion: storage.CurrentSchemaVersion,
		ExportedAt:    s.clock.Now(),
		Areas:         areas,
		Projects:      projects,
		Headings:      headings,
		Tags:          tags,
		Tasks:         tasks,
	}, nil
}

// ImportJSON replaces repository content with the snapshot in r.
// Caller is responsible for obtaining user confirmation before invocation.
func (s *Service) ImportJSON(ctx context.Context, r io.Reader) (Snapshot, error) {
	var snap Snapshot
	dec := json.NewDecoder(r)
	if err := dec.Decode(&snap); err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrInvalidImport, err)
	}
	if snap.SchemaVersion > storage.CurrentSchemaVersion {
		return Snapshot{}, ErrSchemaTooNew
	}
	if snap.SchemaVersion < 1 {
		return Snapshot{}, fmt.Errorf("%w: schema_version %d", ErrInvalidImport, snap.SchemaVersion)
	}
	return snap, s.applySnapshot(ctx, snap)
}

func (s *Service) applySnapshot(ctx context.Context, snap Snapshot) error {
	if err := s.wipe(ctx); err != nil {
		return err
	}
	for _, a := range snap.Areas {
		if err := s.repo.AreaCreate(ctx, a); err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			return err
		}
	}
	for _, p := range snap.Projects {
		if err := s.repo.ProjectCreate(ctx, p); err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			return err
		}
	}
	for _, h := range snap.Headings {
		if err := s.repo.HeadingCreate(ctx, h); err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			return err
		}
	}
	for _, tg := range snap.Tags {
		if err := s.repo.TagCreate(ctx, tg); err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			return err
		}
	}
	for _, t := range snap.Tasks {
		if err := s.repo.TaskCreate(ctx, t); err != nil && !errors.Is(err, storage.ErrAlreadyExists) {
			return err
		}
	}
	return nil
}

func (s *Service) wipe(ctx context.Context) error {
	tasks, err := s.repo.TaskList(ctx, storage.TaskFilter{IncludeDeleted: true})
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := s.repo.TaskDelete(ctx, t.ID, false); err != nil {
			return err
		}
	}
	projects, err := s.repo.ProjectList(ctx, storage.ProjectFilter{IncludeDeleted: true})
	if err != nil {
		return err
	}
	for _, p := range projects {
		hs, err := s.repo.HeadingList(ctx, p.ID)
		if err != nil {
			return err
		}
		for _, h := range hs {
			if err := s.repo.HeadingDelete(ctx, h.ID); err != nil {
				return err
			}
		}
		if err := s.repo.ProjectDelete(ctx, p.ID, false); err != nil {
			return err
		}
	}
	areas, err := s.repo.AreaList(ctx)
	if err != nil {
		return err
	}
	for _, a := range areas {
		if err := s.repo.AreaDelete(ctx, a.ID, false); err != nil {
			return err
		}
	}
	tags, err := s.repo.TagList(ctx)
	if err != nil {
		return err
	}
	for _, tg := range tags {
		if err := s.repo.TagDelete(ctx, tg.ID); err != nil {
			return err
		}
	}
	return nil
}

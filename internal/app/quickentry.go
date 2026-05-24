package app

import (
	"context"
	"strings"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/quickentry"
	"github.com/jtprogru/todushka/internal/domain/task"
)

// QuickEntry parses raw input, resolves project / tag references against the
// repository, and creates the resulting Task. Returns ErrAmbiguousProject if
// a `@<name>` token matches multiple projects.
func (s *Service) QuickEntry(ctx context.Context, raw string) (task.Task, error) {
	if quickentry.IsEmpty(raw) {
		return task.Task{}, ErrEmptyInput
	}
	parsed, err := quickentry.Parse(raw)
	if err != nil {
		return task.Task{}, err
	}

	in := AddTaskInput{
		Title:    strings.TrimSpace(parsed.Title),
		Deadline: parsed.Deadline,
	}
	if parsed.StartDate != nil {
		d := task.NewDate(s.clock.Now())
		in.StartDate = &d
	}
	if parsed.ProjectRef != nil {
		projects, err := s.repo.ProjectFindByName(ctx, *parsed.ProjectRef)
		if err != nil {
			return task.Task{}, err
		}
		switch len(projects) {
		case 0:
			return task.Task{}, ErrUnknownProject
		case 1:
			pid := projects[0].ID
			in.ProjectID = &pid
			if projects[0].AreaID != nil {
				aid := *projects[0].AreaID
				in.AreaID = &aid
			}
		default:
			return task.Task{}, ErrAmbiguousProject
		}
	}
	tagIDs := make([]id.ID, 0, len(parsed.Tags))
	for _, name := range parsed.Tags {
		t, err := s.repo.TagUpsert(ctx, name)
		if err != nil {
			return task.Task{}, err
		}
		tagIDs = append(tagIDs, t.ID)
	}
	in.Tags = tagIDs

	return s.AddTask(ctx, in)
}

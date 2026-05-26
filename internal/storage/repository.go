// Package storage defines the persistence contract used by the application
// service. Concrete implementations live in subpackages (bbolt, fakes).
package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/tag"
	"github.com/jtprogru/todushka/internal/domain/task"
)

// CurrentSchemaVersion is the schema version this binary understands.
const CurrentSchemaVersion = 1

var (
	ErrNotFound       = errors.New("storage: not found")
	ErrAlreadyExists  = errors.New("storage: already exists")
	ErrDatabaseLocked = errors.New("storage: database is locked by another todushka process")
	ErrSchemaTooNew   = errors.New("storage: database schema is newer than this binary supports")
	ErrInvalidImport  = errors.New("storage: invalid import payload")
	ErrReadOnly       = errors.New("storage: repository is read-only")
)

// TaskFilter narrows TaskList results. All fields are optional; zero-values
// mean "any".
type TaskFilter struct {
	Statuses       []task.Status
	AreaID         *id.ID
	ProjectID      *id.ID
	HeadingID      *id.ID
	TagsAll        []id.ID
	Someday        *bool
	HasStartDate   *bool
	StartDateFrom  *time.Time
	StartDateTo    *time.Time
	DeadlineTo     *time.Time
	PinnedTodayOn  *time.Time
	IncludeDeleted bool
}

// ProjectFilter narrows ProjectList results.
type ProjectFilter struct {
	AreaID         *id.ID
	Statuses       []project.Status
	IncludeDeleted bool
}

// Repository is the persistence interface used by the application service.
type Repository interface {
	Close() error
	SchemaVersion(ctx context.Context) (int, error)
	Migrate(ctx context.Context, target int) error

	// Tasks
	TaskCreate(ctx context.Context, t task.Task) error
	TaskGet(ctx context.Context, id id.ID) (task.Task, error)
	TaskList(ctx context.Context, f TaskFilter) ([]task.Task, error)
	TaskUpdate(ctx context.Context, t task.Task) error
	TaskDelete(ctx context.Context, id id.ID, soft bool) error
	TaskMatchShort(ctx context.Context, prefix string) ([]task.Task, error)

	// Projects
	ProjectCreate(ctx context.Context, p project.Project) error
	ProjectGet(ctx context.Context, id id.ID) (project.Project, error)
	ProjectList(ctx context.Context, f ProjectFilter) ([]project.Project, error)
	ProjectUpdate(ctx context.Context, p project.Project) error
	ProjectDelete(ctx context.Context, id id.ID, soft bool) error
	ProjectFindByName(ctx context.Context, name string) ([]project.Project, error)
	HeadingCreate(ctx context.Context, h project.Heading) error
	HeadingUpdate(ctx context.Context, h project.Heading) error
	HeadingDelete(ctx context.Context, id id.ID) error
	HeadingList(ctx context.Context, projectID id.ID) ([]project.Heading, error)

	// Areas
	AreaCreate(ctx context.Context, a area.Area) error
	AreaGet(ctx context.Context, id id.ID) (area.Area, error)
	AreaList(ctx context.Context) ([]area.Area, error)
	AreaUpdate(ctx context.Context, a area.Area) error
	AreaDelete(ctx context.Context, id id.ID, soft bool) error
	AreaFindByNormalized(ctx context.Context, normalized string) (area.Area, error)

	// Tags
	TagCreate(ctx context.Context, t tag.Tag) error
	TagUpsert(ctx context.Context, name string) (tag.Tag, error)
	TagGet(ctx context.Context, id id.ID) (tag.Tag, error)
	TagList(ctx context.Context) ([]tag.Tag, error)
	TagRename(ctx context.Context, id id.ID, newName string) error
	TagDelete(ctx context.Context, id id.ID) error

	// ReadOnly reports whether this repository was opened in read-only mode.
	// When true, all write methods return ErrReadOnly. Read methods work
	// regardless.
	ReadOnly() bool
}

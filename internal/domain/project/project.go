package project

import (
	"errors"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
)

type Status string

const (
	StatusOpen      Status = "open"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

type Project struct {
	ID          id.ID      `json:"id"`
	Name        string     `json:"name"`
	Notes       string     `json:"notes,omitempty"`
	AreaID      *id.ID     `json:"area_id,omitempty"`
	Deadline    *task.Date `json:"deadline,omitempty"`
	Status      Status     `json:"status"`
	AutoClose   bool       `json:"auto_close,omitempty"`
	Position    int64      `json:"position"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type Heading struct {
	ID        id.ID     `json:"id"`
	ProjectID id.ID     `json:"project_id"`
	Name      string    `json:"name"`
	Position  int64     `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrEmptyName     = errors.New("project: empty name")
	ErrInvalidStatus = errors.New("project: invalid status")
	ErrEmptyProject  = errors.New("heading: project id required")
)

func (p Project) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyName
	}
	switch p.Status {
	case StatusOpen, StatusCompleted, StatusCancelled:
	default:
		return ErrInvalidStatus
	}
	return nil
}

func (h Heading) Validate() error {
	if strings.TrimSpace(h.Name) == "" {
		return ErrEmptyName
	}
	if h.ProjectID == "" {
		return ErrEmptyProject
	}
	return nil
}

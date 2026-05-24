package task

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/repeat"
)

type Status string

const (
	StatusOpen      Status = "open"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

// Date represents a calendar day in local time, normalised to 00:00:00.
type Date struct {
	time.Time
}

// NewDate constructs a Date by truncating the given time to a local day.
func NewDate(t time.Time) Date {
	return Date{Time: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format("2006-01-02") + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return fmt.Errorf("date: %w", err)
	}
	d.Time = t
	return nil
}

// Before reports whether d is strictly before other (date-only comparison).
func (d Date) Before(other Date) bool { return d.Time.Before(other.Time) }

// Equal reports whether d and other represent the same calendar day.
func (d Date) Equal(other Date) bool {
	return d.Year() == other.Year() && d.Month() == other.Month() && d.Day() == other.Day()
}

type ChecklistItem struct {
	ID   id.ID  `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type Task struct {
	ID          id.ID           `json:"id"`
	Title       string          `json:"title"`
	Notes       string          `json:"notes,omitempty"`
	Status      Status          `json:"status"`
	AreaID      *id.ID          `json:"area_id,omitempty"`
	ProjectID   *id.ID          `json:"project_id,omitempty"`
	HeadingID   *id.ID          `json:"heading_id,omitempty"`
	Tags        []id.ID         `json:"tags,omitempty"`
	StartDate   *Date           `json:"start_date,omitempty"`
	Deadline    *Date           `json:"deadline,omitempty"`
	Someday     bool            `json:"someday,omitempty"`
	PinnedToday *Date           `json:"pinned_today,omitempty"`
	Checklist   []ChecklistItem `json:"checklist,omitempty"`
	Repeat      *repeat.Rule    `json:"repeat,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	CancelledAt *time.Time      `json:"cancelled_at,omitempty"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
	Position    int64           `json:"position"`
}

var (
	ErrEmptyTitle           = errors.New("task: empty title")
	ErrInvalidStatus        = errors.New("task: invalid status")
	ErrStatusInvariant      = errors.New("task: status invariant violated")
	ErrEmptyChecklistText   = errors.New("task: empty checklist text")
	ErrInvalidChecklistID   = errors.New("task: invalid checklist item id")
	ErrChecklistDuplicateID = errors.New("task: duplicate checklist item id")
)

// Validate checks all per-Task invariants. Cross-entity rules (deadline vs
// start_date, project existence) are enforced at the service layer.
func (t Task) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return ErrEmptyTitle
	}
	switch t.Status {
	case StatusOpen:
		if t.CompletedAt != nil || t.CancelledAt != nil {
			return ErrStatusInvariant
		}
	case StatusCompleted:
		if t.CompletedAt == nil || t.CancelledAt != nil {
			return ErrStatusInvariant
		}
	case StatusCancelled:
		if t.CancelledAt == nil || t.CompletedAt != nil {
			return ErrStatusInvariant
		}
	default:
		return ErrInvalidStatus
	}
	seen := make(map[id.ID]struct{}, len(t.Checklist))
	for _, ci := range t.Checklist {
		if strings.TrimSpace(ci.Text) == "" {
			return ErrEmptyChecklistText
		}
		if ci.ID == "" {
			return ErrInvalidChecklistID
		}
		if _, dup := seen[ci.ID]; dup {
			return ErrChecklistDuplicateID
		}
		seen[ci.ID] = struct{}{}
	}
	return nil
}

package area

import (
	"errors"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
)

type Area struct {
	ID        id.ID      `json:"id"`
	Name      string     `json:"name"`
	Position  int64      `json:"position"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

var ErrEmptyName = errors.New("area: empty name")

// Normalize returns the case-folded form used for uniqueness comparison.
func Normalize(name string) string { return strings.ToLower(strings.TrimSpace(name)) }

func (a Area) Validate() error {
	if strings.TrimSpace(a.Name) == "" {
		return ErrEmptyName
	}
	return nil
}

package tag

import (
	"errors"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
)

type Tag struct {
	ID         id.ID     `json:"id"`
	Name       string    `json:"name"`
	Normalized string    `json:"normalized"`
	CreatedAt  time.Time `json:"created_at"`
}

var ErrEmptyName = errors.New("tag: empty name")

// Normalize returns the canonical form used for uniqueness comparison.
func Normalize(name string) string { return strings.ToLower(strings.TrimSpace(name)) }

func (t Tag) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrEmptyName
	}
	return nil
}

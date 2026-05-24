package project

import (
	"testing"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/stretchr/testify/require"
)

func TestProject_ValidateScenarios(t *testing.T) {
	cases := []struct {
		name    string
		project Project
		wantErr error
	}{
		{"empty_name", Project{Status: StatusOpen}, ErrEmptyName},
		{"whitespace_name", Project{Name: "   ", Status: StatusOpen}, ErrEmptyName},
		{"unknown_status", Project{Name: "x", Status: "weird"}, ErrInvalidStatus},
		{"empty_status", Project{Name: "x"}, ErrInvalidStatus},
		{"ok_open", Project{Name: "x", Status: StatusOpen}, nil},
		{"ok_completed", Project{Name: "x", Status: StatusCompleted}, nil},
		{"ok_cancelled", Project{Name: "x", Status: StatusCancelled}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.project.Validate()
			if c.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, c.wantErr)
			}
		})
	}
}

func TestHeading_Validate(t *testing.T) {
	pid := id.New()
	t.Run("empty_name", func(t *testing.T) {
		h := Heading{ProjectID: pid}
		require.ErrorIs(t, h.Validate(), ErrEmptyName)
	})
	t.Run("missing_project_id", func(t *testing.T) {
		h := Heading{Name: "x"}
		require.ErrorIs(t, h.Validate(), ErrEmptyProject)
	})
	t.Run("ok", func(t *testing.T) {
		h := Heading{Name: "x", ProjectID: pid}
		require.NoError(t, h.Validate())
	})
}

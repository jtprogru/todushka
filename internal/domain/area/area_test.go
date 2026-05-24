package area

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArea_ValidateEmptyName(t *testing.T) {
	cases := []string{"", "   ", "\t\n"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			a := Area{Name: c}
			require.ErrorIs(t, a.Validate(), ErrEmptyName)
		})
	}
}

func TestArea_ValidateOK(t *testing.T) {
	a := Area{Name: "Work"}
	require.NoError(t, a.Validate())
}

func TestArea_NormalizeIdempotent(t *testing.T) {
	require.Equal(t, "work", Normalize("  WORK "))
	require.Equal(t, Normalize("Work"), Normalize(Normalize("Work")))
}

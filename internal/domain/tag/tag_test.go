package tag

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTag_NormalizeLowercaseAndTrim(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Work", "work"},
		{"  Work ", "work"},
		{"\thome\n", "home"},
		{"", ""},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			require.Equal(t, c.want, Normalize(c.in))
		})
	}
}

func TestTag_NormalizeIdempotent(t *testing.T) {
	for _, s := range []string{"Work", "  HOME  ", "Mixed-Case"} {
		require.Equal(t, Normalize(s), Normalize(Normalize(s)))
	}
}

func TestTag_ValidateEmptyName(t *testing.T) {
	require.ErrorIs(t, Tag{Name: ""}.Validate(), ErrEmptyName)
	require.ErrorIs(t, Tag{Name: "   "}.Validate(), ErrEmptyName)
	require.NoError(t, Tag{Name: "x"}.Validate())
}

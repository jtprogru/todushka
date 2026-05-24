package id

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestID_NewIsULID(t *testing.T) {
	got := New()
	require.Len(t, string(got), ulidLen)
	_, err := ulid.Parse(string(got))
	require.NoError(t, err)
}

func TestID_NewMonotonic(t *testing.T) {
	a := New()
	b := New()
	require.NotEqual(t, a, b)
}

func TestID_ShortStable(t *testing.T) {
	x := New()
	require.Equal(t, Short(x), Short(x))
	require.Len(t, Short(x), ShortLen)
}

func TestID_ParseRejectsInvalid(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := Parse("")
		require.ErrorIs(t, err, ErrEmpty)
	})
	t.Run("too_short", func(t *testing.T) {
		_, err := Parse("abc")
		require.ErrorIs(t, err, ErrInvalid)
	})
	t.Run("non_ulid_chars", func(t *testing.T) {
		_, err := Parse("**************************")
		require.ErrorIs(t, err, ErrInvalid)
	})
}

func TestID_ParseAcceptsValid(t *testing.T) {
	new := New()
	got, err := Parse(string(new))
	require.NoError(t, err)
	require.Equal(t, new, got)
}

package quickentry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestQuickEntry_IsEmpty(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"x", false},
		{" x ", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			require.Equal(t, c.want, IsEmpty(c.in))
		})
	}
}

func TestQuickEntry_BareTitle(t *testing.T) {
	p, err := Parse("buy milk")
	require.NoError(t, err)
	require.Equal(t, "buy milk", p.Title)
	require.Nil(t, p.StartDate)
	require.Nil(t, p.Deadline)
	require.Nil(t, p.ProjectRef)
	require.Empty(t, p.Tags)
}

func TestQuickEntry_TitleWithTags(t *testing.T) {
	p, err := Parse("buy milk #shop #fast")
	require.NoError(t, err)
	require.Equal(t, "buy milk", p.Title)
	require.Equal(t, []string{"shop", "fast"}, p.Tags)
}

func TestQuickEntry_TitleWithTodayAndDeadline(t *testing.T) {
	p, err := Parse("report @today !2026-06-01")
	require.NoError(t, err)
	require.Equal(t, "report", p.Title)
	require.NotNil(t, p.StartDate)
	require.NotNil(t, p.Deadline)
	require.Equal(t, "2026-06-01", p.Deadline.Format("2006-01-02"))
}

func TestQuickEntry_ProjectRef(t *testing.T) {
	p, err := Parse("review pr @work")
	require.NoError(t, err)
	require.NotNil(t, p.ProjectRef)
	require.Equal(t, "work", *p.ProjectRef)
	require.Equal(t, "review pr", p.Title)
}

func TestQuickEntry_InvalidDateRejected(t *testing.T) {
	_, err := Parse("x !2026-13-40")
	var pe *ParseError
	require.True(t, errors.As(err, &pe))
	require.Equal(t, "!2026-13-40", pe.Token)
}

func TestQuickEntry_EmptyTagRejected(t *testing.T) {
	_, err := Parse("x #")
	var pe *ParseError
	require.True(t, errors.As(err, &pe))
}

func TestQuickEntry_EmptyMentionRejected(t *testing.T) {
	_, err := Parse("x @")
	var pe *ParseError
	require.True(t, errors.As(err, &pe))
}

func TestQuickEntry_EmptyInputRejected(t *testing.T) {
	_, err := Parse("   ")
	require.ErrorIs(t, err, ErrEmptyInput)
}

func TestQuickEntry_NoTitleAfterTokens(t *testing.T) {
	_, err := Parse("#shop @today")
	var pe *ParseError
	require.True(t, errors.As(err, &pe))
}

func TestQuickEntry_TokenOrderIndependent(t *testing.T) {
	a, err := Parse("buy milk #shop @today !2026-06-01")
	require.NoError(t, err)
	b, err := Parse("#shop !2026-06-01 buy milk @today")
	require.NoError(t, err)
	require.Equal(t, a.Tags, b.Tags)
	require.Equal(t, a.Deadline.Format("2006-01-02"), b.Deadline.Format("2006-01-02"))
	// titles preserve word order, so "buy milk" vs "buy milk" both
	require.Equal(t, "buy milk", a.Title)
	require.Equal(t, "buy milk", b.Title)
}

// PBT covers CP-7 fragment: invalid date tokens always rejected.
func TestProp_QuickEntryInvalidAbsence(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		month := rapid.IntRange(13, 99).Draw(rt, "month")
		day := rapid.IntRange(32, 99).Draw(rt, "day")
		input := "x !2026-" + intToString(month) + "-" + intToString(day)
		_, err := Parse(input)
		require.Error(rt, err)
	})
}

func intToString(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	d1 := n / 10
	d2 := n % 10
	return string(rune('0'+d1)) + string(rune('0'+d2))
}

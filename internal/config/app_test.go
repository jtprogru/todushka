package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaults_AreValid(t *testing.T) {
	c, warns := Defaults().Validate()
	require.Empty(t, warns)
	require.Equal(t, "macchiato", c.Theme)
	require.Equal(t, 100, c.DualPaneMinWidth)
	require.InDelta(t, 0.45, c.ListPaneShare, 1e-9)
	require.Equal(t, 5, c.BulkConfirmThreshold)
	require.Equal(t, 8, c.NotesMaxLines)
	require.True(t, c.ConfirmDelete)
}

func TestDefaults_ConfirmDeleteTrue(t *testing.T) {
	require.True(t, Defaults().ConfirmDelete)
}

func TestValidate_ThemeFallback(t *testing.T) {
	c, warns := AppConfig{Theme: "unknown"}.Validate()
	require.Equal(t, "macchiato", c.Theme)
	require.Len(t, warns, 1)
}

func TestValidate_NumericRanges(t *testing.T) {
	c, warns := AppConfig{
		Theme:                "macchiato",
		DualPaneMinWidth:     10,
		ListPaneShare:        1.5,
		BulkConfirmThreshold: -1,
		NotesMaxLines:        -5,
	}.Validate()
	require.Len(t, warns, 4)
	require.Equal(t, 100, c.DualPaneMinWidth)
	require.InDelta(t, 0.45, c.ListPaneShare, 1e-9)
	require.Equal(t, 5, c.BulkConfirmThreshold)
	require.Equal(t, 8, c.NotesMaxLines)
}

func TestValidate_AutoThemeIsValid(t *testing.T) {
	c, warns := AppConfig{Theme: "auto"}.Validate()
	require.Empty(t, warns)
	require.Equal(t, "auto", c.Theme)
}

func TestValidate_SystemThemeIsValid(t *testing.T) {
	c, warns := AppConfig{Theme: "system"}.Validate()
	require.Empty(t, warns)
	require.Equal(t, "system", c.Theme)
}

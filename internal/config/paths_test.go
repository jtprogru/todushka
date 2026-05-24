package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaths_XDGFallback(t *testing.T) {
	env := func(k string) string {
		switch k {
		case "HOME":
			return "/tmp/u"
		default:
			return ""
		}
	}

	dataDir, err := DataDir(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/tmp/u", ".local", "share", "todushka"), dataDir)

	stateDir, err := StateDir(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/tmp/u", ".local", "state", "todushka"), stateDir)

	logPath, err := LogPath(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(stateDir, "log"), logPath)
}

func TestPaths_RespectsXDGOverride(t *testing.T) {
	env := func(k string) string {
		switch k {
		case "XDG_DATA_HOME":
			return "/custom/data"
		case "XDG_STATE_HOME":
			return "/custom/state"
		case "HOME":
			return "/tmp/u"
		}
		return ""
	}

	dataDir, err := DataDir(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/custom/data", "todushka"), dataDir)

	stateDir, err := StateDir(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/custom/state", "todushka"), stateDir)
}

func TestPaths_NoHomeReturnsError(t *testing.T) {
	env := func(string) string { return "" }
	_, err := DataDir(env)
	require.ErrorIs(t, err, ErrNoHome)
}

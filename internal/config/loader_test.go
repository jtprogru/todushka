package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// mockEnv returns an Env closure backed by the given map.
func mockEnv(m map[string]string) Env {
	return func(k string) string {
		return m[k]
	}
}

func TestLoad_FileMissingCreatesDefault(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	env := mockEnv(nil)

	cfg, warns, err := Load(path, env)
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, Defaults(), cfg)

	// File should now exist on disk with default contents.
	data, rerr := os.ReadFile(path)
	require.NoError(t, rerr)
	require.Contains(t, string(data), "theme: macchiato")
}

func TestLoad_FileValidParses(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	contents := "theme: latte\n" +
		"dual_pane_min_width: 120\n" +
		"list_pane_share: 0.5\n" +
		"bulk_confirm_threshold: 3\n" +
		"notes_max_lines: 12\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))

	cfg, warns, err := Load(path, mockEnv(nil))
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, "latte", cfg.Theme)
	require.Equal(t, 120, cfg.DualPaneMinWidth)
	require.InDelta(t, 0.5, cfg.ListPaneShare, 1e-9)
	require.Equal(t, 3, cfg.BulkConfirmThreshold)
	require.Equal(t, 12, cfg.NotesMaxLines)
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	contents := "theme: macchiato\nnotes_max_lines: 8\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))

	env := mockEnv(map[string]string{
		"TODUSHKA_NOTES_MAX_LINES": "20",
	})
	cfg, warns, err := Load(path, env)
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, 20, cfg.NotesMaxLines)
}

func TestLoad_FlagOverridesEnv(t *testing.T) {
	// ResolvePath: flag takes precedence over TODUSHKA_CONFIG env var.
	env := mockEnv(map[string]string{
		"TODUSHKA_CONFIG": "/env/config.yaml",
		"HOME":            "/home/user",
	})
	resolved, err := ResolvePath(env, "/flag/config.yaml")
	require.NoError(t, err)
	require.Equal(t, "/flag/config.yaml", resolved)
}

func TestLoad_UnknownYAMLFieldsIgnored(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	contents := "feature_x: true\ntheme: macchiato\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))

	cfg, warns, err := Load(path, mockEnv(nil))
	require.NoError(t, err)
	require.Empty(t, warns)
	require.Equal(t, "macchiato", cfg.Theme)
}

func TestLoad_MalformedYAMLReturnsDefaults(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	contents := "not valid: : ::\n  - [unbalanced\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))

	cfg, warns, err := Load(path, mockEnv(nil))
	require.NoError(t, err)
	require.Equal(t, Defaults(), cfg)
	require.NotEmpty(t, warns)
	// Ensure at least one warning mentions parse failure.
	var found bool
	for _, w := range warns {
		if strings.Contains(w, "config:") || strings.Contains(w, "parse") {
			found = true
			break
		}
	}
	require.True(t, found, "expected a parse-related warning, got %v", warns)
}

func TestLoad_InvalidEnvFallsBackToFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.yaml")
	contents := "notes_max_lines: 8\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))

	env := mockEnv(map[string]string{
		"TODUSHKA_NOTES_MAX_LINES": "abc",
	})
	cfg, warns, err := Load(path, env)
	require.NoError(t, err)
	require.Equal(t, 8, cfg.NotesMaxLines)
	// One warning should mention the bad env var.
	var found bool
	for _, w := range warns {
		if strings.Contains(w, "TODUSHKA_NOTES_MAX_LINES") {
			found = true
			break
		}
	}
	require.True(t, found, "expected a warning about invalid env var, got %v", warns)
}

func TestResolvePath_FlagAbsolute(t *testing.T) {
	resolved, err := ResolvePath(mockEnv(nil), "/abs/path.yaml")
	require.NoError(t, err)
	require.Equal(t, "/abs/path.yaml", resolved)
}

func TestResolvePath_EnvFallback(t *testing.T) {
	env := mockEnv(map[string]string{
		"TODUSHKA_CONFIG": "/env/path.yaml",
	})
	resolved, err := ResolvePath(env, "")
	require.NoError(t, err)
	require.Equal(t, "/env/path.yaml", resolved)
}

func TestResolvePath_XDGDefault(t *testing.T) {
	env := mockEnv(map[string]string{
		"HOME": "/home/user",
	})
	resolved, err := ResolvePath(env, "")
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/user", ".config", "todushka", "config.yaml"), resolved)

	// With XDG_CONFIG_HOME explicitly set.
	env2 := mockEnv(map[string]string{
		"HOME":            "/home/user",
		"XDG_CONFIG_HOME": "/custom/config",
	})
	resolved2, err := ResolvePath(env2, "")
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/custom/config", "todushka", "config.yaml"), resolved2)
}

// TestProp_LoadPrecedence verifies CP-9 (REQ-4.1, 4.6, 4.8): for any
// (fileVal, envVal) pair the env value wins for notes_max_lines.
func TestProp_LoadPrecedence(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.yaml")
		fileVal := rapid.IntRange(1, 30).Draw(rt, "fileVal")
		envVal := rapid.IntRange(1, 30).Draw(rt, "envVal")
		content := fmt.Sprintf("notes_max_lines: %d\n", fileVal)
		require.NoError(rt, os.WriteFile(path, []byte(content), 0600))
		env := mockEnv(map[string]string{
			"TODUSHKA_NOTES_MAX_LINES": fmt.Sprintf("%d", envVal),
		})
		cfg, _, err := Load(path, env)
		require.NoError(rt, err)
		require.Equal(rt, envVal, cfg.NotesMaxLines)
	})
}

// TestProp_ValidateCorrectsInvalid verifies CP-10 (REQ-4.5, 4.7): for any
// AppConfig with all-invalid values, Validate emits warnings and returns
// a config equal to Defaults() in every field.
func TestProp_ValidateCorrectsInvalid(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		bad := AppConfig{
			Theme:                rapid.SampledFrom([]string{"unknown", "x", "bad"}).Draw(rt, "theme"),
			DualPaneMinWidth:     rapid.IntRange(-10, 39).Draw(rt, "dpw"),
			ListPaneShare:        rapid.Float64Range(1.1, 5.0).Draw(rt, "lps"),
			BulkConfirmThreshold: rapid.IntRange(-5, 0).Draw(rt, "bct"),
			NotesMaxLines:        rapid.IntRange(-10, 0).Draw(rt, "nml"),
		}
		cfg, warns := bad.Validate()
		require.NotEmpty(rt, warns)
		def := Defaults()
		require.Equal(rt, def.Theme, cfg.Theme)
		require.Equal(rt, def.DualPaneMinWidth, cfg.DualPaneMinWidth)
		require.Equal(rt, def.ListPaneShare, cfg.ListPaneShare)
		require.Equal(rt, def.BulkConfirmThreshold, cfg.BulkConfirmThreshold)
		require.Equal(rt, def.NotesMaxLines, cfg.NotesMaxLines)
	})
}

// TestProp_AutoCreateRoundTrip verifies CP-11 (REQ-4.2): the first Load
// auto-creates the file on disk; a second Load returns the exact same
// AppConfig and finds the file present.
func TestProp_AutoCreateRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "sub", "config.yaml")
		env := mockEnv(nil)
		cfg1, _, err := Load(path, env)
		require.NoError(rt, err)
		cfg2, _, err := Load(path, env)
		require.NoError(rt, err)
		require.Equal(rt, cfg1, cfg2)
		_, err = os.Stat(path)
		require.NoError(rt, err)
	})
}

// TestProp_FlagBeatsEnv verifies CP-13 (REQ-4.1): a non-empty --config
// flag always beats TODUSHKA_CONFIG in ResolvePath.
func TestProp_FlagBeatsEnv(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		flag := "/abs/flag/path.yaml"
		env := mockEnv(map[string]string{"TODUSHKA_CONFIG": "/abs/env/path.yaml"})
		p, err := ResolvePath(env, flag)
		require.NoError(rt, err)
		require.Equal(rt, flag, p)
	})
}

// TestProp_UnknownYAMLFieldsIgnored verifies CP-14 (REQ-4.4): unknown
// fields in the YAML file are ignored without affecting the parsed
// AppConfig.
func TestProp_UnknownYAMLFieldsIgnored(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "config.yaml")
		unknownKey := rapid.StringMatching(`[a-z_]{3,15}`).Draw(rt, "key")
		content := fmt.Sprintf("theme: macchiato\n%s: true\n", unknownKey)
		require.NoError(rt, os.WriteFile(path, []byte(content), 0600))
		cfg, _, err := Load(path, mockEnv(nil))
		require.NoError(rt, err)
		require.Equal(rt, "macchiato", cfg.Theme)
	})
}

package tui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestResolveAutoTheme_DarkMode(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return true, nil }
	require.Equal(t, "macchiato", resolveAutoTheme())
}

func TestResolveAutoTheme_LightMode(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return false, nil }
	require.Equal(t, "latte", resolveAutoTheme())
}

func TestResolveAutoTheme_ErrorFallsBackToDark(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return false, errors.New("boom") }
	require.Equal(t, "macchiato", resolveAutoTheme())
}

func TestSelectThemeFromConfig_AutoDarkUsesMacchiato(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return true, nil }
	th := selectThemeFromConfig("auto", func(string) string { return "" })
	require.Equal(t, "catppuccin-macchiato", th.Name)
}

func TestSelectThemeFromConfig_AutoLightUsesLatte(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return false, nil }
	th := selectThemeFromConfig("auto", func(string) string { return "" })
	require.Equal(t, "catppuccin-latte", th.Name)
}

func TestSelectThemeFromConfig_NoColorOverridesAuto(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return true, nil }
	env := func(k string) string {
		if k == "NO_COLOR" {
			return "1"
		}
		return ""
	}
	th := selectThemeFromConfig("auto", env)
	require.Equal(t, "monochrome", th.Name)
}

func TestSelectThemeFromConfig_SystemAliasMatchesAuto(t *testing.T) {
	orig := detectDarkModeFn
	defer func() { detectDarkModeFn = orig }()
	detectDarkModeFn = func() (bool, error) { return true, nil }
	th1 := selectThemeFromConfig("auto", func(string) string { return "" })
	th2 := selectThemeFromConfig("system", func(string) string { return "" })
	require.Equal(t, th1.Name, th2.Name)
}

// TestProp_AutoThemeResolution verifies CP-1 (REQ-1.1, 1.2, 1.3, 1.4):
// resolveAutoTheme maps detected mode dark/err → macchiato, light → latte.
func TestProp_AutoThemeResolution(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		isDark := rapid.Bool().Draw(rt, "isDark")
		wantErr := rapid.Bool().Draw(rt, "wantErr")
		orig := detectDarkModeFn
		defer func() { detectDarkModeFn = orig }()
		detectDarkModeFn = func() (bool, error) {
			if wantErr {
				return false, errors.New("simulated")
			}
			return isDark, nil
		}
		got := resolveAutoTheme()
		if wantErr || isDark {
			require.Equal(rt, "macchiato", got)
		} else {
			require.Equal(rt, "latte", got)
		}
	})
}

// TestProp_NoColorOverridesAuto verifies CP-2 (REQ-1.6): regardless of the
// configured theme name, NO_COLOR=1 forces the monochrome theme.
func TestProp_NoColorOverridesAuto(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		themes := []string{"auto", "system", "macchiato", "latte", "mono"}
		name := rapid.SampledFrom(themes).Draw(rt, "themeName")
		orig := detectDarkModeFn
		defer func() { detectDarkModeFn = orig }()
		detectDarkModeFn = func() (bool, error) { return true, nil }
		env := func(k string) string {
			if k == "NO_COLOR" {
				return "1"
			}
			return ""
		}
		th := selectThemeFromConfig(name, env)
		require.Equal(rt, "monochrome", th.Name)
	})
}

// TestProp_AutoSystemAliases verifies CP-11 (REQ-1.1, 1.5): "auto" and
// "system" produce identical themes for the same detected mode.
func TestProp_AutoSystemAliases(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		isDark := rapid.Bool().Draw(rt, "isDark")
		orig := detectDarkModeFn
		defer func() { detectDarkModeFn = orig }()
		detectDarkModeFn = func() (bool, error) { return isDark, nil }
		env := func(string) string { return "" }
		th1 := selectThemeFromConfig("auto", env)
		th2 := selectThemeFromConfig("system", env)
		require.Equal(rt, th1.Name, th2.Name)
	})
}

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// ResolvePath returns the path to use for the config file.
// Precedence: flag (if non-empty) > env("TODUSHKA_CONFIG") > $XDG_CONFIG_HOME/todushka/config.yaml
func ResolvePath(env Env, flag string) (string, error) {
	if flag != "" {
		return filepath.Abs(flag)
	}
	if env == nil {
		env = OSEnv
	}
	if p := env("TODUSHKA_CONFIG"); p != "" {
		return filepath.Abs(p)
	}
	dir, err := resolveDir(env, "XDG_CONFIG_HOME", ".config")
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load resolves the config path and returns the final AppConfig
// (after precedence cascade + validation) and any warnings encountered.
func Load(path string, env Env) (AppConfig, []string, error) {
	if env == nil {
		env = OSEnv
	}
	cfg, warns, err := loadFromFile(path)
	if err != nil {
		// file read/parse error → warn, return defaults
		warns = append(warns, fmt.Sprintf("config: %v; using defaults", err))
		cfg = Defaults()
	}
	cfg, envWarns := applyEnv(cfg, env)
	warns = append(warns, envWarns...)
	cfg, validateWarns := cfg.Validate()
	warns = append(warns, validateWarns...)
	return cfg, warns, nil
}

// loadFromFile reads `path`. If file does not exist, auto-creates it with
// defaults + inline comments and returns Defaults().
func loadFromFile(path string) (AppConfig, []string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is the resolved user config file location
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if cerr := createDefaultConfig(path); cerr != nil {
				return Defaults(), []string{fmt.Sprintf("could not create %s: %v", path, cerr)}, nil
			}
			return Defaults(), nil, nil
		}
		return Defaults(), nil, err
	}
	cfg := Defaults()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Defaults(), nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil, nil
}

func createDefaultConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(defaultConfigYAML()), 0600)
}

func defaultConfigYAML() string {
	return `# todushka configuration. See https://github.com/jtprogru/todushka
# for documentation. Edit values below to customize.

# Color theme: macchiato | latte | mono
theme: macchiato

# Minimum terminal width (columns) to activate dual-pane layout.
dual_pane_min_width: 100

# Fraction of width allocated to the list pane in dual-pane mode (0.0 - 1.0).
list_pane_share: 0.45

# Minimum number of selected tasks before a bulk operation requires confirmation.
bulk_confirm_threshold: 5

# Maximum lines of task notes displayed in the details pane.
notes_max_lines: 8

# Show a confirmation modal before deleting a single task. Bulk deletes
# (>= bulk_confirm_threshold) always require confirmation regardless.
confirm_delete: true
`
}

func applyEnv(cfg AppConfig, env Env) (AppConfig, []string) {
	var warns []string
	if v := env("TODUSHKA_THEME"); v != "" {
		cfg.Theme = v
	}
	if v := env("TODUSHKA_DUAL_PANE_MIN_WIDTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.DualPaneMinWidth = n
		} else {
			warns = append(warns, "TODUSHKA_DUAL_PANE_MIN_WIDTH="+v+" not an integer")
		}
	}
	if v := env("TODUSHKA_LIST_PANE_SHARE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.ListPaneShare = f
		} else {
			warns = append(warns, "TODUSHKA_LIST_PANE_SHARE="+v+" not a float")
		}
	}
	if v := env("TODUSHKA_BULK_CONFIRM_THRESHOLD"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.BulkConfirmThreshold = n
		} else {
			warns = append(warns, "TODUSHKA_BULK_CONFIRM_THRESHOLD="+v+" not an integer")
		}
	}
	if v := env("TODUSHKA_NOTES_MAX_LINES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.NotesMaxLines = n
		} else {
			warns = append(warns, "TODUSHKA_NOTES_MAX_LINES="+v+" not an integer")
		}
	}
	if v := env("TODUSHKA_CONFIRM_DELETE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.ConfirmDelete = b
		} else {
			warns = append(warns, "TODUSHKA_CONFIRM_DELETE="+v+" not a bool")
		}
	}
	return cfg, warns
}

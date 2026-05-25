package cli

import (
	"errors"
	"fmt"

	"github.com/jtprogru/todushka/internal/config"
	"github.com/spf13/cobra"
)

// NewRootCmd builds the root Cobra command, with all subcommands attached.
// When invoked without arguments, it launches the TUI via deps.LaunchTUI.
func NewRootCmd(deps Deps) *cobra.Command {
	var configFlag string
	root := &cobra.Command{
		Use:           "todushka",
		Short:         "Things 3-style todo TUI for the terminal",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			path, err := config.ResolvePath(deps.Env, configFlag)
			if err != nil {
				// Path resolution failed (e.g. HOME unset in tests).
				// Fall back to defaults silently — non-fatal.
				deps.Config = config.Defaults()
				return nil
			}
			cfg, warns, _ := config.Load(path, deps.Env)
			for _, w := range warns {
				_, _ = fmt.Fprintln(deps.Stderr, "warning:", w)
			}
			deps.Config = cfg
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if deps.LaunchTUI == nil {
				return errors.New("todushka: TUI launcher not configured")
			}
			return deps.LaunchTUI(deps.Service, deps.Config)
		},
	}
	root.PersistentFlags().StringVar(&configFlag, "config", "", "Path to config file (overrides default)")
	root.SetOut(deps.Stdout)
	root.SetErr(deps.Stderr)
	root.SetIn(deps.Stdin)

	root.AddCommand(newAddCmd(deps))
	root.AddCommand(newTodayCmd(deps))
	root.AddCommand(newCompleteCmd(deps))
	root.AddCommand(newExportCmd(deps))
	root.AddCommand(newImportCmd(deps))
	root.AddCommand(newAreaCmd(deps))
	root.AddCommand(newProjectCmd(deps))
	root.AddCommand(newVersionCmd(deps))
	return root
}

// Execute is a convenience that builds the root command and runs it.
func Execute(deps Deps) error {
	root := NewRootCmd(deps)
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(deps.Stderr, err.Error())
		return err
	}
	return nil
}

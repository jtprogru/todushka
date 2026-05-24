package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newTodayCmd(deps Deps) *cobra.Command {
	var jsonMode bool
	cmd := &cobra.Command{
		Use:   "today",
		Short: "List tasks for today",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			tasks, err := deps.Service.ListToday(ctx)
			if err != nil {
				return err
			}
			return FormatTaskList(deps.Stdout, tasks, jsonMode, ColorEnabled(deps.Env))
		},
	}
	cmd.Flags().BoolVar(&jsonMode, "json", false, "Output as JSON")
	return cmd
}

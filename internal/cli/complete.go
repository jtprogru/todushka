package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/spf13/cobra"
)

func newCompleteCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "complete <task-id>",
		Short: "Mark a task as completed (by short or full ID)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			candidates, err := deps.Service.FindTaskByShort(ctx, args[0])
			if err != nil {
				return err
			}
			switch len(candidates) {
			case 0:
				return fmt.Errorf("%w: %q", app.ErrTaskNotFound, args[0])
			case 1:
				next, err := deps.Service.CompleteTask(ctx, candidates[0].ID)
				if err != nil {
					return err
				}
				if next != nil {
					_, _ = fmt.Fprintf(deps.Stdout, "completed (recurring next: %s)\n", id.Short(next.ID))
				} else {
					_, _ = fmt.Fprintln(deps.Stdout, "completed")
				}
				return nil
			default:
				_, _ = fmt.Fprintln(deps.Stderr, "ambiguous prefix, candidates:")
				for _, c := range candidates {
					_, _ = fmt.Fprintf(deps.Stderr, "  %s  %s\n", id.Short(c.ID), c.Title)
				}
				return errors.New("ambiguous task id")
			}
		},
	}
}

package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jtprogru/todushka/internal/app"
	domainarea "github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/spf13/cobra"
)

func newAreaCmd(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "area",
		Short: "Manage Areas (work / home / etc.)",
	}
	root.AddCommand(newAreaAddCmd(deps))
	root.AddCommand(newAreaListCmd(deps))
	root.AddCommand(newAreaDeleteCmd(deps))
	return root
}

func newAreaAddCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Create an Area",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			name := strings.Join(args, " ")
			a, err := deps.Service.AddArea(ctx, name)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(deps.Stdout, id.Short(a.ID))
			return nil
		},
	}
}

func newAreaListCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Areas",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			areas, err := deps.Service.ListAreas(ctx)
			if err != nil {
				return err
			}
			if len(areas) == 0 {
				_, _ = fmt.Fprintln(deps.Stdout, "(no areas)")
				return nil
			}
			for _, a := range areas {
				_, _ = fmt.Fprintf(deps.Stdout, "%s  %s\n", id.Short(a.ID), a.Name)
			}
			return nil
		},
	}
}

func newAreaDeleteCmd(deps Deps) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an Area (moves children to Inbox with --force when not empty)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			name := strings.Join(args, " ")
			a, err := resolveAreaByName(ctx, deps.Service, name)
			if err != nil {
				return err
			}
			err = deps.Service.DeleteArea(ctx, a.ID, force)
			if errors.Is(err, app.ErrAreaNotEmpty) {
				return fmt.Errorf("%w: pass --force to move children to Inbox and delete", app.ErrAreaNotEmpty)
			}
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(deps.Stdout, "deleted")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Move children to Inbox before deleting")
	return cmd
}

func resolveAreaByName(ctx context.Context, svc *app.Service, name string) (domainarea.Area, error) {
	a, err := svc.Repo().AreaFindByNormalized(ctx, domainarea.Normalize(name))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return domainarea.Area{}, fmt.Errorf("area %q: %w", name, storage.ErrNotFound)
		}
		return domainarea.Area{}, err
	}
	return a, nil
}

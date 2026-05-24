package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/spf13/cobra"
)

func newProjectCmd(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "project",
		Short: "Manage Projects",
	}
	root.AddCommand(newProjectAddCmd(deps))
	root.AddCommand(newProjectListCmd(deps))
	root.AddCommand(newProjectDeleteCmd(deps))
	return root
}

func newProjectAddCmd(deps Deps) *cobra.Command {
	var (
		areaName    string
		deadlineStr string
		autoClose   bool
	)
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a Project",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			name := strings.Join(args, " ")
			in := app.AddProjectInput{Name: name, AutoClose: autoClose}
			if areaName != "" {
				a, err := resolveAreaByName(ctx, deps.Service, areaName)
				if err != nil {
					return fmt.Errorf("--area: %w", err)
				}
				aid := a.ID
				in.AreaID = &aid
			}
			if deadlineStr != "" {
				d, err := parseDate(deadlineStr)
				if err != nil {
					return fmt.Errorf("--deadline: %w", err)
				}
				in.Deadline = &d
			}
			p, err := deps.Service.AddProject(ctx, in)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(deps.Stdout, id.Short(p.ID))
			return nil
		},
	}
	cmd.Flags().StringVar(&areaName, "area", "", "Parent area name")
	cmd.Flags().StringVar(&deadlineStr, "deadline", "", "Project deadline YYYY-MM-DD")
	cmd.Flags().BoolVar(&autoClose, "auto-close", false, "Close project automatically when all tasks are done")
	return cmd
}

func newProjectListCmd(deps Deps) *cobra.Command {
	var areaName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			var areaID *id.ID
			if areaName != "" {
				a, err := resolveAreaByName(ctx, deps.Service, areaName)
				if err != nil {
					return fmt.Errorf("--area: %w", err)
				}
				aid := a.ID
				areaID = &aid
			}
			projects, err := deps.Service.ListProjects(ctx, areaID)
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				_, _ = fmt.Fprintln(deps.Stdout, "(no projects)")
				return nil
			}
			for _, p := range projects {
				suffix := ""
				if p.Deadline != nil {
					suffix = "  due:" + p.Deadline.Format("2006-01-02")
				}
				_, _ = fmt.Fprintf(deps.Stdout, "%s  %s  [%s]%s\n", id.Short(p.ID), p.Name, p.Status, suffix)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&areaName, "area", "", "Filter by area name")
	return cmd
}

func newProjectDeleteCmd(deps Deps) *cobra.Command {
	var soft bool
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a Project",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			name := strings.Join(args, " ")
			projects, err := deps.Service.Repo().ProjectFindByName(ctx, name)
			if err != nil {
				return err
			}
			switch len(projects) {
			case 0:
				return fmt.Errorf("project %q: not found", name)
			case 1:
				if err := deps.Service.Repo().ProjectDelete(ctx, projects[0].ID, soft); err != nil {
					return err
				}
				_, _ = fmt.Fprintln(deps.Stdout, "deleted")
				return nil
			default:
				_, _ = fmt.Fprintln(deps.Stderr, "ambiguous project name, candidates:")
				for _, p := range projects {
					line := id.Short(p.ID) + "  " + p.Name
					if p.AreaID != nil {
						a, _ := deps.Service.Repo().AreaGet(ctx, *p.AreaID)
						line += "  (area: " + a.Name + ")"
					}
					_, _ = fmt.Fprintln(deps.Stderr, "  "+line)
				}
				return errors.New("ambiguous project")
			}
		},
	}
	cmd.Flags().BoolVar(&soft, "soft", false, "Soft-delete (move to Trash)")
	return cmd
}

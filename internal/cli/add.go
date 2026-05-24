package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/app"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/spf13/cobra"
)

func newAddCmd(deps Deps) *cobra.Command {
	var (
		projectName string
		areaName    string
		tags        []string
		startStr    string
		deadlineStr string
		someday     bool
	)
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := strings.Join(args, " ")
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			in := app.AddTaskInput{Title: title, Someday: someday}

			if startStr != "" {
				d, err := parseDate(startStr)
				if err != nil {
					return fmt.Errorf("--start: %w", err)
				}
				in.StartDate = &d
			}
			if deadlineStr != "" {
				d, err := parseDate(deadlineStr)
				if err != nil {
					return fmt.Errorf("--deadline: %w", err)
				}
				in.Deadline = &d
			}
			if projectName != "" {
				projects, err := deps.Service.Repo().ProjectFindByName(ctx, projectName)
				if err != nil {
					return err
				}
				switch len(projects) {
				case 0:
					return fmt.Errorf("--project: %w (%q)", app.ErrUnknownProject, projectName)
				case 1:
					pid := projects[0].ID
					in.ProjectID = &pid
					if projects[0].AreaID != nil {
						aid := *projects[0].AreaID
						in.AreaID = &aid
					}
				default:
					return fmt.Errorf("--project: %w (%q)", app.ErrAmbiguousProject, projectName)
				}
			}
			if areaName != "" {
				a, err := deps.Service.Repo().AreaFindByNormalized(ctx, strings.ToLower(strings.TrimSpace(areaName)))
				if err != nil {
					return fmt.Errorf("--area: %w", err)
				}
				aid := a.ID
				in.AreaID = &aid
			}
			tagIDs := make([]id.ID, 0, len(tags))
			for _, name := range tags {
				t, err := deps.Service.UpsertTag(ctx, name)
				if err != nil {
					return fmt.Errorf("--tag: %w", err)
				}
				tagIDs = append(tagIDs, t.ID)
			}
			in.Tags = tagIDs

			tk, err := deps.Service.AddTask(ctx, in)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(deps.Stdout, id.Short(tk.ID))
			return nil
		},
	}
	cmd.Flags().StringVar(&projectName, "project", "", "Project name (exact match)")
	cmd.Flags().StringVar(&areaName, "area", "", "Area name")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tag (repeatable)")
	cmd.Flags().StringVar(&startStr, "start", "", "Start date YYYY-MM-DD")
	cmd.Flags().StringVar(&deadlineStr, "deadline", "", "Deadline YYYY-MM-DD")
	cmd.Flags().BoolVar(&someday, "someday", false, "Add to Someday list")
	return cmd
}

func parseDate(s string) (task.Date, error) {
	if strings.TrimSpace(s) == "" {
		return task.Date{}, errors.New("empty date")
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return task.Date{}, err
	}
	return task.NewDate(t), nil
}

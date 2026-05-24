package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

func newExportCmd(deps Deps) *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the entire database as JSON",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			out := deps.Stdout
			if path != "" {
				f, err := os.Create(path) //nolint:gosec // path is user-supplied flag
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				out = f
			}
			return deps.Service.ExportJSON(ctx, out)
		},
	}
	cmd.Flags().StringVar(&path, "json", "", "Output file (stdout if empty)")
	return cmd
}

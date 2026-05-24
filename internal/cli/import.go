package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newImportCmd(deps Deps) *cobra.Command {
	var (
		path string
		yes  bool
	)
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Replace database content from a JSON snapshot",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if path == "" {
				return errors.New("--json <path> is required")
			}
			if !yes {
				return errors.New("import will REPLACE the database; pass --yes to confirm")
			}
			f, err := os.Open(path) //nolint:gosec // path is user-supplied flag
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			snap, err := deps.Service.ImportJSON(ctx, f)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(deps.Stdout, "imported %d areas, %d projects, %d tags, %d tasks\n",
				len(snap.Areas), len(snap.Projects), len(snap.Tags), len(snap.Tasks))
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "json", "", "Snapshot file to import")
	cmd.Flags().BoolVar(&yes, "yes", false, "Confirm DB replacement")
	return cmd
}

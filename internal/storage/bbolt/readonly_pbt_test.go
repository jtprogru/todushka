package bbolt

// T-8 — Property-based tests for the read-only mode feature.
//
// Coverage map (design.md §2.6):
//   - CP-1 (writes return ErrReadOnly): exhaustively covered by
//     TestBbolt_AllWritesReturnErrReadOnly in bbolt_test.go. A rapid PBT
//     over the 16 write methods adds no signal beyond the existing
//     table-driven test, so no PBT is added here.
//   - CP-2 (reads work in both modes): covered below by
//     TestProp_Bbolt_ReadsEqualAcrossModes — random N tasks written via RW,
//     reopened RO, TaskList returns the same set.
//   - CP-3 (OpenAuto fallback): N/A — OpenAuto was removed during T-3
//     (Option B pivot). bbolt's flock cannot grant LOCK_SH while another
//     process holds LOCK_EX, so the property is unrealisable.
//   - CP-4, CP-5, CP-6, CP-7: covered by existing unit tests
//     (TestBbolt_OpenReadOnlyTrue, TestBbolt_OpenRWReadOnlyFalse,
//     TestBbolt_MigrateNoOpInRO, TestFakes_ReadOnlyAlwaysFalse).
//   - CP-15 (auto-fallback warning to stderr): N/A — no auto-fallback
//     exists post-pivot. main.go prints a clear locked-DB error and
//     exits; no warning channel to property-check.

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// TestProp_Bbolt_ReadsEqualAcrossModes verifies CP-2: for any number of
// tasks written via a writable Repo, reopening the same database in
// read-only mode yields the identical task set via TaskList.
func TestProp_Bbolt_ReadsEqualAcrossModes(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		n := rapid.IntRange(1, 10).Draw(rt, "taskCount")
		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "todushka.db")

		rw, err := Open(path)
		require.NoError(rt, err)

		titles := make([]string, 0, n)
		for i := 0; i < n; i++ {
			title := fmt.Sprintf("task-%02d", i)
			titles = append(titles, title)
			tk := task.Task{
				ID:        id.New(),
				Title:     title,
				Status:    task.StatusOpen,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			require.NoError(rt, rw.TaskCreate(ctx, tk))
		}
		require.NoError(rt, rw.Close())

		ro, err := OpenReadOnly(path)
		require.NoError(rt, err)
		defer ro.Close()

		require.True(rt, ro.ReadOnly())

		got, err := ro.TaskList(ctx, storage.TaskFilter{})
		require.NoError(rt, err)
		require.Len(rt, got, n)

		gotTitles := make([]string, 0, len(got))
		for _, tk := range got {
			gotTitles = append(gotTitles, tk.Title)
		}
		sort.Strings(gotTitles)
		sort.Strings(titles)
		require.Equal(rt, titles, gotTitles)
	})
}

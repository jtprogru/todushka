package bbolt

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"testing"
	"time"

	domainarea "github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/tag"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func tempDB(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "todushka.db")
}

func TestBbolt_OpenCreatesFreshDB(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()
	v, err := r.SchemaVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, storage.CurrentSchemaVersion, v)
}

func TestBbolt_OpenLocked(t *testing.T) {
	path := tempDB(t)
	r1, err := Open(path)
	require.NoError(t, err)
	defer r1.Close()

	_, err = Open(path)
	require.ErrorIs(t, err, storage.ErrDatabaseLocked)
}

func TestBbolt_FutureSchemaRejected(t *testing.T) {
	path := tempDB(t)
	{
		r, err := Open(path)
		require.NoError(t, err)
		err = r.db.Update(func(tx *bolt.Tx) error {
			bkt := tx.Bucket([]byte(bktMeta))
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, uint32(storage.CurrentSchemaVersion+10))
			return bkt.Put([]byte(keySchemaVersion), buf)
		})
		require.NoError(t, err)
		require.NoError(t, r.Close())
	}
	_, err := Open(path)
	require.ErrorIs(t, err, storage.ErrSchemaTooNew)
}

func TestBbolt_MigrationRunsFromZero(t *testing.T) {
	path := tempDB(t)
	{
		db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
		require.NoError(t, err)
		err = db.Update(func(tx *bolt.Tx) error {
			bkt, err := tx.CreateBucketIfNotExists([]byte(bktMeta))
			if err != nil {
				return err
			}
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, 0)
			return bkt.Put([]byte(keySchemaVersion), buf)
		})
		require.NoError(t, err)
		require.NoError(t, db.Close())
	}
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()
	v, err := r.SchemaVersion(context.Background())
	require.NoError(t, err)
	require.Equal(t, storage.CurrentSchemaVersion, v)
}

func TestBbolt_TaskRoundTripAcrossReopen(t *testing.T) {
	// CP-24: synchronous commit durability — survive close + reopen.
	ctx := context.Background()
	path := tempDB(t)

	tid := id.New()
	tk := task.Task{
		ID:        tid,
		Title:     "persisted",
		Status:    task.StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	r1, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, r1.TaskCreate(ctx, tk))
	require.NoError(t, r1.Close())

	r2, err := Open(path)
	require.NoError(t, err)
	defer r2.Close()
	got, err := r2.TaskGet(ctx, tid)
	require.NoError(t, err)
	require.Equal(t, "persisted", got.Title)
	require.Equal(t, task.StatusOpen, got.Status)
}

func TestBbolt_TaskRoundTrip(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()

	tid := id.New()
	tk := task.Task{
		ID:        tid,
		Title:     "milk",
		Status:    task.StatusOpen,
		Tags:      []id.ID{id.New()},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, r.TaskCreate(ctx, tk))
	got, err := r.TaskGet(ctx, tid)
	require.NoError(t, err)
	require.Equal(t, tk.ID, got.ID)
	require.Equal(t, tk.Title, got.Title)
	require.Equal(t, tk.Status, got.Status)
	require.Equal(t, tk.Tags, got.Tags)
}

func TestBbolt_TaskMatchShortDisambiguates(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()

	// Craft IDs that share a 6-char prefix.
	id1 := id.ID("00000000000000000000000001")
	id2 := id.ID("00000000000000000000000002")
	require.NoError(t, r.TaskCreate(ctx, task.Task{ID: id1, Title: "a", Status: task.StatusOpen}))
	require.NoError(t, r.TaskCreate(ctx, task.Task{ID: id2, Title: "b", Status: task.StatusOpen}))

	got, err := r.TaskMatchShort(ctx, "000000")
	require.NoError(t, err)
	require.Len(t, got, 2)
}

func TestBbolt_TagUpsertIdempotent(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()

	a, err := r.TagUpsert(ctx, "Shop")
	require.NoError(t, err)
	b, err := r.TagUpsert(ctx, "  shop ")
	require.NoError(t, err)
	require.Equal(t, a.ID, b.ID)
	require.Equal(t, "shop", a.Normalized)
}

func TestBbolt_TagDeletePreservesTasks(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()

	tg, err := r.TagUpsert(ctx, "shop")
	require.NoError(t, err)
	tid := id.New()
	require.NoError(t, r.TaskCreate(ctx, task.Task{
		ID: tid, Title: "milk", Status: task.StatusOpen,
		Tags: []id.ID{tg.ID},
	}))

	require.NoError(t, r.TagDelete(ctx, tg.ID))

	got, err := r.TaskGet(ctx, tid)
	require.NoError(t, err)
	require.NotContains(t, got.Tags, tg.ID)
}

func TestBbolt_OpenRWReadOnlyFalse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()
	require.False(t, r.ReadOnly())
}

func TestBbolt_OpenReadOnlyTrue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	rw, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, rw.Close())

	ro, err := OpenReadOnly(path)
	require.NoError(t, err)
	defer ro.Close()
	require.True(t, ro.ReadOnly())
}

func TestBbolt_OpenReadOnlyMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.db")
	_, err := OpenReadOnly(path)
	require.Error(t, err)
}

// NOTE: OpenAuto was removed during T-3 (Option B pivot). bbolt's flock
// semantics prevent LOCK_SH (read-only) from being granted while LOCK_EX
// (writer) is held by another process — so auto-fallback to RO while the
// primary instance is alive is impossible without snapshotting. The
// feature pivoted to "explicit --readonly flag only; on lock conflict
// print a clear error". See readonly-mode/explore.md "Option B" section.

func TestBbolt_OpenLockedReturnsErrDatabaseLocked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	first, err := Open(path)
	require.NoError(t, err)
	defer first.Close()
	_, err = Open(path)
	require.ErrorIs(t, err, storage.ErrDatabaseLocked)
}

func TestBbolt_OpenReadOnlyLockedReturnsErrDatabaseLocked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	first, err := Open(path)
	require.NoError(t, err)
	defer first.Close()
	_, err = OpenReadOnly(path)
	require.ErrorIs(t, err, storage.ErrDatabaseLocked)
}

func TestBbolt_MigrateNoOpInRO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	rw, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, rw.Close())
	ro, err := OpenReadOnly(path)
	require.NoError(t, err)
	defer ro.Close()
	require.NoError(t, ro.Migrate(context.Background(), storage.CurrentSchemaVersion))
}

func TestBbolt_AreaNameCollision(t *testing.T) {
	ctx := context.Background()
	path := tempDB(t)
	r, err := Open(path)
	require.NoError(t, err)
	defer r.Close()

	a1 := domainarea.Area{ID: id.New(), Name: "Work", CreatedAt: time.Now()}
	require.NoError(t, r.AreaCreate(ctx, a1))

	a2 := domainarea.Area{ID: id.New(), Name: "  WORK ", CreatedAt: time.Now()}
	err = r.AreaCreate(ctx, a2)
	require.ErrorIs(t, err, storage.ErrAlreadyExists)
}

func TestBbolt_AllWritesReturnErrReadOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	rw, err := Open(path)
	require.NoError(t, err)
	require.NoError(t, rw.Close())
	ro, err := OpenReadOnly(path)
	require.NoError(t, err)
	defer ro.Close()
	ctx := context.Background()
	fakeID := id.New()
	fakeTask := task.Task{ID: fakeID, Title: "x", Status: task.StatusOpen, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	tests := []struct {
		name string
		op   func() error
	}{
		{"TaskCreate", func() error { return ro.TaskCreate(ctx, fakeTask) }},
		{"TaskUpdate", func() error { return ro.TaskUpdate(ctx, fakeTask) }},
		{"TaskDelete", func() error { return ro.TaskDelete(ctx, fakeID, false) }},
		{"ProjectCreate", func() error {
			return ro.ProjectCreate(ctx, project.Project{ID: fakeID, Name: "p", CreatedAt: time.Now(), UpdatedAt: time.Now()})
		}},
		{"ProjectUpdate", func() error {
			return ro.ProjectUpdate(ctx, project.Project{ID: fakeID, Name: "p"})
		}},
		{"ProjectDelete", func() error { return ro.ProjectDelete(ctx, fakeID, false) }},
		{"HeadingCreate", func() error {
			return ro.HeadingCreate(ctx, project.Heading{ID: fakeID, ProjectID: fakeID, Name: "h"})
		}},
		{"HeadingUpdate", func() error {
			return ro.HeadingUpdate(ctx, project.Heading{ID: fakeID, ProjectID: fakeID, Name: "h"})
		}},
		{"HeadingDelete", func() error { return ro.HeadingDelete(ctx, fakeID) }},
		{"AreaCreate", func() error {
			return ro.AreaCreate(ctx, domainarea.Area{ID: fakeID, Name: "a", CreatedAt: time.Now(), UpdatedAt: time.Now()})
		}},
		{"AreaUpdate", func() error {
			return ro.AreaUpdate(ctx, domainarea.Area{ID: fakeID, Name: "a"})
		}},
		{"AreaDelete", func() error { return ro.AreaDelete(ctx, fakeID, false) }},
		{"TagCreate", func() error {
			return ro.TagCreate(ctx, tag.Tag{ID: fakeID, Name: "t"})
		}},
		{"TagUpsert", func() error {
			_, err := ro.TagUpsert(ctx, "t")
			return err
		}},
		{"TagRename", func() error { return ro.TagRename(ctx, fakeID, "x") }},
		{"TagDelete", func() error { return ro.TagDelete(ctx, fakeID) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.op(), storage.ErrReadOnly)
		})
	}
}

func TestBbolt_ReadsWorkInRO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	rw, err := Open(path)
	require.NoError(t, err)
	// Add some data while writable
	ctx := context.Background()
	tk := task.Task{ID: id.New(), Title: "from-writer", Status: task.StatusOpen, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	require.NoError(t, rw.TaskCreate(ctx, tk))
	require.NoError(t, rw.Close())

	// Reopen read-only
	ro, err := OpenReadOnly(path)
	require.NoError(t, err)
	defer ro.Close()

	// Reads should succeed
	got, err := ro.TaskGet(ctx, tk.ID)
	require.NoError(t, err)
	require.Equal(t, "from-writer", got.Title)

	list, err := ro.TaskList(ctx, storage.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, list, 1)
}

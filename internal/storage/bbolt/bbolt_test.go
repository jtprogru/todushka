package bbolt

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"testing"
	"time"

	domainarea "github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
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

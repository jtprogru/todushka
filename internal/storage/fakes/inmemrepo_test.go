package fakes

import (
	"context"
	"testing"

	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestInMemRepo_TaskRoundTrip(t *testing.T) {
	ctx := context.Background()
	r := New()
	tid := id.New()
	tk := task.Task{ID: tid, Title: "hello", Status: task.StatusOpen}
	require.NoError(t, r.TaskCreate(ctx, tk))
	got, err := r.TaskGet(ctx, tid)
	require.NoError(t, err)
	require.Equal(t, tk, got)
}

func TestInMemRepo_TaskNotFound(t *testing.T) {
	ctx := context.Background()
	r := New()
	_, err := r.TaskGet(ctx, id.New())
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestInMemRepo_TagUpsertIdempotent(t *testing.T) {
	ctx := context.Background()
	r := New()
	a, err := r.TagUpsert(ctx, "Shop")
	require.NoError(t, err)
	b, err := r.TagUpsert(ctx, "  shop ")
	require.NoError(t, err)
	require.Equal(t, a.ID, b.ID)
}

func TestInMemRepo_TagDeletePreservesTasks(t *testing.T) {
	ctx := context.Background()
	r := New()
	tg, err := r.TagUpsert(ctx, "shop")
	require.NoError(t, err)
	tid := id.New()
	tk := task.Task{ID: tid, Title: "milk", Status: task.StatusOpen, Tags: []id.ID{tg.ID}}
	require.NoError(t, r.TaskCreate(ctx, tk))

	require.NoError(t, r.TagDelete(ctx, tg.ID))

	got, err := r.TaskGet(ctx, tid)
	require.NoError(t, err)
	require.NotContains(t, got.Tags, tg.ID)
}

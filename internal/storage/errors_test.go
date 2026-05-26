package storage_test

import (
	"errors"
	"testing"

	"github.com/jtprogru/todushka/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestErrReadOnly_IsSentinel(t *testing.T) {
	require.True(t, errors.Is(storage.ErrReadOnly, storage.ErrReadOnly))
	require.NotEmpty(t, storage.ErrReadOnly.Error())
	require.Contains(t, storage.ErrReadOnly.Error(), "read-only")
}

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/cache"
)

func TestNewNoop(t *testing.T) {
	t.Parallel()

	require.Equal(t, cache.Noop{}, cache.NewNoop())
}

func TestNoop_Get(t *testing.T) {
	t.Parallel()

	_, err := cache.Noop{}.Get(context.Background(), "key")
	require.ErrorIs(t, err, cache.ErrNotFound)
}

func TestNoop_Set(t *testing.T) {
	t.Parallel()

	require.NoError(t, cache.Noop{}.Set(context.Background(), "key", []byte("value"), time.Minute))
}

func TestNoop_Delete(t *testing.T) {
	t.Parallel()

	require.NoError(t, cache.Noop{}.Delete(context.Background(), "key"))
}

func TestNoop_Interface(t *testing.T) {
	t.Parallel()

	var _ cache.Cache = cache.Noop{}
}

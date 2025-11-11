package dynratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimpleCache(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()
	c := NewSimpleCache[int]("test", 10, time.Hour)

	// get from empty cache, expect not found
	v, ok, err := c.Get(ctx, "k1")
	require.NoError(err)
	require.False(ok) // not found
	require.Nil(v)

	// normal set then get
	v1 := 1
	require.NoError(c.SetNX(ctx, "k1", &v1))
	v, ok, err = c.Get(ctx, "k1")
	require.NoError(err)
	require.True(ok) // found
	require.NotNil(v)
	require.Same(&v1, v) // expected value

	// set an existed key, expect no change
	v2 := 2
	require.NoError(c.SetNX(ctx, "k1", &v2)) // set an existed key
	v, ok, err = c.Get(ctx, "k1")
	require.NoError(err)
	require.True(ok) // found
	require.NotNil(v)
	require.Same(&v1, v) // no change

	// get a not existed key, expect not found
	v, ok, err = c.Get(ctx, "k2")
	require.NoError(err)
	require.False(ok) // not found
	require.Nil(v)

	// set a nil pointer
	require.NoError(c.SetNX(ctx, "k2", nil))
	v, ok, err = c.Get(ctx, "k2")
	require.NoError(err)
	require.True(ok) // found
	require.Nil(v)   // but nil

	// set an existed key has nil pointer value, expect it keep nil
	require.NoError(c.SetNX(ctx, "k2", &v2))
	v, ok, err = c.Get(ctx, "k2")
	require.NoError(err)
	require.True(ok)
	require.Nil(v) // keep nil
}

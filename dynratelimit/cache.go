package dynratelimit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluele/gcache"
)

type SimpleCache[T any] struct {
	cache gcache.Cache
}

func NewSimpleCache[T any](name string, maxKey int, ttl time.Duration) *SimpleCache[T] {
	return &SimpleCache[T]{
		cache: gcache.New(maxKey).LRU().Expiration(time.Hour).Build(),
	}
}

func (c *SimpleCache[T]) isNotFound(err error) bool {
	return err == gcache.KeyNotFoundError
}

func (c *SimpleCache[T]) Get(ctx context.Context, key string) (*T, bool, error) {
	v, err := c.cache.Get(key)
	if err != nil {
		if c.isNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	t, ok := v.(*T)
	if !ok {
		return nil, false, fmt.Errorf("failed to cast interface{} as T")
	}
	return t, true, nil
}

func (c *SimpleCache[T]) SetNX(ctx context.Context, key string, val *T) error {
	if _, err := c.cache.Get(key); err != nil {
		if c.isNotFound(err) {
			if err := c.cache.Set(key, val); err != nil {
				return err
			}
			return nil
		}
		return err
	}
	return nil
}

func (r *CachedDynamicRatelimit[K, V]) Get(ctx context.Context, k K) (*V, error) {
	v, ok, err := r.cache.Get(ctx, k.Key())
	if err == nil && ok {
		return v, nil
	}
	if err != nil {
		slog.Error("failed to get from cache of ratelimit", slog.Any("error", err))
	}

	if err := r.Wait(ctx); err != nil {
		return nil, err
	}

	v, err = r.load(ctx, k)
	if err != nil {
		return nil, err
	}

	if err := r.cache.SetNX(ctx, k.Key(), v); err != nil {
		slog.Warn("failed to set into cache of ratelimit", slog.Any("error", err))
	}

	return v, nil
}

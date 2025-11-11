package dynratelimit

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

func NewCachedDynamicRatelimit[K CacheKey, V any](ratelimit *DynamicRatelimit, cache *SimpleCache[V], load func(context.Context, K) (*V, error)) *CachedDynamicRatelimit[K, V] {
	return &CachedDynamicRatelimit[K, V]{
		DynamicRatelimit: ratelimit,
		cache:            cache,
		load:             load,
	}
}

func NewDynamicRatelimit(
	name string,
	refreshInterval time.Duration,
	limit func() rate.Limit,
	onChangeLimit func(rate.Limit),
) *DynamicRatelimit {
	return NewDynamicRatelimit2(name, refreshInterval, newDynamicLimiter(limit, onChangeLimit))
}

func NewDynamicRatelimit2(name string, refreshInterval time.Duration, dynLimiter DynamicLimiter) *DynamicRatelimit {
	l := &DynamicRatelimit{
		name:       name,
		stop:       make(chan struct{}),
		ticker:     *time.NewTicker(refreshInterval),
		lastLimit:  dynLimiter.Limit(),
		dynLimiter: dynLimiter,
	}
	l.Limiter = rate.NewLimiter(l.lastLimit, max(int(l.lastLimit), 1))
	l.asyncRun()
	return l
}

type DynamicLimiter interface {
	Limit() rate.Limit
	OnChange(rate.Limit)
}

type DynamicRatelimit struct {
	*rate.Limiter

	name       string
	stop       chan struct{}
	ticker     time.Ticker
	lastLimit  rate.Limit
	dynLimiter DynamicLimiter
}

func (l *DynamicRatelimit) Close() {
	l.ticker.Stop()
	close(l.stop)
}

type CacheKey interface {
	Key() string
}

type CachedDynamicRatelimit[K CacheKey, V any] struct {
	*DynamicRatelimit
	cache *SimpleCache[V]
	load  func(context.Context, K) (*V, error)
}

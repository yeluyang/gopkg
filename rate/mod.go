package rate

import (
	"time"

	"golang.org/x/time/rate"
)

func NewDynamicLimiter(
	name string,
	refreshInterval time.Duration,
	limit func() rate.Limit,
	onChangeLimit func(rate.Limit),
) *Limiter {
	return NewDynamicLimiter2(name, refreshInterval, newDynamicLimit(limit, onChangeLimit))
}

func NewDynamicLimiter2(name string, refreshInterval time.Duration, dynLimiter DynamicLimit) *Limiter {
	l := &Limiter{
		name:       name,
		stop:       make(chan struct{}),
		ticker:     time.NewTicker(refreshInterval),
		lastLimit:  dynLimiter.Limit(),
		dynLimiter: dynLimiter,
	}
	l.Limiter = rate.NewLimiter(l.lastLimit, max(int(l.lastLimit), 1))
	l.asyncRun()
	return l
}

type DynamicLimit interface {
	Limit() rate.Limit
	OnChange(rate.Limit)
}

type Limiter struct {
	*rate.Limiter

	name       string
	stop       chan struct{}
	ticker     *time.Ticker
	lastLimit  rate.Limit
	dynLimiter DynamicLimit
}

func (l *Limiter) Close() {
	l.ticker.Stop()
	close(l.stop)
}

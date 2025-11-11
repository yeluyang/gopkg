package dynratelimit

import (
	"github.com/yeluyang/gopkg/routine"
	"golang.org/x/time/rate"
)

type dynamicLimiter struct {
	limit    func() rate.Limit
	onChange func(rate.Limit)
}

func (l *dynamicLimiter) Limit() rate.Limit { return l.limit() }
func (l *dynamicLimiter) OnChange(limit rate.Limit) {
	if l.onChange != nil {
		l.onChange(limit)
	}
}

func newDynamicLimiter(
	limit func() rate.Limit,
	onChange func(rate.Limit),
) DynamicLimiter {
	return &dynamicLimiter{limit: limit, onChange: onChange}
}

func (l *DynamicRatelimit) asyncRun() { routine.Go(func() { l.run() }) }
func (l *DynamicRatelimit) run() {
	for {
		select {
		case <-l.stop:
			return
		case <-l.ticker.C:
			curLimit := l.dynLimiter.Limit()
			if curLimit != l.lastLimit {
				l.Limiter.SetLimit(curLimit)
				l.Limiter.SetBurst(max(int(curLimit), 1))
				l.lastLimit = curLimit
				l.dynLimiter.OnChange(curLimit)
			}
		}
	}
}

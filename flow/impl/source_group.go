package impl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/yeluyang/gopkg/flow/errs"
	"github.com/yeluyang/gopkg/flow/impl/utils"
	"github.com/yeluyang/gopkg/flow/interfaces"
)

type sourceGroup[E comparable, V any] struct {
	fanIn   *utils.InFlightChan[Msg[E, V]]
	wg      sync.WaitGroup
	active  atomic.Int64
	notify  func()
	onError func(error)
}

func newSourceGroup[E comparable, V any](
	fanIn *utils.InFlightChan[Msg[E, V]],
	notify func(),
	onError func(error),
) *sourceGroup[E, V] {
	return &sourceGroup[E, V]{
		fanIn:   fanIn,
		notify:  notify,
		onError: onError,
	}
}

func (sg *sourceGroup[E, V]) Start(ctx context.Context, src interfaces.Source[E, V]) {
	sg.active.Add(1)
	sg.wg.Add(1)
	go func() {
		defer sg.wg.Done()
		defer func() {
			sg.active.Add(-1)
			sg.notify()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			msgs, more, err := src.Next(ctx)
			if err != nil {
				sg.onError(errs.CodeSource.From(err))
				return
			}
			for _, msg := range msgs {
				if !sg.fanIn.Send(ctx, msg) {
					return
				}
			}
			if !more {
				return
			}
		}
	}()
}

func (sg *sourceGroup[E, V]) Exhausted() bool {
	return sg.active.Load() <= 0
}

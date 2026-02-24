package impl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/conc/pool"
	"github.com/yeluyang/gopkg/flow/errs"
	"github.com/yeluyang/gopkg/flow/impl/utils"
	"github.com/yeluyang/gopkg/routine"
)

type actorPool[E comparable, V any] struct {
	visitor   V
	endpoints *endpointGroup[E, V]
	onError   func(error)

	fanIn   *utils.InFlightChan[Msg[E, V]]
	active  atomic.Int64
	requeue utils.Queue[Msg[E, V]]
	wake    chan struct{}

	pool        *pool.ErrorPool
	schedulerWg sync.WaitGroup
}

func newActorPool[E comparable, V any](
	concurrency int,
	visitor V,
	endpoints *endpointGroup[E, V],
	fanIn *utils.InFlightChan[Msg[E, V]],
	wake chan struct{},
	onError func(error),
) *actorPool[E, V] {
	return &actorPool[E, V]{
		visitor:   visitor,
		endpoints: endpoints,
		onError:   onError,
		fanIn:     fanIn,
		wake:      wake,
		pool:      pool.New().WithErrors().WithMaxGoroutines(concurrency),
	}
}

func (ap *actorPool[E, V]) Start(ctx context.Context) {
	ap.schedulerWg.Add(1)
	go ap.run(ctx)
}

func (ap *actorPool[E, V]) run(ctx context.Context) {
	defer ap.schedulerWg.Done()

	for {
		// Priority: drain requeue first (non-blocking)
		if msg, ok := ap.requeue.Pop(); ok {
			ap.submit(func() error { return ap.processChain(ctx, msg) })
			continue
		}

		// Queue is empty here. If no pending channel messages, no active
		// workers, and all sources exhausted, there is nothing left to do.
		if ap.endpoints.Exhausted() && ap.fanIn.Load() <= 0 && ap.active.Load() <= 0 {
			return
		}

		// Wait for new work
		select {
		case msg := <-ap.fanIn.C():
			ap.fanIn.Decr()
			ap.submit(func() error { return ap.processChain(ctx, msg) })
		case <-ap.wake:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (ap *actorPool[E, V]) submit(fn func() error) {
	fn = routine.Recover(fn)

	ap.active.Add(1)
	ap.pool.Go(func() error {
		defer func() {
			if ap.active.Add(-1) <= 0 {
				select {
				case ap.wake <- struct{}{}:
				default:
				}
			}
		}()
		err := fn()
		if err != nil {
			ap.onError(err)
		}
		return err
	})
}

func (ap *actorPool[E, V]) processChain(ctx context.Context, msg Msg[E, V]) error {
	for {
		// 1. Activate lazy endpoints
		for _, id := range msg.Activate(ctx) {
			if err := ap.endpoints.Activate(ctx, id); err != nil {
				return errs.CodeActivate.From(err)
			}
		}

		// 2. Check drain targets
		drainTargets := msg.DrainTo(ctx)
		if len(drainTargets) > 0 {
			for _, target := range drainTargets {
				if err := ap.endpoints.DrainTo(ctx, msg, target); err != nil {
					return err
				}
			}
			return nil
		}

		// 3. Accept with visitor
		newMsgs, err := msg.Accept(ctx, ap.visitor)
		if err != nil {
			return errs.CodeAccept.From(err)
		}
		if len(newMsgs) == 0 {
			return nil
		}

		// Emit K-1 to requeue (non-blocking)
		if len(newMsgs) > 1 {
			ap.enqueue(newMsgs[:len(newMsgs)-1])
		}

		// Continue processing the last child in this worker
		msg = newMsgs[len(newMsgs)-1]
	}
}

func (ap *actorPool[E, V]) enqueue(msgs []Msg[E, V]) {
	ap.requeue.Push(msgs)
	select {
	case ap.wake <- struct{}{}:
	default:
	}
}

func (ap *actorPool[E, V]) Wait() {
	ap.schedulerWg.Wait()
	if err := ap.pool.Wait(); err != nil {
		ap.onError(err)
	}
}

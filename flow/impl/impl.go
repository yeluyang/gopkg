package impl

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/yeluyang/gopkg/flow/errs"
	"github.com/yeluyang/gopkg/flow/impl/utils"
	"github.com/yeluyang/gopkg/flow/interfaces"
)

type Msg[E comparable, V any] = interfaces.Message[E, V]

type Config[E comparable, V any] struct {
	Visitor       V
	Concurrency   int
	EagerSources  []interfaces.Source[E, V]
	LazySources   []interfaces.Source[E, V]
	EagerSinks    []interfaces.Sink[E, V]
	LazySinks     []interfaces.Sink[E, V]
	EagerDuplexes []interfaces.Duplex[E, V]
	LazyDuplexes  []interfaces.Duplex[E, V]
}

type Flow[E comparable, V any] struct {
	cfg         Config[E, V]
	concurrency int

	ctx    context.Context
	cancel context.CancelFunc

	endpoints *endpointGroup[E, V]
	actors    *actorPool[E, V]

	firstErr atomic.Pointer[error]
	errOnce  sync.Once
	running  atomic.Bool
}

func New[E comparable, V any](cfg Config[E, V]) *Flow[E, V] {
	concurrency := cfg.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	f := &Flow[E, V]{
		cfg:         cfg,
		concurrency: concurrency,
	}

	wake := make(chan struct{}, 1)
	fanIn := utils.NewInFlightChan[Msg[E, V]]()
	sources := newSourceGroup[E, V](fanIn, func() {
		select {
		case wake <- struct{}{}:
		default:
		}
	}, f.handleError)
	sinks := newSinkGroup[E, V](cfg, f.handleError)
	f.endpoints = newEndpointGroup(sources, sinks)
	f.actors = newActorPool(
		concurrency, cfg.Visitor,
		f.endpoints,
		fanIn, wake,
		f.handleError,
	)

	return f
}

func (f *Flow[E, V]) Run(ctx context.Context) error {
	if !f.running.CompareAndSwap(false, true) {
		return errs.CodeAlreadyRunning.With("flow: already running")
	}
	if len(f.cfg.EagerSources)+len(f.cfg.EagerDuplexes) == 0 {
		return nil
	}

	f.ctx, f.cancel = context.WithCancel(ctx)

	if err := f.endpoints.Init(f.ctx, f.cfg); err != nil {
		f.cancel()
		return errs.CodeActivate.From(err)
	}

	f.actors.Start(f.ctx)

	return nil
}

func (f *Flow[E, V]) handleError(err error) {
	f.errOnce.Do(func() {
		f.firstErr.Store(&err)
		f.cancel()
	})
}

func (f *Flow[E, V]) Wait() error {
	f.actors.Wait()

	f.endpoints.CloseSinks()
	f.endpoints.WaitSinks()

	if errPtr := f.firstErr.Load(); errPtr != nil {
		f.endpoints.Close(f.ctx)
		return *errPtr
	}

	// Parent context cancelled — no internal error, but f.ctx was cancelled
	// via the context chain. handleError is the only caller of f.cancel(),
	// so if firstErr is nil, cancellation came from the parent.
	if f.ctx != nil && f.ctx.Err() != nil {
		f.endpoints.Close(f.ctx)
		return errs.CodeCancelled.From(f.ctx.Err())
	}

	f.endpoints.Close(context.Background())

	return nil
}

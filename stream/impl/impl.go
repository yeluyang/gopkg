package impl

import (
	"context"

	"github.com/yeluyang/gopkg/stream/errs"
	"github.com/yeluyang/gopkg/stream/interfaces"
)

type (
	Source[V, M any]  = interfaces.Source[V, M]
	Message[V, M any] = interfaces.Message[V, M]
	Sink[M any]       = interfaces.Sink[M]
)

type Stream[V, M any] struct {
	actorPool *actorPool[V, M]
	sink Sink[M]
	shutdown  *shutdown
}

// New creates a stream that reads from source, accepts via visitor, and drains to sink.
func New[V, M any](src Source[V, M], visitor V, routines int, sink Sink[M]) *Stream[V, M] {
	sd := newShutdown()
	return &Stream[V, M]{
		actorPool: newActorPool(src, visitor, routines, sd),
		sink:    sink,
		shutdown:  sd,
	}
}

func (s *Stream[V, M]) Run(ctx context.Context) error {
	s.shutdown.WithContext(ctx)
	s.actorPool.run()
	for v := range s.actorPool.C() {
		if err := s.sink.Drain(s.shutdown.ctx, v); err != nil {
			s.shutdown.Abort(errs.CodeDrain.From(err))
			break
		}
	}
	// Drain remaining messages until the channel closes.
	// The channel close is the natural synchronization point —
	// it happens after workers exit, source closes, and coordinator finishes.
	for range s.actorPool.C() {
	}
	return s.shutdown.Err()
}

func (s *Stream[V, M]) Close() error {
	s.shutdown.Stop()
	return s.shutdown.Err()
}

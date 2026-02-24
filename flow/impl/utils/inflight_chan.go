package utils

import (
	"context"
	"sync/atomic"
)

// InFlightChan couples an unbuffered channel with a counter that tracks
// messages sent but not yet received. Send increments before the channel
// send; the receiver calls Decr after reading from C().
type InFlightChan[T any] struct {
	ch       chan T
	inFlight atomic.Int64
}

func NewInFlightChan[T any]() *InFlightChan[T] {
	return &InFlightChan[T]{ch: make(chan T)}
}

// Send increments the counter, then sends msg. If ctx is cancelled during
// the blocking send, the increment is rolled back.
func (c *InFlightChan[T]) Send(ctx context.Context, msg T) bool {
	select {
	case c.ch <- msg:
		c.inFlight.Add(1)
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *InFlightChan[T]) C() <-chan T { return c.ch }
func (c *InFlightChan[T]) Decr() int64 { return c.inFlight.Add(-1) }
func (c *InFlightChan[T]) Load() int64 { return c.inFlight.Load() }

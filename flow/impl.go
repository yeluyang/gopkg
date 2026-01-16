package flow

import (
	"context"
)

type seqMessage struct{}

type flow[S, O comparable, V any] struct {
	source Source[Message[S, O, V]]
	sink   Sink[O, Message[S, O, V]]
}

func (f *flow[S, O, V]) Run(ctx context.Context) error

type source[M any] struct {
	source Source[M]
	ch     chan M
}

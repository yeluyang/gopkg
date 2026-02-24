package interfaces

import (
	"context"
)

type Source[V, M any] interface {
	Next(ctx context.Context) ([]Message[V, M], bool, error)
}

type Message[V, M any] interface {
	Accept(ctx context.Context, visitor V) (M, error)
}

type Sink[M any] interface {
	Drain(ctx context.Context, msgs ...M) error
}

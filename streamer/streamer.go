package streamer

import "context"

type (
	Message[C comparable, V any] interface {
		Code() C
		Accept(ctx context.Context, visitor V) (Message[C, V], error)
	}

	Source[M any] interface {
		Next(ctx context.Context) (M, bool, error)
	}

	Sink[M any] interface {
		Drain(ctx context.Context, msg M) error
	}

	Actor interface {
		Run(ctx context.Context) error
	}

	msgBus[C comparable, M interface{ Code() C }] struct {
		buff map[C]msgBuff[M]
	}

	msgBuff[M any] struct {
		buff   []M
		closed bool
	}

	sourceActor[M any] struct {
		s      Source[M]
		ch     chan M
		closed bool
	}

	sinkActor[M any] struct {
		s Sink[M]
	}

	transAcotr struct{}
)

func NewSource[M any](source Source[M]) {}

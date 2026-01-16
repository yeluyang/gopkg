package interfaces

import (
	"context"

	"github.com/samber/mo"
)

type (
	Message[Seq, SinkID comparable, Visitor any] interface {
		DrainTo() mo.Option[SinkID]
		Sequence() Seq
		Accept(ctx context.Context, visitor Visitor) ([]Message[Seq, SinkID, Visitor], error)
	}

	Source[Msg any] interface {
		Next(ctx context.Context) ([]Msg, bool, error)
	}

	Sink[SinkID comparable, Msg any] interface {
		ID() SinkID
		Drain(ctx context.Context, msg []Msg) error
	}
)

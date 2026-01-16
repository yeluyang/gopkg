package flow

import (
	"context"

	"github.com/samber/mo"
)

type (
	Message[Seq, SinkID comparable, Visitor any] interface {
		Outlet() mo.Option[SinkID]
		Sequence() Seq
		Accept(ctx context.Context, visitor Visitor) (Message[Seq, SinkID, Visitor], error)
	}

	Source[Msg any] interface {
		Next(ctx context.Context) (Msg, bool, error)
	}

	Sink[SinkID comparable, Msg any] interface {
		Inlet() SinkID
		Drain(ctx context.Context, msg Msg) error
	}
)

func New[Seq, SinkID comparable, Visitor any](sources []Source[Message[Seq, SinkID, Visitor]], sinks []Sink[SinkID, Message[Seq, SinkID, Visitor]]) *flow[Seq, SinkID, Visitor] {
	return &flow[Seq, SinkID, Visitor]{}
}

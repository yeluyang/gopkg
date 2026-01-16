package impl

import (
	"context"

	"github.com/yeluyang/gopkg/flow/interfaces"
)

type (
	Message[Seq, SinkID comparable, Visitor any] = interfaces.Message[Seq, SinkID, Visitor]
	Source[Msg any]                              = interfaces.Source[Msg]
	Sink[SinkID comparable, Msg any]             = interfaces.Sink[SinkID, Msg]
)

type seqMessage struct{}

type Flow[S, O comparable, V any] struct {
	source Source[Message[S, O, V]]
	sink   Sink[O, Message[S, O, V]]
}

func (f *Flow[S, O, V]) Run(ctx context.Context) error

type source[M any] struct {
	source Source[M]
	ch     chan M
}

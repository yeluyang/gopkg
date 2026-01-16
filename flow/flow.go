package flow

import (
	"github.com/yeluyang/gopkg/flow/impl"
	"github.com/yeluyang/gopkg/flow/interfaces"
)

type (
	Message[Seq, SinkID comparable, Visitor any] = interfaces.Message[Seq, SinkID, Visitor]
	Source[Msg any]                              = interfaces.Source[Msg]
	Sink[SinkID comparable, Msg any]             = interfaces.Sink[SinkID, Msg]
)

func New[Seq, SinkID comparable, Visitor any](sources []Source[Message[Seq, SinkID, Visitor]], sinks []Sink[SinkID, Message[Seq, SinkID, Visitor]]) *impl.Flow[Seq, SinkID, Visitor] {
	return &impl.Flow[Seq, SinkID, Visitor]{}
}

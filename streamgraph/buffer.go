package streamgraph

import (
	"context"
)

type buffer[Code MsgCode, Visitor MsgVisitor] struct {
	ch chan Message[Code, Visitor]
}

func (b *buffer[Code, Visitor]) Send(ctx context.Context, msg Message[Code, Visitor]) error
func (b *buffer[Code, Visitor]) Recv(ctx context.Context) (Message[Code, Visitor], bool, error)

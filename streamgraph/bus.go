package streamgraph

import (
	"context"
	"fmt"
)

type msgBus[Code MsgCode, Visitor MsgVisitor] struct {
	buffers map[Code]MsgBuffer[Code, Visitor]
}

func (m *msgBus[Code, Visitor]) Send(ctx context.Context, msg Message[Code, Visitor]) error {
	buf, err := m.buffer(msg.Code())
	if err != nil {
		return err
	}
	return buf.Send(ctx, msg)
}

func (m *msgBus[Code, Visitor]) Recv(ctx context.Context, code Code) (Message[Code, Visitor], bool, error) {
	buf, err := m.buffer(code)
	if err != nil {
		return nil, false, err
	}
	return buf.Recv(ctx)
}

func (m *msgBus[Code, Visitor]) buffer(code Code) (MsgBuffer[Code, Visitor], error) {
	buf, ok := m.buffers[code]
	if !ok {
		return nil, fmt.Errorf("failed to get buffer: not exists, code=%v", code)
	}
	return buf, nil
}

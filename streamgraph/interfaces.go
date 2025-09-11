package streamgraph

import "context"

type (
	MsgCode interface {
		comparable
	}

	MsgVisitor interface {
		any
	}

	Message[Code MsgCode, Visitor MsgVisitor] interface {
		Code() Code
		Accept(ctx context.Context, v Visitor) (Message[Code, Visitor], error)
	}

	MsgBuffer[Code MsgCode, Visitor MsgVisitor] interface {
		Send(ctx context.Context, msg Message[Code, Visitor]) error
		Recv(ctx context.Context) (Message[Code, Visitor], bool, error)
	}

	MsgBus[Code MsgCode, Visitor MsgVisitor] interface {
		Send(ctx context.Context, msg Message[Code, Visitor]) error
		Recv(ctx context.Context, code Code) (Message[Code, Visitor], bool, error)
	}

	Source[Code MsgCode, Visitor MsgVisitor] interface {
		Next(ctx context.Context) (Message[Code, Visitor], bool, error)
	}
)

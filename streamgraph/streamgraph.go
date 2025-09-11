package streamgraph

import (
	"context"
)

type (
	Event[Code comparable, Visitor any] interface {
		RoutingCode() Code
		Accept(ctx context.Context, visitor Visitor) ([]Event[Code, Visitor], error)
	}

	Bus[Code comparable, Visitor any] interface {
		Send(ctx context.Context, events ...Event[Code, Visitor]) error
		Recv(ctx context.Context) ([]Event[Code, Visitor], error)
	}

	Source[Code comparable, Visitor any] interface {
		Produce(ctx context.Context) ([]Event[Code, Visitor], error)
	}
)

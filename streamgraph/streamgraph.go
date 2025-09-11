package streamgraph

import (
	"context"
)

type streamGraph[Code MsgCode, Visitor MsgVisitor] struct {
	buffer  MsgBus[Code, Visitor]
	visitor Visitor
	sources []*srcConfig[Code, Visitor]
	sinks   map[Code]*config
	actors  map[Code]*config
}

func (s *streamGraph[Code, Visitor]) runAllSource(ctx context.Context) error {
	// TODO go routines runSourceGroup
	return nil
}

func (s *streamGraph[Code, Visitor]) runSource(ctx context.Context, src *srcConfig[Code, Visitor]) error {
	for {
		select {
		case <-ctx.Done():
		default:
			out, ok, err := src.Next(ctx)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			if err := s.buffer.Send(ctx, out); err != nil {
				return err
			}
		}
	}
}

func (s *streamGraph[Code, Visitor]) runActor(ctx context.Context, code Code) error {
	for {
		select {
		case <-ctx.Done():
		default:
			in, ok, err := s.buffer.Recv(ctx, code)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			out, err := in.Accept(ctx, s.visitor)
			if err != nil {
				return err
			}

			if err := s.buffer.Send(ctx, out); err != nil {
				return err
			}
		}
	}
}

func (s *streamGraph[Code, Visitor]) runSink(ctx context.Context, code Code) error {
	for {
		select {
		case <-ctx.Done():
		default:
			in, ok, err := s.buffer.Recv(ctx, code)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			if _, err = in.Accept(ctx, s.visitor); err != nil {
				return err
			}
		}
	}
}

package impl

import (
	"github.com/yeluyang/gopkg/routine"
	"github.com/yeluyang/gopkg/stream/errs"
)

type sourceRunner[V, M any] struct {
	source   Source[V, M]
	shutdown *shutdown
	c        chan Message[V, M]
}

func (s *sourceRunner[V, M]) C() <-chan Message[V, M] { return s.c }

func newSource[V, M any](src Source[V, M], sd *shutdown) *sourceRunner[V, M] {
	return &sourceRunner[V, M]{
		source:  src,
		shutdown: sd,
		c:       make(chan Message[V, M]),
	}
}

func (s *sourceRunner[V, M]) run() {
	routine.Go(func() {
		s.provide()
	})
}

func (s *sourceRunner[V, M]) send(msg Message[V, M]) bool {
	select {
	case <-s.shutdown.Done():
		return false
	case s.c <- msg:
		return true
	}
}

func (s *sourceRunner[V, M]) provide() {
	defer close(s.c)
	for {
		select {
		case <-s.shutdown.Done():
			return
		default:
			msgs, ok, err := s.source.Next(s.shutdown.ctx)
			if err != nil {
				s.shutdown.Abort(errs.CodeSource.From(err))
				return
			}
			if !ok {
				return
			}
			for _, msg := range msgs {
				if !s.send(msg) {
					return
				}
			}
		}
	}
}

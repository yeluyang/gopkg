package impl

import (
	"sync"

	"github.com/yeluyang/gopkg/routine"
	"github.com/yeluyang/gopkg/stream/errs"
)

type actorPool[V, M any] struct {
	source   *sourceRunner[V, M]
	visitor  V
	shutdown *shutdown
	c        chan M
	routines int
}

func (m *actorPool[V, M]) C() <-chan M { return m.c }

func newActorPool[V, M any](src Source[V, M], visitor V, routines int, sd *shutdown) *actorPool[V, M] {
	routines = max(routines, 1)
	return &actorPool[V, M]{
		source:   newSource(src, sd),
		visitor:  visitor,
		shutdown: sd,
		c:        make(chan M),
		routines: routines,
	}
}

func (m *actorPool[V, M]) run() {
	m.source.run()
	routine.Go(func() {
		defer close(m.c)

		var wg sync.WaitGroup
		for range m.routines {
			wg.Add(1)
			routine.Go(func() {
				defer wg.Done()
				m.provide()
			})
		}
		wg.Wait()
	})
}

func (m *actorPool[V, M]) send(v M) bool {
	select {
	case <-m.shutdown.Done():
		return false
	case m.c <- v:
		return true
	}
}

func (m *actorPool[V, M]) provide() {
	for {
		select {
		case <-m.shutdown.Done():
			return
		case msg, ok := <-m.source.C():
			if !ok {
				return
			}
			result, err := msg.Accept(m.shutdown.ctx, m.visitor)
			if err != nil {
				m.shutdown.Abort(errs.CodeAccept.From(err))
				return
			}
			if !m.send(result) {
				return
			}
		}
	}
}

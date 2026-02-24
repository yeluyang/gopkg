package impl

import (
	"context"
	"sync"

	"github.com/yeluyang/gopkg/flow/errs"
	"github.com/yeluyang/gopkg/flow/interfaces"
	"github.com/yeluyang/gopkg/routine"
)

type sinkGroup[E comparable, V any] struct {
	visitor  V
	channels map[E]chan Msg[E, V]
	wg       sync.WaitGroup
	onError  func(error)
}

func newSinkGroup[E comparable, V any](cfg Config[E, V], onError func(error)) *sinkGroup[E, V] {
	sg := &sinkGroup[E, V]{
		visitor:  cfg.Visitor,
		channels: make(map[E]chan Msg[E, V]),
		onError:  onError,
	}
	for _, sink := range cfg.EagerSinks {
		sg.channels[sink.ID()] = make(chan Msg[E, V])
	}
	for _, sink := range cfg.LazySinks {
		sg.channels[sink.ID()] = make(chan Msg[E, V])
	}
	for _, dup := range cfg.EagerDuplexes {
		sg.channels[dup.ID()] = make(chan Msg[E, V])
	}
	for _, dup := range cfg.LazyDuplexes {
		sg.channels[dup.ID()] = make(chan Msg[E, V])
	}
	return sg
}

func (sg *sinkGroup[E, V]) Start(ctx context.Context, sink interfaces.Sink[E, V]) {
	ch := sg.channels[sink.ID()]
	sg.wg.Add(1)
	routine.Go(func() {
		defer sg.wg.Done()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if err := sink.Drain(ctx, []Msg[E, V]{msg}); err != nil {
					sg.onError(errs.CodeDrain.From(err))
					return
				}
			case <-ctx.Done():
				return
			}
		}
	})
}

func (sg *sinkGroup[E, V]) Has(target E) bool {
	_, ok := sg.channels[target]
	return ok
}

func (sg *sinkGroup[E, V]) Send(ctx context.Context, msg Msg[E, V], target E) {
	select {
	case sg.channels[target] <- msg:
	case <-ctx.Done():
	}
}

func (sg *sinkGroup[E, V]) Close() {
	for _, ch := range sg.channels {
		close(ch)
	}
}

func (sg *sinkGroup[E, V]) Wait() {
	sg.wg.Wait()
}

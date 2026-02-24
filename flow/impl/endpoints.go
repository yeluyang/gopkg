package impl

import (
	"context"
	"sync"

	"github.com/yeluyang/gopkg/flow/errs"
	"github.com/yeluyang/gopkg/flow/interfaces"
)

type endpointState[E comparable] struct {
	interfaces.Endpoint[E]
	activated bool
}

func (es *endpointState[E]) activate(ctx context.Context) error {
	if es.activated {
		return nil
	}
	if err := es.Endpoint.Activate(ctx); err != nil {
		return err
	}
	es.activated = true
	return nil
}

func (es *endpointState[E]) close(ctx context.Context) error {
	if !es.activated {
		return nil
	}
	es.activated = false
	return es.Endpoint.Close(ctx)
}

type endpointGroup[E comparable, V any] struct {
	sources *sourceGroup[E, V]
	sinks   *sinkGroup[E, V]

	endpoints   map[E]*endpointState[E]
	lazySources map[E]interfaces.Source[E, V]
	lazySinks   map[E]interfaces.Sink[E, V]
	lazyMu      sync.Mutex
}

func newEndpointGroup[E comparable, V any](
	sources *sourceGroup[E, V],
	sinks *sinkGroup[E, V],
) *endpointGroup[E, V] {
	return &endpointGroup[E, V]{
		sources:     sources,
		sinks:       sinks,
		endpoints:   make(map[E]*endpointState[E]),
		lazySources: make(map[E]interfaces.Source[E, V]),
		lazySinks:   make(map[E]interfaces.Sink[E, V]),
	}
}

func (eg *endpointGroup[E, V]) Init(ctx context.Context, cfg Config[E, V]) error {
	for _, src := range cfg.EagerSources {
		if err := eg.activate(ctx, src); err != nil {
			return err
		}
		eg.sources.Start(ctx, src)
	}
	for _, sink := range cfg.EagerSinks {
		if err := eg.activate(ctx, sink); err != nil {
			return err
		}
		eg.sinks.Start(ctx, sink)
	}
	for _, dup := range cfg.EagerDuplexes {
		if err := eg.activate(ctx, dup); err != nil {
			return err
		}
		eg.sources.Start(ctx, dup)
		eg.sinks.Start(ctx, dup)
	}
	for _, src := range cfg.LazySources {
		eg.register(src)
		eg.lazySources[src.ID()] = src
	}
	for _, sink := range cfg.LazySinks {
		eg.register(sink)
		eg.lazySinks[sink.ID()] = sink
	}
	for _, dup := range cfg.LazyDuplexes {
		id := dup.ID()
		eg.register(dup)
		eg.lazySources[id] = dup
		eg.lazySinks[id] = dup
	}
	return nil
}

func (eg *endpointGroup[E, V]) register(ep interfaces.Endpoint[E]) {
	eg.endpoints[ep.ID()] = &endpointState[E]{Endpoint: ep}
}

func (eg *endpointGroup[E, V]) activate(ctx context.Context, ep interfaces.Endpoint[E]) error {
	eg.register(ep)
	return eg.endpoints[ep.ID()].activate(ctx)
}

// Activate finds lazy endpoints by ID, activates, and starts them.
func (eg *endpointGroup[E, V]) Activate(ctx context.Context, id E) error {
	src, sink, found := eg.takeLazy(id)
	if !found {
		return nil
	}
	if err := eg.endpoints[id].activate(ctx); err != nil {
		return err
	}
	if src != nil {
		eg.sources.Start(ctx, src)
	}
	if sink != nil {
		eg.sinks.Start(ctx, sink)
	}
	return nil
}

// DrainTo delivers a message to a target sink, lazily activating if needed.
func (eg *endpointGroup[E, V]) DrainTo(ctx context.Context, msg Msg[E, V], sinkID E) error {
	if !eg.sinks.Has(sinkID) {
		return errs.CodeDrain.Format("no such sink: id=%v", sinkID)
	}
	if err := eg.Activate(ctx, sinkID); err != nil {
		return err
	}
	eg.sinks.Send(ctx, msg, sinkID)
	return nil
}

// Exhausted reports whether all sources have finished.
func (eg *endpointGroup[E, V]) Exhausted() bool {
	return eg.sources.Exhausted()
}

func (eg *endpointGroup[E, V]) takeLazy(id E) (src interfaces.Source[E, V], sink interfaces.Sink[E, V], found bool) {
	eg.lazyMu.Lock()
	defer eg.lazyMu.Unlock()

	src, srcOK := eg.lazySources[id]
	if srcOK {
		delete(eg.lazySources, id)
		found = true
	}
	sink, sinkOK := eg.lazySinks[id]
	if sinkOK {
		delete(eg.lazySinks, id)
		found = true
	}
	return
}

// CloseSinks closes all sink channels, causing sink goroutines to exit.
func (eg *endpointGroup[E, V]) CloseSinks() {
	eg.sinks.Close()
}

// WaitSinks waits for all sink goroutines to finish.
func (eg *endpointGroup[E, V]) WaitSinks() {
	eg.sinks.Wait()
}

// Close deactivates all endpoints.
func (eg *endpointGroup[E, V]) Close(ctx context.Context) {
	for _, es := range eg.endpoints {
		_ = es.close(ctx)
	}
}

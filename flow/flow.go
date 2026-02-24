package flow

import (
	"github.com/yeluyang/gopkg/flow/impl"
	"github.com/yeluyang/gopkg/flow/interfaces"
)

type builder[E comparable, V any] struct {
	visitor       V
	concurrency   int
	eagerSources  []interfaces.Source[E, V]
	lazySources   []interfaces.Source[E, V]
	eagerSinks    []interfaces.Sink[E, V]
	lazySinks     []interfaces.Sink[E, V]
	eagerDuplexes []interfaces.Duplex[E, V]
	lazyDuplexes  []interfaces.Duplex[E, V]
}

func New[E comparable, V any](visitor V) *builder[E, V] {
	return &builder[E, V]{
		visitor:     visitor,
		concurrency: 1,
	}
}

func (b *builder[E, V]) ActivateSource(s interfaces.Source[E, V]) *builder[E, V] {
	b.eagerSources = append(b.eagerSources, s)
	return b
}

func (b *builder[E, V]) ActivateSink(s interfaces.Sink[E, V]) *builder[E, V] {
	b.eagerSinks = append(b.eagerSinks, s)
	return b
}

func (b *builder[E, V]) ActivateDuplex(s interfaces.Duplex[E, V]) *builder[E, V] {
	b.eagerDuplexes = append(b.eagerDuplexes, s)
	return b
}

func (b *builder[E, V]) Source(s interfaces.Source[E, V]) *builder[E, V] {
	b.lazySources = append(b.lazySources, s)
	return b
}

func (b *builder[E, V]) Sink(s interfaces.Sink[E, V]) *builder[E, V] {
	b.lazySinks = append(b.lazySinks, s)
	return b
}

func (b *builder[E, V]) Duplex(s interfaces.Duplex[E, V]) *builder[E, V] {
	b.lazyDuplexes = append(b.lazyDuplexes, s)
	return b
}

func (b *builder[E, V]) Concurrency(n int) *builder[E, V] {
	b.concurrency = n
	return b
}

func (b *builder[E, V]) Build() *impl.Flow[E, V] {
	return impl.New[E, V](impl.Config[E, V]{
		Visitor:       b.visitor,
		Concurrency:   b.concurrency,
		EagerSources:  b.eagerSources,
		LazySources:   b.lazySources,
		EagerSinks:    b.eagerSinks,
		LazySinks:     b.lazySinks,
		EagerDuplexes: b.eagerDuplexes,
		LazyDuplexes:  b.lazyDuplexes,
	})
}

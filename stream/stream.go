package stream

import (
	"fmt"

	"github.com/yeluyang/gopkg/errorx"
	"github.com/yeluyang/gopkg/stream/errs"
	"github.com/yeluyang/gopkg/stream/impl"
	"github.com/yeluyang/gopkg/stream/interfaces"
)

// Re-export interfaces so users only need to import "stream".
type (
	Source[V, M any]  = interfaces.Source[V, M]
	Message[V, M any] = interfaces.Message[V, M]
	Sink[M any]       = interfaces.Sink[M]
	// Stream is the public handle for a stream pipeline.
	Stream[V, M any] = impl.Stream[V, M]
)

// Error codes for stream pipeline failures.
const (
	CodeSource errorx.Code = errs.CodeSource
	CodeAccept errorx.Code = errs.CodeAccept
	CodeDrain  errorx.Code = errs.CodeDrain
)

// Builder guides the construction of a stream pipeline.
//
// Usage:
//
//	s := stream.From(source).Visitor(visitor).To(sink).Routines(25).Build()
//	err := s.Run(ctx)
type Builder[V, M any] struct {
	source   Source[V, M]
	visitor  V
	sink     Sink[M]
	routines int
}

// From creates a Builder starting from the given source.
func From[V, M any](source Source[V, M]) *Builder[V, M] {
	return &Builder[V, M]{
		source:   source,
		routines: 1,
	}
}

// Visitor sets the visitor used to accept messages.
func (b *Builder[V, M]) Visitor(visitor V) *Builder[V, M] {
	b.visitor = visitor
	return b
}

// To sets the sink that receives accepted results.
func (b *Builder[V, M]) To(sink Sink[M]) *Builder[V, M] {
	b.sink = sink
	return b
}

// Routines sets the number of concurrent message processors. Defaults to 1.
func (b *Builder[V, M]) Routines(n int) *Builder[V, M] {
	b.routines = n
	return b
}

// Build creates the stream pipeline. Returns error if required fields are missing.
func (b *Builder[V, M]) Build() (*Stream[V, M], error) {
	if b.sink == nil {
		return nil, fmt.Errorf("stream: sink is required, call .To() before .Build()")
	}
	return impl.New(b.source, b.visitor, b.routines, b.sink), nil
}

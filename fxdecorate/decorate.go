package fxdecorate

import "go.uber.org/fx"

type (
	Decorator[I any] func(I) I

	Result[I any] struct {
		fx.Out
		Decorator Decorator[I] `group:"decorators"`
	}
)

func New[I any](
	p struct {
		fx.In
		Impl       I
		Decorators []Decorator[I] `group:"decorators"`
	},
) I {
	i := p.Impl
	for _, d := range p.Decorators {
		i = d(i)
	}
	return i
}

func With[I any](d Decorator[I]) Result[I] {
	return Result[I]{Decorator: d}
}

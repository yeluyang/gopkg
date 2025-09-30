package ctxkv

import (
	"context"

	"github.com/samber/mo"
)

type (
	With[V any] func(ctx context.Context, v V) context.Context
	From[V any] func(ctx context.Context) mo.Option[V]
)

func New[K, V any]() (With[V], From[V]) {
	k := *new(K)
	return func(ctx context.Context, v V) context.Context {
			return context.WithValue(ctx, k, v)
		},
		func(ctx context.Context) mo.Option[V] {
			val, ok := ctx.Value(k).(V)
			if !ok {
				return mo.None[V]()
			}
			return mo.Some(val)
		}
}

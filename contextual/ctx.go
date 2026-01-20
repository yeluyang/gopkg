package contextual

import (
	"context"

	"github.com/samber/mo"
)

type (
	With[V any] func(ctx context.Context, v V) context.Context

	From[V any] func(ctx context.Context) mo.Option[V]

	key string
)

func New[V any](name string) (With[V], From[V]) {
	k := newKey(name)
	return func(ctx context.Context, v V) context.Context {
			return context.WithValue(ctx, k, v)
		},
		func(ctx context.Context) mo.Option[V] {
			if v := ctx.Value(k); v != nil {
				if v, ok := v.(V); ok {
					return mo.Some(v)
				}
			}
			return mo.None[V]()
		}
}

func newKey(name string) *key {
	k := new(string)
	*k = name
	return (*key)(k)
}

func (k *key) String() string {
	return string(*k)
}

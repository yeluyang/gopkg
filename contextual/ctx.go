package contextual

import "context"

type (
	With[V any] func(ctx context.Context, v V) context.Context

	From[V any] func(ctx context.Context) (V, bool)

	key string
)

func New[V any](name string) (With[V], From[V]) {
	k := newKey(name)
	return func(ctx context.Context, v V) context.Context {
			return context.WithValue(ctx, k, v)
		},
		func(ctx context.Context) (V, bool) {
			val, ok := ctx.Value(k).(V)
			return val, ok
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

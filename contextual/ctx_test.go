package contextual_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeluyang/gopkg/contextual"
)

// define unexported key
type (
	reqIDKey struct{}
	envKey   struct{}
)

// define context setters and getters
var (
	withReqID, ReqIDFrom = contextual.New[reqIDKey, string]()
	WithEnv, EnvFrom     = contextual.New[envKey, string]()
)

// feel free to custom context setters and getters
func WithReqID(ctx context.Context) context.Context {
	return withReqID(ctx, "request-id-generated-directly-without-passed-argument")
}

func TestContextual(t *testing.T) {
	ctx := context.Background()

	reqID := ReqIDFrom(ctx)
	assert.Truef(t, reqID.IsNone(), "reqID=%+v", reqID)

	ctx = WithReqID(ctx)
	reqID = ReqIDFrom(ctx)
	assert.Truef(t, reqID.IsSome(), "reqID=%+v", reqID)
	assert.Equal(t, "request-id-generated-directly-without-passed-argument", reqID.MustGet())

	env := EnvFrom(ctx)
	assert.Truef(t, env.IsNone(), "env=%+v", env)

	ctx = WithEnv(ctx, "yly")
	env = EnvFrom(ctx)
	assert.Truef(t, env.IsSome(), "env=%+v", env)
	assert.Equal(t, "yly", env.MustGet())

	ctx = WithEnv(ctx, "yly2")
	assert.Equal(t, "yly2", EnvFrom(ctx).MustGet())

	reqID = ReqIDFrom(ctx)
	assert.Truef(t, reqID.IsSome(), "reqID=%+v", reqID)
	assert.Equal(t, "request-id-generated-directly-without-passed-argument", reqID.MustGet())
}

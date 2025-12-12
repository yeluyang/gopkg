package contextual_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/yeluyang/gopkg/contextual"
)

// define context setters and getters
var (
	WithEnv, EnvFrom     = contextual.New[string]("env")
	WithEnv2, Env2From   = contextual.New[string]("env") // duplicated name
	withReqID, reqIDFrom = contextual.New[string]("request_id")
)

// feel free to custom context setters and getters
func WithReqID(ctx context.Context) context.Context {
	return withReqID(ctx, "request-id-generated-directly-without-passed-argument")
}

func ReqIDFrom(ctx context.Context) string {
	v, ok := reqIDFrom(ctx)
	if !ok {
		return "-"
	}
	return v
}

func TestContextual(t *testing.T) {
	suite.Run(t, new(testSuiteContextual))
}

type testSuiteContextual struct {
	suite.Suite
}

func (s *testSuiteContextual) TestNormalCase() {
	ctx := context.Background()
	env, ok := EnvFrom(ctx)
	s.Falsef(ok, "env=%+v", env)

	ctx = WithEnv(ctx, "yly")
	env, ok = EnvFrom(ctx)
	s.Truef(ok, "env=%+v", env)
	s.Equal("yly", env)

	ctx = WithEnv(ctx, "yly2")
	env, ok = EnvFrom(ctx)
	s.Truef(ok, "env=%+v", env)
	s.Equal("yly2", env)
}

func (s *testSuiteContextual) TestCustom() {
	ctx := context.Background()
	reqID := ReqIDFrom(ctx)
	s.Equal("-", reqID)

	ctx = WithReqID(ctx)
	reqID = ReqIDFrom(ctx)
	s.Equal("request-id-generated-directly-without-passed-argument", reqID)
}

func (s *testSuiteContextual) TestConflictName() {
	ctx := context.Background()
	ctx = WithEnv(ctx, "yly")
	env, ok := EnvFrom(ctx)
	s.Truef(ok, "env=%+v", env)
	s.Equal("yly", env)

	ctx = WithEnv2(ctx, "env2-value")
	env2, ok := Env2From(ctx)
	s.True(ok)
	s.Equal("env2-value", env2)

	env, ok = EnvFrom(ctx)
	s.True(ok)
	s.Equal("yly", env)
}

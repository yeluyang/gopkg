// Contextual Logging with `log/slog`
package ctxzap

import (
	"context"

	"go.uber.org/zap"

	"github.com/yeluyang/gopkg/ctxkv"
)

type ctxKey struct{}

var With, from = ctxkv.New[ctxKey, *zap.Logger]()

func From(ctx context.Context) *zap.Logger {
	return from(ctx).OrElse(new())
}

func New(ctx context.Context) (*zap.Logger, context.Context) {
	logger := new()
	return logger, With(ctx, logger)
}

func new() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

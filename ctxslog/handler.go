// Contextual Logging with `log/slog`
package ctxslog

import (
	"context"
	"log/slog"

	"github.com/yeluyang/gopkg/ctxkv"
)

type logKey struct{}

var With, from = ctxkv.New[logKey, *slog.Logger]()

func From(ctx context.Context) *slog.Logger {
	return from(ctx).OrElse(slog.Default())
}

func New(ctx context.Context, h slog.Handler) (*slog.Logger, context.Context) {
	logger := slog.New(h)
	return logger, With(ctx, logger)
}

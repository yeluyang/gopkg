package ctxslog_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yeluyang/gopkg/ctxslog"
)

func TestCtxSlog(t *testing.T) {
	ctx := context.Background()
	log, ctx := ctxslog.New(ctx, slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true}))

	log.Info("start")

	assert.NotPanics(t, func() { top(ctx) })
}

func top(ctx context.Context) {
	log := ctxslog.From(ctx).With("top", "yes")
	ctx = ctxslog.With(ctx, log)

	log.Info("start", slog.String("level", "top"))
	mid(ctx)
	log.Info("exit", slog.String("level", "top"))
}

func mid(ctx context.Context) {
	log := ctxslog.From(ctx).With("mid", "yes")
	ctx = ctxslog.With(ctx, log)

	log.Info("start", slog.String("level", "mid"))
	bottom(ctx)
	log.Info("exit", slog.String("level", "mid"))
}

func bottom(ctx context.Context) {
	log := ctxslog.From(ctx).With("bottom", "yes")
	ctx = ctxslog.With(ctx, log)

	log.Info("start", slog.String("level", "bottom"))
	log.Info("exit", slog.String("level", "bottom"))
}

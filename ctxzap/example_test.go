package ctxzap_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/yeluyang/gopkg/ctxzap"
)

func TestCtxZap(t *testing.T) {
	ctx := context.Background()
	log, ctx := ctxzap.New(ctx)

	log.Info("start")
	assert.NotPanics(t, func() { top(ctx) })
	log.Info("exit")
}

func top(ctx context.Context) {
	level := zap.NewAtomicLevel()
	level.SetLevel(zapcore.DebugLevel)
	log := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(os.Stderr),
			level,
		),
	)
	ctx = ctxzap.With(ctx, log)

	log.Info("start", zap.String("level", "top"))
	mid(ctx)
	log.Info("exit", zap.String("level", "top"))
}

func mid(ctx context.Context) {
	log := ctxzap.From(ctx).With(zap.String("mid", "yes"))
	ctx = ctxzap.With(ctx, log)

	log.Info("start", zap.String("level", "mid"))
	bottom(ctx)
	log.Info("exit", zap.String("level", "mid"))
}

func bottom(ctx context.Context) {
	log := ctxzap.From(ctx).With(zap.String("bottom", "yes"))

	log.Info("start", zap.String("level", "bottom"))
	log.Info("exit", zap.String("level", "bottom"))
}

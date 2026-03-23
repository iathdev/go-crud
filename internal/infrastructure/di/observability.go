package di

import (
	"context"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/infrastructure/logging"
	infrasentry "learning-go/internal/infrastructure/sentry"
	"learning-go/internal/infrastructure/tracing"
	"learning-go/internal/shared/ctxlog"
	"learning-go/internal/shared/logger"

	"go.uber.org/zap"
)

type observability struct {
	tracerShutdown func(context.Context) error
	sentryCleanup  func()
}

func initObservability(cfg *config.Config) (*observability, error) {
	log, err := logging.NewLogger(cfg)
	if err != nil {
		return nil, err
	}
	logger.Init(log)

	ctx := context.Background()
	tracerShutdown, err := tracing.InitTracer(ctx, cfg)
	if err != nil {
		logger.Warn("[SERVER] failed to init tracer, continuing without tracing", zap.Error(err))
		tracerShutdown = func(context.Context) error { return nil }
	}

	logger.RegisterContextExtractor(tracing.OTELContextExtractor())
	logger.RegisterContextExtractor(ctxlog.Fields)

	sentryCleanup, err := infrasentry.Init(cfg)
	if err != nil {
		logger.Warn("[SERVER] failed to init sentry, continuing without error tracking", zap.Error(err))
		sentryCleanup = func() {}
	}

	return &observability{
		tracerShutdown: tracerShutdown,
		sentryCleanup:  sentryCleanup,
	}, nil
}

func (obs *observability) cleanup() {
	obs.sentryCleanup()
	obs.tracerShutdown(context.Background())
	logger.Sync()
}

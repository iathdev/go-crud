package tracing

import (
	"context"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/shared/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// InitTracer sets up the OpenTelemetry tracer provider with OTLP gRPC exporter.
// Returns a shutdown function. If OTLP_ENDPOINT is empty, returns a no-op shutdown.
func InitTracer(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	if cfg.OTLPEndpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	}
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.GetServiceName()),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// OTELContextExtractor returns a logger.ContextExtractor that extracts
// trace_id and span_id from the OTEL span in context.
func OTELContextExtractor() logger.ContextExtractor {
	return func(ctx context.Context) []zap.Field {
		span := trace.SpanFromContext(ctx)
		sc := span.SpanContext()
		if !sc.IsValid() {
			return nil
		}
		return []zap.Field{
			zap.String("trace_id", sc.TraceID().String()),
			zap.String("span_id", sc.SpanID().String()),
		}
	}
}

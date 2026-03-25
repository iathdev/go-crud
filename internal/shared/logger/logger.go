package logger

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

// ContextExtractor extracts additional fields from context (e.g., trace_id, span_id).
type ContextExtractor func(ctx context.Context) []zap.Field

// --- global instance ---

var globalPtr atomic.Pointer[zap.Logger]

var extractorMu sync.RWMutex
var contextExtractors []ContextExtractor

func init() {
	globalPtr.Store(zap.NewNop())
}

// Init sets the global logger. Call once at startup.
func Init(l *zap.Logger) { globalPtr.Store(l) }

func getLogger() *zap.Logger { return globalPtr.Load() }

// RegisterContextExtractor adds a function that extracts fields from context.
// Thread-safe. Call at startup to register extractors (e.g., OTEL trace IDs).
func RegisterContextExtractor(fn ContextExtractor) {
	extractorMu.Lock()
	contextExtractors = append(contextExtractors, fn)
	extractorMu.Unlock()
}

// WithContext returns a *zap.Logger enriched with fields extracted from context.
func WithContext(ctx context.Context) *zap.Logger {
	extractorMu.RLock()
	extractors := contextExtractors
	extractorMu.RUnlock()

	log := getLogger()
	var fields []zap.Field
	for _, extractor := range extractors {
		fields = append(fields, extractor(ctx)...)
	}
	if len(fields) == 0 {
		return log
	}
	return log.With(fields...)
}

func Debug(first any, rest ...any)         { dispatch((*zap.Logger).Debug, first, rest...) }
func Info(first any, rest ...any)          { dispatch((*zap.Logger).Info, first, rest...) }
func Warn(first any, rest ...any)          { dispatch((*zap.Logger).Warn, first, rest...) }
func Error(first any, rest ...any)         { dispatch((*zap.Logger).Error, first, rest...) }
func Fatal(first any, rest ...any)         { dispatch((*zap.Logger).Fatal, first, rest...) }
func With(fields ...zap.Field) *zap.Logger { return getLogger().With(fields...) }
func Sync() error                          { return getLogger().Sync() }

func dispatch(fn func(*zap.Logger, string, ...zap.Field), first any, rest ...any) {
	var log *zap.Logger
	var msg string
	var fieldArgs []any

	switch v := first.(type) {
	case context.Context:
		log = WithContext(v)
		if len(rest) > 0 {
			msg, _ = rest[0].(string)
			fieldArgs = rest[1:]
		}
	case string:
		log = getLogger()
		msg = v
		fieldArgs = rest
	default:
		getLogger().Warn("logger: invalid first argument", zap.String("type", fmt.Sprintf("%T", first)))
		return
	}

	fields := make([]zap.Field, 0, len(fieldArgs))
	for _, a := range fieldArgs {
		if f, ok := a.(zap.Field); ok {
			fields = append(fields, f)
		} else {
			fields = append(fields, zap.Any("_dropped", a))
		}
	}
	fn(log, msg, fields...)
}

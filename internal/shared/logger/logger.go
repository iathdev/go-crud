package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// ContextExtractor extracts additional fields from context (e.g., trace_id, span_id).
type ContextExtractor func(ctx context.Context) []zap.Field

// --- global instance ---

var global = zap.NewNop()

var extractorMu sync.RWMutex
var contextExtractors []ContextExtractor

// Init sets the global logger. Call once at startup.
func Init(l *zap.Logger) { global = l }

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

	var fields []zap.Field
	for _, extractor := range extractors {
		fields = append(fields, extractor(ctx)...)
	}
	if len(fields) == 0 {
		return global
	}
	return global.With(fields...)
}

// Package-level functions delegate to the global logger.

func Debug(msg string, fields ...zap.Field) { global.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { global.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { global.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { global.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { global.Fatal(msg, fields...) }
func With(fields ...zap.Field) *zap.Logger  { return global.With(fields...) }
func Sync() error                           { return global.Sync() }

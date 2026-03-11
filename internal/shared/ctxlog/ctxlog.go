package ctxlog

import (
	"context"
	"strings"

	"go.uber.org/zap"
)

type ctxKey struct{}

var sensitiveKeys = map[string]bool{
	"password":      true,
	"token":         true,
	"secret":        true,
	"authorization": true,
	"cookie":        true,
	"credit_card":   true,
	"ssn":           true,
	"api_key":       true,
}

func isSensitive(key string) bool {
	return sensitiveKeys[strings.ToLower(key)]
}

func sanitize(fields []zap.Field) []zap.Field {
	for i, f := range fields {
		if isSensitive(f.Key) {
			fields[i] = zap.String(f.Key, "*")
		}
	}
	return fields
}

func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	existing := Fields(ctx)
	return context.WithValue(ctx, ctxKey{}, append(existing, sanitize(fields)...))
}

func Fields(ctx context.Context) []zap.Field {
	if fields, ok := ctx.Value(ctxKey{}).([]zap.Field); ok {
		copied := make([]zap.Field, len(fields))
		copy(copied, fields)
		return copied
	}
	return nil
}

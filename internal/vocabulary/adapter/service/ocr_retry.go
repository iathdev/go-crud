package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"learning-go/internal/shared/logger"
	"learning-go/internal/vocabulary/application/port"
)

// OCRRetryDecorator wraps an OCRServicePort with retry + exponential backoff.
type OCRRetryDecorator struct {
	inner      port.OCRServicePort
	maxRetries int
	baseDelay  time.Duration
}

func NewOCRRetryDecorator(inner port.OCRServicePort, maxRetries int, baseDelay time.Duration) port.OCRServicePort {
	return &OCRRetryDecorator{
		inner:      inner,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}
}

func (d *OCRRetryDecorator) Recognize(ctx context.Context, req port.OCRRequest) (*port.OCRResult, error) {
	var lastErr error
	for attempt := range d.maxRetries {
		result, err := d.inner.Recognize(ctx, req)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt == d.maxRetries-1 {
			break
		}

		backoff := d.baseDelay * time.Duration(1<<attempt)
		logger.Warn(ctx, "[OCR] retrying",
			zap.Int("attempt", attempt+1),
			zap.Int("max", d.maxRetries),
			zap.Duration("backoff", backoff),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, lastErr
}

package cache

import (
	"context"
	"learning-go/internal/shared/logger"

	"go.uber.org/zap"
)

// MultiLevelCache composes L1 (fast, e.g. memory) and L2 (durable, e.g. Redis).
// Reads cascade L1 -> L2 with automatic L1 backfill on L2 hit.
// Writes populate both levels. Errors in L1 are non-fatal.
type MultiLevelCache[T any] struct {
	l1 Cache[T]
	l2 Cache[T]
}

func NewMultiLevelCache[T any](l1, l2 Cache[T]) *MultiLevelCache[T] {
	return &MultiLevelCache[T]{l1: l1, l2: l2}
}

func (multiCache *MultiLevelCache[T]) Get(ctx context.Context, key string) (*T, error) {
	value, err := multiCache.l1.Get(ctx, key)
	if err != nil {
		logger.Debug(ctx, "[CACHE] L1 read failed", zap.String("key", key), zap.Error(err))
	}
	if value != nil {
		return value, nil
	}

	value, err = multiCache.l2.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}

	// Backfill L1
	if backfillErr := multiCache.l1.Set(ctx, key, value); backfillErr != nil {
		logger.Debug(ctx, "[CACHE] L1 backfill failed", zap.String("key", key), zap.Error(backfillErr))
	}

	return value, nil
}

func (multiCache *MultiLevelCache[T]) Set(ctx context.Context, key string, value *T) error {
	if err := multiCache.l2.Set(ctx, key, value); err != nil {
		return err
	}

	if err := multiCache.l1.Set(ctx, key, value); err != nil {
		logger.Debug(ctx, "[CACHE] L1 write failed", zap.String("key", key), zap.Error(err))
	}

	return nil
}

func (multiCache *MultiLevelCache[T]) Delete(ctx context.Context, key string) error {
	// Delete L2 first (authoritative). If L2 fails, L1 stays consistent
	// and a subsequent read won't backfill a deleted entry from L2.
	l2Err := multiCache.l2.Delete(ctx, key)
	l1Err := multiCache.l1.Delete(ctx, key)

	if l2Err != nil {
		return l2Err
	}
	if l1Err != nil {
		logger.Debug(ctx, "[CACHE] L1 delete failed", zap.String("key", key), zap.Error(l1Err))
	}

	return nil
}

func (multiCache *MultiLevelCache[T]) Close() error {
	l1Err := multiCache.l1.Close()
	l2Err := multiCache.l2.Close()

	if l1Err != nil {
		return l1Err
	}
	return l2Err
}

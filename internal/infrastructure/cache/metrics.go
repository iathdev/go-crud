package cache

import (
	"context"
	"sync/atomic"
)

// Stats holds cache hit/miss/error counters.
type Stats struct {
	Hits   uint64
	Misses uint64
	Errors uint64
}

// MetricsCache wraps a Cache and tracks Get hit/miss/error counts.
// Use Stats() to read counters — wire them to Prometheus, OTel, or logs as needed.
type MetricsCache[T any] struct {
	inner  Cache[T]
	hits   atomic.Uint64
	misses atomic.Uint64
	errors atomic.Uint64
}

func WithMetrics[T any](inner Cache[T]) *MetricsCache[T] {
	return &MetricsCache[T]{inner: inner}
}

func (mc *MetricsCache[T]) Get(ctx context.Context, key string) (*T, error) {
	value, err := mc.inner.Get(ctx, key)
	if err != nil {
		mc.errors.Add(1)
		return nil, err
	}
	if value == nil {
		mc.misses.Add(1)
	} else {
		mc.hits.Add(1)
	}
	return value, nil
}

func (mc *MetricsCache[T]) Set(ctx context.Context, key string, value *T) error {
	return mc.inner.Set(ctx, key, value)
}

func (mc *MetricsCache[T]) Delete(ctx context.Context, key string) error {
	return mc.inner.Delete(ctx, key)
}

func (mc *MetricsCache[T]) Close() error {
	return mc.inner.Close()
}

func (mc *MetricsCache[T]) Stats() Stats {
	return Stats{
		Hits:   mc.hits.Load(),
		Misses: mc.misses.Load(),
		Errors: mc.errors.Load(),
	}
}

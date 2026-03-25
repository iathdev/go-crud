package cache

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// Loader wraps a Cache with singleflight-protected loading to prevent cache stampede.
// Concurrent GetOrLoad calls for the same key are deduplicated — only one load runs.
type Loader[T any] struct {
	cache Cache[T]
	group singleflight.Group
}

func NewLoader[T any](cache Cache[T]) *Loader[T] {
	return &Loader[T]{cache: cache}
}

// GetOrLoad returns the cached value for key. On miss, calls load exactly once
// (even under concurrent access) and populates the cache with the result.
func (loader *Loader[T]) GetOrLoad(ctx context.Context, key string, load func(ctx context.Context) (*T, error)) (*T, error) {
	value, err := loader.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if value != nil {
		return value, nil
	}

	// Detach from caller's context so cancellation of the first caller
	// doesn't fail all waiting goroutines sharing this singleflight.
	loadCtx := context.WithoutCancel(ctx)

	result, err, _ := loader.group.Do(key, func() (any, error) {
		// Double-check: another goroutine may have populated the cache
		if cached, cacheErr := loader.cache.Get(loadCtx, key); cacheErr == nil && cached != nil {
			return cached, nil
		}

		loaded, loadErr := load(loadCtx)
		if loadErr != nil {
			return nil, loadErr
		}
		if loaded != nil {
			_ = loader.cache.Set(loadCtx, key, loaded)
		}
		return loaded, nil
	})
	if err != nil {
		return nil, err
	}

	value, _ = result.(*T)
	return value, nil
}

// Delegated Cache[T] methods — Loader can be used wherever a Cache is needed.

func (loader *Loader[T]) Get(ctx context.Context, key string) (*T, error) {
	return loader.cache.Get(ctx, key)
}

func (loader *Loader[T]) Set(ctx context.Context, key string, value *T) error {
	return loader.cache.Set(ctx, key, value)
}

func (loader *Loader[T]) Delete(ctx context.Context, key string) error {
	return loader.cache.Delete(ctx, key)
}

func (loader *Loader[T]) Close() error {
	return loader.cache.Close()
}

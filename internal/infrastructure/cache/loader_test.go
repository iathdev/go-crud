package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoader_CacheHit(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	ctx := context.Background()
	mc.Set(ctx, "key1", &testItem{Name: "cached"})

	loader := NewLoader[testItem](mc)
	got, err := loader.GetOrLoad(ctx, "key1", func(_ context.Context) (*testItem, error) {
		t.Fatal("load should not be called on cache hit")
		return nil, nil
	})

	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if got == nil || got.Name != "cached" {
		t.Errorf("expected cached value, got %v", got)
	}
}

func TestLoader_CacheMissLoads(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	loader := NewLoader[testItem](mc)
	ctx := context.Background()

	got, err := loader.GetOrLoad(ctx, "key1", func(_ context.Context) (*testItem, error) {
		return &testItem{Name: "loaded"}, nil
	})

	if err != nil {
		t.Fatalf("GetOrLoad failed: %v", err)
	}
	if got == nil || got.Name != "loaded" {
		t.Errorf("expected loaded value, got %v", got)
	}

	// Verify value was cached
	cached, _ := mc.Get(ctx, "key1")
	if cached == nil || cached.Name != "loaded" {
		t.Error("expected value to be cached after load")
	}
}

func TestLoader_LoadError(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	loader := NewLoader[testItem](mc)
	loadErr := errors.New("load failed")

	_, err := loader.GetOrLoad(context.Background(), "key1", func(_ context.Context) (*testItem, error) {
		return nil, loadErr
	})

	if !errors.Is(err, loadErr) {
		t.Errorf("expected load error, got %v", err)
	}
}

func TestLoader_Singleflight(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	loader := NewLoader[testItem](mc)
	ctx := context.Background()

	var loadCount atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loader.GetOrLoad(ctx, "key1", func(_ context.Context) (*testItem, error) {
				loadCount.Add(1)
				time.Sleep(50 * time.Millisecond) // simulate slow load
				return &testItem{Name: "loaded"}, nil
			})
		}()
	}

	wg.Wait()

	if count := loadCount.Load(); count != 1 {
		t.Errorf("expected exactly 1 load call (singleflight), got %d", count)
	}
}

func TestLoader_DelegatesCacheInterface(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	loader := NewLoader[testItem](mc)
	ctx := context.Background()

	// Set via loader
	loader.Set(ctx, "key1", &testItem{Name: "test"})

	// Get via loader
	got, _ := loader.Get(ctx, "key1")
	if got == nil || got.Name != "test" {
		t.Error("expected delegated Get to work")
	}

	// Delete via loader
	loader.Delete(ctx, "key1")
	got, _ = loader.Get(ctx, "key1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

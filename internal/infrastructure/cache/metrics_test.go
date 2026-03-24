package cache

import (
	"context"
	"testing"
	"time"
)

func TestMetricsCache_TracksHitsAndMisses(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	metricsCache := WithMetrics[testItem](mc)
	ctx := context.Background()

	metricsCache.Set(ctx, "key1", &testItem{Name: "test"})

	// Hit
	metricsCache.Get(ctx, "key1")
	// Miss
	metricsCache.Get(ctx, "nonexistent")
	// Another hit
	metricsCache.Get(ctx, "key1")

	stats := metricsCache.Stats()
	if stats.Hits != 2 {
		t.Errorf("expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", stats.Errors)
	}
}

func TestMetricsCache_DelegatesOperations(t *testing.T) {
	mc := NewMemoryCache[testItem](100, time.Minute)
	defer mc.Close()

	metricsCache := WithMetrics[testItem](mc)
	ctx := context.Background()

	metricsCache.Set(ctx, "key1", &testItem{Name: "test"})

	got, _ := metricsCache.Get(ctx, "key1")
	if got == nil || got.Name != "test" {
		t.Error("expected delegated Get to return cached value")
	}

	metricsCache.Delete(ctx, "key1")
	got, _ = metricsCache.Get(ctx, "key1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

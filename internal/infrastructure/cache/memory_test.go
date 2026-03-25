package cache

import (
	"context"
	"testing"
	"time"
)

type testItem struct {
	Name string
}

func TestMemoryCache_GetSet(t *testing.T) {
	memCache := NewMemoryCache[testItem](100, time.Minute)
	defer memCache.Close()

	ctx := context.Background()
	item := &testItem{Name: "test"}

	if err := memCache.Set(ctx, "key1", item); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := memCache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || got.Name != "test" {
		t.Errorf("expected {Name: test}, got %v", got)
	}
}

func TestMemoryCache_NoCopyAliasing(t *testing.T) {
	memCache := NewMemoryCache[testItem](100, time.Minute)
	defer memCache.Close()

	ctx := context.Background()
	original := &testItem{Name: "original"}
	memCache.Set(ctx, "key1", original)

	// Mutate the original — cache should not be affected
	original.Name = "mutated"

	got, _ := memCache.Get(ctx, "key1")
	if got.Name != "original" {
		t.Errorf("cache should not alias caller pointer, got %q", got.Name)
	}

	// Mutate the returned value — cache should not be affected
	got.Name = "also-mutated"

	got2, _ := memCache.Get(ctx, "key1")
	if got2.Name != "original" {
		t.Errorf("cache should return independent copies, got %q", got2.Name)
	}
}

func TestMemoryCache_Miss(t *testing.T) {
	memCache := NewMemoryCache[testItem](100, time.Minute)
	defer memCache.Close()

	got, err := memCache.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil on miss, got %v", got)
	}
}

func TestMemoryCache_TTLExpiry(t *testing.T) {
	memCache := NewMemoryCache[testItem](100, 50*time.Millisecond)
	defer memCache.Close()

	ctx := context.Background()
	memCache.Set(ctx, "key1", &testItem{Name: "expiring"})

	time.Sleep(60 * time.Millisecond)

	got, err := memCache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after TTL expiry, got %v", got)
	}
}

func TestMemoryCache_ExpiredEntryDeletedOnGet(t *testing.T) {
	memCache := NewMemoryCache[testItem](2, 50*time.Millisecond)
	defer memCache.Close()

	ctx := context.Background()
	memCache.Set(ctx, "a", &testItem{Name: "a"})

	time.Sleep(60 * time.Millisecond)

	// Get should delete expired entry, freeing the slot
	memCache.Get(ctx, "a")

	memCache.mu.Lock()
	count := memCache.order.Len()
	memCache.mu.Unlock()

	if count != 0 {
		t.Errorf("expected expired entry to be removed, got %d entries", count)
	}
}

func TestMemoryCache_LRUEviction(t *testing.T) {
	memCache := NewMemoryCache[testItem](2, time.Minute)
	defer memCache.Close()

	ctx := context.Background()

	memCache.Set(ctx, "a", &testItem{Name: "a"})
	memCache.Set(ctx, "b", &testItem{Name: "b"})

	// Access "a" to make it recently used — "b" becomes LRU
	memCache.Get(ctx, "a")

	// Insert "c" — should evict "b" (LRU)
	memCache.Set(ctx, "c", &testItem{Name: "c"})

	got, _ := memCache.Get(ctx, "b")
	if got != nil {
		t.Error("expected 'b' (LRU) to be evicted")
	}

	got, _ = memCache.Get(ctx, "a")
	if got == nil {
		t.Error("expected 'a' (recently used) to still exist")
	}

	got, _ = memCache.Get(ctx, "c")
	if got == nil {
		t.Error("expected 'c' to exist")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	memCache := NewMemoryCache[testItem](100, time.Minute)
	defer memCache.Close()

	ctx := context.Background()

	memCache.Set(ctx, "key1", &testItem{Name: "test"})
	memCache.Delete(ctx, "key1")

	got, _ := memCache.Get(ctx, "key1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

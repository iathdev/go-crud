package cache

import (
	"context"
	"testing"
	"time"
)

func newTestMultiLevel() (*MemoryCache[testItem], *MemoryCache[testItem], *MultiLevelCache[testItem]) {
	l1 := NewMemoryCache[testItem](100, 30*time.Second)
	l2 := NewMemoryCache[testItem](100, time.Minute)
	multi := NewMultiLevelCache[testItem](l1, l2)
	return l1, l2, multi
}

func TestMultiLevelCache_L1Hit(t *testing.T) {
	l1, l2, multi := newTestMultiLevel()
	defer l1.Close()
	defer l2.Close()

	ctx := context.Background()
	l1.Set(ctx, "key1", &testItem{Name: "from-l1"})

	got, err := multi.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || got.Name != "from-l1" {
		t.Errorf("expected L1 hit, got %v", got)
	}
}

func TestMultiLevelCache_L2HitBackfillsL1(t *testing.T) {
	l1, l2, multi := newTestMultiLevel()
	defer l1.Close()
	defer l2.Close()

	ctx := context.Background()
	l2.Set(ctx, "key1", &testItem{Name: "from-l2"})

	got, err := multi.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || got.Name != "from-l2" {
		t.Errorf("expected L2 hit, got %v", got)
	}

	l1Val, _ := l1.Get(ctx, "key1")
	if l1Val == nil || l1Val.Name != "from-l2" {
		t.Error("expected L1 to be backfilled from L2")
	}
}

func TestMultiLevelCache_BothMiss(t *testing.T) {
	l1, l2, multi := newTestMultiLevel()
	defer l1.Close()
	defer l2.Close()

	got, err := multi.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil on both miss, got %v", got)
	}
}

func TestMultiLevelCache_SetPopulatesBothLevels(t *testing.T) {
	l1, l2, multi := newTestMultiLevel()
	defer l1.Close()
	defer l2.Close()

	ctx := context.Background()
	multi.Set(ctx, "key1", &testItem{Name: "both"})

	l1Val, _ := l1.Get(ctx, "key1")
	if l1Val == nil || l1Val.Name != "both" {
		t.Error("expected L1 to have the value")
	}

	l2Val, _ := l2.Get(ctx, "key1")
	if l2Val == nil || l2Val.Name != "both" {
		t.Error("expected L2 to have the value")
	}
}

func TestMultiLevelCache_DeleteClearsBothLevels(t *testing.T) {
	l1, l2, multi := newTestMultiLevel()
	defer l1.Close()
	defer l2.Close()

	ctx := context.Background()
	multi.Set(ctx, "key1", &testItem{Name: "delete-me"})
	multi.Delete(ctx, "key1")

	l1Val, _ := l1.Get(ctx, "key1")
	if l1Val != nil {
		t.Error("expected L1 to be cleared")
	}

	l2Val, _ := l2.Get(ctx, "key1")
	if l2Val != nil {
		t.Error("expected L2 to be cleared")
	}
}

func TestMultiLevelCache_L1ExpiresBeforeL2(t *testing.T) {
	l1 := NewMemoryCache[testItem](100, 50*time.Millisecond)
	l2 := NewMemoryCache[testItem](100, time.Minute)
	multi := NewMultiLevelCache[testItem](l1, l2)
	defer l1.Close()
	defer l2.Close()

	ctx := context.Background()
	multi.Set(ctx, "key1", &testItem{Name: "test"})

	time.Sleep(60 * time.Millisecond)

	l1Val, _ := l1.Get(ctx, "key1")
	if l1Val != nil {
		t.Error("expected L1 to be expired")
	}

	l2Val, _ := l2.Get(ctx, "key1")
	if l2Val == nil {
		t.Error("expected L2 to still have value")
	}
}

func TestMultiLevelCache_Close(t *testing.T) {
	_, _, multi := newTestMultiLevel()
	if err := multi.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

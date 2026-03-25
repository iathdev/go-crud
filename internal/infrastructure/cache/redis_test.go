package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestRedisCache_GetSet(t *testing.T) {
	client, _ := newTestRedis(t)

	rc := NewRedisCache[testItem](client, "test:", time.Minute)
	ctx := context.Background()

	if err := rc.Set(ctx, "key1", &testItem{Name: "test"}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := rc.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil || got.Name != "test" {
		t.Errorf("expected {Name: test}, got %v", got)
	}
}

func TestRedisCache_Miss(t *testing.T) {
	client, _ := newTestRedis(t)

	rc := NewRedisCache[testItem](client, "test:", time.Minute)

	got, err := rc.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil on miss, got %v", got)
	}
}

func TestRedisCache_TTLExpiry(t *testing.T) {
	client, mr := newTestRedis(t)

	rc := NewRedisCache[testItem](client, "test:", time.Minute)
	ctx := context.Background()

	rc.Set(ctx, "key1", &testItem{Name: "expiring"})

	mr.FastForward(2 * time.Minute)

	got, err := rc.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil after TTL expiry, got %v", got)
	}
}

func TestRedisCache_Delete(t *testing.T) {
	client, _ := newTestRedis(t)

	rc := NewRedisCache[testItem](client, "test:", time.Minute)
	ctx := context.Background()

	rc.Set(ctx, "key1", &testItem{Name: "test"})
	if err := rc.Delete(ctx, "key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	got, _ := rc.Get(ctx, "key1")
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestRedisCache_KeyPrefix(t *testing.T) {
	client, mr := newTestRedis(t)

	rc := NewRedisCache[testItem](client, "prefix:", time.Minute)
	rc.Set(context.Background(), "key1", &testItem{Name: "test"})

	if !mr.Exists("prefix:key1") {
		t.Error("expected key to be stored with prefix")
	}
}

package cache

import (
	"container/list"
	"context"
	"encoding/json"
	"sync"
	"time"
)

const defaultCleanupInterval = time.Minute

type lruEntry struct {
	key       string
	data      []byte // JSON-serialized to prevent pointer aliasing
	expiresAt time.Time
}

// MemoryCache is an in-memory LRU cache with TTL expiration.
// Values are stored as JSON bytes to prevent pointer aliasing between
// caller and cache — mutations to cached objects won't affect the original.
type MemoryCache[T any] struct {
	mu      sync.Mutex
	items   map[string]*list.Element
	order   *list.List // front = most recently used, back = LRU
	maxSize int
	ttl     time.Duration
	stop    chan struct{}
}

func NewMemoryCache[T any](maxSize int, ttl time.Duration) *MemoryCache[T] {
	if maxSize <= 0 {
		maxSize = 5000
	}

	if ttl <= 0 {
		ttl = time.Minute
	}

	memCache := &MemoryCache[T]{
		items:   make(map[string]*list.Element, maxSize),
		order:   list.New(),
		maxSize: maxSize,
		ttl:     ttl,
		stop:    make(chan struct{}),
	}

	go memCache.cleanupLoop(defaultCleanupInterval)

	return memCache
}

func (memCache *MemoryCache[T]) Get(_ context.Context, key string) (*T, error) {
	memCache.mu.Lock()
	defer memCache.mu.Unlock()

	elem, exists := memCache.items[key]
	if !exists {
		return nil, nil
	}

	entry := elem.Value.(*lruEntry)
	if time.Now().After(entry.expiresAt) {
		memCache.removeLocked(elem, key)
		return nil, nil
	}

	memCache.order.MoveToFront(elem)

	var value T
	if err := json.Unmarshal(entry.data, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (memCache *MemoryCache[T]) Set(_ context.Context, key string, value *T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	memCache.mu.Lock()
	defer memCache.mu.Unlock()

	if elem, exists := memCache.items[key]; exists {
		entry := elem.Value.(*lruEntry)
		entry.data = data
		entry.expiresAt = time.Now().Add(memCache.ttl)
		memCache.order.MoveToFront(elem)
		return nil
	}

	for memCache.order.Len() >= memCache.maxSize {
		memCache.evictLRU()
	}

	entry := &lruEntry{key: key, data: data, expiresAt: time.Now().Add(memCache.ttl)}
	elem := memCache.order.PushFront(entry)
	memCache.items[key] = elem
	return nil
}

func (memCache *MemoryCache[T]) Delete(_ context.Context, key string) error {
	memCache.mu.Lock()
	defer memCache.mu.Unlock()

	if elem, exists := memCache.items[key]; exists {
		memCache.removeLocked(elem, key)
	}
	return nil
}

func (memCache *MemoryCache[T]) Close() error {
	close(memCache.stop)
	return nil
}

// evictLRU removes the least recently used entry. Caller must hold mu.
func (memCache *MemoryCache[T]) evictLRU() {
	back := memCache.order.Back()
	if back == nil {
		return
	}
	memCache.removeLocked(back, back.Value.(*lruEntry).key)
}

// removeLocked removes an entry from both map and list. Caller must hold mu.
func (memCache *MemoryCache[T]) removeLocked(elem *list.Element, key string) {
	memCache.order.Remove(elem)
	delete(memCache.items, key)
}

func (memCache *MemoryCache[T]) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-memCache.stop:
			return
		case <-ticker.C:
			memCache.deleteExpired()
		}
	}
}

func (memCache *MemoryCache[T]) deleteExpired() {
	memCache.mu.Lock()
	defer memCache.mu.Unlock()

	now := time.Now()
	for key, elem := range memCache.items {
		if now.After(elem.Value.(*lruEntry).expiresAt) {
			memCache.removeLocked(elem, key)
		}
	}
}

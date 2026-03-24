package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a generic key-value cache.
// Returns (*T, nil) on hit, (nil, nil) on miss, (nil, error) on failure.
// T must be JSON-serializable.
type Cache[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value *T) error
	Delete(ctx context.Context, key string) error
	Close() error
}

// NoExpire means entries never expire. Use as TTL value.
const NoExpire time.Duration = -1

type Options struct {
	Level     string        // "L1", "L2", "multi" (default: "L2")
	KeyPrefix string
	TTL       time.Duration // cache TTL (default: 5m, -1 = no expire)
	MemoryMax int           // L1 max entries (default: 5000)
	L1TTL     time.Duration // L1 TTL for multi-level (default: 30s, -1 = no expire)
}

func (options Options) withDefaults() Options {
	if options.Level == "" {
		options.Level = "L2"
	}
	if options.TTL == 0 {
		options.TTL = 5 * time.Minute
	}
	if options.MemoryMax == 0 {
		options.MemoryMax = 5000
	}
	if options.L1TTL == 0 {
		options.L1TTL = 30 * time.Second
	}
	return options
}

func New[T any](redisClient *redis.Client, options Options) Cache[T] {
	options = options.withDefaults()

	switch options.Level {
	case "L1":
		return NewMemoryCache[T](options.MemoryMax, options.TTL)
	case "multi":
		l1 := NewMemoryCache[T](options.MemoryMax, options.L1TTL)
		l2 := NewRedisCache[T](redisClient, options.KeyPrefix, options.TTL)
		return NewMultiLevelCache[T](l1, l2)
	default:
		return NewRedisCache[T](redisClient, options.KeyPrefix, options.TTL)
	}
}

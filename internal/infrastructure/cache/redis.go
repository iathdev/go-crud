package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-backed cache with JSON serialization.
// T must be JSON-serializable.
type RedisCache[T any] struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRedisCache[T any](client *redis.Client, keyPrefix string, ttl time.Duration) *RedisCache[T] {
	return &RedisCache[T]{client: client, keyPrefix: keyPrefix, ttl: ttl}
}

func (redisCache *RedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
	data, err := redisCache.client.Get(ctx, redisCache.keyPrefix+key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return &value, nil
}

func (redisCache *RedisCache[T]) Set(ctx context.Context, key string, value *T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	ttl := redisCache.ttl
	if ttl == NoExpire {
		ttl = 0 // Redis 0 = no expiration
	}

	return redisCache.client.Set(ctx, redisCache.keyPrefix+key, data, ttl).Err()
}

func (redisCache *RedisCache[T]) Delete(ctx context.Context, key string) error {
	return redisCache.client.Del(ctx, redisCache.keyPrefix+key).Err()
}

func (redisCache *RedisCache[T]) Close() error {
	return nil
}

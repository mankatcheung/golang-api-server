// Package cache provides a generic caching interface and Redis implementation.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines a generic key-value cache with TTL support.
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
}

type RedisCache struct {
	client redis.UniversalClient
}

// NewRedisCache returns a Cache backed by Redis. When addrs contains more than
// one address a ClusterClient is created; otherwise a single-node client is used.
func NewRedisCache(addrs []string, password string, db int) Cache {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    addrs,
		Password: password,
		DB:       db,
	})
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("redis get: %w", err)
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	if cluster, ok := c.client.(*redis.ClusterClient); ok {
		return cluster.ForEachMaster(ctx, func(ctx context.Context, master *redis.Client) error {
			return scanAndDelete(ctx, master, pattern)
		})
	}
	return scanAndDelete(ctx, c.client, pattern)
}

func scanAndDelete(ctx context.Context, client redis.Cmdable, pattern string) error {
	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("redis del %s: %w", iter.Val(), err)
		}
	}
	return iter.Err()
}

// Close closes the Redis client connection.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// ErrCacheMiss is returned when a key is not found in the cache.
var ErrCacheMiss = fmt.Errorf("cache miss")

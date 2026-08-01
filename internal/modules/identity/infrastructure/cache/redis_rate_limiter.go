package cache

import (
	"context"
	"fmt"
	"time"

	ports "sipon-be/internal/modules/identity/application/ports"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (l *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (ports.RateLimitResult, error) {
	fullKey := fmt.Sprintf("ratelimit:%s", key)

	val, err := l.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return ports.RateLimitResult{}, fmt.Errorf("rate limit incr: %w", err)
	}

	if val == 1 {
		l.client.Expire(ctx, fullKey, window)
	}

	ttl, err := l.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return ports.RateLimitResult{}, fmt.Errorf("rate limit ttl: %w", err)
	}

	remaining := limit - int(val)
	if remaining < 0 {
		remaining = 0
	}

	allowed := val <= int64(limit)

	return ports.RateLimitResult{
		Allowed:   allowed,
		Remaining: remaining,
		ResetAt:   time.Now().Add(ttl),
	}, nil
}

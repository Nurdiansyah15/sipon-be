package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSessionRevocationStore struct {
	client *redis.Client
}

func NewRedisSessionRevocationStore(client *redis.Client) *RedisSessionRevocationStore {
	return &RedisSessionRevocationStore{client: client}
}

func (s *RedisSessionRevocationStore) RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:revoked:%s", sessionID)
	return s.client.Set(ctx, key, "1", ttl).Err()
}

func (s *RedisSessionRevocationStore) IsSessionRevoked(ctx context.Context, sessionID string) (bool, error) {
	key := fmt.Sprintf("session:revoked:%s", sessionID)
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check session revoked: %w", err)
	}
	return val == "1", nil
}

func (s *RedisSessionRevocationStore) RevokeAllBefore(ctx context.Context, userID string, before time.Time, ttl time.Duration) error {
	key := fmt.Sprintf("session:revoked_before:%s", userID)
	return s.client.Set(ctx, key, strconv.FormatInt(before.Unix(), 10), ttl).Err()
}

func (s *RedisSessionRevocationStore) RevokedBefore(ctx context.Context, userID string) (*time.Time, error) {
	key := fmt.Sprintf("session:revoked_before:%s", userID)
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get revoked before: %w", err)
	}

	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse revoked before: %w", err)
	}

	t := time.Unix(ts, 0)
	return &t, nil
}

func (s *RedisSessionRevocationStore) RevokeDeviceBefore(ctx context.Context, userID, deviceID string, before time.Time, ttl time.Duration) error {
	key := fmt.Sprintf("session:revoked_before:%s:%s", userID, deviceID)
	return s.client.Set(ctx, key, strconv.FormatInt(before.Unix(), 10), ttl).Err()
}

func (s *RedisSessionRevocationStore) DeviceRevokedBefore(ctx context.Context, userID, deviceID string) (*time.Time, error) {
	key := fmt.Sprintf("session:revoked_before:%s:%s", userID, deviceID)
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get device revoked before: %w", err)
	}

	ts, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse device revoked before: %w", err)
	}

	t := time.Unix(ts, 0)
	return &t, nil
}

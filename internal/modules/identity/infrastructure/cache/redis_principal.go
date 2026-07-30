package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PrincipalData struct {
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	Scopes      []Scope  `json:"scopes"`
}

type Scope struct {
	ScopeType string  `json:"scope_type"`
	ScopeID   *string `json:"scope_id"`
}

type RedisPrincipalCache struct {
	client *redis.Client
}

func NewRedisPrincipalCache(client *redis.Client) *RedisPrincipalCache {
	return &RedisPrincipalCache{client: client}
}

func (c *RedisPrincipalCache) Get(ctx context.Context, userID string) (*PrincipalData, error) {
	key := fmt.Sprintf("principal:%s", userID)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get principal cache: %w", err)
	}

	var data PrincipalData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("unmarshal principal: %w", err)
	}

	return &data, nil
}

func (c *RedisPrincipalCache) Set(ctx context.Context, userID string, data *PrincipalData, ttl time.Duration) error {
	key := fmt.Sprintf("principal:%s", userID)

	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal principal: %w", err)
	}

	return c.client.Set(ctx, key, string(val), ttl).Err()
}

func (c *RedisPrincipalCache) Delete(ctx context.Context, userID string) error {
	key := fmt.Sprintf("principal:%s", userID)
	return c.client.Del(ctx, key).Err()
}

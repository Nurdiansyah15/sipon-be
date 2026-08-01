package ports

import (
	"context"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (RateLimitResult, error)
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

package ports

import (
	"context"
	"time"
)

type SessionRevocationStore interface {
	RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error
	IsSessionRevoked(ctx context.Context, sessionID string) (bool, error)
	RevokeAllBefore(ctx context.Context, userID string, before time.Time, ttl time.Duration) error
	RevokedBefore(ctx context.Context, userID string) (*time.Time, error)
	RevokeDeviceBefore(ctx context.Context, userID, deviceID string, before time.Time, ttl time.Duration) error
	DeviceRevokedBefore(ctx context.Context, userID, deviceID string) (*time.Time, error)
}

package application

import (
	"context"
	"time"

	"sipon-be/internal/modules/identity/domain"
)

type TokenGenerator interface {
	GenerateAccessToken(userID, sessionID, deviceID string) (string, error)
	GenerateRefreshToken(userID, deviceID string) (string, error)
	ParseAccessToken(token string) (*TokenClaims, error)
	ParseRefreshToken(token string) (*RefreshTokenClaims, error)
}

type TokenClaims struct {
	UserID    string
	SessionID string
	DeviceID  string
	IssuedAt  time.Time
}

type RefreshTokenClaims struct {
	UserID   string
	DeviceID string
	IssuedAt time.Time
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hashed, plain string) error
}

type EmailSender interface {
	SendOTP(toEmail, username, otp string) error
	SendPasswordResetOTP(toEmail, username, otp string) error
}

type SMSSender interface {
	SendOTP(toPhone, otp string) error
}

type OTPGenerator interface {
	Generate() (string, error)
}

type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	MarkDeleted(ctx context.Context, key string) error
	DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
	PublicURL(key string) string
	KeyFromURL(url string) string
}

type PrivacyRule string

const (
	PrivacyPublic  PrivacyRule = "PUBLIC"
	PrivacyPrivate PrivacyRule = "PRIVATE"
)

type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type SessionRevocationStore interface {
	RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error
	IsSessionRevoked(ctx context.Context, sessionID string) (bool, error)
	RevokeAllBefore(ctx context.Context, userID string, before time.Time, ttl time.Duration) error
	RevokedBefore(ctx context.Context, userID string) (*time.Time, error)
	RevokeDeviceBefore(ctx context.Context, userID, deviceID string, before time.Time, ttl time.Duration) error
	DeviceRevokedBefore(ctx context.Context, userID, deviceID string) (*time.Time, error)
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (RateLimitResult, error)
}

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

type UserListRepository interface {
	List(ctx context.Context, status string, roleID string, search string, page, limit int) ([]*domain.User, int64, error)
	FindByIDWithRoles(ctx context.Context, userID string) (*domain.User, []string, error)
}

package ports

import "time"

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

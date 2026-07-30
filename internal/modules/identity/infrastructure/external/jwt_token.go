package external

import (
	"time"

	"sipon-be/internal/modules/identity/application"

	"github.com/golang-jwt/jwt/v5"
)

type JWTTokenGenerator struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTTokenGenerator(secret string, accessTTL, refreshTTL time.Duration) *JWTTokenGenerator {
	return &JWTTokenGenerator{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

type accessClaims struct {
	UserID    string `json:"sub"`
	SessionID string `json:"sid"`
	DeviceID  string `json:"did"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UserID   string `json:"sub"`
	DeviceID string `json:"did"`
	jwt.RegisteredClaims
}

func (g *JWTTokenGenerator) GenerateAccessToken(userID, sessionID, deviceID string) (string, error) {
	now := time.Now()
	claims := accessClaims{
		UserID:    userID,
		SessionID: sessionID,
		DeviceID:  deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(g.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}

func (g *JWTTokenGenerator) GenerateRefreshToken(userID, deviceID string) (string, error) {
	now := time.Now()
	claims := refreshClaims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(g.refreshTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}

func (g *JWTTokenGenerator) ParseAccessToken(tokenString string) (*application.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(t *jwt.Token) (interface{}, error) {
		return g.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return &application.TokenClaims{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		DeviceID:  claims.DeviceID,
		IssuedAt:  claims.IssuedAt.Time,
	}, nil
}

func (g *JWTTokenGenerator) ParseRefreshToken(tokenString string) (*application.RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &refreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		return g.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*refreshClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return &application.RefreshTokenClaims{
		UserID:   claims.UserID,
		DeviceID: claims.DeviceID,
		IssuedAt: claims.IssuedAt.Time,
	}, nil
}

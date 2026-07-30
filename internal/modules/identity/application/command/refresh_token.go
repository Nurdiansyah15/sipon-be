package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RefreshTokenUseCase struct {
	tokenGen               application.TokenGenerator
	sessionRevocationStore application.SessionRevocationStore
}

func NewRefreshTokenUseCase(
	tokenGen application.TokenGenerator,
	sessionRevocationStore application.SessionRevocationStore,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		tokenGen:               tokenGen,
		sessionRevocationStore: sessionRevocationStore,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	claims, err := uc.tokenGen.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInvalidToken, err)
	}

	revokedBefore, err := uc.sessionRevocationStore.RevokedBefore(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if revokedBefore != nil && claims.IssuedAt.Before(*revokedBefore) {
		return nil, kernel.New(application.ErrCodeSessionRevoked)
	}

	deviceRevokedBefore, err := uc.sessionRevocationStore.DeviceRevokedBefore(ctx, claims.UserID, claims.DeviceID)
	if err != nil {
		return nil, err
	}
	if deviceRevokedBefore != nil && claims.IssuedAt.Before(*deviceRevokedBefore) {
		return nil, kernel.New(application.ErrCodeSessionRevoked)
	}

	sessionID := uuid.NewString()
	deviceID := uuid.NewString()

	accessToken, err := uc.tokenGen.GenerateAccessToken(claims.UserID, sessionID, deviceID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenGen.GenerateRefreshToken(claims.UserID, deviceID)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

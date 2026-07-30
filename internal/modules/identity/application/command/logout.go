package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/shared/kernel"
)

const (
	sessionRevokeTTL = 30 * 24 * time.Hour
)

type LogoutUseCase struct {
	tokenGen               application.TokenGenerator
	sessionRevocationStore application.SessionRevocationStore
}

func NewLogoutUseCase(
	tokenGen application.TokenGenerator,
	sessionRevocationStore application.SessionRevocationStore,
) *LogoutUseCase {
	return &LogoutUseCase{
		tokenGen:               tokenGen,
		sessionRevocationStore: sessionRevocationStore,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, accessToken string) error {
	claims, err := uc.tokenGen.ParseAccessToken(accessToken)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInvalidToken, err)
	}

	return uc.sessionRevocationStore.RevokeSession(ctx, claims.SessionID, sessionRevokeTTL)
}

func (uc *LogoutUseCase) ExecuteRevokeAll(ctx context.Context, accessToken string) error {
	claims, err := uc.tokenGen.ParseAccessToken(accessToken)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInvalidToken, err)
	}

	return uc.sessionRevocationStore.RevokeAllBefore(ctx, claims.UserID, time.Now(), sessionRevokeTTL)
}

func (uc *LogoutUseCase) ExecuteRevokeDevice(ctx context.Context, accessToken string) error {
	claims, err := uc.tokenGen.ParseAccessToken(accessToken)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInvalidToken, err)
	}

	return uc.sessionRevocationStore.RevokeDeviceBefore(ctx, claims.UserID, claims.DeviceID, time.Now(), sessionRevokeTTL)
}

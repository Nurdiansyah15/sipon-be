package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/shared/kernel"
)

type LogoutUseCase struct {
	revocationStore application.SessionRevocationStore
	accessTokenTTL  time.Duration
}

func NewLogoutUseCase(revocationStore application.SessionRevocationStore, accessTokenTTL time.Duration) *LogoutUseCase {
	return &LogoutUseCase{revocationStore: revocationStore, accessTokenTTL: accessTokenTTL}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	if err := uc.revocationStore.RevokeSession(ctx, sessionID, uc.accessTokenTTL); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	return nil
}

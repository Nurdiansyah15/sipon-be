package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/identity/application"
	ports "sipon-be/internal/modules/identity/application/ports"
	"sipon-be/internal/shared/kernel"
)

type LogoutUseCase struct {
	revocationStore ports.SessionRevocationStore
	accessTokenTTL  time.Duration
}

func NewLogoutUseCase(revocationStore ports.SessionRevocationStore, accessTokenTTL time.Duration) *LogoutUseCase {
	return &LogoutUseCase{revocationStore: revocationStore, accessTokenTTL: accessTokenTTL}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, sessionID string) error {
	if err := uc.revocationStore.RevokeSession(ctx, sessionID, uc.accessTokenTTL); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal mencabut sesi", err)
	}
	return nil
}

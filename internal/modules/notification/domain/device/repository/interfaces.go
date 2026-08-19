package repository

import (
	"context"

	"sipon-be/internal/modules/notification/domain/device/entity"
)

type DeviceRegistrationRepository interface {
	Save(ctx context.Context, dr *entity.DeviceRegistration) error
	Update(ctx context.Context, dr *entity.DeviceRegistration) error
	FindByID(ctx context.Context, id string) (*entity.DeviceRegistration, error)
	FindByToken(ctx context.Context, token string) (*entity.DeviceRegistration, error)
	FindByUserIDAndToken(ctx context.Context, userID, token string) (*entity.DeviceRegistration, error)
	FindByUserID(ctx context.Context, userID string, includeInactive bool) ([]*entity.DeviceRegistration, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.DeviceRegistration, error)
	FindActiveByUserIDs(ctx context.Context, userIDs []string) (map[string][]*entity.DeviceRegistration, error)
	DeactivateByToken(ctx context.Context, token string) error
	DeactivateAllByUserID(ctx context.Context, userID string) error
}

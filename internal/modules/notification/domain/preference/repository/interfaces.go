package repository

import (
	"context"

	"sipon-be/internal/modules/notification/domain/preference/entity"
)

// NotificationPreferenceRepository menyimpan dan mengambil NotificationPreference.
type NotificationPreferenceRepository interface {
	FindOrCreateByUserID(ctx context.Context, userID string) (*entity.NotificationPreference, error)
	Update(ctx context.Context, pref *entity.NotificationPreference) error
	FindByUserIDs(ctx context.Context, userIDs []string) (map[string]*entity.NotificationPreference, error)
}

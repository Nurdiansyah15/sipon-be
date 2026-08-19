package repository

import (
	"context"

	"sipon-be/internal/modules/notification/domain/notification/entity"
)

// NotificationRepository menyimpan dan mengambil blueprint Notification.
type NotificationRepository interface {
	Save(ctx context.Context, n *entity.Notification) error
	FindByID(ctx context.Context, id string) (*entity.Notification, error)
}

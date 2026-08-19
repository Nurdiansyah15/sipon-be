package command

import (
	"context"

	"sipon-be/internal/modules/notification/application/dto"
	deliveryRepo "sipon-be/internal/modules/notification/domain/delivery/repository"
)

// MarkAllNotificationsReadUseCase menandai semua notifikasi in_app sebagai sudah dibaca.
type MarkAllNotificationsReadUseCase struct {
	deliveryRepo deliveryRepo.DeliveryAttemptRepository
}

func NewMarkAllNotificationsReadUseCase(repo deliveryRepo.DeliveryAttemptRepository) *MarkAllNotificationsReadUseCase {
	return &MarkAllNotificationsReadUseCase{deliveryRepo: repo}
}

func (uc *MarkAllNotificationsReadUseCase) Execute(ctx context.Context, userID string) (*dto.MarkAllReadResponse, error) {
	marked, err := uc.deliveryRepo.MarkAllRead(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.MarkAllReadResponse{Marked: marked}, nil
}

package query

import (
	"context"

	"sipon-be/internal/modules/notification/application/dto"
	deliveryRepo "sipon-be/internal/modules/notification/domain/delivery/repository"
)

type UnreadCountUseCase struct {
	deliveryRepo deliveryRepo.DeliveryAttemptRepository
}

func NewUnreadCountUseCase(repo deliveryRepo.DeliveryAttemptRepository) *UnreadCountUseCase {
	return &UnreadCountUseCase{deliveryRepo: repo}
}

func (uc *UnreadCountUseCase) Execute(ctx context.Context, userID string) (*dto.UnreadCountResponse, error) {
	count, err := uc.deliveryRepo.CountUnreadInApp(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.UnreadCountResponse{Count: count}, nil
}

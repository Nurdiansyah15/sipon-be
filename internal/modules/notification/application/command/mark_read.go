package command

import (
	"context"

	"sipon-be/internal/modules/notification/application"
	deliveryRepo "sipon-be/internal/modules/notification/domain/delivery/repository"
	notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
)

// MarkNotificationReadUseCase menandai satu DeliveryAttempt (in_app) sebagai sudah dibaca.
type MarkNotificationReadUseCase struct {
	deliveryRepo deliveryRepo.DeliveryAttemptRepository
}

func NewMarkNotificationReadUseCase(repo deliveryRepo.DeliveryAttemptRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{deliveryRepo: repo}
}

func (uc *MarkNotificationReadUseCase) Execute(ctx context.Context, notifID, userID string) error {
	if err := uc.deliveryRepo.MarkRead(ctx, notifID, userID); err != nil {
		return application.WrapDomainErr(err, notifconstant.CodeDeliveryAttemptNotFound)
	}
	return nil
}

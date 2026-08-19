package command

import (
	"context"

	"sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/notification/application/dto"
notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
notifvo "sipon-be/internal/modules/notification/domain/valueobject"
)

// SendNotificationUseCase mengirim notifikasi via dispatcher.
type SendNotificationUseCase struct {
	dispatcher *application.Dispatcher
}

func NewSendNotificationUseCase(dispatcher *application.Dispatcher) *SendNotificationUseCase {
	return &SendNotificationUseCase{dispatcher: dispatcher}
}

func (uc *SendNotificationUseCase) Execute(
	ctx context.Context,
	mode application.TargetMode,
	recipientID string,
	recipientIDs []string,
	req dto.BroadcastRequest,
) error {
	tmpl := application.NotificationTemplate{
		Type: notifconstant.NotificationType(req.Type),
		Title: req.Title,
		Body:  req.Body,
		Payload: notifvo.NotificationPayload{
			Module:    "announcement",
			EventType: "broadcast",
			Bypass:    true,
		},
		Channels: []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp},
		Bypass:   true,
	}

	target := application.Target{
		Mode:         mode,
		RecipientID:  recipientID,
		RecipientIDs: recipientIDs,
	}

	return uc.dispatcher.Dispatch(ctx, tmpl, target)
}

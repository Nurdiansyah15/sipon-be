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
	channels := parseChannels(req.Channels)

	tmpl := application.NotificationTemplate{
		Type: notifconstant.NotificationType(req.Type),
		Title: req.Title,
		Body:  req.Body,
		Payload: notifvo.NotificationPayload{
			Module:    "announcement",
			EventType: "broadcast",
			Bypass:    true,
		},
		Channels: channels,
		Bypass:   true,
	}

	target := application.Target{
		Mode:         mode,
		RecipientID:  recipientID,
		RecipientIDs: recipientIDs,
	}

	return uc.dispatcher.Dispatch(ctx, tmpl, target)
}

func parseChannels(raw []string) []notifconstant.NotificationChannel {
	if len(raw) == 0 {
		return []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp}
	}
	channels := make([]notifconstant.NotificationChannel, 0, len(raw))
	for _, c := range raw {
		ch := notifconstant.NotificationChannel(c)
		if ch.IsValid() {
			channels = append(channels, ch)
		}
	}
	if len(channels) == 0 {
		return []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp}
	}
	return channels
}

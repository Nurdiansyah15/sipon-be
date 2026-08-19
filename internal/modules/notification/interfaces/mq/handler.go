package mq

import (
	"context"
	"encoding/json"
	"fmt"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	notifvo "sipon-be/internal/modules/notification/domain/valueobject"
	"sipon-be/internal/modules/notification/application"
	"sipon-be/internal/modules/messaging"
	"github.com/google/uuid"
)

type Dependencies struct {
	Dispatcher *application.Dispatcher
}

type handlers struct {
	deps Dependencies
}

func (h handlers) handleLoginSucceeded(ctx context.Context, msg messaging.Message) error {
	var p LoginSucceededPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingLoginSucceeded, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Login Berhasil",
		Body:  "Anda telah berhasil masuk ke sistem.",
		Payload: notifvo.NotificationPayload{
			Module:    "identity",
			EventType: "login_succeeded",
			EntityID:  p.UserID,
		},
		Channels: []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp},
		Bypass:   true,
	}

	target := application.Target{
		Mode:        application.TargetModeUnicast,
		RecipientID: p.UserID,
	}

	id := uuid.NewString()
	_ = id

	if err := h.deps.Dispatcher.Dispatch(ctx, tmpl, target); err != nil {
		return messaging.NewRetryableError(err)
	}
	return nil
}

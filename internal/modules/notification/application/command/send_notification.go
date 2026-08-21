package command

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"sipon-be/internal/modules/notification/application/dto"
	"sipon-be/internal/modules/notification/application/ports"
)

// SendNotificationUseCase mempublish broadcast admin ke outbox; fanout ke
// semua user aktif dieksekusi oleh worker lewat handler notification.broadcast.
type SendNotificationUseCase struct {
	outboxWriter ports.OutboxWriter
}

func NewSendNotificationUseCase() *SendNotificationUseCase {
	return &SendNotificationUseCase{}
}

func (uc *SendNotificationUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *SendNotificationUseCase) Execute(ctx context.Context, req dto.BroadcastRequest) error {
	if uc.outboxWriter == nil {
		return errors.New("notification: outbox writer belum dipasang")
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	if err := uc.outboxWriter.Save(ctx, RoutingAdminBroadcast, payload); err != nil {
		slog.Warn("notification: gagal publish broadcast admin", slog.Any("error", err))
		return err
	}
	return nil
}

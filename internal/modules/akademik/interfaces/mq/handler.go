package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/messaging/domain/message/valueobject"
	messagingerrors "sipon-be/internal/modules/messaging/domain/message_job/errors"
)

// Dependencies adalah usecase yang dipanggil handler MQ akademik. Hanya berisi
// dependency yang benar-benar dipakai layer transport; business logic tetap di
// application/command.
type Dependencies struct {
	FingerprintSync  *command.SyncAttendanceFromFingerprintUseCase
	SessionAutoClose *command.AutoCloseSessionUseCase
}

type handlers struct {
	deps Dependencies
}

func (h handlers) handleFingerprintSync(ctx context.Context, msg valueobject.Message) error {
	var p FingerprintSyncPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messagingerrors.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingFingerprintSync, err))
	}
	if err := p.Validate(); err != nil {
		return messagingerrors.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	if _, err := h.deps.FingerprintSync.Execute(ctx, p.SessionID); err != nil {
		return messagingerrors.NewRetryableError(err)
	}
	return nil
}

func (h handlers) handleSessionAutoClose(ctx context.Context, msg valueobject.Message) error {
	var p SessionAutoClosePayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messagingerrors.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingSessionAutoClose, err))
	}
	if err := p.Validate(); err != nil {
		return messagingerrors.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	if err := h.deps.SessionAutoClose.Execute(ctx, p.SessionID); err != nil {
		return messagingerrors.NewRetryableError(err)
	}
	return nil
}

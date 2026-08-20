package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/messaging"
	"sipon-be/internal/shared/kernel"
)

// Dependencies adalah usecase yang dipanggil handler MQ akademik. Hanya berisi
// dependency yang benar-benar dipakai layer transport; business logic tetap di
// application/command.
type Dependencies struct {
	FingerprintSync  *command.SyncAttendanceFromFingerprintUseCase
	SessionAutoClose *command.AutoCloseSessionUseCase
	SessionAutoOpen  *command.AutoOpenSessionUseCase
}

type handlers struct {
	deps Dependencies
}

func (h handlers) handleFingerprintSync(ctx context.Context, msg messaging.Message) error {
	var p FingerprintSyncPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingFingerprintSync, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	if _, err := h.deps.FingerprintSync.Execute(ctx, p.SessionID); err != nil {
		if isUnprocessable(err) {
			return messaging.NewFatalError(err)
		}
		return messaging.NewRetryableError(err)
	}
	return nil
}

func (h handlers) handleSessionAutoClose(ctx context.Context, msg messaging.Message) error {
	var p SessionAutoClosePayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingSessionAutoClose, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	if err := h.deps.SessionAutoClose.Execute(ctx, p.SessionID); err != nil {
		if isUnprocessable(err) {
			return messaging.NewFatalError(err)
		}
		return messaging.NewRetryableError(err)
	}
	return nil
}

func (h handlers) handleSessionAutoOpen(ctx context.Context, msg messaging.Message) error {
	var p SessionAutoOpenPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingSessionAutoOpen, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	if err := h.deps.SessionAutoOpen.Execute(ctx, p.SessionID); err != nil {
		if isUnprocessable(err) {
			return messaging.NewFatalError(err)
		}
		return messaging.NewRetryableError(err)
	}
	return nil
}

func isUnprocessable(err error) bool {
	var ke *kernel.AppError
	return errors.As(err, &ke) && ke.Code == application.ErrCodeUnprocessableEntity
}

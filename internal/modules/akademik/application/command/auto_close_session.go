package command

import (
	"context"
	"log/slog"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
)

// SessionCompleter adalah boundary untuk menyelesaikan sesi. Dipenuhi oleh
// CompleteSessionUseCase; interface ini memudahkan unit test AutoCloseSessionUseCase.
type SessionCompleter interface {
	Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error)
}

var _ SessionCompleter = (*CompleteSessionUseCase)(nil)

// AutoCloseSessionUseCase menyelesaikan sesi saat waktu berakhir, lalu mem-pause
// recurring fingerprint sync job milik sesi tersebut supaya tidak lagi ditarik
// oleh scheduler setelah sesi ditutup.
type AutoCloseSessionUseCase struct {
	completeSessionUC   SessionCompleter
	scheduler           ports.Scheduler
	fingerprintSyncType string
}

func NewAutoCloseSessionUseCase(
	completeSessionUC SessionCompleter,
	scheduler ports.Scheduler,
	fingerprintSyncType string,
) *AutoCloseSessionUseCase {
	return &AutoCloseSessionUseCase{
		completeSessionUC:   completeSessionUC,
		scheduler:           scheduler,
		fingerprintSyncType: fingerprintSyncType,
	}
}

func (uc *AutoCloseSessionUseCase) Execute(ctx context.Context, sessionID string) error {
	if _, err := uc.completeSessionUC.Execute(ctx, sessionID); err != nil {
		return err
	}

	if err := uc.scheduler.PauseByTypeAndReferenceID(ctx, uc.fingerprintSyncType, sessionID); err != nil {
		slog.Warn("akademik: gagal pause recurring sync job",
			"session_id", sessionID, "error", err)
	}
	return nil
}

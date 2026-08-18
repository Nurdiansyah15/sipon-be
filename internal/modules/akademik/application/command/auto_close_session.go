package command

import (
	"context"
	"log/slog"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/constant"
	sjRepo "sipon-be/internal/shared/scheduler/domain/scheduled_job/repository"
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
	scheduledJobRepo    sjRepo.Repository
	fingerprintSyncType string
}

func NewAutoCloseSessionUseCase(
	completeSessionUC SessionCompleter,
	scheduledJobRepo sjRepo.Repository,
	fingerprintSyncType string,
) *AutoCloseSessionUseCase {
	return &AutoCloseSessionUseCase{
		completeSessionUC:   completeSessionUC,
		scheduledJobRepo:    scheduledJobRepo,
		fingerprintSyncType: fingerprintSyncType,
	}
}

func (uc *AutoCloseSessionUseCase) Execute(ctx context.Context, sessionID string) error {
	if _, err := uc.completeSessionUC.Execute(ctx, sessionID); err != nil {
		return err
	}

	syncJob, err := uc.scheduledJobRepo.FindByTypeAndReferenceID(ctx, uc.fingerprintSyncType, sessionID)
	if err != nil {
		slog.Warn("akademik: gagal cari recurring sync job untuk di-pause",
			"session_id", sessionID, "error", err)
		return nil
	}
	if syncJob != nil && syncJob.Status == constant.StatusActive {
		syncJob.Pause()
		if err := uc.scheduledJobRepo.Update(ctx, syncJob); err != nil {
			slog.Warn("akademik: gagal pause recurring sync job",
				"session_id", sessionID, "error", err)
		}
	}
	return nil
}

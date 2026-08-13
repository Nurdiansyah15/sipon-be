package command

import (
	"context"
	"log/slog"
	"time"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
)

type OpenSessionUseCase struct {
	sessionRepo      sesRepo.ActivitySessionRepository
	scheduleJobsUC   *ScheduleSessionJobsUseCase
}

func NewOpenSessionUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleJobsUC *ScheduleSessionJobsUseCase,
) *OpenSessionUseCase {
	return &OpenSessionUseCase{
		sessionRepo:    sessionRepo,
		scheduleJobsUC: scheduleJobsUC,
	}
}

func (uc *OpenSessionUseCase) Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivitySessionNotFound)
	}

	now := time.Now()
	if now.Before(session.StartsAt) {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "sesi belum waktunya dibuka (sebelum waktu mulai)", nil)
	}
	if now.After(session.EndsAt) {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "sesi sudah lewat waktunya (setelah waktu selesai)", nil)
	}

	if err := session.Open(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivitySessionInvalidStatus)
	}
	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if uc.scheduleJobsUC != nil {
		if err := uc.scheduleJobsUC.Execute(ctx, session.ID, session.EndsAt); err != nil {
			slog.Warn("akademik: gagal schedule session jobs",
				"session_id", session.ID, "error", err)
		}
	}

	return MapSessionToResponse(session), nil
}

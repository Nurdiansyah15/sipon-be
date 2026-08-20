package command

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	"sipon-be/internal/modules/akademik/domain/activity_session/entity"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type CreateSessionUseCase struct {
	sessionRepo       sesRepo.ActivitySessionRepository
	scheduleRepo      schRepo.ActivityScheduleRepository
	scheduleAutoOpenUC *ScheduleAutoOpenUseCase
}

func NewCreateSessionUseCase(sessionRepo sesRepo.ActivitySessionRepository, scheduleRepo schRepo.ActivityScheduleRepository, scheduleAutoOpenUC *ScheduleAutoOpenUseCase) *CreateSessionUseCase {
	return &CreateSessionUseCase{sessionRepo: sessionRepo, scheduleRepo: scheduleRepo, scheduleAutoOpenUC: scheduleAutoOpenUC}
}

func (uc *CreateSessionUseCase) Execute(ctx context.Context, req dto.CreateSessionRequest) (*dto.ActivitySessionResponse, error) {
	schedule, err := uc.scheduleRepo.FindByID(ctx, req.ActivityScheduleID)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivitySessionNotFound)
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	// Aplikasikan early/late minutes dari jadwal: start dimajukan (early) dan
	// end dimundurkan (late).
	startsAt = startsAt.Add(-time.Duration(schedule.EarlyMinutes) * time.Minute)
	endsAt = endsAt.Add(time.Duration(schedule.LateMinutes) * time.Minute)

	session, err := entity.NewActivitySession(uuid.NewString(), req.ActivityScheduleID, startsAt, endsAt)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivitySessionInvalidTime)
	}
	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if uc.scheduleAutoOpenUC != nil {
		if err := uc.scheduleAutoOpenUC.Execute(ctx, session.ID, session.StartsAt); err != nil {
			slog.Warn("akademik: gagal schedule auto-open",
				"session_id", session.ID, "starts_at", session.StartsAt, "error", err)
		}
	}

	return MapSessionToResponse(session), nil
}

func MapSessionToResponse(s *entity.ActivitySession) *dto.ActivitySessionResponse {
	return &dto.ActivitySessionResponse{
		ID:                 s.ID,
		ActivityScheduleID: s.ActivityScheduleID,
		StartsAt:           timeutil.ToPlatform(s.StartsAt),
		EndsAt:             timeutil.ToPlatform(s.EndsAt),
		Status:             string(s.Status),
		CreatedAt:          timeutil.ToPlatform(s.CreatedAt),
		UpdatedAt:          timeutil.ToPlatform(s.UpdatedAt),
	}
}

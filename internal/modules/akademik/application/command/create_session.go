package command

import (
	"context"
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
	sessionRepo  sesRepo.ActivitySessionRepository
	scheduleRepo schRepo.ActivityScheduleRepository
}

func NewCreateSessionUseCase(sessionRepo sesRepo.ActivitySessionRepository, scheduleRepo schRepo.ActivityScheduleRepository) *CreateSessionUseCase {
	return &CreateSessionUseCase{sessionRepo: sessionRepo, scheduleRepo: scheduleRepo}
}

func (uc *CreateSessionUseCase) Execute(ctx context.Context, req dto.CreateSessionRequest) (*dto.ActivitySessionResponse, error) {
	if _, err := uc.scheduleRepo.FindByID(ctx, req.ActivityScheduleID); err != nil {
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

	session, err := entity.NewActivitySession(uuid.NewString(), req.ActivityScheduleID, startsAt, endsAt)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivitySessionInvalidTime)
	}
	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
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

package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	"sipon-be/internal/modules/akademik/domain/activity_period/constant"
	"sipon-be/internal/modules/akademik/domain/activity_period/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type CreateActivityPeriodUseCase struct {
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
	periodRepo         periodRepo.AcademicPeriodRepository
}

func NewCreateActivityPeriodUseCase(
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	periodRepo periodRepo.AcademicPeriodRepository,
) *CreateActivityPeriodUseCase {
	return &CreateActivityPeriodUseCase{activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo, periodRepo: periodRepo}
}

func (uc *CreateActivityPeriodUseCase) Execute(ctx context.Context, req dto.CreateActivityPeriodRequest) (*dto.ActivityPeriodResponse, error) {
	activity, err := uc.activityRepo.FindByID(ctx, req.ActivityID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}
	period, err := uc.periodRepo.FindByID(ctx, req.AcademicPeriodID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	existing, err := uc.activityPeriodRepo.FindByActivityAndPeriod(ctx, req.ActivityID, req.AcademicPeriodID)
	if err != nil && !application.IsNotFoundErr(err, constant.CodeActivityPeriodNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	ap, err := entity.NewActivityPeriod(uuid.NewString(), req.ActivityID, req.AcademicPeriodID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if err := uc.activityPeriodRepo.Save(ctx, ap); err != nil {
		return nil, application.WrapConflictErr(err, constant.CodeActivityPeriodDuplicate)
	}

	resp := MapActivityPeriodToResponse(ap)
	resp.ActivityCode = activity.Code
	resp.ActivityName = activity.Name
	resp.PeriodName = period.Name
	return resp, nil
}

func MapActivityPeriodToResponse(p *entity.ActivityPeriod) *dto.ActivityPeriodResponse {
	return &dto.ActivityPeriodResponse{
		ID:               p.ID,
		ActivityID:       p.ActivityID,
		AcademicPeriodID: p.AcademicPeriodID,
		Status:           string(p.Status),
		CreatedAt:        timeutil.ToPlatform(p.CreatedAt),
		UpdatedAt:        timeutil.ToPlatform(p.UpdatedAt),
	}
}

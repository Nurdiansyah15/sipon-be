package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity/repository"
	"sipon-be/internal/shared/kernel"
)

type GetActivityUseCase struct {
	activityRepo repository.ActivityRepository
}

func NewGetActivityUseCase(activityRepo repository.ActivityRepository) *GetActivityUseCase {
	return &GetActivityUseCase{activityRepo: activityRepo}
}

func (uc *GetActivityUseCase) Execute(ctx context.Context, id string) (*dto.ActivityResponse, error) {
	activity, err := uc.activityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, kernel.Code("ACTIVITY_NOT_FOUND"))
	}
	return command.MapActivityToResponse(activity), nil
}

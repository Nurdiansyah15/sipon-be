package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity_period/constant"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/shared/kernel"
)

type ActivateActivityPeriodUseCase struct {
	activityPeriodRepo apRepo.ActivityPeriodRepository
}

func NewActivateActivityPeriodUseCase(activityPeriodRepo apRepo.ActivityPeriodRepository) *ActivateActivityPeriodUseCase {
	return &ActivateActivityPeriodUseCase{activityPeriodRepo: activityPeriodRepo}
}

func (uc *ActivateActivityPeriodUseCase) Execute(ctx context.Context, id string) (*dto.ActivityPeriodResponse, error) {
	ap, err := uc.activityPeriodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivityPeriodNotFound)
	}
	if err := ap.Activate(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivityPeriodInvalidStatus)
	}
	if err := uc.activityPeriodRepo.Update(ctx, ap); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapActivityPeriodToResponse(ap), nil
}

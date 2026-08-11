package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity/constant"
	repo "sipon-be/internal/modules/akademik/domain/activity/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateActivityUseCase struct {
	activityRepo repo.ActivityRepository
}

func NewUpdateActivityUseCase(activityRepo repo.ActivityRepository) *UpdateActivityUseCase {
	return &UpdateActivityUseCase{activityRepo: activityRepo}
}

func (uc *UpdateActivityUseCase) Execute(ctx context.Context, id string, req dto.UpdateActivityRequest) (*dto.ActivityResponse, error) {
	activity, err := uc.activityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivityNotFound)
	}

	if req.Code != nil && *req.Code != "" && *req.Code != activity.Code {
		existing, _ := uc.activityRepo.FindByCode(ctx, *req.Code)
		if existing != nil && existing.ID != id {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		activity.Code = *req.Code
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	status := ""
	if req.Status != nil {
		status = *req.Status
	}
	if err := activity.Update(name, status); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivityInvalidStatus)
	}

	if err := uc.activityRepo.Update(ctx, activity); err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivityNotFound)
	}
	return MapActivityToResponse(activity), nil
}

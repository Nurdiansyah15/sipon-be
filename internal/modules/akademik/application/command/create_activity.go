package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity/constant"
	"sipon-be/internal/modules/akademik/domain/activity/entity"
	repo "sipon-be/internal/modules/akademik/domain/activity/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateActivityUseCase struct {
	activityRepo repo.ActivityRepository
}

func NewCreateActivityUseCase(activityRepo repo.ActivityRepository) *CreateActivityUseCase {
	return &CreateActivityUseCase{activityRepo: activityRepo}
}

func (uc *CreateActivityUseCase) Execute(ctx context.Context, req dto.CreateActivityRequest) (*dto.ActivityResponse, error) {
	existing, err := uc.activityRepo.FindByCode(ctx, req.Code)
	if err != nil && !application.IsNotFoundErr(err, constant.CodeActivityNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	activity, err := entity.NewActivity(uuid.NewString(), req.Code, req.Name)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivityNotFound)
	}
	if err := uc.activityRepo.Save(ctx, activity); err != nil {
		return nil, application.WrapConflictErr(err, constant.CodeActivityDuplicate)
	}
	return MapActivityToResponse(activity), nil
}

func MapActivityToResponse(a *entity.Activity) *dto.ActivityResponse {
	return &dto.ActivityResponse{
		ID:        a.ID,
		Code:      a.Code,
		Name:      a.Name,
		Status:    string(a.Status),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

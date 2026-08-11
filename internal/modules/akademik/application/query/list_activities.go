package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity/repository"
	"sipon-be/internal/shared/kernel"
)

type ListActivitiesUseCase struct {
	activityRepo repository.ActivityRepository
}

func NewListActivitiesUseCase(activityRepo repository.ActivityRepository) *ListActivitiesUseCase {
	return &ListActivitiesUseCase{activityRepo: activityRepo}
}

func (uc *ListActivitiesUseCase) Execute(ctx context.Context, q dto.ActivityListQuery) ([]dto.ActivityResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.activityRepo.List(ctx, repository.ActivityListQuery{
		Status: q.Status,
		Search: q.Search,
		Page:   q.Page,
		Limit:  q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ActivityResponse, len(result.Items))
	for i, a := range result.Items {
		items[i] = *command.MapActivityToResponse(a)
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}

package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

type ListAcademicPeriodsUseCase struct {
	periodRepo repository.AcademicPeriodRepository
}

func NewListAcademicPeriodsUseCase(periodRepo repository.AcademicPeriodRepository) *ListAcademicPeriodsUseCase {
	return &ListAcademicPeriodsUseCase{periodRepo: periodRepo}
}

func (uc *ListAcademicPeriodsUseCase) Execute(ctx context.Context, q dto.AcademicPeriodListQuery) ([]dto.AcademicPeriodResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.periodRepo.List(ctx, repository.AcademicPeriodListQuery{
		Status: q.Status,
		Search: q.Search,
		Page:   q.Page,
		Limit:  q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.AcademicPeriodResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = *command.MapAcademicPeriodToResponse(p)
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}

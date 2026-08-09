package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ListPeriodsUseCase struct {
	periodRepo periodRepo.AccountingPeriodRepository
}

func NewListPeriodsUseCase(periodRepo periodRepo.AccountingPeriodRepository) *ListPeriodsUseCase {
	return &ListPeriodsUseCase{periodRepo: periodRepo}
}

func (uc *ListPeriodsUseCase) Execute(ctx context.Context, query dto.PeriodListQuery) ([]dto.PeriodResponse, *dto.Meta, error) {
	repoQuery := periodRepo.PeriodListQuery{
		Status: query.Status,
		Page:   query.Page,
		Limit:  query.Limit,
	}
	if repoQuery.Page == 0 {
		repoQuery.Page = 1
	}
	if repoQuery.Limit == 0 {
		repoQuery.Limit = 20
	}

	result, err := uc.periodRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.PeriodResponse, len(result.Items))
	for i, p := range result.Items {
		resp := dto.PeriodResponse{
			ID:        p.ID,
			Name:      p.Name,
			StartDate: p.StartDate.Format("2006-01-02"),
			EndDate:   p.EndDate.Format("2006-01-02"),
			Status:    string(p.Status),
			ClosedBy:  p.ClosedBy,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if p.ClosedAt != nil {
			s := p.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
			resp.ClosedAt = &s
		}
		items[i] = resp
	}

	totalPages := (result.Total + int64(repoQuery.Limit) - 1) / int64(repoQuery.Limit)
	meta := &dto.Meta{
		Page:       repoQuery.Page,
		Limit:      repoQuery.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}

	return items, meta, nil
}

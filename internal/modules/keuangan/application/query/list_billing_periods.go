package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	"sipon-be/internal/shared/kernel"
)

type ListBillingPeriodsUseCase struct {
	billingPeriodRepo bpRepo.BillingPeriodRepository
}

func NewListBillingPeriodsUseCase(billingPeriodRepo bpRepo.BillingPeriodRepository) *ListBillingPeriodsUseCase {
	return &ListBillingPeriodsUseCase{billingPeriodRepo: billingPeriodRepo}
}

func (uc *ListBillingPeriodsUseCase) Execute(ctx context.Context, query dto.BillingPeriodListQuery) ([]dto.BillingPeriodResponse, *dto.Meta, error) {
	repoQuery := bpRepo.BillingPeriodListQuery{
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

	result, err := uc.billingPeriodRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.BillingPeriodResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = buildBillingPeriodResponse(p)
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

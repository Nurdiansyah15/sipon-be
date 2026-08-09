package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bbRepo "sipon-be/internal/modules/keuangan/domain/billingbatch/repository"
	"sipon-be/internal/shared/kernel"
)

type ListBillingBatchesUseCase struct {
	batchRepo bbRepo.BillingBatchRepository
}

func NewListBillingBatchesUseCase(batchRepo bbRepo.BillingBatchRepository) *ListBillingBatchesUseCase {
	return &ListBillingBatchesUseCase{batchRepo: batchRepo}
}

func (uc *ListBillingBatchesUseCase) Execute(ctx context.Context, query dto.BillingBatchListQuery) ([]dto.BillingBatchResponse, *dto.Meta, error) {
	repoQuery := bbRepo.BillingBatchListQuery{
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

	result, err := uc.batchRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.BillingBatchResponse, len(result.Items))
	for i, b := range result.Items {
		items[i] = buildBillingBatchResponse(b)
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

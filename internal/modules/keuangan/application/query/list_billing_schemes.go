package query

import (
	"context"

	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type ListBillingSchemesUseCase struct {
	billingSchemeRepo bsRepo.BillingSchemeRepository
}

func NewListBillingSchemesUseCase(billingSchemeRepo bsRepo.BillingSchemeRepository) *ListBillingSchemesUseCase {
	return &ListBillingSchemesUseCase{billingSchemeRepo: billingSchemeRepo}
}

func (uc *ListBillingSchemesUseCase) Execute(ctx context.Context, query dto.BillingSchemeListQuery) ([]dto.BillingSchemeResponse, *dto.Meta, error) {
	repoQuery := bsRepo.BillingSchemeListQuery{
		Active: query.Active,
		Page:   query.Page,
		Limit:  query.Limit,
	}
	if repoQuery.Page == 0 {
		repoQuery.Page = 1
	}
	if repoQuery.Limit == 0 {
		repoQuery.Limit = 20
	}

	result, err := uc.billingSchemeRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeQueryFailed)
	}

	items := make([]dto.BillingSchemeResponse, len(result.Items))
	for i, bs := range result.Items {
		resp := dto.BillingSchemeResponse{
			ID:          bs.ID,
			Name:        bs.Name,
			Description: bs.Description,
			IsActive:    bs.IsActive,
			CreatedAt:   bs.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   bs.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if bs.Items != nil {
			resp.Items = make([]dto.BillingSchemeItemResponse, len(bs.Items))
			for j, item := range bs.Items {
				resp.Items[j] = dto.BillingSchemeItemResponse{
					ID:             item.ID,
					FeeComponentID: item.FeeComponentID,
					AmountOverride: item.AmountOverride,
					IsRequired:     item.IsRequired,
					SortOrder:      item.SortOrder,
				}
			}
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

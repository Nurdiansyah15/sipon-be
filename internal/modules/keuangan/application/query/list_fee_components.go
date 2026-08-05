package query

import (
	"context"

	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type ListFeeComponentsUseCase struct {
	feeComponentRepo feeRepo.FeeComponentRepository
}

func NewListFeeComponentsUseCase(feeComponentRepo feeRepo.FeeComponentRepository) *ListFeeComponentsUseCase {
	return &ListFeeComponentsUseCase{feeComponentRepo: feeComponentRepo}
}

func (uc *ListFeeComponentsUseCase) Execute(ctx context.Context, query dto.FeeComponentListQuery) ([]dto.FeeComponentResponse, *dto.Meta, error) {
	repoQuery := feeRepo.FeeComponentListQuery{
		Type:   query.Type,
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

	result, err := uc.feeComponentRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, application.WrapRepoErr(err, feeConst.CodeFeeComponentQueryFailed)
	}

	items := make([]dto.FeeComponentResponse, len(result.Items))
	for i, fc := range result.Items {
		resp := dto.FeeComponentResponse{
			ID:         fc.ID,
			Code:       fc.Code,
			Name:       fc.Name,
			Type:       string(fc.Type),
			Amount:     fc.Amount,
			IsPeriodic: fc.IsPeriodic,
			IsActive:   fc.IsActive,
			CreatedAt:  fc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  fc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if fc.PeriodType != nil {
			s := string(*fc.PeriodType)
			resp.PeriodType = &s
		}
		if fc.Description != nil {
			resp.Description = fc.Description
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

package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/shared/kernel"
)

type ListFeeComponentsUseCase struct {
	feeComponentRepo feeRepo.FeeComponentRepository
	accountRepo      accRepo.AccountRepository
}

func NewListFeeComponentsUseCase(feeComponentRepo feeRepo.FeeComponentRepository, accountRepo accRepo.AccountRepository) *ListFeeComponentsUseCase {
	return &ListFeeComponentsUseCase{feeComponentRepo: feeComponentRepo, accountRepo: accountRepo}
}

func (uc *ListFeeComponentsUseCase) Execute(ctx context.Context, query dto.FeeComponentListQuery) ([]dto.FeeComponentResponse, *dto.Meta, error) {
	repoQuery := feeRepo.FeeComponentListQuery{
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
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	accounts, err := uc.accountRepo.ListAll(ctx)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	accountByID := make(map[string]*dto.AccountBriefResponse, len(accounts))
	for _, acc := range accounts {
		accountByID[acc.ID] = &dto.AccountBriefResponse{
			ID:      acc.ID,
			Code:    acc.Code,
			Name:    acc.Name,
			Type:    string(acc.Type),
			SubType: subTypeStr(acc.SubType),
		}
	}

	items := make([]dto.FeeComponentResponse, len(result.Items))
	for i, fc := range result.Items {
		resp := dto.FeeComponentResponse{
			ID:         fc.ID,
			Code:       fc.Code,
			Name:       fc.Name,
			Amount:     fc.Amount,
			IsPeriodic: fc.IsPeriodic,
			IsActive:   fc.IsActive,
			CreatedAt:  fc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  fc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if a, ok := accountByID[fc.RevenueAccountID]; ok {
			resp.RevenueAccount = a
		}
		if a, ok := accountByID[fc.ReceivableAccountID]; ok {
			resp.ReceivableAccount = a
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

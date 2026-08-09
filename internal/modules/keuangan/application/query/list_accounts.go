package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/shared/kernel"
)

type ListAccountsUseCase struct {
	accountRepo accRepo.AccountRepository
}

func NewListAccountsUseCase(accountRepo accRepo.AccountRepository) *ListAccountsUseCase {
	return &ListAccountsUseCase{accountRepo: accountRepo}
}

func (uc *ListAccountsUseCase) Execute(ctx context.Context, query dto.AccountListQuery) ([]dto.AccountResponse, *dto.Meta, error) {
	repoQuery := accRepo.AccountListQuery{
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

	result, err := uc.accountRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.AccountResponse, len(result.Items))
	for i, acc := range result.Items {
		items[i] = dto.AccountResponse{
			ID:            acc.ID,
			Code:          acc.Code,
			Name:          acc.Name,
			Type:          string(acc.Type),
			ParentID:      acc.ParentID,
			Level:         acc.Level,
			IsPostable:    acc.IsPostable,
			NormalBalance: string(acc.NormalBalance),
			Description:   acc.Description,
			IsActive:      acc.IsActive,
			IsSystem:      acc.IsSystem,
			CreatedAt:     acc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     acc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
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

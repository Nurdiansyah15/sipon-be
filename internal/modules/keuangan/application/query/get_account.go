package query

import (
	"context"

	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type GetAccountUseCase struct {
	accountRepo accRepo.AccountRepository
}

func NewGetAccountUseCase(accountRepo accRepo.AccountRepository) *GetAccountUseCase {
	return &GetAccountUseCase{accountRepo: accountRepo}
}

func (uc *GetAccountUseCase) Execute(ctx context.Context, id string) (*dto.AccountResponse, error) {
	acc, err := uc.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
	}

	return &dto.AccountResponse{
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
	}, nil
}

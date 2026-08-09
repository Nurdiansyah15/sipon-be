package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/shared/kernel"
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
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case accConst.CodeAccountNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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

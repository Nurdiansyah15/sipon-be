package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateAccountUseCase struct {
	accountRepo accRepo.AccountRepository
}

func NewUpdateAccountUseCase(accountRepo accRepo.AccountRepository) *UpdateAccountUseCase {
	return &UpdateAccountUseCase{accountRepo: accountRepo}
}

func (uc *UpdateAccountUseCase) Execute(ctx context.Context, id string, req dto.UpdateAccountRequest) (*dto.AccountResponse, error) {
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

	if err := acc.Update(req.Name, req.Description, req.IsPostable); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case accConst.CodeAccountIsSystem:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.accountRepo.Update(ctx, acc); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case accConst.CodeAccountNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toAccountResponse(acc), nil
}

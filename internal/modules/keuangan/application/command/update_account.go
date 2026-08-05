package command

import (
	"context"

	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
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
		return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
	}

	if err := acc.Update(req.Name, req.Description, req.IsPostable); err != nil {
		return nil, application.WrapRepoErr(err, accConst.CodeAccountIsSystem)
	}

	if err := uc.accountRepo.Update(ctx, acc); err != nil {
		return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
	}

	return toAccountResponse(acc), nil
}

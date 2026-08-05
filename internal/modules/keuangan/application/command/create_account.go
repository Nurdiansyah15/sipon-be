package command

import (
	"context"

	"github.com/google/uuid"

	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accEntity "sipon-be/internal/modules/keuangan/domain/account/entity"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/shared/kernel"
)

type CreateAccountUseCase struct {
	accountRepo accRepo.AccountRepository
}

func NewCreateAccountUseCase(accountRepo accRepo.AccountRepository) *CreateAccountUseCase {
	return &CreateAccountUseCase{accountRepo: accountRepo}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, req dto.CreateAccountRequest, createdBy string) (*dto.AccountResponse, error) {
	exists, err := uc.accountRepo.ExistsByCode(ctx, req.Code, "")
	if err != nil {
		return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
	}
	if exists {
		return nil, kernel.New(accConst.CodeAccountDuplicate)
	}

	if req.ParentID != nil && *req.ParentID != "" {
		parent, err := uc.accountRepo.FindByID(ctx, *req.ParentID)
		if err != nil {
			return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
		}
		level := parent.Level + 1
		accType := accConst.AccountType(req.Type)
		normalBalance := accConst.NormalBalance(req.NormalBalance)
		acc, err := accEntity.NewAccount(uuid.New().String(), req.Code, req.Name, accType, req.ParentID, level, normalBalance, createdBy)
		if err != nil {
			return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
		}
		acc.Description = req.Description
		acc.IsPostable = req.IsPostable
		if err := uc.accountRepo.Save(ctx, acc); err != nil {
			return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
		}
		return toAccountResponse(acc), nil
	}

	accType := accConst.AccountType(req.Type)
	normalBalance := accConst.NormalBalance(req.NormalBalance)
	acc, err := accEntity.NewAccount(uuid.New().String(), req.Code, req.Name, accType, req.ParentID, 1, normalBalance, createdBy)
	if err != nil {
		return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
	}
	acc.Description = req.Description
	acc.IsPostable = req.IsPostable
	if err := uc.accountRepo.Save(ctx, acc); err != nil {
		return nil, application.WrapRepoErr(err, accConst.CodeAccountNotFound)
	}
	return toAccountResponse(acc), nil
}

func toAccountResponse(acc *accEntity.Account) *dto.AccountResponse {
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
	}
}

package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateBillingSchemeUseCase struct {
	billingSchemeRepo bsRepo.BillingSchemeRepository
}

func NewUpdateBillingSchemeUseCase(billingSchemeRepo bsRepo.BillingSchemeRepository) *UpdateBillingSchemeUseCase {
	return &UpdateBillingSchemeUseCase{billingSchemeRepo: billingSchemeRepo}
}

func (uc *UpdateBillingSchemeUseCase) Execute(ctx context.Context, id string, req dto.UpdateBillingSchemeRequest) (*dto.BillingSchemeResponse, error) {
	scheme, err := uc.billingSchemeRepo.FindByID(ctx, id)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeBillingSchemeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	scheme.Update(req.Name, req.Description)

	if err := uc.billingSchemeRepo.Update(ctx, scheme); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeBillingSchemeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toBillingSchemeResponse(scheme), nil
}

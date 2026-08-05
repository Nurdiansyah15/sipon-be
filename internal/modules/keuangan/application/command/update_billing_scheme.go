package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
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
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	scheme.Update(req.Name, req.Description)

	if err := uc.billingSchemeRepo.Update(ctx, scheme); err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	return toBillingSchemeResponse(scheme), nil
}

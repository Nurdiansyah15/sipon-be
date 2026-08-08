package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
)

type GetBillingPeriodUseCase struct {
	billingPeriodRepo bpRepo.BillingPeriodRepository
}

func NewGetBillingPeriodUseCase(billingPeriodRepo bpRepo.BillingPeriodRepository) *GetBillingPeriodUseCase {
	return &GetBillingPeriodUseCase{billingPeriodRepo: billingPeriodRepo}
}

func (uc *GetBillingPeriodUseCase) Execute(ctx context.Context, id string) (*dto.BillingPeriodResponse, error) {
	period, err := uc.billingPeriodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, bpConst.CodeBillingPeriodNotFound)
	}
	resp := buildBillingPeriodResponse(period)
	return &resp, nil
}

package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
)

type CloseBillingPeriodUseCase struct {
	billingPeriodRepo bpRepo.BillingPeriodRepository
}

func NewCloseBillingPeriodUseCase(billingPeriodRepo bpRepo.BillingPeriodRepository) *CloseBillingPeriodUseCase {
	return &CloseBillingPeriodUseCase{billingPeriodRepo: billingPeriodRepo}
}

func (uc *CloseBillingPeriodUseCase) Execute(ctx context.Context, periodID string) (*dto.BillingPeriodResponse, error) {
	period, err := uc.billingPeriodRepo.FindByID(ctx, periodID)
	if err != nil {
		return nil, application.WrapRepoErr(err, bpConst.CodeBillingPeriodNotFound)
	}

	if err := period.Close(); err != nil {
		return nil, application.WrapRepoErr(err, bpConst.CodeBillingPeriodInvalidStatus)
	}

	if err := uc.billingPeriodRepo.Update(ctx, period); err != nil {
		return nil, application.WrapRepoErr(err, bpConst.CodeBillingPeriodNotFound)
	}

	return toBillingPeriodResponse(period), nil
}

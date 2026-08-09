package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	"sipon-be/internal/shared/kernel"
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
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	resp := buildBillingPeriodResponse(period)
	return &resp, nil
}

package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	"sipon-be/internal/shared/kernel"
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
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := period.Close(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.billingPeriodRepo.Update(ctx, period); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toBillingPeriodResponse(period), nil
}

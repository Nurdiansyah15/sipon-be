package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReopenPeriodUseCase struct {
	periodRepo periodRepo.AccountingPeriodRepository
}

func NewReopenPeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository) *ReopenPeriodUseCase {
	return &ReopenPeriodUseCase{periodRepo: periodRepo}
}

func (uc *ReopenPeriodUseCase) Execute(ctx context.Context, periodID string) (*dto.PeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, periodID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := period.Reopen(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toPeriodResponse(period), nil
}

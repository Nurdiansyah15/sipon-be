package command

import (
	"context"

	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type LockPeriodUseCase struct {
	periodRepo periodRepo.AccountingPeriodRepository
}

func NewLockPeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository) *LockPeriodUseCase {
	return &LockPeriodUseCase{periodRepo: periodRepo}
}

func (uc *LockPeriodUseCase) Execute(ctx context.Context, periodID string, closedBy string) (*dto.PeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, periodID)
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	if err := period.Lock(closedBy); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodInvalidStatus)
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	return toPeriodResponse(period), nil
}

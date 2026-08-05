package command

import (
	"context"

	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
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
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	if err := period.Reopen(); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodInvalidStatus)
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	return toPeriodResponse(period), nil
}

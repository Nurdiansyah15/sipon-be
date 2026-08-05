package command

import (
	"context"

	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type ClosePeriodUseCase struct {
	periodRepo periodRepo.AccountingPeriodRepository
}

func NewClosePeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository) *ClosePeriodUseCase {
	return &ClosePeriodUseCase{periodRepo: periodRepo}
}

func (uc *ClosePeriodUseCase) Execute(ctx context.Context, periodID string, closedBy string) (*dto.PeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, periodID)
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	if err := period.Close(closedBy); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodInvalidStatus)
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	return toPeriodResponse(period), nil
}

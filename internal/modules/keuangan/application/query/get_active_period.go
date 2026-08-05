package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
)

type GetActivePeriodUseCase struct {
	periodRepo periodRepo.AccountingPeriodRepository
}

func NewGetActivePeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository) *GetActivePeriodUseCase {
	return &GetActivePeriodUseCase{periodRepo: periodRepo}
}

func (uc *GetActivePeriodUseCase) Execute(ctx context.Context) (*dto.PeriodResponse, error) {
	p, err := uc.periodRepo.FindActive(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	resp := &dto.PeriodResponse{
		ID:        p.ID,
		Name:      p.Name,
		StartDate: p.StartDate.Format("2006-01-02"),
		EndDate:   p.EndDate.Format("2006-01-02"),
		Status:    string(p.Status),
		ClosedBy:  p.ClosedBy,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.ClosedAt != nil {
		s := p.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ClosedAt = &s
	}
	return resp, nil
}

package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
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
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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

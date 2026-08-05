package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodEntity "sipon-be/internal/modules/keuangan/domain/period/entity"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/shared/kernel"
)

type CreatePeriodUseCase struct {
	periodRepo periodRepo.AccountingPeriodRepository
}

func NewCreatePeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository) *CreatePeriodUseCase {
	return &CreatePeriodUseCase{periodRepo: periodRepo}
}

func (uc *CreatePeriodUseCase) Execute(ctx context.Context, req dto.CreatePeriodRequest, createdBy string) (*dto.PeriodResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	hasOverlap, err := uc.periodRepo.HasOverlap(ctx, startDate, endDate, "")
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}
	if hasOverlap {
		return nil, kernel.New(periodConst.CodePeriodOverlap)
	}

	period, err := periodEntity.NewAccountingPeriod(uuid.New().String(), req.Name, startDate, endDate, createdBy)
	if err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	if err := uc.periodRepo.Save(ctx, period); err != nil {
		return nil, application.WrapRepoErr(err, periodConst.CodePeriodNotFound)
	}

	return toPeriodResponse(period), nil
}

func toPeriodResponse(p *periodEntity.AccountingPeriod) *dto.PeriodResponse {
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
	return resp
}

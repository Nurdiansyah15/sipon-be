package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodEntity "sipon-be/internal/modules/keuangan/domain/period/entity"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
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
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	hasOverlap, err := uc.periodRepo.HasOverlap(ctx, startDate, endDate, "")
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if hasOverlap {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Periode akuntansi saling tumpang tindih", nil)
	}

	period, err := periodEntity.NewAccountingPeriod(uuid.New().String(), req.Name, startDate, endDate, createdBy)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.periodRepo.Save(ctx, period); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodOverlap:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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

package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpEntity "sipon-be/internal/modules/keuangan/domain/billingperiod/entity"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateBillingPeriodUseCase struct {
	billingPeriodRepo bpRepo.BillingPeriodRepository
	periodRepo        periodRepo.AccountingPeriodRepository
}

func NewCreateBillingPeriodUseCase(billingPeriodRepo bpRepo.BillingPeriodRepository, periodRepo periodRepo.AccountingPeriodRepository) *CreateBillingPeriodUseCase {
	return &CreateBillingPeriodUseCase{billingPeriodRepo: billingPeriodRepo, periodRepo: periodRepo}
}

func (uc *CreateBillingPeriodUseCase) Execute(ctx context.Context, req dto.CreateBillingPeriodRequest, createdBy string) (*dto.BillingPeriodResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	// Validasi periode akuntansi: harus ada. Validasi rentang tanggal periode
	// tagihan berada di dalam periode akuntansi dilakukan di domain entity.
	accountingPeriod, err := uc.periodRepo.FindByID(ctx, req.AccountingPeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Periode akuntansi tidak ditemukan", err)
	}

	period, err := bpEntity.NewBillingPeriod(
		uuid.New().String(), req.Name, req.AccountingPeriodID, feeConst.PeriodType(req.PeriodType),
		startDate, endDate, accountingPeriod.StartDate, accountingPeriod.EndDate, createdBy,
	)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodNotFound, bpConst.CodeBillingPeriodInvalidDateRange:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.billingPeriodRepo.Save(ctx, period); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toBillingPeriodResponse(period), nil
}

func toBillingPeriodResponse(p *bpEntity.BillingPeriod) *dto.BillingPeriodResponse {
	return &dto.BillingPeriodResponse{
		ID:                 p.ID,
		Name:               p.Name,
		PeriodType:         string(p.PeriodType),
		AccountingPeriodID: p.AccountingPeriodID,
		StartDate:          p.StartDate.Format("2006-01-02"),
		EndDate:            p.EndDate.Format("2006-01-02"),
		Status:             string(p.Status),
		CreatedAt:          p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

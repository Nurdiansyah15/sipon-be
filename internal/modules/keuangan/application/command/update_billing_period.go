package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bpConst "sipon-be/internal/modules/keuangan/domain/billingperiod/constant"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateBillingPeriodUseCase struct {
	billingPeriodRepo bpRepo.BillingPeriodRepository
	periodRepo        periodRepo.AccountingPeriodRepository
}

func NewUpdateBillingPeriodUseCase(billingPeriodRepo bpRepo.BillingPeriodRepository, periodRepo periodRepo.AccountingPeriodRepository) *UpdateBillingPeriodUseCase {
	return &UpdateBillingPeriodUseCase{billingPeriodRepo: billingPeriodRepo, periodRepo: periodRepo}
}

func (uc *UpdateBillingPeriodUseCase) Execute(ctx context.Context, periodID string, req dto.UpdateBillingPeriodRequest) (*dto.BillingPeriodResponse, error) {
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

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	// Periode akuntansi induk tidak berubah saat edit; hanya dipakai untuk
	// memvalidasi rentang tanggal baru periode tagihan.
	accountingPeriod, err := uc.periodRepo.FindByID(ctx, period.AccountingPeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Periode akuntansi tidak ditemukan", err)
	}

	if err := period.Update(req.Name, feeConst.PeriodType(req.PeriodType), startDate, endDate, accountingPeriod.StartDate, accountingPeriod.EndDate); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bpConst.CodeBillingPeriodInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case bpConst.CodeBillingPeriodInvalidDateRange, bpConst.CodeBillingPeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
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
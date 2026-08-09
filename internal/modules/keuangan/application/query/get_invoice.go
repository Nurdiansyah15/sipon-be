package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	adjRepo "sipon-be/internal/modules/keuangan/domain/adjustment/repository"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

type GetInvoiceUseCase struct {
	invoiceRepo       invRepo.InvoiceRepository
	feeComponentRepo  feeRepo.FeeComponentRepository
	billingPeriodRepo bpRepo.BillingPeriodRepository
	paymentRepo       payRepo.PaymentRepository
	adjustmentRepo    adjRepo.AdjustmentRepository
}

func NewGetInvoiceUseCase(
	invoiceRepo invRepo.InvoiceRepository,
	feeComponentRepo feeRepo.FeeComponentRepository,
	billingPeriodRepo bpRepo.BillingPeriodRepository,
	paymentRepo payRepo.PaymentRepository,
	adjustmentRepo adjRepo.AdjustmentRepository,
) *GetInvoiceUseCase {
	return &GetInvoiceUseCase{
		invoiceRepo:       invoiceRepo,
		feeComponentRepo:  feeComponentRepo,
		billingPeriodRepo: billingPeriodRepo,
		paymentRepo:       paymentRepo,
		adjustmentRepo:    adjustmentRepo,
	}
}

func (uc *GetInvoiceUseCase) Execute(ctx context.Context, id string) (*dto.InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case invConst.CodeInvoiceNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	resp := buildInvoiceResponse(ctx, inv, uc.feeComponentRepo, uc.billingPeriodRepo)

	if payments, err := uc.paymentRepo.FindByInvoiceID(ctx, inv.ID); err == nil {
		resp.Payments = make([]dto.PaymentResponse, len(payments))
		for i, p := range payments {
			resp.Payments[i] = buildPaymentResponse(p)
		}
	}

	if adjustments, err := uc.adjustmentRepo.FindByInvoiceID(ctx, inv.ID); err == nil {
		resp.Adjustments = make([]dto.InvoiceAdjustmentResponse, len(adjustments))
		for i, a := range adjustments {
			resp.Adjustments[i] = buildAdjustmentResponse(a)
		}
	}

	return &resp, nil
}

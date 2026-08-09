package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

type GetPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
	accountRepo accRepo.AccountRepository
}

func NewGetPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, accountRepo accRepo.AccountRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, accountRepo: accountRepo}
}

func (uc *GetPaymentUseCase) Execute(ctx context.Context, id string) (*dto.PaymentResponse, error) {
	p, err := uc.paymentRepo.FindByID(ctx, id)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	resp := buildPaymentResponse(p)
	if inv, err := uc.invoiceRepo.FindByID(ctx, p.InvoiceID); err == nil {
		invResp := buildInvoiceResponse(ctx, inv, nil, nil)
		resp.Invoice = &invResp
	}
	if p.DebitAccountID != nil {
		if acc, err := uc.accountRepo.FindByID(ctx, *p.DebitAccountID); err == nil {
			resp.DebitAccount = &dto.AccountBriefResponse{ID: acc.ID, Code: acc.Code, Name: acc.Name, Type: string(acc.Type)}
		}
	}
	return &resp, nil
}

package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

type ListPaymentsUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
	accountRepo accRepo.AccountRepository
}

func NewListPaymentsUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, accountRepo accRepo.AccountRepository) *ListPaymentsUseCase {
	return &ListPaymentsUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, accountRepo: accountRepo}
}

func (uc *ListPaymentsUseCase) Execute(ctx context.Context, query dto.PaymentListQuery) ([]dto.PaymentResponse, *dto.Meta, error) {
	repoQuery := payRepo.PaymentListQuery{
		InvoiceID: query.InvoiceID,
		Status:    query.Status,
		PeriodID:  query.PeriodID,
		Page:      query.Page,
		Limit:     query.Limit,
	}
	if repoQuery.Page == 0 {
		repoQuery.Page = 1
	}
	if repoQuery.Limit == 0 {
		repoQuery.Limit = 20
	}

	result, err := uc.paymentRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.PaymentResponse, len(result.Items))
	for i, p := range result.Items {
		resp := buildPaymentResponse(p)
		if inv, err := uc.invoiceRepo.FindByID(ctx, p.InvoiceID); err == nil {
			invResp := buildInvoiceResponse(ctx, inv, nil, nil)
			resp.Invoice = &invResp
		}
		if p.DebitAccountID != nil {
			if acc, err := uc.accountRepo.FindByID(ctx, *p.DebitAccountID); err == nil {
				resp.DebitAccount = &dto.AccountBriefResponse{ID: acc.ID, Code: acc.Code, Name: acc.Name, Type: string(acc.Type), SubType: subTypeStr(acc.SubType)}
			}
		}
		items[i] = resp
	}

	totalPages := (result.Total + int64(repoQuery.Limit) - 1) / int64(repoQuery.Limit)
	meta := &dto.Meta{
		Page:       repoQuery.Page,
		Limit:      repoQuery.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}

	return items, meta, nil
}

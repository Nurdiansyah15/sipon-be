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

type MyPaymentsUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
	accountRepo accRepo.AccountRepository
}

func NewMyPaymentsUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, accountRepo accRepo.AccountRepository) *MyPaymentsUseCase {
	return &MyPaymentsUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, accountRepo: accountRepo}
}

func (uc *MyPaymentsUseCase) Execute(ctx context.Context, userID string, query dto.PaymentListQuery) ([]dto.PaymentResponse, *dto.Meta, error) {
	page := query.Page
	limit := query.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}

	invListQuery := invRepo.InvoiceListQuery{
		UserID: &userID,
		Page:   1,
		Limit:  10000,
	}
	invResult, err := uc.invoiceRepo.List(ctx, invListQuery)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if len(invResult.Items) == 0 {
		meta := &dto.Meta{
			Page:       page,
			Limit:      limit,
			Total:      0,
			TotalPages: 0,
		}
		return []dto.PaymentResponse{}, meta, nil
	}

	var allPayments []dto.PaymentResponse
	for _, inv := range invResult.Items {
		repoQuery := payRepo.PaymentListQuery{
			InvoiceID: &inv.ID,
			Status:    query.Status,
			Page:      1,
			Limit:     10000,
		}
		payResult, err := uc.paymentRepo.List(ctx, repoQuery)
		if err != nil {
			return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		invResp := buildInvoiceResponse(ctx, inv, nil, nil)
		for _, p := range payResult.Items {
			resp := buildPaymentResponse(p)
			resp.Invoice = &invResp
			if p.DebitAccountID != nil {
				if acc, err := uc.accountRepo.FindByID(ctx, *p.DebitAccountID); err == nil {
					resp.DebitAccount = &dto.AccountBriefResponse{ID: acc.ID, Code: acc.Code, Name: acc.Name, Type: string(acc.Type), SubType: subTypeStr(acc.SubType)}
				}
			}
			allPayments = append(allPayments, resp)
		}
	}

	total := int64(len(allPayments))
	start := (page - 1) * limit
	end := start + limit
	if start > len(allPayments) {
		start = len(allPayments)
	}
	if end > len(allPayments) {
		end = len(allPayments)
	}
	paginatedPayments := allPayments[start:end]

	totalPages := (total + int64(limit) - 1) / int64(limit)
	meta := &dto.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return paginatedPayments, meta, nil
}

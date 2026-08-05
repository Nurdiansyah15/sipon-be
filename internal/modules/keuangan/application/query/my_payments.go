package query

import (
	"context"

	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type MyPaymentsUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
}

func NewMyPaymentsUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository) *MyPaymentsUseCase {
	return &MyPaymentsUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo}
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
		return nil, nil, application.WrapRepoErr(err, payConst.CodePaymentQueryFailed)
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
			return nil, nil, application.WrapRepoErr(err, payConst.CodePaymentQueryFailed)
		}
		for _, p := range payResult.Items {
			resp := dto.PaymentResponse{
				ID:              p.ID,
				PaymentNumber:   p.PaymentNumber,
				InvoiceID:       p.InvoiceID,
				DebitAccountID:  p.DebitAccountID,
				Amount:          p.Amount,
				Method:          string(p.Method),
				ReferenceNumber: p.ReferenceNumber,
				PaymentDate:     p.PaymentDate.Format("2006-01-02"),
				Status:          string(p.Status),
				VerifiedBy:      p.VerifiedBy,
				Notes:           p.Notes,
				ProofKey:        p.ProofKey,
				CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			if p.VerifiedAt != nil {
				s := p.VerifiedAt.Format("2006-01-02T15:04:05Z07:00")
				resp.VerifiedAt = &s
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

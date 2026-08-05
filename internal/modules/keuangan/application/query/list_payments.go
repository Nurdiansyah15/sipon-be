package query

import (
	"context"

	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type ListPaymentsUseCase struct {
	paymentRepo payRepo.PaymentRepository
}

func NewListPaymentsUseCase(paymentRepo payRepo.PaymentRepository) *ListPaymentsUseCase {
	return &ListPaymentsUseCase{paymentRepo: paymentRepo}
}

func (uc *ListPaymentsUseCase) Execute(ctx context.Context, query dto.PaymentListQuery) ([]dto.PaymentResponse, *dto.Meta, error) {
	repoQuery := payRepo.PaymentListQuery{
		InvoiceID: query.InvoiceID,
		Status:    query.Status,
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
		return nil, nil, application.WrapRepoErr(err, payConst.CodePaymentQueryFailed)
	}

	items := make([]dto.PaymentResponse, len(result.Items))
	for i, p := range result.Items {
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

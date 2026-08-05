package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
)

type GetPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
}

func NewGetPaymentUseCase(paymentRepo payRepo.PaymentRepository) *GetPaymentUseCase {
	return &GetPaymentUseCase{paymentRepo: paymentRepo}
}

func (uc *GetPaymentUseCase) Execute(ctx context.Context, id string) (*dto.PaymentResponse, error) {
	p, err := uc.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	resp := &dto.PaymentResponse{
		ID:              p.ID,
		PaymentNumber:   p.PaymentNumber,
		InvoiceID:       p.InvoiceID,
		DebitAccountID:  p.DebitAccountID,
		Amount:          p.Amount,
		Method:          string(p.Method),
		ReferenceNumber: p.ReferenceNumber,
		PaymentDate:     p.PaymentDate.Format("2006-01-02"),
		Status:          string(p.Status),
		Notes:           p.Notes,
		ProofKey:        p.ProofKey,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.VerifiedBy != nil {
		resp.VerifiedBy = p.VerifiedBy
	}
	if p.VerifiedAt != nil {
		s := p.VerifiedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.VerifiedAt = &s
	}
	return resp, nil
}

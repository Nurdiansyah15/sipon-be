package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	payVO "sipon-be/internal/modules/keuangan/domain/payment/valueobject"
)

type CreateManualPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
}

func NewCreateManualPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository) *CreateManualPaymentUseCase {
	return &CreateManualPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo}
}

func (uc *CreateManualPaymentUseCase) Execute(ctx context.Context, req dto.CreateManualPaymentRequest, createdBy string) (*dto.PaymentResponse, error) {
	_, err := uc.invoiceRepo.FindByID(ctx, req.InvoiceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	method := payConst.PaymentMethod(req.Method)
	payNum := payVO.NewPaymentNumberNow(1)
	payment, err := payEntity.NewPayment(
		uuid.New().String(), payNum.String(), req.InvoiceID,
		req.Amount, method, paymentDate,
		req.DebitAccountID, req.ReferenceNumber, req.Notes, req.ProofKey,
		createdBy,
	)
	if err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	if err := uc.paymentRepo.Save(ctx, payment); err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	return toPaymentResponse(payment), nil
}

func toPaymentResponse(p *payEntity.Payment) *dto.PaymentResponse {
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
	return resp
}

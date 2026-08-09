package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

type CreateManualPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
	accountRepo accRepo.AccountRepository
}

func NewCreateManualPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, accountRepo accRepo.AccountRepository) *CreateManualPaymentUseCase {
	return &CreateManualPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, accountRepo: accountRepo}
}

func (uc *CreateManualPaymentUseCase) Execute(ctx context.Context, req dto.CreateManualPaymentRequest, createdBy string) (*dto.PaymentResponse, error) {
	_, err := uc.invoiceRepo.FindByID(ctx, req.InvoiceID)
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

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	debitAcc, err := uc.accountRepo.FindByID(ctx, req.DebitAccountID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case accConst.CodeAccountNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if err := debitAcc.EnsurePostable(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case accConst.CodeAccountNotPostable:
				return nil, kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun tidak dapat diposting", err)
	}
	if debitAcc.SubType == nil || *debitAcc.SubType != accConst.SubTypeCashBank {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun debit pembayaran harus merupakan akun kas atau bank", nil)
	}

	method := payConst.PaymentMethod(req.Method)
	payNum, err := uc.paymentRepo.NextPaymentNumber(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	debitAccountID := req.DebitAccountID
	payment, err := payEntity.NewPayment(
		uuid.New().String(), payNum.String(), req.InvoiceID,
		req.Amount, method, paymentDate,
		&debitAccountID, req.ReferenceNumber, req.Notes, req.ProofKey,
		createdBy,
	)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.paymentRepo.Save(ctx, payment); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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

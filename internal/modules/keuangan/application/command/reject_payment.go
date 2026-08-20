package command

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

type RejectPaymentUseCase struct {
	paymentRepo  payRepo.PaymentRepository
	invoiceRepo  invRepo.InvoiceRepository
	outboxWriter ports.OutboxWriter
}

func NewRejectPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository) *RejectPaymentUseCase {
	return &RejectPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo}
}

func (uc *RejectPaymentUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *RejectPaymentUseCase) publishPaymentRejected(ctx context.Context, userID, invoiceID string) {
	if uc.outboxWriter == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"user_id":    userID,
		"invoice_id": invoiceID,
	})
	if err := uc.outboxWriter.Save(ctx, RoutingPaymentRejected, payload); err != nil {
		slog.Warn("keuangan: gagal publish event", "routing_key", RoutingPaymentRejected, "invoice_id", invoiceID, "error", err)
	}
}

func (uc *RejectPaymentUseCase) Execute(ctx context.Context, paymentID string) (*dto.PaymentResponse, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, paymentID)
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

	if err := payment.Reject(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	inv, err := uc.invoiceRepo.FindByID(ctx, payment.InvoiceID)
	if err == nil {
		uc.publishPaymentRejected(ctx, inv.UserID, inv.ID)
	}

	return toPaymentResponse(payment), nil
}

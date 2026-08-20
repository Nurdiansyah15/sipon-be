package command

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

type SubmitPaymentUseCase struct {
	paymentRepo  payRepo.PaymentRepository
	invoiceRepo  invRepo.InvoiceRepository
	outboxWriter ports.OutboxWriter
}

func NewSubmitPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository) *SubmitPaymentUseCase {
	return &SubmitPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo}
}

func (uc *SubmitPaymentUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *SubmitPaymentUseCase) publishPaymentSubmitted(ctx context.Context, userID, invoiceID string) {
	if uc.outboxWriter == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"user_id":    userID,
		"invoice_id": invoiceID,
	})
	if err := uc.outboxWriter.Save(ctx, RoutingPaymentSubmitted, payload); err != nil {
		slog.Warn("keuangan: gagal publish event", "routing_key", RoutingPaymentSubmitted, "invoice_id", invoiceID, "error", err)
	}
}

func (uc *SubmitPaymentUseCase) Execute(ctx context.Context, userID string, req dto.SubmitPaymentRequest) (*dto.PaymentResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, req.InvoiceID)
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

	if inv.UserID != userID {
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Invoice bukan milik Anda", nil)
	}

	if inv.Status != invConst.StatusIssued && inv.Status != invConst.StatusPartial {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Invoice tidak dalam status yang dapat dibayar", nil)
	}

	outstanding := inv.RemainingAmount()
	if req.Amount > outstanding {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Nominal pembayaran melebihi sisa tagihan", nil)
	}

	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal tidak valid", err)
	}

	payNum, err := uc.paymentRepo.NextPaymentNumber(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	method := payConst.PaymentMethod(req.Method)
	proofKey := req.ProofKey
	payment, err := payEntity.NewPayment(
		uuid.New().String(), payNum.String(), req.InvoiceID,
		req.Amount, method, paymentDate,
		nil, req.ReferenceNumber, req.Notes, &proofKey,
		userID,
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

	uc.publishPaymentSubmitted(ctx, userID, req.InvoiceID)

	return toPaymentResponse(payment), nil
}

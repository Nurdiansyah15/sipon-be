package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	pgConst "sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	pgEntity "sipon-be/internal/modules/keuangan/domain/paymentgateway/entity"
	pgRepo "sipon-be/internal/modules/keuangan/domain/paymentgateway/repository"
	"sipon-be/internal/shared/kernel"
)

type GetPaymentGatewayStatusUseCase struct {
	paymentGatewayRepo pgRepo.PaymentGatewayRepository
	invoiceRepo        invRepo.InvoiceRepository
}

func NewGetPaymentGatewayStatusUseCase(
	paymentGatewayRepo pgRepo.PaymentGatewayRepository,
	invoiceRepo invRepo.InvoiceRepository,
) *GetPaymentGatewayStatusUseCase {
	return &GetPaymentGatewayStatusUseCase{
		paymentGatewayRepo: paymentGatewayRepo,
		invoiceRepo:        invoiceRepo,
	}
}

func (uc *GetPaymentGatewayStatusUseCase) Execute(ctx context.Context, invoiceID, userID string) (*dto.PaymentGatewayStatusResponse, error) {
	if userID == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "autentikasi diperlukan", nil)
	}

	inv, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
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
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Invoice bukan milik user yang terautentikasi", nil)
	}

	txs, err := uc.paymentGatewayRepo.FindByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	// Transaksi terbaru menang.
	var latest *pgEntity.PaymentGatewayTransaction
	for _, tx := range txs {
		if latest == nil || tx.CreatedAt.After(latest.CreatedAt) {
			latest = tx
		}
	}

	return buildPaymentGatewayStatusResponse(latest), nil
}

func buildPaymentGatewayStatusResponse(tx *pgEntity.PaymentGatewayTransaction) *dto.PaymentGatewayStatusResponse {
	if tx == nil {
		return &dto.PaymentGatewayStatusResponse{
			Status: string(pgConst.GatewayStatusPending),
		}
	}
	return &dto.PaymentGatewayStatusResponse{
		TransactionID: tx.TransactionID,
		InvoiceID:     tx.InvoiceID,
		PaymentID:     tx.PaymentID,
		Amount:        tx.Amount,
		Status:        string(tx.Status),
		PaymentMethod: tx.PaymentMethod,
		SnapToken:     tx.SnapToken,
		RedirectURL:   tx.RedirectURL,
		ExpiresAt:     tx.ExpiredAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

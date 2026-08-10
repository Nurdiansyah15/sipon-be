package command

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	pgConst "sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	pgEntity "sipon-be/internal/modules/keuangan/domain/paymentgateway/entity"
	pgRepo "sipon-be/internal/modules/keuangan/domain/paymentgateway/repository"
	pgVO "sipon-be/internal/modules/keuangan/domain/paymentgateway/valueobject"
	"sipon-be/internal/shared/kernel"
)

type CreateMidtransPaymentUseCase struct {
	paymentGatewayRepo pgRepo.PaymentGatewayRepository
	invoiceRepo        invRepo.InvoiceRepository
	feeRepo            feeRepo.FeeComponentRepository
	midtransGateway    ports.MidtransGateway
	expiryMinutes      int
}

func NewCreateMidtransPaymentUseCase(
	paymentGatewayRepo pgRepo.PaymentGatewayRepository,
	invoiceRepo invRepo.InvoiceRepository,
	feeRepo feeRepo.FeeComponentRepository,
	midtransGateway ports.MidtransGateway,
	expiryMinutes int,
) *CreateMidtransPaymentUseCase {
	return &CreateMidtransPaymentUseCase{
		paymentGatewayRepo: paymentGatewayRepo,
		invoiceRepo:        invoiceRepo,
		feeRepo:            feeRepo,
		midtransGateway:    midtransGateway,
		expiryMinutes:      expiryMinutes,
	}
}

func (uc *CreateMidtransPaymentUseCase) Execute(ctx context.Context, req dto.CreateMidtransPaymentRequest, userID string) (*dto.MidtransPaymentResponse, error) {
	if userID == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "autentikasi diperlukan", nil)
	}

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

	// Hanya pemilik invoice yang boleh membayar online.
	if inv.UserID != userID {
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Invoice bukan milik user yang terautentikasi", nil)
	}

	// Invoice yang dapat dibayar hanya yang berstatus issued/partial.
	if inv.Status != invConst.StatusIssued && inv.Status != invConst.StatusPartial {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Invoice tidak dalam status yang dapat dibayar online", nil)
	}

	// Idempotensi: bila transaksi aktif sudah ada untuk invoice ini, kembalikan
	// transaksi tersebut agar snap token yang sama tetap dipakai.
	existing, err := uc.paymentGatewayRepo.FindActiveByInvoiceID(ctx, inv.ID)
	if err == nil && existing != nil {
		if existing.Status == pgConst.GatewayStatusCapture || existing.Status == pgConst.GatewayStatusSettlement {
			return nil, kernel.WrapMsg(application.ErrCodeConflict, "Invoice sudah dibayar", nil)
		}
		return buildMidtransPaymentResponse(existing), nil
	}
	if err != nil {
		var ke *kernel.AppError
		if !errors.As(err, &ke) || ke.Code != pgConst.CodePaymentGatewayNotFound {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}

	fee, err := uc.feeRepo.FindByID(ctx, inv.FeeComponentID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case feeConst.CodeFeeComponentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	itemName := fee.Name
	if inv.Notes != nil && *inv.Notes != "" {
		itemName = *inv.Notes
	}

	transactionID := pgVO.NewTransactionID()
	amount := inv.RemainingAmount()

	snapResp, err := uc.midtransGateway.CreateSnapTransaction(ctx, ports.SnapTransactionRequest{
		OrderID:       transactionID.String(),
		GrossAmount:   amount,
		CustomerName:  inv.SantriID,
		ExpiryMinutes: uc.expiryMinutes,
		Items: []ports.SnapItem{{
			ID:       inv.InvoiceNumber,
			Price:    amount,
			Quantity: 1,
			Name:     itemName,
		}},
	})
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pgConst.CodePaymentGatewayAPIFailed {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	expiredAt := time.Now().Add(time.Duration(uc.expiryMinutes) * time.Minute)
	metadata, _ := json.Marshal(map[string]string{
		"invoice_number": inv.InvoiceNumber,
		"santri_id":      inv.SantriID,
	})

	gatewayTx, err := pgEntity.NewPaymentGatewayTransaction(
		uuid.New().String(),
		transactionID.String(),
		inv.ID,
		amount,
		snapResp.Token,
		snapResp.RedirectURL,
		metadata,
		expiredAt,
	)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case pgConst.CodePaymentGatewayInvalidStatus, pgConst.CodePaymentGatewayNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.paymentGatewayRepo.Save(ctx, gatewayTx); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return buildMidtransPaymentResponse(gatewayTx), nil
}

func buildMidtransPaymentResponse(tx *pgEntity.PaymentGatewayTransaction) *dto.MidtransPaymentResponse {
	return &dto.MidtransPaymentResponse{
		TransactionID: tx.TransactionID,
		InvoiceID:     tx.InvoiceID,
		Amount:        tx.Amount,
		SnapToken:     tx.SnapToken,
		RedirectURL:   tx.RedirectURL,
		Status:        string(tx.Status),
		ExpiresAt:     tx.ExpiredAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

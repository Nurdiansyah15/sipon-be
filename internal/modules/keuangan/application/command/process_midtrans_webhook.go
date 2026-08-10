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
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	pgConst "sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	pgEntity "sipon-be/internal/modules/keuangan/domain/paymentgateway/entity"
	pgRepo "sipon-be/internal/modules/keuangan/domain/paymentgateway/repository"
	"sipon-be/internal/shared/kernel"
)

// systemActor adalah UUID sentinel untuk aksi yang dipicu sistem (webhook),
// bukan user sungguhan. Kolom created_by / posted_by pada tabel keuangan
// bertipe UUID NOT NULL dan tidak berelasi FK, sehingga nilai sentinel valid.
const systemActor = "00000000-0000-0000-0000-000000000000"

type ProcessMidtransWebhookUseCase struct {
	paymentGatewayRepo  pgRepo.PaymentGatewayRepository
	paymentRepo         payRepo.PaymentRepository
	invoiceRepo         invRepo.InvoiceRepository
	feeRepo             feeRepo.FeeComponentRepository
	midtransGateway     ports.MidtransGateway
	transactor          ports.Transactor
	autoPosting         *journalService.AutoPostingService
	settlementAccountID string
}

func NewProcessMidtransWebhookUseCase(
	paymentGatewayRepo pgRepo.PaymentGatewayRepository,
	paymentRepo payRepo.PaymentRepository,
	invoiceRepo invRepo.InvoiceRepository,
	feeRepo feeRepo.FeeComponentRepository,
	midtransGateway ports.MidtransGateway,
	transactor ports.Transactor,
	autoPosting *journalService.AutoPostingService,
	settlementAccountID string,
) *ProcessMidtransWebhookUseCase {
	return &ProcessMidtransWebhookUseCase{
		paymentGatewayRepo:  paymentGatewayRepo,
		paymentRepo:         paymentRepo,
		invoiceRepo:         invoiceRepo,
		feeRepo:             feeRepo,
		midtransGateway:     midtransGateway,
		transactor:          transactor,
		autoPosting:         autoPosting,
		settlementAccountID: settlementAccountID,
	}
}

// Execute memproses notifikasi webhook Midtrans secara idempotent. Endpoint
// webhook tidak memakai autentikasi user; keamanan dijamin oleh verifikasi
// signature key.
func (uc *ProcessMidtransWebhookUseCase) Execute(ctx context.Context, notif dto.MidtransWebhookNotification) error {
	if !uc.midtransGateway.VerifySignature(
		notif.OrderID, notif.StatusCode, notif.GrossAmount, notif.SignatureKey,
	) {
		return kernel.WrapMsg(application.ErrCodeUnauthorized, "signature webhook Midtrans tidak valid", nil)
	}

	gatewayTx, err := uc.paymentGatewayRepo.FindByTransactionID(ctx, notif.OrderID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pgConst.CodePaymentGatewayNotFound {
			// Transaksi tidak dikenal — jangan gagalkan webhook (Midtrans akan
			// mengulang), tapi catat sebagai sukses agar pengiriman berhenti.
			return nil
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	status := mapMidtransStatus(notif.TransactionStatus, notif.FraudStatus)
	raw, _ := json.Marshal(notif)
	paymentMethod := notif.PaymentType

	if err := gatewayTx.ApplyNotification(status, &paymentMethod, raw); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pgConst.CodePaymentGatewayInvalidStatus {
			return kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	// Sudah pernah diproses menjadi pembayaran — cukup simpan status terbaru.
	if gatewayTx.PaymentID != nil {
		return uc.paymentGatewayRepo.Update(ctx, gatewayTx)
	}

	// Status non-sukses (pending, gagal, kadaluarsa, batal): simpan status saja.
	if !status.IsSuccess() {
		return uc.paymentGatewayRepo.Update(ctx, gatewayTx)
	}

	return uc.settlePayment(ctx, gatewayTx)
}

// settlePayment membuat Payment dari settlement Midtrans lalu (bila akun
// settlement dikonfigurasi) langsung memverifikasinya sehingga invoice
// ter-update dan jurnal ter-posting.
func (uc *ProcessMidtransWebhookUseCase) settlePayment(ctx context.Context, gatewayTx *pgEntity.PaymentGatewayTransaction) error {
	return uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		inv, err := uc.invoiceRepo.FindByID(txCtx, gatewayTx.InvoiceID)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) && ke.Code == invConst.CodeInvoiceNotFound {
				return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}

		// Invoice mungkin sudah lunas dari pembayaran lain; jangan double-post.
		if !inv.HasOutstanding() {
			return nil
		}

		payNum, err := uc.paymentRepo.NextPaymentNumber(txCtx)
		if err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}

		var debitAccountID *string
		if uc.settlementAccountID != "" {
			debitAccountID = &uc.settlementAccountID
		}
		ref := gatewayTx.TransactionID

		payment, err := payEntity.NewPayment(
			uuid.New().String(),
			payNum.String(),
			inv.ID,
			gatewayTx.Amount,
			payConst.MethodTransfer,
			time.Now(),
			debitAccountID,
			&ref,
			nil,
			nil,
			systemActor,
		)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) && ke.Code == payConst.CodePaymentNotFound {
				return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}

		if err := uc.paymentRepo.Save(txCtx, payment); err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}

		// Verifikasi otomatis hanya bila akun settlement tersedia.
		if debitAccountID != nil {
			if err := payment.Verify(systemActor); err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) && ke.Code == payConst.CodePaymentInvalidStatus {
					return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
			if err := inv.AddPayment(payment.Amount); err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) && ke.Code == invConst.CodeInvoiceInvalidStatus {
					return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
				}
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}

			fee, err := uc.feeRepo.FindByID(txCtx, inv.FeeComponentID)
			if err != nil {
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}

			if err := uc.paymentRepo.Update(txCtx, payment); err != nil {
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
			if err := uc.invoiceRepo.Update(txCtx, inv); err != nil {
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
			if err := uc.autoPosting.PostPaymentVerified(
				txCtx, payment.ID, payment.PaymentNumber, "Pembayaran online Midtrans",
				payment.PaymentDate, payment.Amount, *debitAccountID, fee.ReceivableAccountID, systemActor,
			); err != nil {
				return kernel.WrapMsg(application.ErrCodeInternal, "gagal mem-posting jurnal pembayaran", err)
			}
		} else {
			if err := uc.paymentRepo.Update(txCtx, payment); err != nil {
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
		}

		if err := gatewayTx.LinkPayment(payment.ID); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) && ke.Code == pgConst.CodePaymentGatewayInvalidStatus {
				return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}

		return uc.paymentGatewayRepo.Update(txCtx, gatewayTx)
	})
}

// mapMidtransStatus memetakan transaction_status + fraud_status Midtrans ke
// PaymentGatewayStatus internal. Semua status "uang diterima" dianggap sukses;
// challenge tetap menunggu sampai di-approve/deny.
func mapMidtransStatus(transactionStatus, fraudStatus string) pgConst.PaymentGatewayStatus {
	switch transactionStatus {
	case "capture":
		if fraudStatus == "challenge" {
			return pgConst.GatewayStatusPendingChallenge
		}
		return pgConst.GatewayStatusSettlement
	case "settlement":
		return pgConst.GatewayStatusSettlement
	case "pending":
		return pgConst.GatewayStatusPending
	case "deny":
		return pgConst.GatewayStatusDeny
	case "cancel":
		return pgConst.GatewayStatusCancel
	case "expire":
		return pgConst.GatewayStatusExpire
	case "failure":
		return pgConst.GatewayStatusFailure
	default:
		return pgConst.GatewayStatusPending
	}
}

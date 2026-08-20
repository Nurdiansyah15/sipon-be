package command

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/shared/kernel"
)

type VerifyPaymentUseCase struct {
	paymentRepo  payRepo.PaymentRepository
	invoiceRepo  invRepo.InvoiceRepository
	feeRepo      feeRepo.FeeComponentRepository
	accountRepo  accRepo.AccountRepository
	transactor   ports.Transactor
	autoPosting  *journalService.AutoPostingService
	outboxWriter ports.OutboxWriter
}

func NewVerifyPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, feeRepo feeRepo.FeeComponentRepository, accountRepo accRepo.AccountRepository, transactor ports.Transactor, autoPosting *journalService.AutoPostingService) *VerifyPaymentUseCase {
	return &VerifyPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, feeRepo: feeRepo, accountRepo: accountRepo, transactor: transactor, autoPosting: autoPosting}
}

func (uc *VerifyPaymentUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *VerifyPaymentUseCase) publishPaymentVerified(ctx context.Context, userID, invoiceID string) {
	if uc.outboxWriter == nil {
		return
	}
	payload, _ := json.Marshal(map[string]string{
		"user_id":    userID,
		"invoice_id": invoiceID,
	})
	if err := uc.outboxWriter.Save(ctx, RoutingPaymentVerified, payload); err != nil {
		slog.Warn("keuangan: gagal publish event", "routing_key", RoutingPaymentVerified, "invoice_id", invoiceID, "error", err)
	}
}

func (uc *VerifyPaymentUseCase) Execute(ctx context.Context, paymentID string, verifiedBy string, debitAccountID string) (*dto.PaymentResponse, error) {
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

	if debitAccountID == "" {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun debit wajib diisi saat verifikasi", nil)
	}

	debitAcc, err := uc.accountRepo.FindByID(ctx, debitAccountID)
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
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun debit harus postable dan aktif", err)
	}
	if debitAcc.SubType == nil || *debitAcc.SubType != accConst.SubTypeCashBank {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Akun debit pembayaran harus merupakan akun kas atau bank", nil)
	}

	payment.DebitAccountID = &debitAccountID

	inv, err := uc.invoiceRepo.FindByID(ctx, payment.InvoiceID)
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

	if err := payment.Verify(verifiedBy); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := inv.AddPayment(payment.Amount); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case invConst.CodeInvoiceInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.paymentRepo.Update(txCtx, payment); err != nil {
			return err
		}
		if err := uc.invoiceRepo.Update(txCtx, inv); err != nil {
			return err
		}
		if uc.autoPosting != nil {
			if err := uc.autoPosting.PostPaymentVerified(
				txCtx, payment.ID, payment.PaymentNumber, "",
				payment.PaymentDate, payment.Amount,
				accountID(payment.DebitAccountID), fee.ReceivableAccountID, verifiedBy,
			); err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					switch ke.Code {
					case accConst.CodeAccountNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
					case accConst.CodeAccountNotPostable, accConst.CodeAccountInvalidSubType:
						return kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
					case journalConst.CodeJournalAccountMappingNotFound:
						return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
					case periodConst.CodePeriodNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
					case journalConst.CodeJournalPeriodClosed:
						return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
					}
				}
				return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
			}
		}
		return nil
	})
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case invConst.CodeInvoiceNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case invConst.CodeInvoiceInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case application.ErrCodeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case application.ErrCodeBadRequest:
				return nil, kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
			case application.ErrCodeConflict:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	uc.publishPaymentVerified(ctx, inv.UserID, inv.ID)

	return toPaymentResponse(payment), nil
}

func accountID(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

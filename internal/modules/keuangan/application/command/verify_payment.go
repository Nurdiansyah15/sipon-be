package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/shared/kernel"
)

type VerifyPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
	transactor  ports.Transactor
	autoPosting *journalService.AutoPostingService
}

func NewVerifyPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, transactor ports.Transactor, autoPosting *journalService.AutoPostingService) *VerifyPaymentUseCase {
	return &VerifyPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, transactor: transactor, autoPosting: autoPosting}
}

func (uc *VerifyPaymentUseCase) Execute(ctx context.Context, paymentID string, verifiedBy string) (*dto.PaymentResponse, error) {
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
				accountID(payment.DebitAccountID), verifiedBy,
			); err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					switch ke.Code {
					case accConst.CodeAccountNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
					case periodConst.CodePeriodNotFound:
						return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
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
			case application.ErrCodeConflict:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return toPaymentResponse(payment), nil
}

func accountID(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
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
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	inv, err := uc.invoiceRepo.FindByID(ctx, payment.InvoiceID)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	if err := payment.Verify(verifiedBy); err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentInvalidStatus)
	}

	if err := inv.AddPayment(payment.Amount); err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceInvalidStatus)
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
				return application.WrapConflictErr(err, journalConst.CodeJournalAccountMappingNotFound)
			}
		}
		return nil
	})
	if err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	return toPaymentResponse(payment), nil
}

func accountID(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

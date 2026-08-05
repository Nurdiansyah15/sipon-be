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
)

type VerifyPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
	invoiceRepo invRepo.InvoiceRepository
	transactor  ports.Transactor
}

func NewVerifyPaymentUseCase(paymentRepo payRepo.PaymentRepository, invoiceRepo invRepo.InvoiceRepository, transactor ports.Transactor) *VerifyPaymentUseCase {
	return &VerifyPaymentUseCase{paymentRepo: paymentRepo, invoiceRepo: invoiceRepo, transactor: transactor}
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
		return nil
	})
	if err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	return toPaymentResponse(payment), nil
}
